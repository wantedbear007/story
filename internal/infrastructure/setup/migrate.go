package setup

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migration struct {
	Version int64
	Name    string
	UpSQL   string
	DownSQL string
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("reading migrations: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		data, err := fs.ReadFile(migrationFS, filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		content := string(data)
		upSQL, downSQL := splitMigration(content)

		version, name := parseMigrationName(entry.Name())
		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   upSQL,
			DownSQL: downSQL,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationName(filename string) (int64, string) {
	name := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return 0, name
	}
	var version int64
	fmt.Sscanf(parts[0], "%d", &version)
	return version, parts[1]
}

func splitMigration(content string) (up, down string) {
	const upMarker = "-- +goose Up"
	const downMarker = "-- +goose Down"

	if idx := strings.Index(content, upMarker); idx >= 0 {
		afterUp := content[idx+len(upMarker):]
		if idx2 := strings.Index(afterUp, downMarker); idx2 >= 0 {
			up = strings.TrimSpace(afterUp[:idx2])
			down = strings.TrimSpace(afterUp[idx2+len(downMarker):])
		} else {
			up = strings.TrimSpace(afterUp)
		}
	}
	return
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ensureMigrationTable(ctx, pool); err != nil {
		return fmt.Errorf("ensuring migration table: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("getting applied migrations: %w", err)
	}

	// If schema_migrations table has records, use them.
	// If empty but tables already exist (e.g., from goose), mark all as applied.
	if len(applied) == 0 && schemaExists(ctx, pool) {
		for _, m := range migrations {
			if _, err := pool.Exec(ctx,
				"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				m.Version, time.Now(),
			); err != nil {
				return fmt.Errorf("recording existing migration %d: %w", m.Version, err)
			}
		}
		fmt.Println("  Schema already up to date.")
		return nil
	}

	appliedSet := make(map[int64]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	anyApplied := false
	for _, m := range migrations {
		if appliedSet[m.Version] {
			continue
		}

		if err := applyMigration(ctx, pool, m); err != nil {
			return fmt.Errorf("applying migration %d_%s: %w", m.Version, m.Name, err)
		}

		fmt.Printf("  ✓ %s table\n", m.Name)
		anyApplied = true
	}

	if !anyApplied {
		fmt.Println("  Schema already up to date.")
	}

	return nil
}

func schemaExists(ctx context.Context, pool *pgxpool.Pool) bool {
	var exists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = 'users' AND schemaname = 'public')",
	).Scan(&exists)
	return err == nil && exists
}

func ensureMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) ([]int64, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, m.UpSQL); err != nil {
		return fmt.Errorf("executing up SQL: %w", err)
	}

	if _, err := tx.Exec(ctx,
		"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)",
		m.Version, time.Now(),
	); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	return tx.Commit(ctx)
}

func RollbackLast(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("getting applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to roll back")
	}

	lastVersion := applied[len(applied)-1]
	var lastMigration *Migration
	for _, m := range migrations {
		if m.Version == lastVersion {
			lastMigration = &m
			break
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("migration %d not found", lastVersion)
	}

	if lastMigration.DownSQL == "" {
		return fmt.Errorf("migration %d has no down SQL", lastVersion)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, lastMigration.DownSQL); err != nil {
		return fmt.Errorf("executing down SQL: %w", err)
	}

	if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", lastVersion); err != nil {
		return fmt.Errorf("removing migration record: %w", err)
	}

	return tx.Commit(ctx)
}

func MigrationStatus(ctx context.Context, pool *pgxpool.Pool) ([]MigrationStatusRow, error) {
	if err := ensureMigrationTable(ctx, pool); err != nil {
		return nil, err
	}

	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return nil, err
	}

	appliedSet := make(map[int64]time.Time)
	for _, v := range applied {
		var t time.Time
		if err := pool.QueryRow(ctx, "SELECT applied_at FROM schema_migrations WHERE version = $1", v).Scan(&t); err != nil {
			return nil, err
		}
		appliedSet[v] = t
	}

	migrations, err := loadMigrations()
	if err != nil {
		return nil, err
	}

	var rows []MigrationStatusRow
	for _, m := range migrations {
		row := MigrationStatusRow{
			Version: m.Version,
			Name:    m.Name,
		}
		if t, ok := appliedSet[m.Version]; ok {
			row.AppliedAt = &t
		}
		rows = append(rows, row)
	}

	return rows, nil
}

type MigrationStatusRow struct {
	Version   int64
	Name      string
	AppliedAt *time.Time
}

var _ = pgx.ErrNoRows
