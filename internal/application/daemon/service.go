package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/anomalyco/story/internal/application/notification"
	"github.com/anomalyco/story/internal/application/scheduler"
	"github.com/anomalyco/story/internal/domain"
	"github.com/anomalyco/story/internal/infrastructure/config"
)

type Service struct {
	store    domain.DaemonStore
	notifSvc *notification.Service
	sched    *scheduler.Service
	cfg      config.CaptureConfig
}

func NewService(store domain.DaemonStore, notifSvc *notification.Service, sched *scheduler.Service, cfg config.CaptureConfig) *Service {
	return &Service{
		store:    store,
		notifSvc: notifSvc,
		sched:    sched,
		cfg:      cfg,
	}
}

func (s *Service) Start(ctx context.Context) error {
	info, err := s.store.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading daemon info: %w", err)
	}
	if info != nil && info.Status == domain.DaemonStatusRunning {
		if processExists(info.PID) {
			return fmt.Errorf("daemon is already running (PID %d)", info.PID)
		}
	}

	cmd := exec.Command(os.Args[0], "_daemon")
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting daemon: %w", err)
	}

	daemonInfo := domain.DaemonInfo{
		PID:    cmd.Process.Pid,
		Status: domain.DaemonStatusRunning,
		Host:   s.cfg.Host,
		Port:   s.cfg.Port,
	}

	if err := s.store.Save(ctx, daemonInfo); err != nil {
		return fmt.Errorf("saving daemon info: %w", err)
	}

	fmt.Printf("Daemon started (PID %d)\n", cmd.Process.Pid)
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	info, err := s.store.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading daemon info: %w", err)
	}
	if info == nil || info.Status != domain.DaemonStatusRunning {
		return fmt.Errorf("daemon is not running")
	}

	if err := syscall.Kill(info.PID, syscall.SIGTERM); err != nil {
		return fmt.Errorf("stopping daemon: %w", err)
	}

	s.store.Remove(ctx)
	fmt.Println("Daemon stopped")
	return nil
}

func (s *Service) Status(ctx context.Context) (*domain.DaemonInfo, error) {
	info, err := s.store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading daemon info: %w", err)
	}
	if info == nil {
		return &domain.DaemonInfo{Status: domain.DaemonStatusStopped}, nil
	}
	if !processExists(info.PID) {
		info.Status = domain.DaemonStatusStopped
		s.store.Remove(ctx)
	}
	return info, nil
}

func (s *Service) RunDaemon(ctx context.Context) error {
	s.sched.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	s.sched.Stop()
	return nil
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
