package domain

import "time"

type ChatThreadScopeType string

const (
	ChatThreadScopeWorkspace ChatThreadScopeType = "workspace"
	ChatThreadScopeCandidate ChatThreadScopeType = "candidate"
)

type ChatThread struct {
	ID                 string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PairingID          string              `gorm:"type:uuid;not null;index"`
	ScopeType          ChatThreadScopeType `gorm:"type:text;not null"`
	CandidateID        *string             `gorm:"type:uuid"`
	CreatedByUserID    string              `gorm:"type:uuid;not null"`
	LastMessageAt      *time.Time
	LastMessagePreview *string
	CreatedAt          time.Time `gorm:"not null;default:now()"`
	UpdatedAt          time.Time `gorm:"not null;default:now()"`
}

func (ChatThread) TableName() string {
	return "chat_threads"
}

type ChatMessage struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ThreadID     string    `gorm:"type:uuid;not null;index"`
	SenderUserID string    `gorm:"type:uuid;not null;index"`
	Body         string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

type ChatThreadRead struct {
	ID                string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ThreadID          string    `gorm:"type:uuid;not null;index"`
	UserID            string    `gorm:"type:uuid;not null;index"`
	LastReadMessageID *string   `gorm:"type:uuid"`
	LastReadAt        *time.Time
	CreatedAt         time.Time `gorm:"not null;default:now()"`
	UpdatedAt         time.Time `gorm:"not null;default:now()"`
}

func (ChatThreadRead) TableName() string {
	return "chat_thread_reads"
}

type ChatMessageCursor struct {
	CreatedAt time.Time
	ID        string
}

type ChatSenderSummary struct {
	UserID      string
	FullName    string
	CompanyName string
	Role        UserRole
}

type ChatMessageWithSender struct {
	ID           string
	ThreadID     string
	SenderUserID string
	Body         string
	CreatedAt    time.Time
	Sender       ChatSenderSummary
}

type ChatThreadRepository interface {
	ResolveOrCreateWorkspaceThread(pairingID, createdByUserID string) (*ChatThread, error)
	ResolveOrCreateCandidateThread(pairingID, candidateID, createdByUserID string) (*ChatThread, error)
	GetByID(id string) (*ChatThread, error)
	ListByPairingID(pairingID string) ([]*ChatThread, error)
	UpdateLastMessage(threadID string, lastMessageAt time.Time, lastMessagePreview string) error
}

type ChatMessageRepository interface {
	Create(message *ChatMessage) error
	GetByIDWithSender(id string) (*ChatMessageWithSender, error)
	GetLatestByThreadID(threadID string) (*ChatMessageWithSender, error)
	GetLatestByThreadIDs(threadIDs []string) (map[string]*ChatMessageWithSender, error)
	ListByThreadID(threadID string, cursor *ChatMessageCursor, limit int) ([]*ChatMessageWithSender, error)
	CountUnreadByThreadAndUser(threadID, userID string) (int64, error)
	CountUnreadByThreadIDsAndUser(threadIDs []string, userID string) (map[string]int64, error)
	CountUnreadSummaryByPairingAndUser(pairingID, userID string) (int64, int64, error)
}

type ChatThreadReadRepository interface {
	Upsert(threadID, userID string, lastReadMessageID *string, lastReadAt *time.Time) (*ChatThreadRead, error)
	GetByThreadAndUser(threadID, userID string) (*ChatThreadRead, error)
	ListByThreadIDsAndUser(threadIDs []string, userID string) (map[string]*ChatThreadRead, error)
}
