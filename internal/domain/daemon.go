package domain

import "context"

type DaemonStatus string

const (
	DaemonStatusRunning DaemonStatus = "running"
	DaemonStatusStopped DaemonStatus = "stopped"
)

type DaemonInfo struct {
	PID    int          `json:"pid"`
	Status DaemonStatus `json:"status"`
	Host   string       `json:"host"`
	Port   int          `json:"port"`
}

type DaemonStore interface {
	Save(ctx context.Context, info DaemonInfo) error
	Load(ctx context.Context) (*DaemonInfo, error)
	Remove(ctx context.Context) error
}
