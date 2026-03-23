package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type authUserRepoMock struct {
	getByEmailFn func(email string) (*domain.User, error)
	getByIDFn    func(id string) (*domain.User, error)
	createFn     func(user *domain.User) error
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
func (m *authUserRepoMock) Update(user *domain.User) error { return nil }

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

	token, err := svc.Register("Agent@Example.com", "password123", "Agent", string(domain.ForeignAgent), "Agency")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	_, err = svc.Register("agent@example.com", "password123", "Agent", string(domain.ForeignAgent), "Agency")
	require.ErrorIs(t, err, ErrUserExists)

	loginToken, err := svc.Login("agent@example.com", "password123")
	require.ErrorIs(t, err, ErrAccountPendingApproval)

	registeredUser := usersByEmail["agent@example.com"]
	require.NotNil(t, registeredUser)
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
