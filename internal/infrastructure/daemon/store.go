package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anomalyco/story/internal/domain"
)

const (
	pidFileName  = "daemon.pid"
	infoFileName = "daemon.json"
)

type FileStore struct {
	storyDir string
}

func NewStore() (*FileStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}
	dir := filepath.Join(home, ".story")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating story directory: %w", err)
	}
	return &FileStore{storyDir: dir}, nil
}

func (s *FileStore) Save(ctx context.Context, info domain.DaemonInfo) error {
	if err := os.WriteFile(filepath.Join(s.storyDir, pidFileName), []byte(fmt.Sprintf("%d", info.PID)), 0644); err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshaling daemon info: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.storyDir, infoFileName), data, 0644); err != nil {
		return fmt.Errorf("writing daemon info file: %w", err)
	}
	return nil
}

func (s *FileStore) Load(ctx context.Context) (*domain.DaemonInfo, error) {
	data, err := os.ReadFile(filepath.Join(s.storyDir, infoFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading daemon info: %w", err)
	}
	var info domain.DaemonInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshaling daemon info: %w", err)
	}
	return &info, nil
}

func (s *FileStore) Remove(ctx context.Context) error {
	os.Remove(filepath.Join(s.storyDir, pidFileName))
	os.Remove(filepath.Join(s.storyDir, infoFileName))
	return nil
}
