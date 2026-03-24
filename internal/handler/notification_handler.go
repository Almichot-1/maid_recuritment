package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/pkg/utils"
)

type NotificationResponse struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Message           string  `json:"message"`
	Type              string  `json:"type"`
	IsRead            bool    `json:"is_read"`
	RelatedEntityType string  `json:"related_entity_type"`
	RelatedEntityID   *string `json:"related_entity_id,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

type NotificationListPagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse     `json:"notifications"`
	UnreadCount   int64                      `json:"unread_count"`
	Pagination    NotificationListPagination `json:"pagination"`
}

type NotificationSummaryResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

type notificationPaginator interface {
	GetByUserIDPaginated(userID string, unreadOnly bool, page, pageSize int) ([]*domain.Notification, error)
	CountByUserID(userID string, unreadOnly bool) (int64, error)
	GetByID(id string) (*domain.Notification, error)
}

type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type NotificationHandler struct {
	notificationRepository domain.NotificationRepository
	paginator              notificationPaginator
	upgrader               websocket.Upgrader
	connections            map[string]map[*wsConn]struct{}
	connectionsMu          sync.RWMutex
}

func NewNotificationHandler(notificationRepository domain.NotificationRepository, allowedOrigins []string) *NotificationHandler {
	paginator, _ := notificationRepository.(notificationPaginator)
	normalizedOrigins := normalizeAllowedOrigins(allowedOrigins)
	return &NotificationHandler{
		notificationRepository: notificationRepository,
		paginator:              paginator,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := strings.TrimSpace(r.Header.Get("Origin"))
				if origin == "" {
					return true
				}
				return isAllowedOrigin(origin, normalizedOrigins)
			},
		},
		connections: make(map[string]map[*wsConn]struct{}),
	}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	unreadOnly, page, pageSize := parseNotificationQuery(r)
	if h.paginator == nil {
		items, err := h.notificationRepository.GetByUserID(userID, unreadOnly)
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
		responses := mapNotifications(items)
		unreadCount := int64(0)
		for _, item := range items {
			if !item.IsRead {
				unreadCount++
			}
		}
		_ = utils.WriteJSON(w, http.StatusOK, NotificationListResponse{
			Notifications: responses,
			UnreadCount:   unreadCount,
			Pagination:    NotificationListPagination{Page: 1, PageSize: len(responses), Total: int64(len(responses))},
		})
		return
	}

	items, err := h.paginator.GetByUserIDPaginated(userID, unreadOnly, page, pageSize)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	total, err := h.paginator.CountByUserID(userID, unreadOnly)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	unreadCount, err := h.paginator.CountByUserID(userID, true)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, NotificationListResponse{
		Notifications: mapNotifications(items),
		UnreadCount:   unreadCount,
		Pagination: NotificationListPagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

func (h *NotificationHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.paginator == nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "notification pagination unavailable"})
		return
	}

	unreadCount, err := h.paginator.CountByUserID(userID, true)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, NotificationSummaryResponse{
		UnreadCount: unreadCount,
	})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	notificationID := strings.TrimSpace(chi.URLParam(r, "id"))
	if notificationID == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "notification id is required"})
		return
	}
	if h.paginator == nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "notification pagination unavailable"})
		return
	}

	notification, err := h.paginator.GetByID(notificationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if strings.TrimSpace(notification.UserID) != strings.TrimSpace(userID) {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	if err := h.notificationRepository.MarkAsRead(notificationID); err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "notification marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.notificationRepository.MarkAllAsRead(userID); err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) NotificationsWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &wsConn{conn: conn}
	conn.SetReadLimit(4096)
	h.addConnection(userID, client)
	defer func() {
		h.removeConnection(userID, client)
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *NotificationHandler) PushToUser(userID string, notification *domain.Notification) {
	if notification == nil {
		return
	}
	payload := map[string]NotificationResponse{"notification": mapNotification(notification)}

	h.connectionsMu.RLock()
	userConnections := h.connections[userID]
	h.connectionsMu.RUnlock()
	if len(userConnections) == 0 {
		return
	}

	for client := range userConnections {
		client.mu.Lock()
		err := client.conn.WriteJSON(payload)
		client.mu.Unlock()
		if err != nil {
			h.removeConnection(userID, client)
			_ = client.conn.Close()
		}
	}
}

func (h *NotificationHandler) addConnection(userID string, client *wsConn) {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()
	if h.connections[userID] == nil {
		h.connections[userID] = make(map[*wsConn]struct{})
	}
	h.connections[userID][client] = struct{}{}
}

func (h *NotificationHandler) removeConnection(userID string, client *wsConn) {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()
	if h.connections[userID] == nil {
		return
	}
	delete(h.connections[userID], client)
	if len(h.connections[userID]) == 0 {
		delete(h.connections, userID)
	}
}

func parseNotificationQuery(r *http.Request) (bool, int, int) {
	query := r.URL.Query()
	unreadOnly := strings.EqualFold(strings.TrimSpace(query.Get("unread_only")), "true")

	page := 1
	if raw := strings.TrimSpace(query.Get("page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if raw := strings.TrimSpace(query.Get("page_size")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return unreadOnly, page, pageSize
}

func mapNotifications(notifications []*domain.Notification) []NotificationResponse {
	responses := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		if notification == nil {
			continue
		}
		responses = append(responses, mapNotification(notification))
	}
	return responses
}

func mapNotification(notification *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:                notification.ID,
		Title:             notification.Title,
		Message:           notification.Message,
		Type:              string(notification.Type),
		IsRead:            notification.IsRead,
		RelatedEntityType: notification.RelatedEntityType,
		RelatedEntityID:   notification.RelatedEntityID,
		CreatedAt:         notification.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func normalizeAllowedOrigins(origins []string) []string {
	if len(origins) == 0 {
		return []string{"http://localhost:3000", "http://localhost:3001"}
	}

	normalized := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	if len(normalized) == 0 {
		return []string{"http://localhost:3000", "http://localhost:3001"}
	}

	return normalized
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}
