package email

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/anomalyco/story/internal/infrastructure/config"
)

// SMTPMailer sends emails via SMTP.
// Uses STARTTLS for secure connections.
// Plain auth is used for compatibility with most SMTP providers.
type SMTPMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPMailer(cfg config.SMTPConfig) *SMTPMailer {
	return &SMTPMailer{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
	}
}

func (m *SMTPMailer) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	subject := "Password Reset - Story"
	body := fmt.Sprintf(
		"Hello,\n\nYou requested a password reset for your Story account.\n\n"+
			"Your reset token is: %s\n\n"+
			"This token expires in 1 hour.\n\n"+
			"If you did not request this reset, please ignore this email.\n\n"+
			"- Story Team",
		token,
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.from, email, subject, body)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	if err := smtp.SendMail(addr, auth, m.from, []string{email}, []byte(msg)); err != nil {
		return fmt.Errorf("sending password reset email: %w", err)
	}

	return nil
}
