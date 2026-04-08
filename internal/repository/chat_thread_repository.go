package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrChatThreadNotFound    = errors.New("chat thread not found")
	ErrDuplicateChatThread   = errors.New("chat thread already exists")
)

type GormChatThreadRepository struct {
	db *gorm.DB
}

func NewGormChatThreadRepository(cfg *config.Config) (*GormChatThreadRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormChatThreadRepository{db: db}, nil
}

func (r *GormChatThreadRepository) DB() *gorm.DB {
	return r.db
}

func (r *GormChatThreadRepository) ResolveOrCreateWorkspaceThread(pairingID, createdByUserID string) (*domain.ChatThread, error) {
	pairingID = strings.TrimSpace(pairingID)
	createdByUserID = strings.TrimSpace(createdByUserID)

	if pairingID == "" {
		return nil, fmt.Errorf("resolve workspace chat thread: pairing id is required")
	}
	if createdByUserID == "" {
		return nil, fmt.Errorf("resolve workspace chat thread: created by user id is required")
	}

	thread, err := r.findWorkspaceThread(pairingID)
	if err == nil {
		return thread, nil
	}
	if !errors.Is(err, ErrChatThreadNotFound) {
		return nil, err
	}

	candidateID := (*string)(nil)
	thread = &domain.ChatThread{
		ID:              uuid.NewString(),
		PairingID:       pairingID,
		ScopeType:       domain.ChatThreadScopeWorkspace,
		CandidateID:     candidateID,
		CreatedByUserID: createdByUserID,
	}

	if err := r.db.Create(thread).Error; err != nil {
		if isDuplicateChatThreadError(err) {
			return r.findWorkspaceThread(pairingID)
		}
		return nil, fmt.Errorf("create workspace chat thread: %w", err)
	}

	return thread, nil
}

func (r *GormChatThreadRepository) ResolveOrCreateCandidateThread(pairingID, candidateID, createdByUserID string) (*domain.ChatThread, error) {
	pairingID = strings.TrimSpace(pairingID)
	candidateID = strings.TrimSpace(candidateID)
	createdByUserID = strings.TrimSpace(createdByUserID)

	if pairingID == "" {
		return nil, fmt.Errorf("resolve candidate chat thread: pairing id is required")
	}
	if candidateID == "" {
		return nil, fmt.Errorf("resolve candidate chat thread: candidate id is required")
	}
	if createdByUserID == "" {
		return nil, fmt.Errorf("resolve candidate chat thread: created by user id is required")
	}

	thread, err := r.findCandidateThread(pairingID, candidateID)
	if err == nil {
		return thread, nil
	}
	if !errors.Is(err, ErrChatThreadNotFound) {
		return nil, err
	}

	candidateValue := candidateID
	thread = &domain.ChatThread{
		ID:              uuid.NewString(),
		PairingID:       pairingID,
		ScopeType:       domain.ChatThreadScopeCandidate,
		CandidateID:     &candidateValue,
		CreatedByUserID: createdByUserID,
	}

	if err := r.db.Create(thread).Error; err != nil {
		if isDuplicateChatThreadError(err) {
			return r.findCandidateThread(pairingID, candidateID)
		}
		return nil, fmt.Errorf("create candidate chat thread: %w", err)
	}

	return thread, nil
}

func (r *GormChatThreadRepository) GetByID(id string) (*domain.ChatThread, error) {
	var thread domain.ChatThread
	if err := r.db.Where("id = ?", strings.TrimSpace(id)).First(&thread).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatThreadNotFound
		}
		return nil, fmt.Errorf("get chat thread by id: %w", err)
	}

	return &thread, nil
}

func (r *GormChatThreadRepository) ListByPairingID(pairingID string) ([]*domain.ChatThread, error) {
	pairingID = strings.TrimSpace(pairingID)
	if pairingID == "" {
		return nil, fmt.Errorf("list chat threads: pairing id is required")
	}

	threads := make([]*domain.ChatThread, 0)
	if err := r.db.
		Where("pairing_id = ?", pairingID).
		Order("last_message_at DESC NULLS LAST").
		Order("created_at DESC").
		Find(&threads).Error; err != nil {
		return nil, fmt.Errorf("list chat threads: %w", err)
	}

	return threads, nil
}

func (r *GormChatThreadRepository) UpdateLastMessage(threadID string, lastMessageAt time.Time, lastMessagePreview string) error {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return fmt.Errorf("update chat thread last message: thread id is required")
	}

	preview := strings.TrimSpace(lastMessagePreview)
	result := r.db.Model(&domain.ChatThread{}).
		Where("id = ?", threadID).
		Updates(map[string]any{
			"last_message_at":      lastMessageAt.UTC(),
			"last_message_preview": preview,
			"updated_at":           time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("update chat thread last message: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrChatThreadNotFound
	}

	return nil
}

func (r *GormChatThreadRepository) findWorkspaceThread(pairingID string) (*domain.ChatThread, error) {
	var thread domain.ChatThread
	if err := r.db.
		Where("pairing_id = ? AND scope_type = ?", pairingID, domain.ChatThreadScopeWorkspace).
		First(&thread).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatThreadNotFound
		}
		return nil, fmt.Errorf("find workspace chat thread: %w", err)
	}
	return &thread, nil
}

func (r *GormChatThreadRepository) findCandidateThread(pairingID, candidateID string) (*domain.ChatThread, error) {
	var thread domain.ChatThread
	if err := r.db.
		Where("pairing_id = ? AND scope_type = ? AND candidate_id = ?", pairingID, domain.ChatThreadScopeCandidate, candidateID).
		First(&thread).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatThreadNotFound
		}
		return nil, fmt.Errorf("find candidate chat thread: %w", err)
	}
	return &thread, nil
}

func isDuplicateChatThreadError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "23505" {
			return false
		}
		return strings.Contains(pgErr.Message, "idx_chat_threads_workspace_unique") ||
			strings.Contains(pgErr.Message, "idx_chat_threads_candidate_unique")
	}
	return false
}
