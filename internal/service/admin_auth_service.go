package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrAdminInvalidCredentials = errors.New("invalid admin credentials")
	ErrAdminInvalidToken       = errors.New("invalid admin token")
	ErrAdminAccountLocked      = errors.New("admin account locked")
	ErrAdminInactive           = errors.New("admin account inactive")
	ErrAdminInvalidMFA         = errors.New("invalid mfa code")
	ErrWeakAdminPassword       = errors.New("admin password does not meet complexity requirements")
	ErrAdminPasswordMismatch   = errors.New("current admin password is incorrect")
)

type AdminAuthService struct {
	adminRepository domain.AdminRepository
	auditRepository domain.AuditLogRepository
	jwtSecret       []byte
}

type adminClaims struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Scope   string `json:"scope"`
	jwt.RegisteredClaims
}

func NewAdminAuthService(adminRepository domain.AdminRepository, auditRepository domain.AuditLogRepository, cfg *config.Config) (*AdminAuthService, error) {
	if adminRepository == nil {
		return nil, fmt.Errorf("admin repository is nil")
	}
	if auditRepository == nil {
		return nil, fmt.Errorf("audit repository is nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, fmt.Errorf("jwt secret is empty")
	}

	return &AdminAuthService{
		adminRepository: adminRepository,
		auditRepository: auditRepository,
		jwtSecret:       []byte(cfg.JWTSecret),
	}, nil
}

func (s *AdminAuthService) CreateAdmin(email, password, fullName, role, mfaSecret string) (*domain.Admin, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	mfaSecret = strings.TrimSpace(mfaSecret)

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrAdminInvalidCredentials
	}
	if err := validateAdminPassword(password); err != nil {
		return nil, err
	}
	parsedRole, err := parseAdminRole(role)
	if err != nil {
		return nil, err
	}
	if fullName == "" || mfaSecret == "" {
		return nil, ErrAdminInvalidCredentials
	}

	existing, err := s.adminRepository.GetByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}
	if err != nil && !errors.Is(err, repository.ErrAdminNotFound) {
		return nil, fmt.Errorf("check existing admin: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash admin password: %w", err)
	}

	admin := &domain.Admin{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         parsedRole,
		MFASecret:    mfaSecret,
		IsActive:     true,
	}

	if err := s.adminRepository.Create(admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *AdminAuthService) Login(email, password, otpCode, ipAddress string) (*domain.Admin, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	admin, err := s.adminRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			return nil, "", ErrAdminInvalidCredentials
		}
		return nil, "", fmt.Errorf("get admin by email: %w", err)
	}

	if err := s.ensureAdminCanLogin(admin); err != nil {
		return nil, "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		if updateErr := s.recordFailedLogin(admin); updateErr != nil {
			return nil, "", updateErr
		}
		return nil, "", ErrAdminInvalidCredentials
	}

	if !totp.Validate(strings.TrimSpace(otpCode), admin.MFASecret) {
		if updateErr := s.recordFailedLogin(admin); updateErr != nil {
			return nil, "", updateErr
		}
		return nil, "", ErrAdminInvalidMFA
	}

	now := time.Now().UTC()
	admin.FailedLoginAttempts = 0
	admin.LockedUntil = nil
	admin.LastLogin = &now
	if err := s.adminRepository.Update(admin); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(admin)
	if err != nil {
		return nil, "", err
	}

	_ = s.logAudit(admin.ID, "admin_login", "", "", map[string]any{
		"email": admin.Email,
		"role":  admin.Role,
	}, ipAddress)

	return admin, token, nil
}

func (s *AdminAuthService) ValidateToken(tokenString string) (string, string, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", "", ErrAdminInvalidToken
	}

	claims := &adminClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrAdminInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return "", "", ErrAdminInvalidToken
	}

	if claims.AdminID == "" || claims.Role == "" || claims.Scope != "admin" {
		return "", "", ErrAdminInvalidToken
	}

	admin, err := s.adminRepository.GetByID(claims.AdminID)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			return "", "", ErrAdminInvalidToken
		}
		return "", "", err
	}
	if err := s.ensureAdminCanLogin(admin); err != nil {
		return "", "", err
	}

	return admin.ID, string(admin.Role), nil
}

func (s *AdminAuthService) LogLogout(adminID, ipAddress string) error {
	return s.logAudit(adminID, "admin_logout", "admin", adminID, map[string]any{}, ipAddress)
}

func (s *AdminAuthService) ChangePassword(adminID, currentPassword, newPassword, ipAddress string) error {
	adminID = strings.TrimSpace(adminID)
	if adminID == "" {
		return ErrAdminInvalidCredentials
	}

	admin, err := s.adminRepository.GetByID(adminID)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			return ErrAdminInvalidCredentials
		}
		return fmt.Errorf("get admin by id: %w", err)
	}

	if err := s.ensureAdminCanLogin(admin); err != nil {
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPassword)) != nil {
		return ErrAdminPasswordMismatch
	}

	if err := validateAdminPassword(newPassword); err != nil {
		return err
	}

	if currentPassword == newPassword {
		return ErrWeakAdminPassword
	}

	admin.PasswordHash = newPassword
	admin.ForcePasswordChange = false
	if err := s.adminRepository.Update(admin); err != nil {
		return fmt.Errorf("update admin password: %w", err)
	}

	_ = s.logAudit(admin.ID, "admin_change_password", "admin", admin.ID, map[string]any{
		"email": admin.Email,
	}, ipAddress)

	return nil
}

func (s *AdminAuthService) ensureAdminCanLogin(admin *domain.Admin) error {
	if admin == nil {
		return ErrAdminInvalidCredentials
	}
	if !admin.IsActive {
		return ErrAdminInactive
	}
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now().UTC()) {
		return ErrAdminAccountLocked
	}
	return nil
}

func (s *AdminAuthService) recordFailedLogin(admin *domain.Admin) error {
	if admin == nil {
		return ErrAdminInvalidCredentials
	}
	admin.FailedLoginAttempts++
	if admin.FailedLoginAttempts >= 3 {
		lockUntil := time.Now().UTC().Add(15 * time.Minute)
		admin.LockedUntil = &lockUntil
		admin.FailedLoginAttempts = 0
	}
	if err := s.adminRepository.Update(admin); err != nil {
		return fmt.Errorf("update failed admin login attempts: %w", err)
	}
	return nil
}

func (s *AdminAuthService) generateToken(admin *domain.Admin) (string, error) {
	now := time.Now().UTC()
	claims := adminClaims{
		AdminID: admin.ID,
		Email:   admin.Email,
		Role:    string(admin.Role),
		Scope:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign admin token: %w", err)
	}
	return signed, nil
}

func (s *AdminAuthService) logAudit(adminID, action, targetType, targetID string, details map[string]any, ipAddress string) error {
	if strings.TrimSpace(adminID) == "" {
		return nil
	}
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	var normalizedTargetID *string
	if strings.TrimSpace(targetID) != "" {
		value := strings.TrimSpace(targetID)
		normalizedTargetID = &value
	}
	return s.auditRepository.Create(&domain.AuditLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: strings.TrimSpace(targetType),
		TargetID:   normalizedTargetID,
		Details:    payload,
		IPAddress:  strings.TrimSpace(ipAddress),
	})
}

func validateAdminPassword(password string) error {
	if len(password) < 12 {
		return ErrWeakAdminPassword
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password)
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrWeakAdminPassword
	}
	return nil
}

func parseAdminRole(role string) (domain.AdminRole, error) {
	switch strings.TrimSpace(role) {
	case string(domain.SuperAdmin):
		return domain.SuperAdmin, nil
	case string(domain.SupportAdmin):
		return domain.SupportAdmin, nil
	default:
		return "", repository.ErrInvalidAdminRole
	}
}
