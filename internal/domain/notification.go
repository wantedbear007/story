package domain

import "context"

type NotificationProvider interface {
	Notify(ctx context.Context, req NotificationRequest) error
	Name() string
}

type NotificationRequest struct {
	Title   string
	Message string
	URL     string
}
