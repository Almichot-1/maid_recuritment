package service

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

type chatNoopAuditRepository struct{}

func (r *chatNoopAuditRepository) Create(log *domain.AuditLog) error { return nil }
func (r *chatNoopAuditRepository) List(filters domain.AuditLogFilters) ([]*domain.AuditLog, error) {
	return []*domain.AuditLog{}, nil
}

type chatMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func (r *chatMemoryUserRepository) Create(user *domain.User) error {
	if user == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(user.ID) == "" {
		user.ID = uuid.NewString()
	}
	copied := *user
	r.users[user.ID] = &copied
	return nil
}

func (r *chatMemoryUserRepository) GetByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	trimmed := strings.TrimSpace(strings.ToLower(email))
	for _, user := range r.users {
		if user != nil && strings.TrimSpace(strings.ToLower(user.Email)) == trimmed {
			copied := *user
			return &copied, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (r *chatMemoryUserRepository) GetByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[strings.TrimSpace(id)]
	if !ok || user == nil {
		return nil, repository.ErrUserNotFound
	}
	copied := *user
	return &copied, nil
}

func (r *chatMemoryUserRepository) Update(user *domain.User) error {
	if user == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *user
	r.users[user.ID] = &copied
	return nil
}

type chatMemoryPairingRepository struct {
	mu       sync.RWMutex
	pairings map[string]*domain.AgencyPairing
}

func (r *chatMemoryPairingRepository) Create(pairing *domain.AgencyPairing) error {
	if pairing == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(pairing.ID) == "" {
		pairing.ID = uuid.NewString()
	}
	copied := *pairing
	r.pairings[pairing.ID] = &copied
	return nil
}

func (r *chatMemoryPairingRepository) GetByID(id string) (*domain.AgencyPairing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pairing, ok := r.pairings[strings.TrimSpace(id)]
	if !ok || pairing == nil {
		return nil, repository.ErrAgencyPairingNotFound
	}
	copied := *pairing
	return &copied, nil
}

func (r *chatMemoryPairingRepository) GetActiveByUsers(ethiopianUserID, foreignUserID string) (*domain.AgencyPairing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, pairing := range r.pairings {
		if pairing == nil {
			continue
		}
		if pairing.Status != domain.AgencyPairingActive {
			continue
		}
		if strings.TrimSpace(pairing.EthiopianUserID) == strings.TrimSpace(ethiopianUserID) && strings.TrimSpace(pairing.ForeignUserID) == strings.TrimSpace(foreignUserID) {
			copied := *pairing
			return &copied, nil
		}
	}
	return nil, repository.ErrAgencyPairingNotFound
}

func (r *chatMemoryPairingRepository) List(filters domain.AgencyPairingFilters) ([]*domain.AgencyPairing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.AgencyPairing, 0)
	for _, pairing := range r.pairings {
		if pairing == nil {
			continue
		}

		if strings.TrimSpace(filters.UserID) != "" {
			userID := strings.TrimSpace(filters.UserID)
			if strings.TrimSpace(pairing.EthiopianUserID) != userID && strings.TrimSpace(pairing.ForeignUserID) != userID {
				continue
			}
		}
		if strings.TrimSpace(filters.EthiopianUserID) != "" && strings.TrimSpace(pairing.EthiopianUserID) != strings.TrimSpace(filters.EthiopianUserID) {
			continue
		}
		if strings.TrimSpace(filters.ForeignUserID) != "" && strings.TrimSpace(pairing.ForeignUserID) != strings.TrimSpace(filters.ForeignUserID) {
			continue
		}
		if filters.Status != nil && pairing.Status != *filters.Status {
			continue
		}

		copied := *pairing
		result = append(result, &copied)
	}

	return result, nil
}

func (r *chatMemoryPairingRepository) Update(pairing *domain.AgencyPairing) error {
	if pairing == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.pairings[pairing.ID]; !ok {
		return repository.ErrAgencyPairingNotFound
	}
	copied := *pairing
	r.pairings[pairing.ID] = &copied
	return nil
}

type chatMemoryShareRepository struct {
	mu     sync.RWMutex
	shares map[string]*domain.CandidatePairShare
}

func (r *chatMemoryShareRepository) Create(share *domain.CandidatePairShare) error {
	if share == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(share.ID) == "" {
		share.ID = uuid.NewString()
	}
	if share.SharedAt.IsZero() {
		share.SharedAt = time.Now().UTC()
	}
	if !share.IsActive {
		share.IsActive = true
	}
	copied := *share
	r.shares[chatPairCandidateKey(share.PairingID, share.CandidateID)] = &copied
	return nil
}

func (r *chatMemoryShareRepository) GetActiveByPairingAndCandidate(pairingID, candidateID string) (*domain.CandidatePairShare, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	share, ok := r.shares[chatPairCandidateKey(pairingID, candidateID)]
	if !ok || share == nil || !share.IsActive {
		return nil, repository.ErrCandidatePairShareNotFound
	}
	copied := *share
	return &copied, nil
}

func (r *chatMemoryShareRepository) ListByCandidateID(candidateID string, activeOnly bool) ([]*domain.CandidatePairShare, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.CandidatePairShare, 0)
	for _, share := range r.shares {
		if share == nil {
			continue
		}
		if strings.TrimSpace(share.CandidateID) != strings.TrimSpace(candidateID) {
			continue
		}
		if activeOnly && !share.IsActive {
			continue
		}
		copied := *share
		result = append(result, &copied)
	}
	return result, nil
}

func (r *chatMemoryShareRepository) ListByPairingID(pairingID string, activeOnly bool) ([]*domain.CandidatePairShare, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.CandidatePairShare, 0)
	for _, share := range r.shares {
		if share == nil {
			continue
		}
		if strings.TrimSpace(share.PairingID) != strings.TrimSpace(pairingID) {
			continue
		}
		if activeOnly && !share.IsActive {
			continue
		}
		copied := *share
		result = append(result, &copied)
	}
	return result, nil
}

func (r *chatMemoryShareRepository) Deactivate(pairingID, candidateID string, unsharedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chatPairCandidateKey(pairingID, candidateID)
	share, ok := r.shares[key]
	if !ok || share == nil || !share.IsActive {
		return repository.ErrCandidatePairShareNotFound
	}
	share.IsActive = false
	share.UnsharedAt = &unsharedAt
	return nil
}

type chatMemoryCandidateRepository struct {
	mu         sync.RWMutex
	candidates map[string]*domain.Candidate
}

func (r *chatMemoryCandidateRepository) Create(candidate *domain.Candidate) error {
	if candidate == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(candidate.ID) == "" {
		candidate.ID = uuid.NewString()
	}
	copied := *candidate
	r.candidates[candidate.ID] = &copied
	return nil
}

func (r *chatMemoryCandidateRepository) GetByID(id string) (*domain.Candidate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	candidate, ok := r.candidates[strings.TrimSpace(id)]
	if !ok || candidate == nil {
		return nil, repository.ErrCandidateNotFound
	}
	copied := *candidate
	return &copied, nil
}

func (r *chatMemoryCandidateRepository) List(filters domain.CandidateFilters) ([]*domain.Candidate, error) {
	return []*domain.Candidate{}, nil
}

func (r *chatMemoryCandidateRepository) Update(candidate *domain.Candidate) error {
	if candidate == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *candidate
	r.candidates[candidate.ID] = &copied
	return nil
}

func (r *chatMemoryCandidateRepository) Delete(id string) error { return nil }
func (r *chatMemoryCandidateRepository) Lock(candidateID, lockedBy string, expiresAt time.Time) error {
	return nil
}
func (r *chatMemoryCandidateRepository) Unlock(candidateID string) error { return nil }

type chatMemorySelectionRepository struct {
	mu               sync.RWMutex
	selectionsByID   map[string]*domain.Selection
	selectionsByPair map[string]*domain.Selection
}

func (r *chatMemorySelectionRepository) Create(selection *domain.Selection) error {
	if selection == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(selection.ID) == "" {
		selection.ID = uuid.NewString()
	}
	copied := *selection
	r.selectionsByID[selection.ID] = &copied
	r.selectionsByPair[chatPairCandidateKey(selection.PairingID, selection.CandidateID)] = &copied
	return nil
}

func (r *chatMemorySelectionRepository) GetByID(id string) (*domain.Selection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	selection, ok := r.selectionsByID[strings.TrimSpace(id)]
	if !ok || selection == nil {
		return nil, repository.ErrSelectionNotFound
	}
	copied := *selection
	return &copied, nil
}

func (r *chatMemorySelectionRepository) GetByCandidateID(candidateID string) (*domain.Selection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, selection := range r.selectionsByID {
		if selection != nil && strings.TrimSpace(selection.CandidateID) == strings.TrimSpace(candidateID) {
			copied := *selection
			return &copied, nil
		}
	}
	return nil, repository.ErrSelectionNotFound
}

func (r *chatMemorySelectionRepository) GetByCandidateIDAndPairingID(candidateID, pairingID string) (*domain.Selection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	selection, ok := r.selectionsByPair[chatPairCandidateKey(pairingID, candidateID)]
	if !ok || selection == nil {
		return nil, repository.ErrSelectionNotFound
	}
	copied := *selection
	return &copied, nil
}

func (r *chatMemorySelectionRepository) GetBySelectedBy(userID string) ([]*domain.Selection, error) {
	return []*domain.Selection{}, nil
}

func (r *chatMemorySelectionRepository) GetBySelectedByAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return []*domain.Selection{}, nil
}

func (r *chatMemorySelectionRepository) GetByCandidateOwner(userID string) ([]*domain.Selection, error) {
	return []*domain.Selection{}, nil
}

func (r *chatMemorySelectionRepository) GetByCandidateOwnerAndPairing(userID, pairingID string) ([]*domain.Selection, error) {
	return []*domain.Selection{}, nil
}

func (r *chatMemorySelectionRepository) UpdateStatus(id string, status domain.SelectionStatus) error {
	return nil
}
func (r *chatMemorySelectionRepository) GetExpiredSelections() ([]*domain.Selection, error) {
	return []*domain.Selection{}, nil
}

type chatMemoryThreadRepository struct {
	mu                 sync.RWMutex
	threadsByID        map[string]*domain.ChatThread
	workspaceByPairing map[string]string
	candidateByPairing map[string]string
}

func (r *chatMemoryThreadRepository) ResolveOrCreateWorkspaceThread(pairingID, createdByUserID string) (*domain.ChatThread, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.TrimSpace(pairingID)
	if existingID, ok := r.workspaceByPairing[key]; ok {
		copied := *r.threadsByID[existingID]
		return &copied, nil
	}
	thread := &domain.ChatThread{
		ID:              uuid.NewString(),
		PairingID:       key,
		ScopeType:       domain.ChatThreadScopeWorkspace,
		CreatedByUserID: strings.TrimSpace(createdByUserID),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	r.threadsByID[thread.ID] = thread
	r.workspaceByPairing[key] = thread.ID
	copied := *thread
	return &copied, nil
}

func (r *chatMemoryThreadRepository) ResolveOrCreateCandidateThread(pairingID, candidateID, createdByUserID string) (*domain.ChatThread, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chatPairCandidateKey(pairingID, candidateID)
	if existingID, ok := r.candidateByPairing[key]; ok {
		copied := *r.threadsByID[existingID]
		return &copied, nil
	}
	candidate := strings.TrimSpace(candidateID)
	thread := &domain.ChatThread{
		ID:              uuid.NewString(),
		PairingID:       strings.TrimSpace(pairingID),
		ScopeType:       domain.ChatThreadScopeCandidate,
		CandidateID:     &candidate,
		CreatedByUserID: strings.TrimSpace(createdByUserID),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	r.threadsByID[thread.ID] = thread
	r.candidateByPairing[key] = thread.ID
	copied := *thread
	return &copied, nil
}

func (r *chatMemoryThreadRepository) GetByID(id string) (*domain.ChatThread, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	thread, ok := r.threadsByID[strings.TrimSpace(id)]
	if !ok || thread == nil {
		return nil, repository.ErrChatThreadNotFound
	}
	copied := *thread
	return &copied, nil
}

func (r *chatMemoryThreadRepository) ListByPairingID(pairingID string) ([]*domain.ChatThread, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.ChatThread, 0)
	for _, thread := range r.threadsByID {
		if thread != nil && strings.TrimSpace(thread.PairingID) == strings.TrimSpace(pairingID) {
			copied := *thread
			result = append(result, &copied)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		left := result[i]
		right := result[j]

		if left.LastMessageAt == nil && right.LastMessageAt == nil {
			return left.CreatedAt.After(right.CreatedAt)
		}
		if left.LastMessageAt == nil {
			return false
		}
		if right.LastMessageAt == nil {
			return true
		}
		if left.LastMessageAt.Equal(*right.LastMessageAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}
		return left.LastMessageAt.After(*right.LastMessageAt)
	})

	return result, nil
}

func (r *chatMemoryThreadRepository) UpdateLastMessage(threadID string, lastMessageAt time.Time, lastMessagePreview string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	thread, ok := r.threadsByID[strings.TrimSpace(threadID)]
	if !ok || thread == nil {
		return repository.ErrChatThreadNotFound
	}
	timestamp := lastMessageAt.UTC()
	preview := strings.TrimSpace(lastMessagePreview)
	thread.LastMessageAt = &timestamp
	thread.LastMessagePreview = &preview
	thread.UpdatedAt = time.Now().UTC()
	return nil
}

type chatMemoryReadRepository struct {
	mu    sync.RWMutex
	reads map[string]*domain.ChatThreadRead
}

func (r *chatMemoryReadRepository) Upsert(threadID, userID string, lastReadMessageID *string, lastReadAt *time.Time) (*domain.ChatThreadRead, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chatThreadUserKey(threadID, userID)
	now := time.Now().UTC()

	entry, ok := r.reads[key]
	if !ok || entry == nil {
		entry = &domain.ChatThreadRead{
			ID:        uuid.NewString(),
			ThreadID:  strings.TrimSpace(threadID),
			UserID:    strings.TrimSpace(userID),
			CreatedAt: now,
		}
		r.reads[key] = entry
	}

	entry.LastReadMessageID = chatOptionalTrimmed(lastReadMessageID)
	entry.LastReadAt = chatOptionalTime(lastReadAt)
	entry.UpdatedAt = now

	copied := *entry
	return &copied, nil
}

func (r *chatMemoryReadRepository) GetByThreadAndUser(threadID, userID string) (*domain.ChatThreadRead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.reads[chatThreadUserKey(threadID, userID)]
	if !ok || entry == nil {
		return nil, repository.ErrChatThreadReadNotFound
	}
	copied := *entry
	return &copied, nil
}

func (r *chatMemoryReadRepository) ListByThreadIDsAndUser(threadIDs []string, userID string) (map[string]*domain.ChatThreadRead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]*domain.ChatThreadRead, len(threadIDs))
	for _, threadID := range threadIDs {
		key := chatThreadUserKey(threadID, userID)
		if entry, ok := r.reads[key]; ok && entry != nil {
			copied := *entry
			result[strings.TrimSpace(threadID)] = &copied
		}
	}
	return result, nil
}

type chatMemoryMessageRepository struct {
	mu               sync.RWMutex
	messagesByID     map[string]*domain.ChatMessageWithSender
	messagesByThread map[string][]*domain.ChatMessageWithSender
	usersByID        map[string]*domain.User
	readRepository   *chatMemoryReadRepository
	threadRepository *chatMemoryThreadRepository
}

func (r *chatMemoryMessageRepository) Create(message *domain.ChatMessage) error {
	if message == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(message.ID) == "" {
		message.ID = uuid.NewString()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}

	user := r.usersByID[strings.TrimSpace(message.SenderUserID)]
	sender := domain.ChatSenderSummary{UserID: message.SenderUserID}
	if user != nil {
		sender.FullName = user.FullName
		sender.CompanyName = user.CompanyName
		sender.Role = user.Role
	}

	stored := &domain.ChatMessageWithSender{
		ID:           message.ID,
		ThreadID:     message.ThreadID,
		SenderUserID: message.SenderUserID,
		Body:         message.Body,
		CreatedAt:    message.CreatedAt.UTC(),
		Sender:       sender,
	}

	r.messagesByID[stored.ID] = stored
	r.messagesByThread[stored.ThreadID] = append(r.messagesByThread[stored.ThreadID], stored)

	sort.SliceStable(r.messagesByThread[stored.ThreadID], func(i, j int) bool {
		left := r.messagesByThread[stored.ThreadID][i]
		right := r.messagesByThread[stored.ThreadID][j]
		if left.CreatedAt.Equal(right.CreatedAt) {
			return left.ID > right.ID
		}
		return left.CreatedAt.After(right.CreatedAt)
	})

	return nil
}

func (r *chatMemoryMessageRepository) GetByIDWithSender(id string) (*domain.ChatMessageWithSender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	message, ok := r.messagesByID[strings.TrimSpace(id)]
	if !ok || message == nil {
		return nil, repository.ErrChatMessageNotFound
	}
	copied := *message
	return &copied, nil
}

func (r *chatMemoryMessageRepository) GetLatestByThreadID(threadID string) (*domain.ChatMessageWithSender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	messages := r.messagesByThread[strings.TrimSpace(threadID)]
	if len(messages) == 0 {
		return nil, repository.ErrChatMessageNotFound
	}
	copied := *messages[0]
	return &copied, nil
}

func (r *chatMemoryMessageRepository) GetLatestByThreadIDs(threadIDs []string) (map[string]*domain.ChatMessageWithSender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]*domain.ChatMessageWithSender, len(threadIDs))
	for _, threadID := range threadIDs {
		messages := r.messagesByThread[strings.TrimSpace(threadID)]
		if len(messages) == 0 {
			continue
		}
		copied := *messages[0]
		result[strings.TrimSpace(threadID)] = &copied
	}
	return result, nil
}

func (r *chatMemoryMessageRepository) ListByThreadID(threadID string, cursor *domain.ChatMessageCursor, limit int) ([]*domain.ChatMessageWithSender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages := r.messagesByThread[strings.TrimSpace(threadID)]
	if limit <= 0 {
		limit = 30
	}

	filtered := make([]*domain.ChatMessageWithSender, 0, len(messages))
	for _, message := range messages {
		if message == nil {
			continue
		}
		if cursor != nil {
			if message.CreatedAt.After(cursor.CreatedAt) {
				continue
			}
			if message.CreatedAt.Equal(cursor.CreatedAt) && message.ID >= cursor.ID {
				continue
			}
		}
		copied := *message
		filtered = append(filtered, &copied)
		if len(filtered) == limit {
			break
		}
	}

	return filtered, nil
}

func (r *chatMemoryMessageRepository) CountUnreadByThreadAndUser(threadID, userID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.countUnreadForThread(strings.TrimSpace(threadID), strings.TrimSpace(userID)), nil
}

func (r *chatMemoryMessageRepository) CountUnreadByThreadIDsAndUser(threadIDs []string, userID string) (map[string]int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]int64, len(threadIDs))
	for _, threadID := range threadIDs {
		trimmed := strings.TrimSpace(threadID)
		result[trimmed] = r.countUnreadForThread(trimmed, strings.TrimSpace(userID))
	}
	return result, nil
}

func (r *chatMemoryMessageRepository) CountUnreadSummaryByPairingAndUser(pairingID, userID string) (int64, int64, error) {
	r.threadRepository.mu.RLock()
	threads := make([]*domain.ChatThread, 0)
	for _, thread := range r.threadRepository.threadsByID {
		if thread != nil && strings.TrimSpace(thread.PairingID) == strings.TrimSpace(pairingID) {
			copied := *thread
			threads = append(threads, &copied)
		}
	}
	r.threadRepository.mu.RUnlock()

	unreadThreads := int64(0)
	unreadMessages := int64(0)
	for _, thread := range threads {
		count := r.countUnreadForThread(thread.ID, strings.TrimSpace(userID))
		if count > 0 {
			unreadThreads++
		}
		unreadMessages += count
	}

	return unreadThreads, unreadMessages, nil
}

func (r *chatMemoryMessageRepository) countUnreadForThread(threadID, userID string) int64 {
	messages := r.messagesByThread[threadID]
	if len(messages) == 0 {
		return 0
	}

	state, _ := r.readRepository.GetByThreadAndUser(threadID, userID)
	var boundary *domain.ChatMessageWithSender
	if state != nil && state.LastReadMessageID != nil {
		boundary = r.messagesByID[strings.TrimSpace(*state.LastReadMessageID)]
	}

	var count int64
	for _, message := range messages {
		if message == nil {
			continue
		}
		if strings.TrimSpace(message.SenderUserID) == userID {
			continue
		}
		if boundary == nil {
			count++
			continue
		}
		if message.CreatedAt.After(boundary.CreatedAt) {
			count++
			continue
		}
		if message.CreatedAt.Equal(boundary.CreatedAt) && message.ID > boundary.ID {
			count++
		}
	}

	return count
}

type chatFixture struct {
	service             *ChatService
	threadRepo          *chatMemoryThreadRepository
	messageRepo         *chatMemoryMessageRepository
	ethUserID           string
	foreignUserID       string
	outsiderUserID      string
	activePairingID     string
	inactivePairingID   string
	sharedCandidateID   string
	unsharedCandidateID string
}

func setupChatFixture(t *testing.T) *chatFixture {
	t.Helper()

	ethUserID := "user-eth"
	foreignUserID := "user-for"
	outsiderUserID := "user-out"
	activePairingID := "pair-active"
	inactivePairingID := "pair-inactive"
	sharedCandidateID := "cand-shared"
	unsharedCandidateID := "cand-unshared"

	userRepo := &chatMemoryUserRepository{users: map[string]*domain.User{
		ethUserID: {
			ID:            ethUserID,
			Email:         "eth@example.com",
			Role:          domain.EthiopianAgent,
			AccountStatus: domain.AccountStatusActive,
			IsActive:      true,
			FullName:      "Ethiopian Agent",
			CompanyName:   "Eth Agency",
		},
		foreignUserID: {
			ID:            foreignUserID,
			Email:         "for@example.com",
			Role:          domain.ForeignAgent,
			AccountStatus: domain.AccountStatusActive,
			IsActive:      true,
			FullName:      "Foreign Agent",
			CompanyName:   "Foreign Agency",
		},
		outsiderUserID: {
			ID:            outsiderUserID,
			Email:         "out@example.com",
			Role:          domain.ForeignAgent,
			AccountStatus: domain.AccountStatusActive,
			IsActive:      true,
			FullName:      "Outsider",
			CompanyName:   "Outsider Agency",
		},
	}}

	pairingRepo := &chatMemoryPairingRepository{pairings: map[string]*domain.AgencyPairing{
		activePairingID: {
			ID:              activePairingID,
			EthiopianUserID: ethUserID,
			ForeignUserID:   foreignUserID,
			Status:          domain.AgencyPairingActive,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		},
		inactivePairingID: {
			ID:              inactivePairingID,
			EthiopianUserID: ethUserID,
			ForeignUserID:   foreignUserID,
			Status:          domain.AgencyPairingEnded,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		},
	}}

	shareRepo := &chatMemoryShareRepository{shares: map[string]*domain.CandidatePairShare{
		chatPairCandidateKey(activePairingID, sharedCandidateID): {
			ID:          "share-1",
			PairingID:   activePairingID,
			CandidateID: sharedCandidateID,
			IsActive:    true,
			SharedAt:    time.Now().UTC(),
		},
	}}

	selectionRepo := &chatMemorySelectionRepository{
		selectionsByID:   map[string]*domain.Selection{},
		selectionsByPair: map[string]*domain.Selection{},
	}

	candidateRepo := &chatMemoryCandidateRepository{candidates: map[string]*domain.Candidate{
		sharedCandidateID: {
			ID:        sharedCandidateID,
			CreatedBy: ethUserID,
			FullName:  "Shared Candidate",
			Status:    domain.CandidateStatusAvailable,
		},
		unsharedCandidateID: {
			ID:        unsharedCandidateID,
			CreatedBy: ethUserID,
			FullName:  "Unshared Candidate",
			Status:    domain.CandidateStatusAvailable,
		},
	}}

	pairingService := &PairingService{
		userRepository:      userRepo,
		pairingRepository:   pairingRepo,
		shareRepository:     shareRepo,
		selectionRepository: selectionRepo,
		auditRepository:     &chatNoopAuditRepository{},
	}

	threadRepo := &chatMemoryThreadRepository{
		threadsByID:        map[string]*domain.ChatThread{},
		workspaceByPairing: map[string]string{},
		candidateByPairing: map[string]string{},
	}
	readRepo := &chatMemoryReadRepository{reads: map[string]*domain.ChatThreadRead{}}
	messageRepo := &chatMemoryMessageRepository{
		messagesByID:     map[string]*domain.ChatMessageWithSender{},
		messagesByThread: map[string][]*domain.ChatMessageWithSender{},
		usersByID:        userRepo.users,
		readRepository:   readRepo,
		threadRepository: threadRepo,
	}

	chatService, err := NewChatService(threadRepo, messageRepo, readRepo, candidateRepo, selectionRepo, pairingService)
	require.NoError(t, err)

	return &chatFixture{
		service:             chatService,
		threadRepo:          threadRepo,
		messageRepo:         messageRepo,
		ethUserID:           ethUserID,
		foreignUserID:       foreignUserID,
		outsiderUserID:      outsiderUserID,
		activePairingID:     activePairingID,
		inactivePairingID:   inactivePairingID,
		sharedCandidateID:   sharedCandidateID,
		unsharedCandidateID: unsharedCandidateID,
	}
}

func TestChatService_ResolveWorkspaceThread_Deduplicates(t *testing.T) {
	fixture := setupChatFixture(t)

	first, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, second)

	assert.Equal(t, first.ID, second.ID)
	assert.Len(t, fixture.threadRepo.workspaceByPairing, 1)
}

func TestChatService_ResolveCandidateThread_Deduplicates(t *testing.T) {
	fixture := setupChatFixture(t)

	first, err := fixture.service.ResolveCandidateThread(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID, fixture.sharedCandidateID)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := fixture.service.ResolveCandidateThread(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID, fixture.sharedCandidateID)
	require.NoError(t, err)
	require.NotNil(t, second)

	assert.Equal(t, first.ID, second.ID)
	assert.Len(t, fixture.threadRepo.candidateByPairing, 1)
}

func TestChatService_ForeignCannotAccessUnsharedCandidateThread(t *testing.T) {
	fixture := setupChatFixture(t)

	thread, err := fixture.service.ResolveCandidateThread(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID, fixture.unsharedCandidateID)
	require.Error(t, err)
	assert.Nil(t, thread)
	assert.ErrorIs(t, err, ErrChatThreadForbidden)
}

func TestChatService_OnlyPairParticipantsCanReadWrite(t *testing.T) {
	fixture := setupChatFixture(t)

	thread, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, thread)

	message, err := fixture.service.SendMessage(fixture.outsiderUserID, string(domain.ForeignAgent), fixture.activePairingID, thread.ID, "hello")
	require.Error(t, err)
	assert.Nil(t, message)
	assert.True(t, errors.Is(err, ErrNoActivePairings) || errors.Is(err, ErrPairingAccessDenied))

	messages, nextCursor, err := fixture.service.ListMessages(fixture.outsiderUserID, string(domain.ForeignAgent), fixture.activePairingID, thread.ID, "", 30)
	require.Error(t, err)
	assert.Nil(t, messages)
	assert.Nil(t, nextCursor)
	assert.True(t, errors.Is(err, ErrNoActivePairings) || errors.Is(err, ErrPairingAccessDenied))
}

func TestChatService_SendMessageUpdatesThreadPreviewAndTimestamp(t *testing.T) {
	fixture := setupChatFixture(t)

	thread, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, thread)

	message, err := fixture.service.SendMessage(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID, thread.ID, "  Hello partner  ")
	require.NoError(t, err)
	require.NotNil(t, message)
	assert.Equal(t, "Hello partner", message.Body)

	storedThread, err := fixture.threadRepo.GetByID(thread.ID)
	require.NoError(t, err)
	require.NotNil(t, storedThread.LastMessageAt)
	require.NotNil(t, storedThread.LastMessagePreview)
	assert.Equal(t, "Hello partner", *storedThread.LastMessagePreview)
	assert.False(t, storedThread.UpdatedAt.IsZero())
}

func TestChatService_MarkReadUpdatesUnreadSummary(t *testing.T) {
	fixture := setupChatFixture(t)

	thread, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, thread)

	_, err = fixture.service.SendMessage(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID, thread.ID, "Need confirmation")
	require.NoError(t, err)

	summaryBefore, err := fixture.service.GetSummary(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, summaryBefore)
	assert.EqualValues(t, 1, summaryBefore.UnreadThreads)
	assert.EqualValues(t, 1, summaryBefore.UnreadMessages)

	err = fixture.service.MarkThreadRead(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID, thread.ID)
	require.NoError(t, err)

	summaryAfter, err := fixture.service.GetSummary(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, summaryAfter)
	assert.EqualValues(t, 0, summaryAfter.UnreadThreads)
	assert.EqualValues(t, 0, summaryAfter.UnreadMessages)

	threadForForeign, err := fixture.service.GetThread(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID, thread.ID)
	require.NoError(t, err)
	require.NotNil(t, threadForForeign)
	assert.EqualValues(t, 0, threadForForeign.UnreadCount)
	require.NotNil(t, threadForForeign.ReadState)
	require.NotNil(t, threadForForeign.ReadState.LastReadAt)
}

func TestChatService_InactivePairingBlocksSending(t *testing.T) {
	fixture := setupChatFixture(t)

	thread, err := fixture.threadRepo.ResolveOrCreateWorkspaceThread(fixture.inactivePairingID, fixture.ethUserID)
	require.NoError(t, err)
	require.NotNil(t, thread)

	message, err := fixture.service.SendMessage(fixture.ethUserID, string(domain.EthiopianAgent), fixture.inactivePairingID, thread.ID, "blocked")
	require.Error(t, err)
	assert.Nil(t, message)
	assert.ErrorIs(t, err, ErrPairingAccessDenied)
}

func TestChatService_SendMessageRejectsInvalidBody(t *testing.T) {
	fixture := setupChatFixture(t)

	thread, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, thread)

	message, err := fixture.service.SendMessage(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID, thread.ID, "   ")
	require.Error(t, err)
	assert.Nil(t, message)
	assert.ErrorIs(t, err, ErrChatMessageEmpty)

	message, err = fixture.service.SendMessage(
		fixture.ethUserID,
		string(domain.EthiopianAgent),
		fixture.activePairingID,
		thread.ID,
		strings.Repeat("a", maxChatBodyChars+1),
	)
	require.Error(t, err)
	assert.Nil(t, message)
	assert.ErrorIs(t, err, ErrChatMessageTooLong)
}

func TestChatService_GetSummaryExcludesInaccessibleCandidateThreads(t *testing.T) {
	fixture := setupChatFixture(t)

	workspaceThread, err := fixture.service.ResolveWorkspaceThread(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, workspaceThread)

	_, err = fixture.service.SendMessage(fixture.ethUserID, string(domain.EthiopianAgent), fixture.activePairingID, workspaceThread.ID, "Visible workspace message")
	require.NoError(t, err)

	hiddenThread, err := fixture.threadRepo.ResolveOrCreateCandidateThread(fixture.activePairingID, fixture.unsharedCandidateID, fixture.ethUserID)
	require.NoError(t, err)
	require.NotNil(t, hiddenThread)

	err = fixture.messageRepo.Create(&domain.ChatMessage{
		ThreadID:     hiddenThread.ID,
		SenderUserID: fixture.ethUserID,
		Body:         "Hidden candidate message",
		CreatedAt:    time.Now().UTC(),
	})
	require.NoError(t, err)
	err = fixture.threadRepo.UpdateLastMessage(hiddenThread.ID, time.Now().UTC(), "Hidden candidate message")
	require.NoError(t, err)

	threads, err := fixture.service.ListThreads(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.Len(t, threads, 1)
	assert.Equal(t, workspaceThread.ID, threads[0].ID)

	summary, err := fixture.service.GetSummary(fixture.foreignUserID, string(domain.ForeignAgent), fixture.activePairingID)
	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.EqualValues(t, 1, summary.UnreadThreads)
	assert.EqualValues(t, 1, summary.UnreadMessages)
}

func chatPairCandidateKey(pairingID, candidateID string) string {
	return strings.TrimSpace(pairingID) + "::" + strings.TrimSpace(candidateID)
}

func chatThreadUserKey(threadID, userID string) string {
	return strings.TrimSpace(threadID) + "::" + strings.TrimSpace(userID)
}

func chatOptionalTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func chatOptionalTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
