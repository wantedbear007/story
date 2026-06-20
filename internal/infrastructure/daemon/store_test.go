package daemon_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/story/internal/domain"
	infradaemon "github.com/anomalyco/story/internal/infrastructure/daemon"
)

func newTestStore(t *testing.T) *infradaemon.FileStore {
	t.Helper()
	dir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	store, err := infradaemon.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func TestStore_SaveAndLoad(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	info := domain.DaemonInfo{
		PID:    12345,
		Status: domain.DaemonStatusRunning,
		Host:   "127.0.0.1",
		Port:   8081,
	}

	if err := store.Save(ctx, info); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil info")
	}
	if loaded.PID != 12345 {
		t.Errorf("expected PID 12345, got %d", loaded.PID)
	}
	if loaded.Status != domain.DaemonStatusRunning {
		t.Errorf("expected status running, got %s", loaded.Status)
	}
	if loaded.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", loaded.Host)
	}
	if loaded.Port != 8081 {
		t.Errorf("expected port 8081, got %d", loaded.Port)
	}
}

func TestStore_LoadNotExists(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	info, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if info != nil {
		t.Fatal("expected nil for non-existent store")
	}
}

func TestStore_Remove(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	info := domain.DaemonInfo{
		PID:    12345,
		Status: domain.DaemonStatusRunning,
	}
	if err := store.Save(ctx, info); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := store.Remove(ctx); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	loaded, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load after remove: %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil after remove")
	}

	pidPath := filepath.Join(os.Getenv("HOME"), ".story", "daemon.pid")
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Errorf("expected PID file to be removed")
	}
}
