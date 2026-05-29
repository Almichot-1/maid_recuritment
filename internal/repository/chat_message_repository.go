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

var ErrChatMessageNotFound = errors.New("chat message not found")

type GormChatMessageRepository struct {
	db *gorm.DB
}

type chatMessageWithSenderRow struct {
	ID                string    `gorm:"column:id"`
	ThreadID          string    `gorm:"column:thread_id"`
	SenderUserID      string    `gorm:"column:sender_user_id"`
	Body              string    `gorm:"column:body"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	SenderID          string    `gorm:"column:sender_id"`
	SenderFullName    string    `gorm:"column:sender_full_name"`
	SenderCompanyName string    `gorm:"column:sender_company_name"`
	SenderRole        string    `gorm:"column:sender_role"`
}

type unreadThreadCountRow struct {
	ThreadID     string `gorm:"column:thread_id"`
	UnreadCount  int64  `gorm:"column:unread_count"`
}

type unreadSummaryRow struct {
	UnreadThreads  int64 `gorm:"column:unread_threads"`
	UnreadMessages int64 `gorm:"column:unread_messages"`
}

func NewGormChatMessageRepository(cfg *config.Config) (*GormChatMessageRepository, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &GormChatMessageRepository{db: db}, nil
}

func (r *GormChatMessageRepository) DB() *gorm.DB {
	return r.db
}

func (r *GormChatMessageRepository) Create(message *domain.ChatMessage) error {
	if message == nil {
		return fmt.Errorf("create chat message: message is nil")
	}
	if strings.TrimSpace(message.ID) == "" {
		message.ID = uuid.NewString()
	}
	if strings.TrimSpace(message.ThreadID) == "" {
		return fmt.Errorf("create chat message: thread id is required")
	}
	if strings.TrimSpace(message.SenderUserID) == "" {
		return fmt.Errorf("create chat message: sender user id is required")
	}
	if strings.TrimSpace(message.Body) == "" {
		return fmt.Errorf("create chat message: body is required")
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}

	if err := r.db.Create(message).Error; err != nil {
		return fmt.Errorf("create chat message: %w", err)
	}

	return nil
}

func (r *GormChatMessageRepository) GetByIDWithSender(id string) (*domain.ChatMessageWithSender, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("get chat message by id: id is required")
	}

	row := chatMessageWithSenderRow{}
	err := r.db.Raw(`
		SELECT
			m.id,
			m.thread_id,
			m.sender_user_id,
			m.body,
			m.created_at,
			u.id AS sender_id,
			COALESCE(u.full_name, '') AS sender_full_name,
			COALESCE(u.company_name, '') AS sender_company_name,
			COALESCE(CAST(u.role AS text), '') AS sender_role
		FROM public.chat_messages m
		JOIN public.users u ON u.id = m.sender_user_id
		WHERE m.id = ?
		LIMIT 1
	`, id).Scan(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get chat message by id: %w", err)
	}
	if strings.TrimSpace(row.ID) == "" {
		return nil, ErrChatMessageNotFound
	}

	mapped := mapChatMessageWithSender(row)
	return &mapped, nil
}

func (r *GormChatMessageRepository) GetLatestByThreadID(threadID string) (*domain.ChatMessageWithSender, error) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return nil, fmt.Errorf("get latest chat message by thread: thread id is required")
	}

	rows, err := r.getMessageRowsByQuery(`
		SELECT
			m.id,
			m.thread_id,
			m.sender_user_id,
			m.body,
			m.created_at,
			u.id AS sender_id,
			COALESCE(u.full_name, '') AS sender_full_name,
			COALESCE(u.company_name, '') AS sender_company_name,
			COALESCE(CAST(u.role AS text), '') AS sender_role
		FROM public.chat_messages m
		JOIN public.users u ON u.id = m.sender_user_id
		WHERE m.thread_id = ?
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT 1
	`, threadID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrChatMessageNotFound
	}

	mapped := mapChatMessageWithSender(rows[0])
	return &mapped, nil
}

func (r *GormChatMessageRepository) GetLatestByThreadIDs(threadIDs []string) (map[string]*domain.ChatMessageWithSender, error) {
	if len(threadIDs) == 0 {
		return map[string]*domain.ChatMessageWithSender{}, nil
	}

	sanitized := sanitizeIDs(threadIDs)
	if len(sanitized) == 0 {
		return map[string]*domain.ChatMessageWithSender{}, nil
	}

	rows := make([]chatMessageWithSenderRow, 0)
	err := r.db.Raw(`
		SELECT DISTINCT ON (m.thread_id)
			m.id,
			m.thread_id,
			m.sender_user_id,
			m.body,
			m.created_at,
			u.id AS sender_id,
			COALESCE(u.full_name, '') AS sender_full_name,
			COALESCE(u.company_name, '') AS sender_company_name,
			COALESCE(CAST(u.role AS text), '') AS sender_role
		FROM public.chat_messages m
		JOIN public.users u ON u.id = m.sender_user_id
		WHERE m.thread_id IN ?
		ORDER BY m.thread_id, m.created_at DESC, m.id DESC
	`, sanitized).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get latest chat messages by thread ids: %w", err)
	}

	result := make(map[string]*domain.ChatMessageWithSender, len(rows))
	for _, row := range rows {
		mapped := mapChatMessageWithSender(row)
		value := mapped
		result[row.ThreadID] = &value
	}

	return result, nil
}

func (r *GormChatMessageRepository) ListByThreadID(threadID string, cursor *domain.ChatMessageCursor, limit int) ([]*domain.ChatMessageWithSender, error) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return nil, fmt.Errorf("list chat messages by thread: thread id is required")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.Raw(`
		SELECT
			m.id,
			m.thread_id,
			m.sender_user_id,
			m.body,
			m.created_at,
			u.id AS sender_id,
			COALESCE(u.full_name, '') AS sender_full_name,
			COALESCE(u.company_name, '') AS sender_company_name,
			COALESCE(CAST(u.role AS text), '') AS sender_role
		FROM public.chat_messages m
		JOIN public.users u ON u.id = m.sender_user_id
		WHERE m.thread_id = ?
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT ?
	`, threadID, limit)

	if cursor != nil {
		if cursor.CreatedAt.IsZero() || strings.TrimSpace(cursor.ID) == "" {
			return nil, fmt.Errorf("list chat messages by thread: invalid cursor")
		}
		query = r.db.Raw(`
			SELECT
				m.id,
				m.thread_id,
				m.sender_user_id,
				m.body,
				m.created_at,
				u.id AS sender_id,
				COALESCE(u.full_name, '') AS sender_full_name,
				COALESCE(u.company_name, '') AS sender_company_name,
				COALESCE(CAST(u.role AS text), '') AS sender_role
			FROM public.chat_messages m
			JOIN public.users u ON u.id = m.sender_user_id
			WHERE m.thread_id = ?
				AND (
					m.created_at < ?
					OR (m.created_at = ? AND m.id < ?)
				)
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT ?
		`, threadID, cursor.CreatedAt.UTC(), cursor.CreatedAt.UTC(), strings.TrimSpace(cursor.ID), limit)
	}

	rows := make([]chatMessageWithSenderRow, 0)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list chat messages by thread: %w", err)
	}

	messages := make([]*domain.ChatMessageWithSender, 0, len(rows))
	for _, row := range rows {
		mapped := mapChatMessageWithSender(row)
		value := mapped
		messages = append(messages, &value)
	}

	return messages, nil
}

func (r *GormChatMessageRepository) CountUnreadByThreadAndUser(threadID, userID string) (int64, error) {
	threadID = strings.TrimSpace(threadID)
	userID = strings.TrimSpace(userID)
	if threadID == "" {
		return 0, fmt.Errorf("count unread chat messages by thread: thread id is required")
	}
	if userID == "" {
		return 0, fmt.Errorf("count unread chat messages by thread: user id is required")
	}

	var unreadCount int64
	err := r.db.Raw(`
		SELECT COUNT(m.id)
		FROM public.chat_messages m
		LEFT JOIN public.chat_thread_reads r ON r.thread_id = m.thread_id AND r.user_id = ?
		LEFT JOIN public.chat_messages rm ON rm.id = r.last_read_message_id
		WHERE m.thread_id = ?
			AND m.sender_user_id <> ?
			AND (
				r.last_read_message_id IS NULL
				OR m.created_at > rm.created_at
				OR (m.created_at = rm.created_at AND m.id > rm.id)
			)
	`, userID, threadID, userID).Scan(&unreadCount).Error
	if err != nil {
		return 0, fmt.Errorf("count unread chat messages by thread: %w", err)
	}

	return unreadCount, nil
}

func (r *GormChatMessageRepository) CountUnreadByThreadIDsAndUser(threadIDs []string, userID string) (map[string]int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("count unread chat messages by thread ids: user id is required")
	}

	sanitized := sanitizeIDs(threadIDs)
	if len(sanitized) == 0 {
		return map[string]int64{}, nil
	}

	rows := make([]unreadThreadCountRow, 0)
	err := r.db.Raw(`
		SELECT
			t.id AS thread_id,
			COUNT(m.id) AS unread_count
		FROM public.chat_threads t
		LEFT JOIN public.chat_thread_reads r ON r.thread_id = t.id AND r.user_id = ?
		LEFT JOIN public.chat_messages rm ON rm.id = r.last_read_message_id
		LEFT JOIN public.chat_messages m ON m.thread_id = t.id
			AND m.sender_user_id <> ?
			AND (
				r.last_read_message_id IS NULL
				OR m.created_at > rm.created_at
				OR (m.created_at = rm.created_at AND m.id > rm.id)
			)
		WHERE t.id IN ?
		GROUP BY t.id
	`, userID, userID, sanitized).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("count unread chat messages by thread ids: %w", err)
	}

	result := make(map[string]int64, len(rows))
	for _, row := range rows {
		result[row.ThreadID] = row.UnreadCount
	}

	return result, nil
}

func (r *GormChatMessageRepository) CountUnreadSummaryByPairingAndUser(pairingID, userID string) (int64, int64, error) {
	pairingID = strings.TrimSpace(pairingID)
	userID = strings.TrimSpace(userID)
	if pairingID == "" {
		return 0, 0, fmt.Errorf("count unread chat summary: pairing id is required")
	}
	if userID == "" {
		return 0, 0, fmt.Errorf("count unread chat summary: user id is required")
	}

	row := unreadSummaryRow{}
	err := r.db.Raw(`
		WITH per_thread AS (
			SELECT
				t.id AS thread_id,
				COUNT(m.id) AS unread_count
			FROM public.chat_threads t
			LEFT JOIN public.chat_thread_reads r ON r.thread_id = t.id AND r.user_id = ?
			LEFT JOIN public.chat_messages rm ON rm.id = r.last_read_message_id
			LEFT JOIN public.chat_messages m ON m.thread_id = t.id
				AND m.sender_user_id <> ?
				AND (
					r.last_read_message_id IS NULL
					OR m.created_at > rm.created_at
					OR (m.created_at = rm.created_at AND m.id > rm.id)
				)
			WHERE t.pairing_id = ?
			GROUP BY t.id
		)
		SELECT
			COUNT(*) FILTER (WHERE unread_count > 0) AS unread_threads,
			COALESCE(SUM(unread_count), 0) AS unread_messages
		FROM per_thread
	`, userID, userID, pairingID).Scan(&row).Error
	if err != nil {
		return 0, 0, fmt.Errorf("count unread chat summary: %w", err)
	}

	return row.UnreadThreads, row.UnreadMessages, nil
}

func (r *GormChatMessageRepository) getMessageRowsByQuery(query string, args ...any) ([]chatMessageWithSenderRow, error) {
	rows := make([]chatMessageWithSenderRow, 0)
	if err := r.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get chat message rows: %w", err)
	}
	return rows, nil
}

func mapChatMessageWithSender(row chatMessageWithSenderRow) domain.ChatMessageWithSender {
	return domain.ChatMessageWithSender{
		ID:           row.ID,
		ThreadID:     row.ThreadID,
		SenderUserID: row.SenderUserID,
		Body:         row.Body,
		CreatedAt:    row.CreatedAt,
		Sender: domain.ChatSenderSummary{
			UserID:      row.SenderID,
			FullName:    row.SenderFullName,
			CompanyName: row.SenderCompanyName,
			Role:        domain.UserRole(row.SenderRole),
		},
	}
}

func sanitizeIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
