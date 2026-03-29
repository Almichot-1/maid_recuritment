package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrUserSessionNotFound = errors.New("user session not found")

type GormUserSessionRepository struct {
	db *gorm.DB
}

func NewGormUserSessionRepository(cfg *config.Config) (*GormUserSessionRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormUserSessionRepository{db: db}, nil
}

func (r *GormUserSessionRepository) Create(session *domain.UserSession) error {
	if session == nil {
		return fmt.Errorf("create user session: session is nil")
	}
	if strings.TrimSpace(session.UserID) == "" {
		return fmt.Errorf("create user session: user id is required")
	}
	if session.ExpiresAt.IsZero() {
		return fmt.Errorf("create user session: expires at is required")
	}
	if strings.TrimSpace(session.ID) == "" {
		session.ID = uuid.NewString()
	}
	if session.LastSeenAt.IsZero() {
		session.LastSeenAt = time.Now().UTC()
	}

	if err := r.db.Create(session).Error; err != nil {
		return fmt.Errorf("create user session: %w", err)
	}
	return nil
}

func (r *GormUserSessionRepository) GetByID(id string) (*domain.UserSession, error) {
	var session domain.UserSession
	if err := r.db.Where("id = ?", strings.TrimSpace(id)).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserSessionNotFound
		}
		return nil, fmt.Errorf("get user session by id: %w", err)
	}
	return &session, nil
}

func (r *GormUserSessionRepository) ListActiveByUserID(userID string) ([]*domain.UserSession, error) {
	sessions := make([]*domain.UserSession, 0)
	if err := r.db.
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", strings.TrimSpace(userID), time.Now().UTC()).
		Order("last_seen_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("list active user sessions: %w", err)
	}
	return sessions, nil
}

func (r *GormUserSessionRepository) Touch(id string, lastSeenAt time.Time) error {
	result := r.db.Model(&domain.UserSession{}).
		Where("id = ? AND revoked_at IS NULL", strings.TrimSpace(id)).
		Updates(map[string]any{
			"last_seen_at": lastSeenAt.UTC(),
			"updated_at":   time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("touch user session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserSessionNotFound
	}
	return nil
}

func (r *GormUserSessionRepository) RevokeByID(userID, sessionID string) error {
	now := time.Now().UTC()
	result := r.db.Model(&domain.UserSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", strings.TrimSpace(sessionID), strings.TrimSpace(userID)).
		Updates(map[string]any{
			"revoked_at": now,
			"updated_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("revoke user session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserSessionNotFound
	}
	return nil
}

func (r *GormUserSessionRepository) RevokeAllByUserID(userID, exceptSessionID string) error {
	now := time.Now().UTC()
	query := r.db.Model(&domain.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", strings.TrimSpace(userID))
	if strings.TrimSpace(exceptSessionID) != "" {
		query = query.Where("id <> ?", strings.TrimSpace(exceptSessionID))
	}
	if err := query.Updates(map[string]any{
		"revoked_at": now,
		"updated_at": now,
	}).Error; err != nil {
		return fmt.Errorf("revoke all user sessions: %w", err)
	}
	return nil
}
