package service

import (
	"fmt"
	"net/smtp"
	"strconv"
	"strings"

	"maid-recruitment-tracking/internal/config"
)

type EmailService interface {
	Send(to, subject, body string) error
}

type SMTPEmailService struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPEmailService(cfg *config.Config) (*SMTPEmailService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(cfg.SMTPHost) == "" || strings.TrimSpace(cfg.SMTPPort) == "" || strings.TrimSpace(cfg.SMTPUser) == "" || strings.TrimSpace(cfg.SMTPPass) == "" {
		return nil, fmt.Errorf("smtp configuration is incomplete")
	}
	port, err := strconv.Atoi(cfg.SMTPPort)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp port: %w", err)
	}

	return &SMTPEmailService{
		host: cfg.SMTPHost,
		port: port,
		user: cfg.SMTPUser,
		pass: cfg.SMTPPass,
		from: cfg.SMTPUser,
	}, nil
}

func (s *SMTPEmailService) Send(to, subject, body string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("email recipient is required")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	message := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n" +
		body + "\r\n"

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
