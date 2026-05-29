package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type authUserRepoMock struct {
	getByEmailFn func(email string) (*domain.User, error)
	getByIDFn    func(id string) (*domain.User, error)
	createFn     func(user *domain.User) error
	updateFn     func(user *domain.User) error
}

func (m *authUserRepoMock) Create(user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(user)
	}
	return nil
}

func (m *authUserRepoMock) GetByEmail(email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(email)
	}
	return nil, repository.ErrUserNotFound
}

func (m *authUserRepoMock) GetByID(id string) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, repository.ErrUserNotFound
}

func (m *authUserRepoMock) Update(user *domain.User) error {
	if m.updateFn != nil {
		return m.updateFn(user)
	}
	return nil
}

type authPasswordResetRepoMock struct {
	createFn            func(request *domain.PasswordResetRequest) error
	getLatestByUserIDFn func(userID string) (*domain.PasswordResetRequest, error)
	getLatestActiveFn   func(userID string) (*domain.PasswordResetRequest, error)
	incrementAttemptsFn func(id string) error
	markUsedFn          func(id string) error
	invalidateActiveFn  func(userID string) error
	deleteFn            func(id string) error
}

type authEmailVerificationRepoMock struct {
	createFn             func(token *domain.EmailVerificationToken) error
	getActiveByTokenHash func(tokenHash string) (*domain.EmailVerificationToken, error)
	getLatestByUserIDFn  func(userID string) (*domain.EmailVerificationToken, error)
	invalidateActiveFn   func(userID string) error
	markUsedFn           func(id string) error
	deleteFn             func(id string) error
}

func (m *authEmailVerificationRepoMock) Create(token *domain.EmailVerificationToken) error {
	if m.createFn != nil {
		return m.createFn(token)
	}
	return nil
}

func (m *authEmailVerificationRepoMock) GetActiveByTokenHash(tokenHash string) (*domain.EmailVerificationToken, error) {
	if m.getActiveByTokenHash != nil {
		return m.getActiveByTokenHash(tokenHash)
	}
	return nil, repository.ErrEmailVerificationTokenNotFound
}

func (m *authEmailVerificationRepoMock) GetLatestByUserID(userID string) (*domain.EmailVerificationToken, error) {
	if m.getLatestByUserIDFn != nil {
		return m.getLatestByUserIDFn(userID)
	}
	return nil, repository.ErrEmailVerificationTokenNotFound
}

func (m *authEmailVerificationRepoMock) InvalidateActiveByUserID(userID string) error {
	if m.invalidateActiveFn != nil {
		return m.invalidateActiveFn(userID)
	}
	return nil
}

func (m *authEmailVerificationRepoMock) MarkUsed(id string) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(id)
	}
	return nil
}

func (m *authEmailVerificationRepoMock) Delete(id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *authPasswordResetRepoMock) Create(request *domain.PasswordResetRequest) error {
	if m.createFn != nil {
		return m.createFn(request)
	}
	return nil
}

func (m *authPasswordResetRepoMock) GetLatestByUserID(userID string) (*domain.PasswordResetRequest, error) {
	if m.getLatestByUserIDFn != nil {
		return m.getLatestByUserIDFn(userID)
	}
	return nil, repository.ErrPasswordResetRequestNotFound
}

func (m *authPasswordResetRepoMock) GetLatestActiveByUserID(userID string) (*domain.PasswordResetRequest, error) {
	if m.getLatestActiveFn != nil {
		return m.getLatestActiveFn(userID)
	}
	return nil, repository.ErrPasswordResetRequestNotFound
}

func (m *authPasswordResetRepoMock) IncrementAttempts(id string) error {
	if m.incrementAttemptsFn != nil {
		return m.incrementAttemptsFn(id)
	}
	return nil
}

func (m *authPasswordResetRepoMock) MarkUsed(id string) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(id)
	}
	return nil
}

func (m *authPasswordResetRepoMock) InvalidateActiveByUserID(userID string) error {
	if m.invalidateActiveFn != nil {
		return m.invalidateActiveFn(userID)
	}
	return nil
}

func (m *authPasswordResetRepoMock) Delete(id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

type authEmailServiceMock struct {
	sent   []authEmailRecord
	sendFn func(to, subject, body string) error
}

type authEmailRecord struct {
	to      string
	subject string
	body    string
}

func (m *authEmailServiceMock) Send(to, subject, body string) error {
	if m.sendFn != nil {
		return m.sendFn(to, subject, body)
	}
	m.sent = append(m.sent, authEmailRecord{to: to, subject: subject, body: body})
	return nil
}

func TestAuthService_RegisterLoginValidate(t *testing.T) {
	usersByEmail := map[string]*domain.User{}
	usersByID := map[string]*domain.User{}
	repo := &authUserRepoMock{}
	repo.getByEmailFn = func(email string) (*domain.User, error) {
		if user, ok := usersByEmail[email]; ok {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}
	repo.getByIDFn = func(id string) (*domain.User, error) {
		if user, ok := usersByID[id]; ok {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}
	repo.createFn = func(user *domain.User) error {
		if _, exists := usersByEmail[user.Email]; exists {
			return repository.ErrDuplicateEmail
		}
		user.ID = "user-1"
		usersByEmail[user.Email] = user
		usersByID[user.ID] = user
		return nil
	}

	svc, err := NewAuthService(repo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)
	svc.SetEmailService(&authEmailServiceMock{})
	svc.SetEmailVerificationRepository(&authEmailVerificationRepoMock{})

	token, err := svc.Register("Agent@Example.com", "password123", "Agent", string(domain.ForeignAgent), "Agency")
	require.NoError(t, err)
	assert.Empty(t, token)

	_, err = svc.Register("agent@example.com", "password123", "Agent", string(domain.ForeignAgent), "Agency")
	require.ErrorIs(t, err, ErrUserExists)

	registeredUser := usersByEmail["agent@example.com"]
	require.NotNil(t, registeredUser)
	registeredUser.EmailVerified = true

	loginToken, err := svc.Login("agent@example.com", "password123")
	require.ErrorIs(t, err, ErrAccountPendingApproval)

	registeredUser.AccountStatus = domain.AccountStatusActive

	loginToken, err = svc.Login("agent@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, loginToken)

	userID, role, err := svc.ValidateToken(loginToken)
	require.NoError(t, err)
	assert.Equal(t, "user-1", userID)
	assert.Equal(t, string(domain.ForeignAgent), role)
}

func TestAuthService_LoginInvalidCredentials(t *testing.T) {
	repo := &authUserRepoMock{
		getByEmailFn: func(email string) (*domain.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}
	svc, err := NewAuthService(repo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)

	_, err = svc.Login("missing@example.com", "password123")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_ValidateTokenInvalid(t *testing.T) {
	repo := &authUserRepoMock{}
	svc, err := NewAuthService(repo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)

	_, _, err = svc.ValidateToken("not-a-token")
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthService_NewAuthServiceValidation(t *testing.T) {
	_, err := NewAuthService(nil, &config.Config{JWTSecret: "x"})
	require.Error(t, err)

	_, err = NewAuthService(&authUserRepoMock{}, nil)
	require.Error(t, err)

	_, err = NewAuthService(&authUserRepoMock{}, &config.Config{})
	require.Error(t, err)
}

func TestAuthService_RegisterUnexpectedLookupError(t *testing.T) {
	repo := &authUserRepoMock{
		getByEmailFn: func(email string) (*domain.User, error) {
			return nil, errors.New("db unavailable")
		},
	}
	svc, err := NewAuthService(repo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)

	_, err = svc.Register("test@example.com", "password123", "Agent", string(domain.ForeignAgent), "Agency")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check existing user")
}

func TestAuthService_RequestPasswordResetCreatesOtpAndEmailsUser(t *testing.T) {
	user := &domain.User{
		ID:            "user-1",
		Email:         "agent@example.com",
		EmailVerified: true,
		FullName:      "Agent",
		Role:          domain.ForeignAgent,
		AccountStatus: domain.AccountStatusActive,
		IsActive:      true,
	}

	var created *domain.PasswordResetRequest
	resetRepo := &authPasswordResetRepoMock{
		getLatestByUserIDFn: func(userID string) (*domain.PasswordResetRequest, error) {
			return nil, repository.ErrPasswordResetRequestNotFound
		},
		createFn: func(request *domain.PasswordResetRequest) error {
			created = request
			return nil
		},
	}
	emailService := &authEmailServiceMock{}
	userRepo := &authUserRepoMock{
		getByEmailFn: func(email string) (*domain.User, error) {
			return user, nil
		},
		updateFn: func(updated *domain.User) error {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updated.PasswordHash), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			updated.PasswordHash = string(hashedPassword)
			return nil
		},
	}

	svc, err := NewAuthService(userRepo, &config.Config{JWTSecret: "secret-key", AppBaseURL: "https://app.example.com"})
	require.NoError(t, err)
	svc.SetEmailService(emailService)
	svc.SetPasswordResetRepository(resetRepo)

	message, err := svc.RequestPasswordReset("agent@example.com")
	require.NoError(t, err)
	assert.Equal(t, passwordResetGenericNotice, message)
	require.NotNil(t, created)
	assert.Equal(t, user.ID, created.UserID)
	assert.NotEmpty(t, created.CodeHash)
	assert.True(t, created.ExpiresAt.After(time.Now().UTC()))
	require.Len(t, emailService.sent, 1)
	assert.Equal(t, user.Email, emailService.sent[0].to)
	assert.Contains(t, emailService.sent[0].body, "Reset page: https://app.example.com/reset-password?email=agent%40example.com")
}

func TestAuthService_RequestPasswordResetSkipsUnknownAndPendingAccounts(t *testing.T) {
	emailService := &authEmailServiceMock{}
	resetRepo := &authPasswordResetRepoMock{}
	userRepo := &authUserRepoMock{
		getByEmailFn: func(email string) (*domain.User, error) {
			switch email {
			case "pending@example.com":
				return &domain.User{ID: "user-2", Email: email, Role: domain.EthiopianAgent, AccountStatus: domain.AccountStatusPendingApproval, IsActive: true}, nil
			default:
				return nil, repository.ErrUserNotFound
			}
		},
	}

	svc, err := NewAuthService(userRepo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)
	svc.SetEmailService(emailService)
	svc.SetPasswordResetRepository(resetRepo)

	message, err := svc.RequestPasswordReset("missing@example.com")
	require.NoError(t, err)
	assert.Equal(t, passwordResetGenericNotice, message)

	message, err = svc.RequestPasswordReset("pending@example.com")
	require.NoError(t, err)
	assert.Equal(t, passwordResetGenericNotice, message)
	assert.Empty(t, emailService.sent)
}

func TestAuthService_ResetPasswordUpdatesPasswordAndConsumesRequest(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &domain.User{
		ID:            "user-1",
		Email:         "agent@example.com",
		EmailVerified: true,
		PasswordHash:  string(passwordHash),
		Role:          domain.ForeignAgent,
		AccountStatus: domain.AccountStatusActive,
		IsActive:      true,
	}

	var markedUsed string
	resetRepo := &authPasswordResetRepoMock{}
	userRepo := &authUserRepoMock{
		getByEmailFn: func(email string) (*domain.User, error) {
			return user, nil
		},
		updateFn: func(updated *domain.User) error {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updated.PasswordHash), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			updated.PasswordHash = string(hashedPassword)
			return nil
		},
	}

	svc, err := NewAuthService(userRepo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)
	resetRepo.getLatestActiveFn = func(userID string) (*domain.PasswordResetRequest, error) {
		return &domain.PasswordResetRequest{
			ID:        "reset-1",
			UserID:    userID,
			CodeHash:  svc.hashPasswordResetCode(userID, "123456"),
			ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		}, nil
	}
	resetRepo.markUsedFn = func(id string) error {
		markedUsed = id
		return nil
	}
	svc.SetPasswordResetRepository(resetRepo)

	err = svc.ResetPassword("agent@example.com", "123456", "new-password")
	require.NoError(t, err)
	assert.Equal(t, "reset-1", markedUsed)
	assert.True(t, strings.HasPrefix(user.PasswordHash, "$2"))
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("new-password")))
}

func TestAuthService_ResetPasswordInvalidCodeIncrementsAttempts(t *testing.T) {
	user := &domain.User{
		ID:            "user-1",
		Email:         "agent@example.com",
		EmailVerified: true,
		Role:          domain.ForeignAgent,
		AccountStatus: domain.AccountStatusActive,
		IsActive:      true,
	}

	var incremented string
	resetRepo := &authPasswordResetRepoMock{
		getLatestActiveFn: func(userID string) (*domain.PasswordResetRequest, error) {
			return &domain.PasswordResetRequest{
				ID:           "reset-2",
				UserID:       userID,
				CodeHash:     "wrong",
				ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
				AttemptCount: 0,
			}, nil
		},
		incrementAttemptsFn: func(id string) error {
			incremented = id
			return nil
		},
	}
	userRepo := &authUserRepoMock{
		getByEmailFn: func(email string) (*domain.User, error) {
			return user, nil
		},
	}

	svc, err := NewAuthService(userRepo, &config.Config{JWTSecret: "secret-key"})
	require.NoError(t, err)
	svc.SetPasswordResetRepository(resetRepo)

	err = svc.ResetPassword("agent@example.com", "654321", "new-password")
	require.ErrorIs(t, err, ErrPasswordResetCodeInvalid)
	assert.Equal(t, "reset-2", incremented)
}
