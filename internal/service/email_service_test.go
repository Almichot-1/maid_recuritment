package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
)

func TestNewSMTPEmailService_Validation(t *testing.T) {
	_, err := NewSMTPEmailService(nil)
	require.Error(t, err)

	_, err = NewSMTPEmailService(&config.Config{})
	require.Error(t, err)

	_, err = NewSMTPEmailService(&config.Config{SMTPHost: "smtp.example.com", SMTPPort: "not-int", SMTPUser: "mailer@example.com", SMTPPass: "p"})
	require.Error(t, err)

	service, err := NewSMTPEmailService(&config.Config{SMTPHost: "smtp.example.com", SMTPPort: "587", SMTPUser: "mailer@example.com", SMTPPass: "p"})
	require.NoError(t, err)
	require.NotNil(t, service)
	require.Equal(t, "mailer@example.com", service.fromEmail)
	require.Equal(t, "mailer@example.com", service.fromHeader)

	service, err = NewSMTPEmailService(&config.Config{
		SMTPHost:      "smtp.example.com",
		SMTPPort:      "587",
		SMTPUser:      "mailer@example.com",
		SMTPPass:      "p",
		SMTPFromEmail: "noreply@example.com",
		SMTPFromName:  "Maid Recruitment",
	})
	require.NoError(t, err)
	require.Equal(t, "noreply@example.com", service.fromEmail)
	require.Equal(t, "\"Maid Recruitment\" <noreply@example.com>", service.fromHeader)
}

func TestSMTPEmailService_SendValidation(t *testing.T) {
	service, err := NewSMTPEmailService(&config.Config{SMTPHost: "smtp.example.com", SMTPPort: "587", SMTPUser: "mailer@example.com", SMTPPass: "p"})
	require.NoError(t, err)

	err = service.Send("", "subject", "body")
	require.Error(t, err)
}
