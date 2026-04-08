package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var ErrChatThreadReadNotFound = errors.New("chat thread read not found")

type GormChatReadRepository struct {
	db *gorm.DB
}

func NewGormChatReadRepository(cfg *config.Config) (*GormChatReadRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormChatReadRepository{db: db}, nil
}

func (r *GormChatReadRepository) DB() *gorm.DB {
	return r.db
}

func (r *GormChatReadRepository) Upsert(threadID, userID string, lastReadMessageID *string, lastReadAt *time.Time) (*domain.ChatThreadRead, error) {
	threadID = strings.TrimSpace(threadID)
	userID = strings.TrimSpace(userID)
	if threadID == "" {
		return nil, fmt.Errorf("upsert chat thread read: thread id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("upsert chat thread read: user id is required")
	}

	now := time.Now().UTC()
	entry := &domain.ChatThreadRead{
		ID:                uuid.NewString(),
		ThreadID:          threadID,
		UserID:            userID,
		LastReadMessageID: normalizeOptionalID(lastReadMessageID),
		LastReadAt:        normalizeOptionalTime(lastReadAt),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "thread_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"last_read_message_id": entry.LastReadMessageID,
			"last_read_at":         entry.LastReadAt,
			"updated_at":           now,
		}),
	}).Create(entry).Error
	if err != nil {
		return nil, fmt.Errorf("upsert chat thread read: %w", err)
	}

	return r.GetByThreadAndUser(threadID, userID)
}

func (r *GormChatReadRepository) GetByThreadAndUser(threadID, userID string) (*domain.ChatThreadRead, error) {
	threadID = strings.TrimSpace(threadID)
	userID = strings.TrimSpace(userID)
	if threadID == "" {
		return nil, fmt.Errorf("get chat thread read by thread and user: thread id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("get chat thread read by thread and user: user id is required")
	}

	var entry domain.ChatThreadRead
	err := r.db.Where("thread_id = ? AND user_id = ?", threadID, userID).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatThreadReadNotFound
		}
		return nil, fmt.Errorf("get chat thread read by thread and user: %w", err)
	}

	return &entry, nil
}

func (r *GormChatReadRepository) ListByThreadIDsAndUser(threadIDs []string, userID string) (map[string]*domain.ChatThreadRead, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("list chat thread reads by thread ids and user: user id is required")
	}

	sanitized := sanitizeIDs(threadIDs)
	if len(sanitized) == 0 {
		return map[string]*domain.ChatThreadRead{}, nil
	}

	entries := make([]*domain.ChatThreadRead, 0)
	if err := r.db.Where("user_id = ? AND thread_id IN ?", userID, sanitized).Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("list chat thread reads by thread ids and user: %w", err)
	}

	result := make(map[string]*domain.ChatThreadRead, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		result[entry.ThreadID] = entry
	}

	return result, nil
}

func normalizeOptionalID(id *string) *string {
	if id == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*id)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
