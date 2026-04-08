package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgconn"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

const (
	chatEventConnected            = "connected"
	chatEventMessageCreated       = "message.created"
	chatEventThreadRead           = "thread.read"
	chatEventThreadSummaryUpdated = "thread.summary_updated"

	maxChatConnectionsPerUserPerPairing = 3
	chatWriteWait                       = 10 * time.Second
	chatPongWait                        = 60 * time.Second
	chatPingPeriod                      = 45 * time.Second
)

type ChatSenderResponse struct {
	UserID      string `json:"user_id"`
	FullName    string `json:"full_name"`
	CompanyName string `json:"company_name"`
	Role        string `json:"role"`
}

type ChatMessageResponse struct {
	ID        string             `json:"id"`
	ThreadID  string             `json:"thread_id"`
	Body      string             `json:"body"`
	Sender    ChatSenderResponse `json:"sender"`
	CreatedAt string             `json:"created_at"`
}

type ChatThreadReadStateResponse struct {
	ThreadID          string  `json:"thread_id"`
	UserID            string  `json:"user_id"`
	LastReadMessageID *string `json:"last_read_message_id,omitempty"`
	LastReadAt        *string `json:"last_read_at,omitempty"`
	UpdatedAt         string  `json:"updated_at"`
}

type ChatPartnerAgencyResponse struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	CompanyName string `json:"company_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role"`
}

type ChatThreadSummaryResponse struct {
	ID                 string                       `json:"id"`
	PairingID          string                       `json:"pairing_id"`
	ScopeType          string                       `json:"scope_type"`
	CandidateID        *string                      `json:"candidate_id,omitempty"`
	CandidateName      *string                      `json:"candidate_name,omitempty"`
	PartnerAgency      *ChatPartnerAgencyResponse   `json:"partner_agency,omitempty"`
	CreatedByUserID    string                       `json:"created_by_user_id"`
	LastMessageAt      *string                      `json:"last_message_at,omitempty"`
	LastMessagePreview *string                      `json:"last_message_preview,omitempty"`
	UnreadCount        int64                        `json:"unread_count"`
	LastMessage        *ChatMessageResponse         `json:"last_message,omitempty"`
	ReadState          *ChatThreadReadStateResponse `json:"read_state,omitempty"`
	CreatedAt          string                       `json:"created_at"`
	UpdatedAt          string                       `json:"updated_at"`
}

type resolveCandidateThreadRequest struct {
	CandidateID string `json:"candidate_id"`
}

type sendMessageRequest struct {
	Body string `json:"body"`
}

type chatResolveThreadResponse struct {
	Thread ChatThreadSummaryResponse `json:"thread"`
}

type chatListThreadsResponse struct {
	Threads          []ChatThreadSummaryResponse `json:"threads"`
	UnreadCountTotal int64                       `json:"unread_count_total"`
}

type chatListMessagesResponse struct {
	Messages   []ChatMessageResponse     `json:"messages"`
	NextCursor string                    `json:"next_cursor,omitempty"`
	Thread     ChatThreadSummaryResponse `json:"thread"`
}

type chatSendMessageResponse struct {
	Message ChatMessageResponse `json:"message"`
}

type chatMarkReadResponse struct {
	Message   string                       `json:"message"`
	ReadState *ChatThreadReadStateResponse `json:"read_state,omitempty"`
}

type chatSummaryResponse struct {
	UnreadThreads  int64 `json:"unread_threads"`
	UnreadMessages int64 `json:"unread_messages"`
}

type chatConnectedPayload struct {
	PairingID string              `json:"pairing_id"`
	UserID    string              `json:"user_id"`
	Summary   chatSummaryResponse `json:"summary"`
}

type chatMessageCreatedPayload struct {
	ThreadID string              `json:"thread_id"`
	Message  ChatMessageResponse `json:"message"`
}

type chatThreadReadPayload struct {
	ThreadID          string  `json:"thread_id"`
	UserID            string  `json:"user_id"`
	LastReadMessageID *string `json:"last_read_message_id,omitempty"`
	LastReadAt        *string `json:"last_read_at,omitempty"`
}

type chatThreadSummaryUpdatedPayload struct {
	ThreadID           string  `json:"thread_id"`
	LastMessageAt      *string `json:"last_message_at,omitempty"`
	LastMessagePreview *string `json:"last_message_preview,omitempty"`
	UpdatedAt          string  `json:"updated_at"`
}

type chatWSOutboundEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type chatWSConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type ChatHandler struct {
	chatService         *service.ChatService
	userRepository      domain.UserRepository
	candidateRepository domain.CandidateRepository
	pairingService      *service.PairingService
	upgrader            websocket.Upgrader
	connections         map[string]map[string]map[*chatWSConn]struct{}
	connectionsMu       sync.RWMutex
}

func NewChatHandler(chatService *service.ChatService, allowedOrigins []string) *ChatHandler {
	normalizedOrigins := normalizeAllowedOrigins(allowedOrigins)

	return &ChatHandler{
		chatService: chatService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := strings.TrimSpace(r.Header.Get("Origin"))
				if origin == "" {
					return false
				}
				return isAllowedOrigin(origin, normalizedOrigins)
			},
		},
		connections: make(map[string]map[string]map[*chatWSConn]struct{}),
	}
}

func (h *ChatHandler) SetContextRepositories(
	userRepository domain.UserRepository,
	candidateRepository domain.CandidateRepository,
	pairingService *service.PairingService,
) {
	h.userRepository = userRepository
	h.candidateRepository = candidateRepository
	h.pairingService = pairingService
}

func (h *ChatHandler) ResolveWorkspaceThread(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	if err := decodeOptionalJSONBody(w, r, 8<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	thread, err := h.chatService.ResolveWorkspaceThread(userID, role, pairingID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	response, err := h.mapThreadSummaryResponseFromViewForUser(userID, role, pairingID, thread)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, chatResolveThreadResponse{Thread: response})
}

func (h *ChatHandler) ResolveCandidateThread(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	var req resolveCandidateThreadRequest
	if err := decodeJSONBody(w, r, &req, 16<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.CandidateID) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate_id is required"})
		return
	}

	thread, err := h.chatService.ResolveCandidateThread(userID, role, pairingID, req.CandidateID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	response, err := h.mapThreadSummaryResponseFromViewForUser(userID, role, pairingID, thread)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, chatResolveThreadResponse{Thread: response})
}

func (h *ChatHandler) ListThreads(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	threads, err := h.chatService.ListThreads(userID, role, pairingID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}
	summary, err := h.chatService.GetSummary(userID, role, pairingID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	responses := make([]ChatThreadSummaryResponse, 0, len(threads))
	for _, thread := range threads {
		if thread == nil {
			continue
		}

		mapped, err := h.mapThreadSummaryResponseForUser(userID, role, pairingID, thread)
		if err != nil {
			h.writeChatError(w, err)
			return
		}
		responses = append(responses, mapped)
	}

	_ = utils.WriteJSON(w, http.StatusOK, chatListThreadsResponse{
		Threads:          responses,
		UnreadCountTotal: summary.UnreadMessages,
	})
}

func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	threadID := strings.TrimSpace(chi.URLParam(r, "id"))
	if threadID == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "thread id is required"})
		return
	}

	cursor := strings.TrimSpace(r.URL.Query().Get("cursor"))
	limit := parseOptionalPositiveInt(r.URL.Query().Get("limit"))

	messages, nextCursor, err := h.chatService.ListMessages(userID, role, pairingID, threadID, cursor, limit)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	thread, err := h.chatService.GetThread(userID, role, pairingID, threadID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	messageResponses := make([]ChatMessageResponse, 0, len(messages))
	for _, message := range messages {
		if message == nil {
			continue
		}
		messageResponses = append(messageResponses, mapMessageResponse(message))
	}

	response := chatListMessagesResponse{
		Messages: messageResponses,
	}
	response.Thread, err = h.mapThreadSummaryResponseForUser(userID, role, pairingID, thread)
	if err != nil {
		h.writeChatError(w, err)
		return
	}
	if nextCursor != nil {
		response.NextCursor = strings.TrimSpace(nextCursor.Cursor)
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	threadID := strings.TrimSpace(chi.URLParam(r, "id"))
	if threadID == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "thread id is required"})
		return
	}

	var req sendMessageRequest
	if err := decodeJSONBody(w, r, &req, 32<<10); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	message, err := h.chatService.SendMessage(userID, role, pairingID, threadID, req.Body)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, chatSendMessageResponse{Message: mapMessageResponse(message)})
}

func (h *ChatHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	threadID := strings.TrimSpace(chi.URLParam(r, "id"))
	if threadID == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "thread id is required"})
		return
	}

	if err := h.chatService.MarkThreadRead(userID, role, pairingID, threadID); err != nil {
		h.writeChatError(w, err)
		return
	}

	thread, err := h.chatService.GetThread(userID, role, pairingID, threadID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	response := chatMarkReadResponse{Message: "thread marked as read"}
	if thread != nil && thread.ReadState != nil {
		mapped := mapThreadReadStateResponse(thread.ReadState)
		response.ReadState = &mapped
	}

	_ = utils.WriteJSON(w, http.StatusOK, response)
}

func (h *ChatHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, role, pairingID, ok := requireChatContext(w, r)
	if !ok {
		return
	}

	summary, err := h.chatService.GetSummary(userID, role, pairingID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, chatSummaryResponse{
		UnreadThreads:  summary.UnreadThreads,
		UnreadMessages: summary.UnreadMessages,
	})
}

func (h *ChatHandler) ChatWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	pairingID, _ := middleware.PairingIDFromContext(r.Context())
	if strings.TrimSpace(pairingID) == "" {
		pairingID = strings.TrimSpace(r.URL.Query().Get("pairing_id"))
	}
	if strings.TrimSpace(pairingID) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "pairing_id is required"})
		return
	}

	summary, err := h.chatService.GetSummary(userID, role, pairingID)
	if err != nil {
		h.writeChatError(w, err)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &chatWSConn{conn: conn}
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(chatPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(chatPongWait))
		return nil
	})

	if !h.addConnection(strings.TrimSpace(pairingID), strings.TrimSpace(userID), client) {
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "too many websocket connections"), time.Now().Add(chatWriteWait))
		_ = conn.Close()
		return
	}
	defer func() {
		h.removeConnection(strings.TrimSpace(pairingID), strings.TrimSpace(userID), client)
		_ = conn.Close()
	}()

	h.writeClientEvent(client, chatWSOutboundEvent{
		Type: chatEventConnected,
		Payload: chatConnectedPayload{
			PairingID: strings.TrimSpace(pairingID),
			UserID:    strings.TrimSpace(userID),
			Summary: chatSummaryResponse{
				UnreadThreads:  summary.UnreadThreads,
				UnreadMessages: summary.UnreadMessages,
			},
		},
	})

	pingTicker := time.NewTicker(chatPingPeriod)
	defer pingTicker.Stop()
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-pingTicker.C:
				client.mu.Lock()
				err := client.conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(chatWriteWait))
				client.mu.Unlock()
				if err != nil {
					_ = conn.Close()
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *ChatHandler) PushMessageCreated(pairingID string, recipientUserIDs []string, event *service.ChatMessageCreatedEvent) error {
	if event == nil || event.Message == nil {
		return nil
	}

	payload := chatMessageCreatedPayload{
		ThreadID: strings.TrimSpace(event.ThreadID),
		Message:  mapMessageResponse(event.Message),
	}

	h.broadcastToRecipients(strings.TrimSpace(pairingID), recipientUserIDs, chatWSOutboundEvent{
		Type:    chatEventMessageCreated,
		Payload: payload,
	})

	return nil
}

func (h *ChatHandler) PushThreadRead(pairingID string, recipientUserIDs []string, event *service.ChatThreadReadEvent) error {
	if event == nil {
		return nil
	}

	payload := chatThreadReadPayload{
		ThreadID:          strings.TrimSpace(event.ThreadID),
		UserID:            strings.TrimSpace(event.UserID),
		LastReadMessageID: event.LastReadMessageID,
		LastReadAt:        formatTimePointer(event.LastReadAt),
	}

	h.broadcastToRecipients(strings.TrimSpace(pairingID), recipientUserIDs, chatWSOutboundEvent{
		Type:    chatEventThreadRead,
		Payload: payload,
	})

	return nil
}

func (h *ChatHandler) PushThreadSummaryUpdated(pairingID string, recipientUserIDs []string, event *service.ChatThreadSummaryUpdatedEvent) error {
	if event == nil {
		return nil
	}

	payload := chatThreadSummaryUpdatedPayload{
		ThreadID:           strings.TrimSpace(event.ThreadID),
		LastMessageAt:      formatTimePointer(event.LastMessageAt),
		LastMessagePreview: event.LastMessagePreview,
		UpdatedAt:          event.UpdatedAt.UTC().Format(time.RFC3339),
	}

	h.broadcastToRecipients(strings.TrimSpace(pairingID), recipientUserIDs, chatWSOutboundEvent{
		Type:    chatEventThreadSummaryUpdated,
		Payload: payload,
	})

	return nil
}

func (h *ChatHandler) broadcastToRecipients(pairingID string, recipientUserIDs []string, event chatWSOutboundEvent) {
	recipients := normalizeRecipients(recipientUserIDs)
	for _, userID := range recipients {
		clients := h.snapshotConnections(pairingID, userID)
		for _, client := range clients {
			if !h.writeClientEvent(client, event) {
				h.removeConnection(pairingID, userID, client)
				_ = client.conn.Close()
			}
		}
	}
}

func (h *ChatHandler) writeClientEvent(client *chatWSConn, event chatWSOutboundEvent) bool {
	if client == nil {
		return false
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	_ = client.conn.SetWriteDeadline(time.Now().Add(chatWriteWait))
	if err := client.conn.WriteJSON(event); err != nil {
		return false
	}
	return true
}

func (h *ChatHandler) addConnection(pairingID, userID string, client *chatWSConn) bool {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	if h.connections[pairingID] == nil {
		h.connections[pairingID] = make(map[string]map[*chatWSConn]struct{})
	}
	if h.connections[pairingID][userID] == nil {
		h.connections[pairingID][userID] = make(map[*chatWSConn]struct{})
	}
	if len(h.connections[pairingID][userID]) >= maxChatConnectionsPerUserPerPairing {
		return false
	}

	h.connections[pairingID][userID][client] = struct{}{}
	return true
}

func (h *ChatHandler) removeConnection(pairingID, userID string, client *chatWSConn) {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	pairingConnections := h.connections[pairingID]
	if pairingConnections == nil {
		return
	}
	userConnections := pairingConnections[userID]
	if userConnections == nil {
		return
	}

	delete(userConnections, client)
	if len(userConnections) == 0 {
		delete(pairingConnections, userID)
	}
	if len(pairingConnections) == 0 {
		delete(h.connections, pairingID)
	}
}

func (h *ChatHandler) snapshotConnections(pairingID, userID string) []*chatWSConn {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	pairingConnections := h.connections[pairingID]
	if pairingConnections == nil {
		return nil
	}
	userConnections := pairingConnections[userID]
	if userConnections == nil {
		return nil
	}

	result := make([]*chatWSConn, 0, len(userConnections))
	for client := range userConnections {
		result = append(result, client)
	}
	return result
}

func (h *ChatHandler) writeChatError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrPairingRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "pairing is required"})
	case errors.Is(err, service.ErrNoActivePairings):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "no active pairings"})
	case errors.Is(err, service.ErrPairingAccessDenied):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrPairingNotActive):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "pairing is not active"})
	case errors.Is(err, service.ErrChatThreadNotFound), errors.Is(err, repository.ErrChatThreadNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "chat thread not found"})
	case errors.Is(err, service.ErrChatThreadForbidden):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrChatInvalidCursor):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cursor"})
	case errors.Is(err, service.ErrChatMessageEmpty):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "message body is required"})
	case errors.Is(err, service.ErrChatMessageTooLong):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "message body exceeds 4000 characters"})
	case errors.Is(err, repository.ErrCandidateNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
	case isChatStorageUnavailableError(err):
		log.Printf("chat storage unavailable: %v", err)
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "chat storage is unavailable"})
	default:
		log.Printf("chat request failed: %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func isChatStorageUnavailableError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	if pgErr.Code == "42P01" {
		return true
	}
	if pgErr.Code != "42501" {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(pgErr.Message))
	return strings.Contains(message, "chat_") ||
		strings.Contains(message, "chat ") ||
		strings.Contains(message, "schema public")
}

func requireChatContext(w http.ResponseWriter, r *http.Request) (string, string, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", "", "", false
	}
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", "", "", false
	}

	pairingID, _ := middleware.PairingIDFromContext(r.Context())
	if strings.TrimSpace(pairingID) == "" {
		pairingID = strings.TrimSpace(r.URL.Query().Get("pairing_id"))
	}
	if strings.TrimSpace(pairingID) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "pairing_id is required"})
		return "", "", "", false
	}

	return strings.TrimSpace(userID), strings.TrimSpace(role), strings.TrimSpace(pairingID), true
}

func decodeOptionalJSONBody(w http.ResponseWriter, r *http.Request, maxBytes int64) error {
	if r == nil || r.Body == nil {
		return nil
	}

	var req struct{}
	err := decodeJSONBody(w, r, &req, maxBytes)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
	}
	return err
}

func mapThreadSummaryResponse(thread *service.ThreadSummary) ChatThreadSummaryResponse {
	if thread == nil {
		return ChatThreadSummaryResponse{}
	}

	response := ChatThreadSummaryResponse{
		ID:                 thread.ID,
		PairingID:          thread.PairingID,
		ScopeType:          thread.ScopeType,
		CandidateID:        thread.CandidateID,
		CreatedByUserID:    thread.CreatedByUserID,
		LastMessageAt:      formatTimePointer(thread.LastMessageAt),
		LastMessagePreview: thread.LastMessagePreview,
		UnreadCount:        thread.UnreadCount,
		CreatedAt:          thread.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          thread.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if thread.LastMessage != nil {
		mapped := mapMessageResponse(thread.LastMessage)
		response.LastMessage = &mapped
	}
	if thread.ReadState != nil {
		mapped := mapThreadReadStateResponse(thread.ReadState)
		response.ReadState = &mapped
	}

	return response
}

func mapThreadSummaryResponseFromView(thread *service.ThreadView) ChatThreadSummaryResponse {
	if thread == nil {
		return ChatThreadSummaryResponse{}
	}
	summary := &service.ThreadSummary{
		ID:                 thread.ID,
		PairingID:          thread.PairingID,
		ScopeType:          thread.ScopeType,
		CandidateID:        thread.CandidateID,
		CreatedByUserID:    thread.CreatedByUserID,
		LastMessageAt:      thread.LastMessageAt,
		LastMessagePreview: thread.LastMessagePreview,
		UnreadCount:        thread.UnreadCount,
		LastMessage:        thread.LastMessage,
		ReadState:          thread.ReadState,
		CreatedAt:          thread.CreatedAt,
		UpdatedAt:          thread.UpdatedAt,
	}
	return mapThreadSummaryResponse(summary)
}

func (h *ChatHandler) mapThreadSummaryResponseForUser(userID, role, pairingID string, thread *service.ThreadSummary) (ChatThreadSummaryResponse, error) {
	response := mapThreadSummaryResponse(thread)

	partnerAgency, candidateName, err := h.resolveThreadContext(strings.TrimSpace(userID), strings.TrimSpace(role), strings.TrimSpace(pairingID), response.PairingID, response.CandidateID)
	if err != nil {
		return ChatThreadSummaryResponse{}, err
	}
	response.PartnerAgency = partnerAgency
	response.CandidateName = candidateName

	return response, nil
}

func (h *ChatHandler) mapThreadSummaryResponseFromViewForUser(userID, role, pairingID string, thread *service.ThreadView) (ChatThreadSummaryResponse, error) {
	summary := mapThreadSummaryResponseFromView(thread)

	partnerAgency, candidateName, err := h.resolveThreadContext(strings.TrimSpace(userID), strings.TrimSpace(role), strings.TrimSpace(pairingID), summary.PairingID, summary.CandidateID)
	if err != nil {
		return ChatThreadSummaryResponse{}, err
	}
	summary.PartnerAgency = partnerAgency
	summary.CandidateName = candidateName

	return summary, nil
}

func (h *ChatHandler) resolveThreadContext(userID, role, requestedPairingID, threadPairingID string, candidateID *string) (*ChatPartnerAgencyResponse, *string, error) {
	if h.pairingService == nil || h.userRepository == nil {
		return nil, nil, nil
	}

	pairingLookupID := strings.TrimSpace(threadPairingID)
	if pairingLookupID == "" {
		pairingLookupID = strings.TrimSpace(requestedPairingID)
	}

	pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingLookupID)
	if err != nil {
		return nil, nil, err
	}

	partnerUserID := strings.TrimSpace(pairing.ForeignUserID)
	if strings.TrimSpace(role) == string(domain.ForeignAgent) {
		partnerUserID = strings.TrimSpace(pairing.EthiopianUserID)
	}

	var partnerAgency *ChatPartnerAgencyResponse
	if partnerUserID != "" {
		partnerUser, err := h.userRepository.GetByID(partnerUserID)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, err
		}
		if err == nil && partnerUser != nil {
			partnerAgency = &ChatPartnerAgencyResponse{
				ID:          partnerUser.ID,
				FullName:    partnerUser.FullName,
				CompanyName: partnerUser.CompanyName,
				Email:       partnerUser.Email,
				Role:        string(partnerUser.Role),
			}
		}
	}

	var candidateName *string
	if candidateID != nil && strings.TrimSpace(*candidateID) != "" && h.candidateRepository != nil {
		candidate, err := h.candidateRepository.GetByID(strings.TrimSpace(*candidateID))
		if err != nil && !errors.Is(err, repository.ErrCandidateNotFound) {
			return nil, nil, err
		}
		if err == nil && candidate != nil {
			name := strings.TrimSpace(candidate.FullName)
			if name != "" {
				candidateName = &name
			}
		}
	}

	return partnerAgency, candidateName, nil
}

func mapMessageResponse(message *service.MessageView) ChatMessageResponse {
	if message == nil {
		return ChatMessageResponse{}
	}
	return ChatMessageResponse{
		ID:        message.ID,
		ThreadID:  message.ThreadID,
		Body:      message.Body,
		CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
		Sender: ChatSenderResponse{
			UserID:      message.Sender.UserID,
			FullName:    message.Sender.FullName,
			CompanyName: message.Sender.CompanyName,
			Role:        message.Sender.Role,
		},
	}
}

func mapThreadReadStateResponse(readState *service.ThreadReadState) ChatThreadReadStateResponse {
	if readState == nil {
		return ChatThreadReadStateResponse{}
	}
	updatedAt := readState.UpdatedAt.UTC().Format(time.RFC3339)
	return ChatThreadReadStateResponse{
		ThreadID:          readState.ThreadID,
		UserID:            readState.UserID,
		LastReadMessageID: readState.LastReadMessageID,
		LastReadAt:        formatTimePointer(readState.LastReadAt),
		UpdatedAt:         updatedAt,
	}
}

func parseOptionalPositiveInt(raw string) int {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

func normalizeRecipients(userIDs []string) []string {
	result := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func formatTimePointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
