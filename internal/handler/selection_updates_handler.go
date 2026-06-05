package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"maid-recruitment-tracking/internal/middleware"
)

// SelectionUpdateMessage represents a real-time selection update sent to clients
type SelectionUpdateMessage struct {
	SelectionID string    `json:"selection_id"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
	Action      string    `json:"action"` // "approve", "reject", "expire"
	PairingID   string    `json:"pairing_id"`
}

type selectionUpdateConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type SelectionUpdatesHandler struct {
	upgrader      websocket.Upgrader
	connections   map[string]map[*selectionUpdateConn]struct{} // userID -> set of connections
	connectionsMu sync.RWMutex
	broadcast     chan *SelectionUpdateMessage
}

const (
	maxSelectionConnectionsPerUser = 3
	selectionWriteWait             = 10 * time.Second
	selectionPongWait              = 60 * time.Second
	selectionPingPeriod            = 45 * time.Second
)

// NewSelectionUpdatesHandler creates a new selection updates handler with WebSocket support
func NewSelectionUpdatesHandler(allowedOrigins []string) *SelectionUpdatesHandler {
	normalizedOrigins := normalizeAllowedOrigins(allowedOrigins)
	handler := &SelectionUpdatesHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := strings.TrimSpace(r.Header.Get("Origin"))
				if origin == "" {
					return false
				}
				return isAllowedOrigin(origin, normalizedOrigins)
			},
		},
		connections: make(map[string]map[*selectionUpdateConn]struct{}),
		broadcast:   make(chan *SelectionUpdateMessage, 100),
	}

	// Start broadcast goroutine
	go handler.broadcastUpdates()

	return handler
}

// SelectionsWebSocket handles WebSocket connections for real-time selection updates
func (h *SelectionUpdatesHandler) SelectionsWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &selectionUpdateConn{conn: conn}

	// Add connection with limit check
	if !h.addConnection(userID, client) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"max connections reached"}`))
		_ = conn.Close()
		return
	}

	defer h.removeConnection(userID, client)

	// Setup ping/pong
	conn.SetReadDeadline(time.Now().Add(selectionPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(selectionPongWait))
		return nil
	})

	// Keep connection alive by reading from it (even though we ignore messages)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Connection closed unexpectedly
			}
			break
		}
	}
}

// BroadcastSelectionUpdate sends a selection update to all connected clients of the affected pairing
func (h *SelectionUpdatesHandler) BroadcastSelectionUpdate(update *SelectionUpdateMessage) {
	h.broadcast <- update
}

// broadcastUpdates handles distribution of updates to interested clients
func (h *SelectionUpdatesHandler) broadcastUpdates() {
	ticker := time.NewTicker(selectionPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case update := <-h.broadcast:
			h.sendUpdateToUsers(update)
		case <-ticker.C:
			h.sendPingsToAll()
		}
	}
}

// sendUpdateToUsers broadcasts an update to all users with active connections
// In a real application, you'd filter by pairing_id subscription
func (h *SelectionUpdatesHandler) sendUpdateToUsers(update *SelectionUpdateMessage) {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	for _, userConns := range h.connections {
		for client := range userConns {
			go func(c *selectionUpdateConn) {
				c.mu.Lock()
				defer c.mu.Unlock()

				c.conn.SetWriteDeadline(time.Now().Add(selectionWriteWait))
				if err := c.conn.WriteJSON(update); err != nil {
					c.conn.Close()
				}
			}(client)
		}
	}
}

// sendPingsToAll sends ping messages to keep connections alive
func (h *SelectionUpdatesHandler) sendPingsToAll() {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	for _, userConns := range h.connections {
		for client := range userConns {
			go func(c *selectionUpdateConn) {
				c.mu.Lock()
				defer c.mu.Unlock()

				c.conn.SetWriteDeadline(time.Now().Add(selectionWriteWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.conn.Close()
				}
			}(client)
		}
	}
}

// addConnection registers a new WebSocket connection for a user
func (h *SelectionUpdatesHandler) addConnection(userID string, client *selectionUpdateConn) bool {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	if _, exists := h.connections[userID]; !exists {
		h.connections[userID] = make(map[*selectionUpdateConn]struct{})
	}

	// Check connection limit
	if len(h.connections[userID]) >= maxSelectionConnectionsPerUser {
		return false
	}

	h.connections[userID][client] = struct{}{}
	return true
}

// removeConnection unregisters a WebSocket connection
func (h *SelectionUpdatesHandler) removeConnection(userID string, client *selectionUpdateConn) {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	if userConns, exists := h.connections[userID]; exists {
		delete(userConns, client)
		if len(userConns) == 0 {
			delete(h.connections, userID)
		}
	}

	client.conn.Close()
}

// PushSelectionUpdate broadcasts a selection update event
// Called by selection service when status changes
func (h *SelectionUpdatesHandler) PushSelectionUpdate(selectionID, status, action, pairingID string) {
	update := &SelectionUpdateMessage{
		SelectionID: selectionID,
		Status:      status,
		UpdatedAt:   time.Now(),
		Action:      action,
		PairingID:   pairingID,
	}
	h.BroadcastSelectionUpdate(update)
}

// GetConnectionCount returns the number of active WebSocket connections (for monitoring)
func (h *SelectionUpdatesHandler) GetConnectionCount() int {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	total := 0
	for _, userConns := range h.connections {
		total += len(userConns)
	}
	return total
}
