package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrUserExists                  = errors.New("user already exists")
	ErrInvalidToken                = errors.New("invalid token")
	ErrEmailNotVerified            = errors.New("email address is not verified")
	ErrEmailVerificationInvalid    = errors.New("invalid email verification link")
	ErrEmailVerificationExpired    = errors.New("email verification link has expired")
	ErrEmailVerificationMissing    = errors.New("email verification service is not configured")
	ErrAccountPendingApproval      = errors.New("account pending approval")
	ErrAccountRejected             = errors.New("account rejected")
	ErrAccountSuspended            = errors.New("account suspended")
	ErrAccountInactive             = errors.New("account inactive")
	ErrAuthRateLimited             = errors.New("too many login attempts")
	ErrPasswordResetCodeInvalid    = errors.New("invalid password reset code")
	ErrPasswordResetCodeExpired    = errors.New("password reset code has expired")
	ErrPasswordResetServiceMissing = errors.New("password reset service is not configured")
)

const (
	passwordResetCodeLength    = 6
	passwordResetExpiry        = 15 * time.Minute
	passwordResetCooldown      = 60 * time.Second
	passwordResetMaxAttempts   = 5
	passwordResetGenericNotice = "If an active account exists for that email, we sent a reset code."
	authLoginAttemptLimit      = 8
	authLoginAttemptWindow     = 15 * time.Minute
	emailVerificationExpiry    = 24 * time.Hour
	emailVerificationCooldown  = 60 * time.Second
)

type AuthService struct {
	userRepository          domain.UserRepository
	sessionRepository       domain.UserSessionRepository
	passwordResetRepository domain.PasswordResetRequestRepository
	emailVerificationRepo   domain.EmailVerificationTokenRepository
	emailService            EmailService
	jwtSecret               []byte
	appBaseURL              string
	loginAttemptsMu         sync.Mutex
	loginAttempts           map[string]authLoginAttempt
}

type authLoginAttempt struct {
	count   int
	resetAt time.Time
}

type authClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const authSessionTouchInterval = time.Minute

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
		appBaseURL:     strings.TrimRight(strings.TrimSpace(cfg.AppBaseURL), "/"),
		loginAttempts:  make(map[string]authLoginAttempt),
	}, nil
}

func (s *AuthService) SetEmailService(emailService EmailService) {
	s.emailService = emailService
}

func (s *AuthService) SetSessionRepository(sessionRepository domain.UserSessionRepository) {
	s.sessionRepository = sessionRepository
}

func (s *AuthService) SetPasswordResetRepository(passwordResetRepository domain.PasswordResetRequestRepository) {
	s.passwordResetRepository = passwordResetRepository
}

func (s *AuthService) SetEmailVerificationRepository(emailVerificationRepo domain.EmailVerificationTokenRepository) {
	s.emailVerificationRepo = emailVerificationRepo
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
		EmailVerified: false,
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

	if err := s.sendEmailVerification(user, true); err != nil {
		return "", err
	}

	return "", nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	token, _, err := s.LoginWithSession(email, password, "", "")
	return token, err
}

func (s *AuthService) LoginWithSession(email, password, userAgent, ipAddress string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if s.isLoginRateLimited(email, ipAddress) {
		return "", "", ErrAuthRateLimited
	}

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.recordFailedLogin(email, ipAddress)
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("get user by email: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		s.recordFailedLogin(email, ipAddress)
		return "", "", ErrInvalidCredentials
	}
	if err := validateUserAccess(user); err != nil {
		return "", "", err
	}
	s.clearFailedLogin(email, ipAddress)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	sessionID := ""
	if s.sessionRepository != nil {
		session := &domain.UserSession{
			UserID:     user.ID,
			UserAgent:  strings.TrimSpace(userAgent),
			IPAddress:  strings.TrimSpace(ipAddress),
			LastSeenAt: time.Now().UTC(),
			ExpiresAt:  expiresAt,
		}
		if err := s.sessionRepository.Create(session); err != nil {
			return "", "", fmt.Errorf("create user session: %w", err)
		}
		sessionID = session.ID
	}

	token, err := s.generateToken(user.ID, user.Email, string(user.Role), sessionID, expiresAt)
	if err != nil {
		return "", "", err
	}

	return token, sessionID, nil
}

func (s *AuthService) isLoginRateLimited(email, ipAddress string) bool {
	key := authLoginAttemptKey(email, ipAddress)
	now := time.Now().UTC()

	s.loginAttemptsMu.Lock()
	defer s.loginAttemptsMu.Unlock()

	entry, ok := s.loginAttempts[key]
	if !ok {
		return false
	}
	if now.After(entry.resetAt) {
		delete(s.loginAttempts, key)
		return false
	}
	return entry.count >= authLoginAttemptLimit
}

func (s *AuthService) recordFailedLogin(email, ipAddress string) {
	key := authLoginAttemptKey(email, ipAddress)
	now := time.Now().UTC()

	s.loginAttemptsMu.Lock()
	defer s.loginAttemptsMu.Unlock()

	entry, ok := s.loginAttempts[key]
	if !ok || now.After(entry.resetAt) {
		s.loginAttempts[key] = authLoginAttempt{
			count:   1,
			resetAt: now.Add(authLoginAttemptWindow),
		}
		return
	}

	entry.count++
	s.loginAttempts[key] = entry
}

func (s *AuthService) clearFailedLogin(email, ipAddress string) {
	key := authLoginAttemptKey(email, ipAddress)
	s.loginAttemptsMu.Lock()
	defer s.loginAttemptsMu.Unlock()
	delete(s.loginAttempts, key)
}

func authLoginAttemptKey(email, ipAddress string) string {
	return strings.TrimSpace(strings.ToLower(email)) + "|" + strings.TrimSpace(ipAddress)
}

func (s *AuthService) RequestPasswordReset(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateAuthEmail(email); err != nil {
		return "", err
	}
	if s.passwordResetRepository == nil || s.emailService == nil {
		return "", ErrPasswordResetServiceMissing
	}

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return passwordResetGenericNotice, nil
		}
		return "", fmt.Errorf("get user by email for password reset: %w", err)
	}
	if user.Role != domain.EthiopianAgent && user.Role != domain.ForeignAgent {
		return passwordResetGenericNotice, nil
	}
	if err := validateUserAccess(user); err != nil {
		return passwordResetGenericNotice, nil
	}

	latest, err := s.passwordResetRepository.GetLatestByUserID(user.ID)
	if err != nil && !errors.Is(err, repository.ErrPasswordResetRequestNotFound) {
		return "", fmt.Errorf("get latest password reset request: %w", err)
	}
	if latest != nil && time.Since(latest.CreatedAt.UTC()) < passwordResetCooldown {
		return passwordResetGenericNotice, nil
	}

	if err := s.passwordResetRepository.InvalidateActiveByUserID(user.ID); err != nil {
		return "", fmt.Errorf("invalidate active password reset requests: %w", err)
	}

	code, err := generateNumericResetCode(passwordResetCodeLength)
	if err != nil {
		return "", fmt.Errorf("generate password reset code: %w", err)
	}

	request := &domain.PasswordResetRequest{
		UserID:    user.ID,
		CodeHash:  s.hashPasswordResetCode(user.ID, code),
		ExpiresAt: time.Now().UTC().Add(passwordResetExpiry),
	}
	if err := s.passwordResetRepository.Create(request); err != nil {
		return "", fmt.Errorf("create password reset request: %w", err)
	}

	if err := s.emailService.Send(user.Email, "Your password reset code", s.buildPasswordResetEmailBody(user, code, request.ExpiresAt)); err != nil {
		_ = s.passwordResetRepository.Delete(request.ID)
		return "", fmt.Errorf("send password reset email: %w", err)
	}

	return passwordResetGenericNotice, nil
}

func (s *AuthService) VerifyEmail(token string) (*domain.User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrEmailVerificationInvalid
	}
	if s.emailVerificationRepo == nil || s.emailService == nil {
		return nil, ErrEmailVerificationMissing
	}

	tokenHash := s.hashEmailVerificationToken(token)
	record, err := s.emailVerificationRepo.GetActiveByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrEmailVerificationTokenNotFound) {
			return nil, ErrEmailVerificationExpired
		}
		return nil, fmt.Errorf("get active email verification token: %w", err)
	}

	user, err := s.userRepository.GetByID(record.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrEmailVerificationInvalid
		}
		return nil, fmt.Errorf("get user for email verification: %w", err)
	}

	if user.EmailVerified {
		_ = s.emailVerificationRepo.MarkUsed(record.ID)
		return user, nil
	}

	user.EmailVerified = true
	if err := s.userRepository.Update(user); err != nil {
		return nil, fmt.Errorf("mark user email verified: %w", err)
	}
	if err := s.emailVerificationRepo.MarkUsed(record.ID); err != nil {
		return nil, fmt.Errorf("mark email verification token used: %w", err)
	}

	return user, nil
}

func (s *AuthService) ResendEmailVerification(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateAuthEmail(email); err != nil {
		return "", err
	}
	if s.emailVerificationRepo == nil || s.emailService == nil {
		return "", ErrEmailVerificationMissing
	}

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "If that account exists, a verification email has been sent.", nil
		}
		return "", fmt.Errorf("get user by email for verification resend: %w", err)
	}
	if user.EmailVerified {
		return "That email is already verified.", nil
	}

	if latest, err := s.emailVerificationRepo.GetLatestByUserID(user.ID); err == nil {
		if time.Since(latest.CreatedAt.UTC()) < emailVerificationCooldown {
			return "A verification email was sent recently. Please check your inbox.", nil
		}
	} else if !errors.Is(err, repository.ErrEmailVerificationTokenNotFound) {
		return "", fmt.Errorf("get latest email verification token: %w", err)
	}

	if err := s.sendEmailVerification(user, false); err != nil {
		return "", err
	}
	return "A fresh verification email has been sent.", nil
}

func (s *AuthService) ResetPassword(email, code, newPassword string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	if err := validateAuthEmail(email); err != nil {
		return err
	}
	if len(code) != passwordResetCodeLength {
		return ErrPasswordResetCodeInvalid
	}
	if len(strings.TrimSpace(newPassword)) < 8 {
		return repository.ErrInvalidPassword
	}
	if s.passwordResetRepository == nil {
		return ErrPasswordResetServiceMissing
	}

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrPasswordResetCodeInvalid
		}
		return fmt.Errorf("get user by email for password reset: %w", err)
	}
	if user.Role != domain.EthiopianAgent && user.Role != domain.ForeignAgent {
		return ErrPasswordResetCodeInvalid
	}
	if err := validateUserAccess(user); err != nil {
		return ErrPasswordResetCodeInvalid
	}

	request, err := s.passwordResetRepository.GetLatestActiveByUserID(user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrPasswordResetRequestNotFound) {
			return ErrPasswordResetCodeExpired
		}
		return fmt.Errorf("get active password reset request: %w", err)
	}
	if request.AttemptCount >= passwordResetMaxAttempts {
		return ErrPasswordResetCodeInvalid
	}
	if time.Now().UTC().After(request.ExpiresAt.UTC()) {
		return ErrPasswordResetCodeExpired
	}

	expectedHash := s.hashPasswordResetCode(user.ID, code)
	if !hmac.Equal([]byte(request.CodeHash), []byte(expectedHash)) {
		if err := s.passwordResetRepository.IncrementAttempts(request.ID); err != nil {
			return fmt.Errorf("increment password reset attempts: %w", err)
		}
		return ErrPasswordResetCodeInvalid
	}

	user.PasswordHash = newPassword
	if err := s.userRepository.Update(user); err != nil {
		return err
	}
	if s.sessionRepository != nil {
		_ = s.sessionRepository.RevokeAllByUserID(user.ID, "")
	}

	if err := s.passwordResetRepository.MarkUsed(request.ID); err != nil {
		return fmt.Errorf("mark password reset request used: %w", err)
	}

	return nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, string, error) {
	userID, role, _, err := s.ValidateTokenWithSession(tokenString)
	return userID, role, err
}

func (s *AuthService) ValidateTokenWithSession(tokenString string) (string, string, string, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", "", "", ErrInvalidToken
	}

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return "", "", "", ErrInvalidToken
	}

	if claims.UserID == "" || claims.Role == "" {
		return "", "", "", ErrInvalidToken
	}

	if s.sessionRepository != nil && strings.TrimSpace(claims.ID) != "" {
		session, err := s.sessionRepository.GetByID(claims.ID)
		if err != nil {
			if errors.Is(err, repository.ErrUserSessionNotFound) {
				return "", "", "", ErrInvalidToken
			}
			return "", "", "", fmt.Errorf("get user session by id: %w", err)
		}
		if strings.TrimSpace(session.UserID) != claims.UserID || session.RevokedAt != nil || time.Now().UTC().After(session.ExpiresAt.UTC()) {
			return "", "", "", ErrInvalidToken
		}
		if time.Since(session.LastSeenAt.UTC()) >= authSessionTouchInterval {
			_ = s.sessionRepository.Touch(session.ID, time.Now().UTC())
		}
	}

	user, err := s.userRepository.GetByID(claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", "", "", ErrInvalidToken
		}
		return "", "", "", fmt.Errorf("get user by id: %w", err)
	}
	if err := validateUserAccess(user); err != nil {
		return "", "", "", err
	}

	return claims.UserID, claims.Role, claims.ID, nil
}

func (s *AuthService) generateToken(userID, email, role, sessionID string, expiresAt time.Time) (string, error) {
	now := time.Now().UTC()
	if expiresAt.IsZero() {
		expiresAt = now.Add(24 * time.Hour)
	}
	claims := authClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        strings.TrimSpace(sessionID),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
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

func (s *AuthService) RevokeSession(userID, sessionID string) error {
	if s.sessionRepository == nil || strings.TrimSpace(userID) == "" || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	if err := s.sessionRepository.RevokeByID(userID, sessionID); err != nil && !errors.Is(err, repository.ErrUserSessionNotFound) {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
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
	if !user.EmailVerified {
		return ErrEmailNotVerified
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

func (s *AuthService) sendEmailVerification(user *domain.User, invalidateExisting bool) error {
	if user == nil {
		return ErrEmailVerificationInvalid
	}
	if s.emailVerificationRepo == nil || s.emailService == nil {
		return ErrEmailVerificationMissing
	}
	if invalidateExisting {
		if err := s.emailVerificationRepo.InvalidateActiveByUserID(user.ID); err != nil {
			return fmt.Errorf("invalidate active email verification tokens: %w", err)
		}
	}

	rawToken, err := generateSecureHexToken(32)
	if err != nil {
		return fmt.Errorf("generate email verification token: %w", err)
	}

	record := &domain.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: s.hashEmailVerificationToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(emailVerificationExpiry),
	}
	if err := s.emailVerificationRepo.Create(record); err != nil {
		return fmt.Errorf("create email verification token: %w", err)
	}

	if err := s.emailService.Send(
		user.Email,
		"Verify your email address",
		s.buildEmailVerificationBody(user, rawToken),
	); err != nil {
		_ = s.emailVerificationRepo.Delete(record.ID)
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

func (s *AuthService) hashEmailVerificationToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) buildEmailVerificationBody(user *domain.User, token string) string {
	var builder strings.Builder
	builder.WriteString("Hello ")
	if strings.TrimSpace(user.FullName) != "" {
		builder.WriteString(strings.TrimSpace(user.FullName))
	} else {
		builder.WriteString("there")
	}
	builder.WriteString(",\n\n")
	builder.WriteString("Please verify your email address before your agency registration can continue.\n\n")
	if s.appBaseURL != "" {
		builder.WriteString("Verification link: ")
		builder.WriteString(s.appBaseURL)
		builder.WriteString("/register/verify?token=")
		builder.WriteString(url.QueryEscape(token))
		builder.WriteString("\n\n")
	}
	builder.WriteString("This verification link expires in 24 hours.\n")
	builder.WriteString("If you did not create this account, you can ignore this email.\n")
	return builder.String()
}

func generateSecureHexToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("byte length must be positive")
	}
	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func generateNumericResetCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	var builder strings.Builder
	builder.Grow(length)
	for i := 0; i < length; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + value.Int64()))
	}

	return builder.String(), nil
}

func (s *AuthService) hashPasswordResetCode(userID, code string) string {
	mac := hmac.New(sha256.New, s.jwtSecret)
	_, _ = mac.Write([]byte(strings.TrimSpace(userID)))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *AuthService) buildPasswordResetEmailBody(user *domain.User, code string, expiresAt time.Time) string {
	var builder strings.Builder

	builder.WriteString("Hello ")
	if strings.TrimSpace(user.FullName) != "" {
		builder.WriteString(strings.TrimSpace(user.FullName))
	} else {
		builder.WriteString("there")
	}
	builder.WriteString(",\n\n")
	builder.WriteString("Use this one-time code to reset your password:\n\n")
	builder.WriteString(code)
	builder.WriteString("\n\n")
	builder.WriteString("This code expires in 15 minutes.\n")
	builder.WriteString("If you did not request a password reset, you can ignore this email.\n")

	if s.appBaseURL != "" {
		builder.WriteString("\nReset page: ")
		builder.WriteString(s.appBaseURL)
		builder.WriteString("/reset-password?email=")
		builder.WriteString(url.QueryEscape(user.Email))
		builder.WriteString("\n")
	}

	builder.WriteString("\nExpires at: ")
	builder.WriteString(expiresAt.UTC().Format(time.RFC1123))
	builder.WriteString("\n")

	return builder.String()
}
