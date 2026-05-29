package repository

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrAdminNotFound    = errors.New("admin not found")
	ErrInvalidAdminRole = errors.New("invalid admin role")
)

type GormAdminRepository struct {
	db *gorm.DB
}

func NewGormAdminRepository(cfg *config.Config) (*GormAdminRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormAdminRepository{db: db}, nil
}

func (r *GormAdminRepository) Create(admin *domain.Admin) error {
	if admin == nil {
		return fmt.Errorf("create admin: admin is nil")
	}
	admin.Email = normalizeAdminEmail(admin.Email)
	if err := validateAdminEmail(admin.Email); err != nil {
		return err
	}
	if err := validateAdminRole(admin.Role); err != nil {
		return err
	}
	if strings.TrimSpace(admin.PasswordHash) == "" {
		return fmt.Errorf("create admin: password hash is required")
	}
	if strings.TrimSpace(admin.MFASecret) == "" {
		return fmt.Errorf("create admin: mfa secret is required")
	}
	if admin.ID == "" {
		admin.ID = uuid.NewString()
	}
	if !isBcryptPasswordHash(admin.PasswordHash) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("create admin: hash password: %w", err)
		}
		admin.PasswordHash = string(hashedPassword)
	}
	if err := r.db.Create(admin).Error; err != nil {
		if isDuplicateEmailError(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}

func (r *GormAdminRepository) GetByEmail(email string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.Where("LOWER(email) = ?", normalizeAdminEmail(email)).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("get admin by email: %w", err)
	}
	return &admin, nil
}

func (r *GormAdminRepository) GetByID(id string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.Where("id = ?", strings.TrimSpace(id)).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("get admin by id: %w", err)
	}
	return &admin, nil
}

func (r *GormAdminRepository) List() ([]*domain.Admin, error) {
	admins := make([]*domain.Admin, 0)
	if err := r.db.Order("created_at DESC").Find(&admins).Error; err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	return admins, nil
}

func (r *GormAdminRepository) ListActive() ([]*domain.Admin, error) {
	admins := make([]*domain.Admin, 0)
	now := time.Now().UTC()
	if err := r.db.
		Where("is_active = ?", true).
		Where("locked_until IS NULL OR locked_until < ?", now).
		Order("created_at DESC").
		Find(&admins).Error; err != nil {
		return nil, fmt.Errorf("list active admins: %w", err)
	}
	return admins, nil
}

func (r *GormAdminRepository) Update(admin *domain.Admin) error {
	if admin == nil {
		return fmt.Errorf("update admin: admin is nil")
	}
	if strings.TrimSpace(admin.ID) == "" {
		return fmt.Errorf("update admin: id is required")
	}
	admin.Email = normalizeAdminEmail(admin.Email)
	if err := validateAdminEmail(admin.Email); err != nil {
		return err
	}
	if err := validateAdminRole(admin.Role); err != nil {
		return err
	}
	updates := map[string]any{
		"email":                 admin.Email,
		"full_name":             admin.FullName,
		"role":                  admin.Role,
		"mfa_secret":            admin.MFASecret,
		"is_active":             admin.IsActive,
		"failed_login_attempts": admin.FailedLoginAttempts,
		"locked_until":          admin.LockedUntil,
		"last_login":            admin.LastLogin,
		"force_password_change": admin.ForcePasswordChange,
	}
	if strings.TrimSpace(admin.PasswordHash) != "" {
		if !isBcryptPasswordHash(admin.PasswordHash) {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.PasswordHash), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("update admin: hash password: %w", err)
			}
			admin.PasswordHash = string(hashedPassword)
		}
		updates["password_hash"] = admin.PasswordHash
	}
	result := r.db.Model(&domain.Admin{}).Where("id = ?", admin.ID).Updates(updates)
	if result.Error != nil {
		if isDuplicateEmailError(result.Error) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("update admin: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAdminNotFound
	}
	return nil
}

func validateAdminEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func normalizeAdminEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func validateAdminRole(role domain.AdminRole) error {
	switch role {
	case domain.SuperAdmin, domain.SupportAdmin:
		return nil
	default:
		return ErrInvalidAdminRole
	}
}

func isBcryptPasswordHash(password string) bool {
	return strings.HasPrefix(password, "$2a$") || strings.HasPrefix(password, "$2b$") || strings.HasPrefix(password, "$2y$")
}
