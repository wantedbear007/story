package email

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/anomalyco/story/internal/infrastructure/config"
)

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

func (m *SMTPMailer) SendVerificationEmail(ctx context.Context, to, token, displayName string) error {
	subject := "Verify your email - Story"
	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Thank you for creating a Story account.\n\n"+
			"Your email verification token is: %s\n\n"+
			"This token expires in 24 hours.\n\n"+
			"If you did not create an account, please ignore this email.\n\n"+
			"- Story Team",
		displayName, token,
	)
	return m.send(to, subject, body)
}

func (m *SMTPMailer) SendPasswordResetEmail(ctx context.Context, to, token, displayName string) error {
	subject := "Password Reset - Story"
	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"You requested a password reset for your Story account.\n\n"+
			"Your reset token is: %s\n\n"+
			"This token expires in 1 hour.\n\n"+
			"If you did not request this reset, please ignore this email.\n\n"+
			"- Story Team",
		displayName, token,
	)
	return m.send(to, subject, body)
}

func (m *SMTPMailer) send(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	return nil
}
