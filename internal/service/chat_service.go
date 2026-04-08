package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

const (
	defaultChatMessagePageSize = 30
	maxChatMessagePageSize     = 100
	maxChatBodyChars           = 4000
	chatPreviewMaxChars        = 180
)

var (
	ErrChatThreadNotFound   = errors.New("chat thread not found")
	ErrChatInvalidCursor    = errors.New("invalid chat cursor")
	ErrChatMessageEmpty     = errors.New("chat message body is required")
	ErrChatMessageTooLong   = errors.New("chat message body is too long")
	ErrChatThreadForbidden  = errors.New("chat thread access denied")
	ErrChatUnsupportedScope = errors.New("unsupported chat thread scope")
)

type ChatSenderView struct {
	UserID      string
	FullName    string
	CompanyName string
	Role        string
}

type MessageView struct {
	ID        string
	ThreadID  string
	Body      string
	Sender    ChatSenderView
	CreatedAt time.Time
}

type ThreadReadState struct {
	ThreadID          string
	UserID            string
	LastReadMessageID *string
	LastReadAt        *time.Time
	UpdatedAt         time.Time
}

type ThreadSummary struct {
	ID                 string
	PairingID          string
	ScopeType          string
	CandidateID        *string
	CreatedByUserID    string
	LastMessageAt      *time.Time
	LastMessagePreview *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	UnreadCount        int64
	LastMessage        *MessageView
	ReadState          *ThreadReadState
}

type ThreadView struct {
	ID                 string
	PairingID          string
	ScopeType          string
	CandidateID        *string
	CreatedByUserID    string
	LastMessageAt      *time.Time
	LastMessagePreview *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	UnreadCount        int64
	LastMessage        *MessageView
	ReadState          *ThreadReadState
}

type NextCursor struct {
	Cursor string
}

type ChatSummary struct {
	UnreadThreads  int64
	UnreadMessages int64
}

type ChatMessageCreatedEvent struct {
	ThreadID string
	Message  *MessageView
}

type ChatThreadReadEvent struct {
	ThreadID          string
	UserID            string
	LastReadMessageID *string
	LastReadAt        *time.Time
}

type ChatThreadSummaryUpdatedEvent struct {
	ThreadID           string
	LastMessageAt      *time.Time
	LastMessagePreview *string
	UpdatedAt          time.Time
}

type RealtimeChatNotifier interface {
	PushMessageCreated(pairingID string, recipientUserIDs []string, event *ChatMessageCreatedEvent) error
	PushThreadRead(pairingID string, recipientUserIDs []string, event *ChatThreadReadEvent) error
	PushThreadSummaryUpdated(pairingID string, recipientUserIDs []string, event *ChatThreadSummaryUpdatedEvent) error
}

type ChatService struct {
	threadRepository    domain.ChatThreadRepository
	messageRepository   domain.ChatMessageRepository
	readRepository      domain.ChatThreadReadRepository
	candidateRepository domain.CandidateRepository
	selectionRepository domain.SelectionRepository
	pairingService      *PairingService
	realtimeNotifier    RealtimeChatNotifier
}

func NewChatService(
	threadRepository domain.ChatThreadRepository,
	messageRepository domain.ChatMessageRepository,
	readRepository domain.ChatThreadReadRepository,
	candidateRepository domain.CandidateRepository,
	selectionRepository domain.SelectionRepository,
	pairingService *PairingService,
) (*ChatService, error) {
	if threadRepository == nil {
		return nil, fmt.Errorf("chat thread repository is nil")
	}
	if messageRepository == nil {
		return nil, fmt.Errorf("chat message repository is nil")
	}
	if readRepository == nil {
		return nil, fmt.Errorf("chat read repository is nil")
	}
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if pairingService == nil {
		return nil, fmt.Errorf("pairing service is nil")
	}

	return &ChatService{
		threadRepository:    threadRepository,
		messageRepository:   messageRepository,
		readRepository:      readRepository,
		candidateRepository: candidateRepository,
		selectionRepository: selectionRepository,
		pairingService:      pairingService,
	}, nil
}

func (s *ChatService) SetRealtimeNotifier(realtimeNotifier RealtimeChatNotifier) {
	s.realtimeNotifier = realtimeNotifier
}

func (s *ChatService) ResolveWorkspaceThread(userID, role, pairingID string) (*ThreadView, error) {
	pairing, err := s.resolveActivePairing(userID, role, pairingID)
	if err != nil {
		return nil, err
	}

	thread, err := s.threadRepository.ResolveOrCreateWorkspaceThread(pairing.ID, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	return s.buildThreadViewForUser(thread, strings.TrimSpace(userID))
}

func (s *ChatService) ResolveCandidateThread(userID, role, pairingID, candidateID string) (*ThreadView, error) {
	pairing, err := s.resolveActivePairing(userID, role, pairingID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureCandidateAccessibleInPairing(strings.TrimSpace(userID), strings.TrimSpace(role), pairing.ID, strings.TrimSpace(candidateID)); err != nil {
		return nil, err
	}

	thread, err := s.threadRepository.ResolveOrCreateCandidateThread(pairing.ID, strings.TrimSpace(candidateID), strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	return s.buildThreadViewForUser(thread, strings.TrimSpace(userID))
}

func (s *ChatService) GetThread(userID, role, pairingID, threadID string) (*ThreadSummary, error) {
	_, thread, err := s.resolveAccessibleThread(userID, role, pairingID, threadID)
	if err != nil {
		return nil, err
	}
	return s.buildThreadSummaryForUser(thread, strings.TrimSpace(userID))
}

func (s *ChatService) ListThreads(userID, role, pairingID string) ([]*ThreadSummary, error) {
	pairing, err := s.resolveActivePairing(userID, role, pairingID)
	if err != nil {
		return nil, err
	}

	filtered, err := s.listAccessibleThreads(strings.TrimSpace(userID), strings.TrimSpace(role), pairing)
	if err != nil {
		return nil, err
	}
	if len(filtered) == 0 {
		return []*ThreadSummary{}, nil
	}

	threadIDs := make([]string, 0, len(filtered))
	for _, thread := range filtered {
		threadIDs = append(threadIDs, thread.ID)
	}

	latestMessagesByThreadID, err := s.messageRepository.GetLatestByThreadIDs(threadIDs)
	if err != nil {
		return nil, err
	}
	unreadByThreadID, err := s.messageRepository.CountUnreadByThreadIDsAndUser(threadIDs, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	readStatesByThreadID, err := s.readRepository.ListByThreadIDsAndUser(threadIDs, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	result := make([]*ThreadSummary, 0, len(filtered))
	for _, thread := range filtered {
		if thread == nil {
			continue
		}

		summary := mapThreadSummary(thread)
		summary.UnreadCount = unreadByThreadID[thread.ID]

		if latestMessage, ok := latestMessagesByThreadID[thread.ID]; ok && latestMessage != nil {
			messageView := mapMessageView(latestMessage)
			summary.LastMessage = &messageView
		}
		if readState, ok := readStatesByThreadID[thread.ID]; ok && readState != nil {
			mappedReadState := mapThreadReadState(readState)
			summary.ReadState = &mappedReadState
		}

		result = append(result, &summary)
	}

	return result, nil
}

func (s *ChatService) ListMessages(userID, role, pairingID, threadID, cursor string, limit int) ([]*MessageView, *NextCursor, error) {
	_, thread, err := s.resolveAccessibleThread(userID, role, pairingID, threadID)
	if err != nil {
		return nil, nil, err
	}

	parsedLimit := normalizeMessageLimit(limit)
	parsedCursor, err := decodeChatCursor(cursor)
	if err != nil {
		return nil, nil, ErrChatInvalidCursor
	}

	messages, err := s.messageRepository.ListByThreadID(thread.ID, parsedCursor, parsedLimit)
	if err != nil {
		return nil, nil, err
	}

	result := make([]*MessageView, 0, len(messages))
	for _, message := range messages {
		if message == nil {
			continue
		}
		mapped := mapMessageView(message)
		result = append(result, &mapped)
	}

	var next *NextCursor
	if len(messages) == parsedLimit {
		last := messages[len(messages)-1]
		next = &NextCursor{Cursor: encodeChatCursor(last.CreatedAt, last.ID)}
	}

	return result, next, nil
}

func (s *ChatService) SendMessage(userID, role, pairingID, threadID, body string) (*MessageView, error) {
	pairing, thread, err := s.resolveAccessibleThread(userID, role, pairingID, threadID)
	if err != nil {
		return nil, err
	}

	normalizedBody, err := validateChatBody(body)
	if err != nil {
		return nil, err
	}

	message := &domain.ChatMessage{
		ThreadID:     thread.ID,
		SenderUserID: strings.TrimSpace(userID),
		Body:         normalizedBody,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.messageRepository.Create(message); err != nil {
		return nil, err
	}

	createdMessage, err := s.messageRepository.GetByIDWithSender(message.ID)
	if err != nil {
		return nil, err
	}

	if err := s.threadRepository.UpdateLastMessage(thread.ID, createdMessage.CreatedAt, buildMessagePreview(createdMessage.Body)); err != nil {
		return nil, err
	}

	messageView := mapMessageView(createdMessage)
	recipients := pairingParticipants(pairing)
	updatedAt := time.Now().UTC()
	preview := buildMessagePreview(createdMessage.Body)

	s.pushMessageCreated(pairing.ID, recipients, &ChatMessageCreatedEvent{
		ThreadID: thread.ID,
		Message:  &messageView,
	})
	s.pushThreadSummaryUpdated(pairing.ID, recipients, &ChatThreadSummaryUpdatedEvent{
		ThreadID:           thread.ID,
		LastMessageAt:      &createdMessage.CreatedAt,
		LastMessagePreview: &preview,
		UpdatedAt:          updatedAt,
	})

	return &messageView, nil
}

func (s *ChatService) MarkThreadRead(userID, role, pairingID, threadID string) error {
	pairing, thread, err := s.resolveAccessibleThread(userID, role, pairingID, threadID)
	if err != nil {
		return err
	}

	latestMessage, err := s.messageRepository.GetLatestByThreadID(thread.ID)
	if err != nil && !errors.Is(err, repository.ErrChatMessageNotFound) {
		return err
	}

	now := time.Now().UTC()
	var lastReadMessageID *string
	if latestMessage != nil {
		id := strings.TrimSpace(latestMessage.ID)
		if id != "" {
			lastReadMessageID = &id
		}
	}

	readState, err := s.readRepository.Upsert(thread.ID, strings.TrimSpace(userID), lastReadMessageID, &now)
	if err != nil {
		return err
	}

	recipients := pairingParticipants(pairing)
	s.pushThreadRead(pairing.ID, recipients, &ChatThreadReadEvent{
		ThreadID:          thread.ID,
		UserID:            strings.TrimSpace(userID),
		LastReadMessageID: readState.LastReadMessageID,
		LastReadAt:        readState.LastReadAt,
	})
	s.pushThreadSummaryUpdated(pairing.ID, recipients, &ChatThreadSummaryUpdatedEvent{
		ThreadID:  thread.ID,
		UpdatedAt: time.Now().UTC(),
	})

	return nil
}

func (s *ChatService) GetSummary(userID, role, pairingID string) (*ChatSummary, error) {
	pairing, err := s.resolveActivePairing(userID, role, pairingID)
	if err != nil {
		return nil, err
	}

	threads, err := s.listAccessibleThreads(strings.TrimSpace(userID), strings.TrimSpace(role), pairing)
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return &ChatSummary{}, nil
	}

	threadIDs := make([]string, 0, len(threads))
	for _, thread := range threads {
		if thread == nil {
			continue
		}
		threadIDs = append(threadIDs, thread.ID)
	}
	if len(threadIDs) == 0 {
		return &ChatSummary{}, nil
	}

	unreadByThreadID, err := s.messageRepository.CountUnreadByThreadIDsAndUser(threadIDs, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}

	unreadThreads := int64(0)
	unreadMessages := int64(0)
	for _, threadID := range threadIDs {
		count := unreadByThreadID[threadID]
		if count > 0 {
			unreadThreads++
		}
		unreadMessages += count
	}

	return &ChatSummary{
		UnreadThreads:  unreadThreads,
		UnreadMessages: unreadMessages,
	}, nil
}

func (s *ChatService) listAccessibleThreads(userID, role string, pairing *domain.AgencyPairing) ([]*domain.ChatThread, error) {
	if pairing == nil || strings.TrimSpace(pairing.ID) == "" {
		return []*domain.ChatThread{}, nil
	}

	threads, err := s.threadRepository.ListByPairingID(pairing.ID)
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return []*domain.ChatThread{}, nil
	}

	filtered := make([]*domain.ChatThread, 0, len(threads))
	for _, thread := range threads {
		if thread == nil {
			continue
		}
		if thread.ScopeType == domain.ChatThreadScopeCandidate {
			if thread.CandidateID == nil || strings.TrimSpace(*thread.CandidateID) == "" {
				continue
			}
			if err := s.ensureCandidateAccessibleInPairing(userID, role, pairing.ID, strings.TrimSpace(*thread.CandidateID)); err != nil {
				if errors.Is(err, ErrChatThreadForbidden) || errors.Is(err, repository.ErrSelectionNotFound) || errors.Is(err, repository.ErrCandidateNotFound) {
					continue
				}
				return nil, err
			}
		}
		filtered = append(filtered, thread)
	}

	return filtered, nil
}

func (s *ChatService) resolveActivePairing(userID, role, pairingID string) (*domain.AgencyPairing, error) {
	trimmedUserID := strings.TrimSpace(userID)
	trimmedRole := strings.TrimSpace(role)

	if trimmedUserID == "" {
		return nil, ErrForbidden
	}
	if trimmedRole != string(domain.EthiopianAgent) && trimmedRole != string(domain.ForeignAgent) {
		return nil, ErrForbidden
	}

	return s.pairingService.ResolveActivePairing(trimmedUserID, trimmedRole, strings.TrimSpace(pairingID))
}

func (s *ChatService) resolveAccessibleThread(userID, role, pairingID, threadID string) (*domain.AgencyPairing, *domain.ChatThread, error) {
	pairing, err := s.resolveActivePairing(userID, role, pairingID)
	if err != nil {
		return nil, nil, err
	}

	thread, err := s.threadRepository.GetByID(strings.TrimSpace(threadID))
	if err != nil {
		if errors.Is(err, repository.ErrChatThreadNotFound) {
			return nil, nil, ErrChatThreadNotFound
		}
		return nil, nil, err
	}
	if strings.TrimSpace(thread.PairingID) != strings.TrimSpace(pairing.ID) {
		return nil, nil, ErrChatThreadForbidden
	}

	scope := strings.TrimSpace(string(thread.ScopeType))
	switch scope {
	case string(domain.ChatThreadScopeWorkspace):
		return pairing, thread, nil
	case string(domain.ChatThreadScopeCandidate):
		if thread.CandidateID == nil || strings.TrimSpace(*thread.CandidateID) == "" {
			return nil, nil, ErrChatUnsupportedScope
		}
		if err := s.ensureCandidateAccessibleInPairing(strings.TrimSpace(userID), strings.TrimSpace(role), pairing.ID, strings.TrimSpace(*thread.CandidateID)); err != nil {
			return nil, nil, err
		}
		return pairing, thread, nil
	default:
		return nil, nil, ErrChatUnsupportedScope
	}
}

func (s *ChatService) ensureCandidateAccessibleInPairing(userID, role, pairingID, candidateID string) error {
	candidate, err := s.candidateRepository.GetByID(strings.TrimSpace(candidateID))
	if err != nil {
		return err
	}

	canAccess, err := s.pairingService.CanUserAccessCandidate(candidate, strings.TrimSpace(userID), strings.TrimSpace(role), strings.TrimSpace(pairingID))
	if err != nil {
		return err
	}
	if !canAccess {
		return ErrChatThreadForbidden
	}

	shared, err := s.pairingService.IsCandidateSharedWithPairing(candidate.ID, strings.TrimSpace(pairingID))
	if err != nil {
		return err
	}
	if shared {
		return nil
	}

	selection, err := s.selectionRepository.GetByCandidateIDAndPairingID(candidate.ID, strings.TrimSpace(pairingID))
	if err != nil {
		if errors.Is(err, repository.ErrSelectionNotFound) {
			return ErrChatThreadForbidden
		}
		return err
	}
	if selection == nil {
		return ErrChatThreadForbidden
	}

	return nil
}

func (s *ChatService) buildThreadViewForUser(thread *domain.ChatThread, userID string) (*ThreadView, error) {
	summary, err := s.buildThreadSummaryForUser(thread, userID)
	if err != nil {
		return nil, err
	}

	if summary == nil {
		return nil, nil
	}

	view := &ThreadView{
		ID:                 summary.ID,
		PairingID:          summary.PairingID,
		ScopeType:          summary.ScopeType,
		CandidateID:        summary.CandidateID,
		CreatedByUserID:    summary.CreatedByUserID,
		LastMessageAt:      summary.LastMessageAt,
		LastMessagePreview: summary.LastMessagePreview,
		CreatedAt:          summary.CreatedAt,
		UpdatedAt:          summary.UpdatedAt,
		UnreadCount:        summary.UnreadCount,
		LastMessage:        summary.LastMessage,
		ReadState:          summary.ReadState,
	}

	return view, nil
}

func (s *ChatService) buildThreadSummaryForUser(thread *domain.ChatThread, userID string) (*ThreadSummary, error) {
	if thread == nil {
		return nil, nil
	}

	summary := mapThreadSummary(thread)

	latestMessage, err := s.messageRepository.GetLatestByThreadID(thread.ID)
	if err != nil && !errors.Is(err, repository.ErrChatMessageNotFound) {
		return nil, err
	}
	if latestMessage != nil {
		mapped := mapMessageView(latestMessage)
		summary.LastMessage = &mapped
	}

	unreadCount, err := s.messageRepository.CountUnreadByThreadAndUser(thread.ID, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	summary.UnreadCount = unreadCount

	readState, err := s.readRepository.GetByThreadAndUser(thread.ID, strings.TrimSpace(userID))
	if err != nil && !errors.Is(err, repository.ErrChatThreadReadNotFound) {
		return nil, err
	}
	if readState != nil {
		mapped := mapThreadReadState(readState)
		summary.ReadState = &mapped
	}

	return &summary, nil
}

func validateChatBody(body string) (string, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return "", ErrChatMessageEmpty
	}
	if strings.ContainsRune(trimmed, '\x00') {
		return "", ErrChatMessageEmpty
	}
	if utf8.RuneCountInString(trimmed) > maxChatBodyChars {
		return "", ErrChatMessageTooLong
	}
	return trimmed, nil
}

func normalizeMessageLimit(limit int) int {
	if limit <= 0 {
		return defaultChatMessagePageSize
	}
	if limit > maxChatMessagePageSize {
		return maxChatMessagePageSize
	}
	return limit
}

func mapThreadSummary(thread *domain.ChatThread) ThreadSummary {
	return ThreadSummary{
		ID:                 thread.ID,
		PairingID:          thread.PairingID,
		ScopeType:          string(thread.ScopeType),
		CandidateID:        thread.CandidateID,
		CreatedByUserID:    thread.CreatedByUserID,
		LastMessageAt:      thread.LastMessageAt,
		LastMessagePreview: thread.LastMessagePreview,
		CreatedAt:          thread.CreatedAt,
		UpdatedAt:          thread.UpdatedAt,
	}
}

func mapMessageView(message *domain.ChatMessageWithSender) MessageView {
	if message == nil {
		return MessageView{}
	}
	return MessageView{
		ID:        message.ID,
		ThreadID:  message.ThreadID,
		Body:      message.Body,
		CreatedAt: message.CreatedAt,
		Sender: ChatSenderView{
			UserID:      message.Sender.UserID,
			FullName:    message.Sender.FullName,
			CompanyName: message.Sender.CompanyName,
			Role:        string(message.Sender.Role),
		},
	}
}

func mapThreadReadState(readState *domain.ChatThreadRead) ThreadReadState {
	return ThreadReadState{
		ThreadID:          readState.ThreadID,
		UserID:            readState.UserID,
		LastReadMessageID: readState.LastReadMessageID,
		LastReadAt:        readState.LastReadAt,
		UpdatedAt:         readState.UpdatedAt,
	}
}

func buildMessagePreview(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	if utf8.RuneCountInString(trimmed) <= chatPreviewMaxChars {
		return trimmed
	}
	runes := []rune(trimmed)
	return string(runes[:chatPreviewMaxChars]) + "..."
}

func pairingParticipants(pairing *domain.AgencyPairing) []string {
	if pairing == nil {
		return []string{}
	}
	candidates := []string{strings.TrimSpace(pairing.EthiopianUserID), strings.TrimSpace(pairing.ForeignUserID)}
	result := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

func encodeChatCursor(createdAt time.Time, id string) string {
	payload := fmt.Sprintf("%d|%s", createdAt.UTC().UnixNano(), strings.TrimSpace(id))
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func decodeChatCursor(cursor string) (*domain.ChatMessageCursor, error) {
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return nil, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("cursor payload malformed")
	}

	nanos, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(parts[1])
	if id == "" {
		return nil, fmt.Errorf("cursor id is empty")
	}

	return &domain.ChatMessageCursor{
		CreatedAt: time.Unix(0, nanos).UTC(),
		ID:        id,
	}, nil
}

func (s *ChatService) pushMessageCreated(pairingID string, recipients []string, event *ChatMessageCreatedEvent) {
	if s.realtimeNotifier == nil || event == nil || len(recipients) == 0 {
		return
	}
	if err := s.realtimeNotifier.PushMessageCreated(strings.TrimSpace(pairingID), recipients, event); err != nil {
		log.Printf("chat realtime message push failed: %v", err)
	}
}

func (s *ChatService) pushThreadRead(pairingID string, recipients []string, event *ChatThreadReadEvent) {
	if s.realtimeNotifier == nil || event == nil || len(recipients) == 0 {
		return
	}
	if err := s.realtimeNotifier.PushThreadRead(strings.TrimSpace(pairingID), recipients, event); err != nil {
		log.Printf("chat realtime read push failed: %v", err)
	}
}

func (s *ChatService) pushThreadSummaryUpdated(pairingID string, recipients []string, event *ChatThreadSummaryUpdatedEvent) {
	if s.realtimeNotifier == nil || event == nil || len(recipients) == 0 {
		return
	}
	if err := s.realtimeNotifier.PushThreadSummaryUpdated(strings.TrimSpace(pairingID), recipients, event); err != nil {
		log.Printf("chat realtime summary push failed: %v", err)
	}
}
