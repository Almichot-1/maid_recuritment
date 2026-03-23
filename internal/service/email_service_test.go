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

	_, err = NewSMTPEmailService(&config.Config{SMTPHost: "smtp", SMTPPort: "not-int", SMTPUser: "u", SMTPPass: "p"})
	require.Error(t, err)

	service, err := NewSMTPEmailService(&config.Config{SMTPHost: "smtp", SMTPPort: "587", SMTPUser: "u", SMTPPass: "p"})
	require.NoError(t, err)
	require.NotNil(t, service)
}

func TestSMTPEmailService_SendValidation(t *testing.T) {
	service, err := NewSMTPEmailService(&config.Config{SMTPHost: "smtp", SMTPPort: "587", SMTPUser: "u", SMTPPass: "p"})
	require.NoError(t, err)

	err = service.Send("", "subject", "body")
	require.Error(t, err)
}