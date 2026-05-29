package repository

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateEmail  = errors.New("email already exists")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidRole     = errors.New("invalid user role")
)

type GormUserRepository struct {
	db *gorm.DB
}

func (r *GormUserRepository) DB() *gorm.DB {
	return r.db
}

func NewGormUserRepository(cfg *config.Config) (*GormUserRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormUserRepository{db: db}, nil
}

func (r *GormUserRepository) Create(user *domain.User) error {
	if user == nil {
		return fmt.Errorf("create user: user is nil")
	}

	if err := validateEmail(user.Email); err != nil {
		return err
	}
	if err := validatePassword(user.PasswordHash); err != nil {
		return err
	}
	if err := validateRole(user.Role); err != nil {
		return err
	}
	if err := validateAccountStatus(user.AccountStatus); err != nil {
		return err
	}

	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	if user.AccountStatus == "" {
		user.AccountStatus = domain.AccountStatusActive
	}

	if !isBcryptHash(user.PasswordHash) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := r.db.Create(user).Error; err != nil {
		if isDuplicateEmailError(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *GormUserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *GormUserRepository) GetByID(id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *GormUserRepository) Update(user *domain.User) error {
	if user == nil {
		return fmt.Errorf("update user: user is nil")
	}

	if user.ID == "" {
		return fmt.Errorf("update user: id is required")
	}
	if err := validateEmail(user.Email); err != nil {
		return err
	}
	if err := validateRole(user.Role); err != nil {
		return err
	}
	if err := validateAccountStatus(user.AccountStatus); err != nil {
		return err
	}
	if user.PasswordHash != "" {
		if err := validatePassword(user.PasswordHash); err != nil {
			return err
		}
		if !isBcryptHash(user.PasswordHash) {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("hash password: %w", err)
			}
			user.PasswordHash = string(hashedPassword)
		}
	}

	updates := map[string]any{
		"email":                      user.Email,
		"email_verified":             user.EmailVerified,
		"full_name":                  user.FullName,
		"role":                       user.Role,
		"company_name":               user.CompanyName,
		"avatar_url":                 user.AvatarURL,
		"auto_share_candidates":      user.AutoShareCandidates,
		"default_foreign_pairing_id": user.DefaultForeignPairingID,
		"account_status":             user.AccountStatus,
		"is_active":                  user.IsActive,
	}
	if user.PasswordHash != "" {
		updates["password_hash"] = user.PasswordHash
	}

	result := r.db.Model(&domain.User{}).Where("id = ?", user.ID).Updates(updates)
	if result.Error != nil {
		if isDuplicateEmailError(result.Error) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func validateRole(role domain.UserRole) error {
	switch role {
	case domain.EthiopianAgent, domain.ForeignAgent:
		return nil
	default:
		return ErrInvalidRole
	}
}

func validateAccountStatus(status domain.AccountStatus) error {
	switch status {
	case domain.AccountStatusPendingApproval, domain.AccountStatusActive, domain.AccountStatusRejected, domain.AccountStatusSuspended:
		return nil
	default:
		return fmt.Errorf("invalid account status")
	}
}

func isDuplicateEmailError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.Message, "email")
	}
	return false
}

func isBcryptHash(password string) bool {
	return strings.HasPrefix(password, "$2a$") || strings.HasPrefix(password, "$2b$") || strings.HasPrefix(password, "$2y$")
}
