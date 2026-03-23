package service

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidToken           = errors.New("invalid token")
	ErrAccountPendingApproval = errors.New("account pending approval")
	ErrAccountRejected        = errors.New("account rejected")
	ErrAccountSuspended       = errors.New("account suspended")
	ErrAccountInactive        = errors.New("account inactive")
)

type AuthService struct {
	userRepository domain.UserRepository
	jwtSecret      []byte
}

type authClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepository domain.UserRepository, cfg *config.Config) (*AuthService, error) {
	if userRepository == nil {
		return nil, fmt.Errorf("user repository is nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, fmt.Errorf("jwt secret is empty")
	}

	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      []byte(cfg.JWTSecret),
	}, nil
}

func (s *AuthService) Register(email, password, fullName, role, companyName string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateAuthEmail(email); err != nil {
		return "", err
	}
	if err := validateAuthPassword(password); err != nil {
		return "", err
	}
	parsedRole, err := parseUserRole(role)
	if err != nil {
		return "", err
	}

	existingUser, err := s.userRepository.GetByEmail(email)
	if err == nil && existingUser != nil {
		return "", ErrUserExists
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return "", fmt.Errorf("check existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:         email,
		PasswordHash:  string(hashedPassword),
		FullName:      strings.TrimSpace(fullName),
		Role:          parsedRole,
		CompanyName:   strings.TrimSpace(companyName),
		AccountStatus: domain.AccountStatusPendingApproval,
		IsActive:      true,
	}

	if err := s.userRepository.Create(user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return "", ErrUserExists
		}
		return "", fmt.Errorf("create user: %w", err)
	}

	token, err := s.generateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("get user by email: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	if err := validateUserAccess(user); err != nil {
		return "", err
	}

	token, err := s.generateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, string, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", "", ErrInvalidToken
	}

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return "", "", ErrInvalidToken
	}

	if claims.UserID == "" || claims.Role == "" {
		return "", "", ErrInvalidToken
	}

	user, err := s.userRepository.GetByID(claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", "", ErrInvalidToken
		}
		return "", "", fmt.Errorf("get user by id: %w", err)
	}
	if err := validateUserAccess(user); err != nil {
		return "", "", err
	}

	return claims.UserID, claims.Role, nil
}

func (s *AuthService) generateToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := authClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func validateAuthEmail(email string) error {
	if email == "" {
		return ErrInvalidCredentials
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func validateAuthPassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidCredentials
	}
	return nil
}

func parseUserRole(role string) (domain.UserRole, error) {
	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		return domain.EthiopianAgent, nil
	case string(domain.ForeignAgent):
		return domain.ForeignAgent, nil
	default:
		return "", ErrInvalidCredentials
	}
}

func validateUserAccess(user *domain.User) error {
	if user == nil {
		return ErrInvalidCredentials
	}
	if !user.IsActive {
		return ErrAccountInactive
	}
	switch user.AccountStatus {
	case domain.AccountStatusActive:
		return nil
	case domain.AccountStatusPendingApproval:
		return ErrAccountPendingApproval
	case domain.AccountStatusRejected:
		return ErrAccountRejected
	case domain.AccountStatusSuspended:
		return ErrAccountSuspended
	default:
		return ErrAccountInactive
	}
}
