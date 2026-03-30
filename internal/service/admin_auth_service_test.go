package service

import (
	"encoding/json"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type adminRepoMock struct {
	admins map[string]*domain.Admin
}

func (m *adminRepoMock) Create(admin *domain.Admin) error { return nil }

func (m *adminRepoMock) GetByEmail(email string) (*domain.Admin, error) {
	for _, admin := range m.admins {
		if admin.Email == email {
			return cloneAdmin(admin), nil
		}
	}
	return nil, repository.ErrAdminNotFound
}

func (m *adminRepoMock) GetByID(id string) (*domain.Admin, error) {
	admin, ok := m.admins[id]
	if !ok {
		return nil, repository.ErrAdminNotFound
	}
	return cloneAdmin(admin), nil
}

func (m *adminRepoMock) List() ([]*domain.Admin, error) { return nil, nil }

func (m *adminRepoMock) ListActive() ([]*domain.Admin, error) { return nil, nil }

func (m *adminRepoMock) Update(admin *domain.Admin) error {
	updated := cloneAdmin(admin)
	hash, err := bcrypt.GenerateFromPassword([]byte(updated.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	updated.PasswordHash = string(hash)
	m.admins[updated.ID] = updated
	return nil
}

type auditRepoMock struct {
	entries []*domain.AuditLog
}

func (m *auditRepoMock) Create(log *domain.AuditLog) error {
	m.entries = append(m.entries, log)
	return nil
}

func (m *auditRepoMock) List(filters domain.AuditLogFilters) ([]*domain.AuditLog, error) {
	return nil, nil
}

func cloneAdmin(admin *domain.Admin) *domain.Admin {
	if admin == nil {
		return nil
	}
	copy := *admin
	return &copy
}

func TestAdminAuthServiceChangePasswordUpdatesPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("AdminPortal#2026!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	adminRepo := &adminRepoMock{
		admins: map[string]*domain.Admin{
			"admin-1": {
				ID:                  "admin-1",
				Email:               "super.admin@example.com",
				PasswordHash:        string(hash),
				FullName:            "Super Admin",
				Role:                domain.SuperAdmin,
				MFASecret:           "SECRET",
				IsActive:            true,
				ForcePasswordChange: true,
			},
		},
	}
	auditRepo := &auditRepoMock{}
	svc, err := NewAdminAuthService(adminRepo, auditRepo, &config.Config{JWTSecret: "test-secret"})
	if err != nil {
		t.Fatalf("new admin auth service: %v", err)
	}

	if err := svc.ChangePassword("admin-1", "AdminPortal#2026!", "NewPortal#2026!", "127.0.0.1"); err != nil {
		t.Fatalf("change password: %v", err)
	}

	updated := adminRepo.admins["admin-1"]
	if updated.ForcePasswordChange {
		t.Fatalf("expected force password change to be cleared")
	}
	if bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("NewPortal#2026!")) != nil {
		t.Fatalf("expected updated password hash to match new password")
	}
	if len(auditRepo.entries) != 1 {
		t.Fatalf("expected one audit log entry, got %d", len(auditRepo.entries))
	}

	var details map[string]any
	if err := json.Unmarshal(auditRepo.entries[0].Details, &details); err != nil {
		t.Fatalf("unmarshal audit details: %v", err)
	}
	if details["email"] != "super.admin@example.com" {
		t.Fatalf("expected audit email to be recorded")
	}
}

func TestAdminAuthServiceChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("AdminPortal#2026!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	adminRepo := &adminRepoMock{
		admins: map[string]*domain.Admin{
			"admin-1": {
				ID:           "admin-1",
				Email:        "super.admin@example.com",
				PasswordHash: string(hash),
				FullName:     "Super Admin",
				Role:         domain.SuperAdmin,
				MFASecret:    "SECRET",
				IsActive:     true,
				LastLogin:    func() *time.Time { now := time.Now().UTC(); return &now }(),
			},
		},
	}
	svc, err := NewAdminAuthService(adminRepo, &auditRepoMock{}, &config.Config{JWTSecret: "test-secret"})
	if err != nil {
		t.Fatalf("new admin auth service: %v", err)
	}

	if err := svc.ChangePassword("admin-1", "wrong-password", "NewPortal#2026!", "127.0.0.1"); err != ErrAdminPasswordMismatch {
		t.Fatalf("expected ErrAdminPasswordMismatch, got %v", err)
	}
}
