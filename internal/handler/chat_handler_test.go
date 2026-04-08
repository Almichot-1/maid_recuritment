package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/service"
)

func TestChatHandlerResolveWorkspaceThreadUnauthorized(t *testing.T) {
	h := NewChatHandler(nil, []string{"http://localhost:3000"})
	req := httptest.NewRequest(http.MethodPost, "/chat/threads/resolve-workspace", strings.NewReader(`{}`))
	recorder := httptest.NewRecorder()

	h.ResolveWorkspaceThread(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	body := decodeChatHandlerBody(t, recorder)
	assert.Equal(t, "unauthorized", body["error"])
}

func TestChatHandlerChatWebSocketUnauthorized(t *testing.T) {
	h := NewChatHandler(nil, []string{"http://localhost:3000"})
	req := httptest.NewRequest(http.MethodGet, "/ws/chat?pairing_id=pair-1", nil)
	recorder := httptest.NewRecorder()

	h.ChatWebSocket(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	body := decodeChatHandlerBody(t, recorder)
	assert.Equal(t, "unauthorized", body["error"])
}

func TestChatHandlerWriteChatErrorPairingAccessDenied(t *testing.T) {
	h := NewChatHandler(nil, nil)
	recorder := httptest.NewRecorder()

	h.writeChatError(recorder, service.ErrPairingAccessDenied)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	body := decodeChatHandlerBody(t, recorder)
	assert.Equal(t, "forbidden", body["error"])
}

func TestChatHandlerWriteChatErrorInvalidMessageBody(t *testing.T) {
	h := NewChatHandler(nil, nil)
	recorder := httptest.NewRecorder()

	h.writeChatError(recorder, service.ErrChatMessageEmpty)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	body := decodeChatHandlerBody(t, recorder)
	assert.Equal(t, "message body is required", body["error"])
}

func TestChatHandlerWriteChatErrorStorageUnavailable(t *testing.T) {
	h := NewChatHandler(nil, nil)
	recorder := httptest.NewRecorder()

	err := fmt.Errorf("list chat threads: %w", &pgconn.PgError{Code: "42P01", Message: "relation \"chat_threads\" does not exist"})
	h.writeChatError(recorder, err)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	body := decodeChatHandlerBody(t, recorder)
	assert.Equal(t, "chat storage is unavailable", body["error"])
}

func TestIsChatStorageUnavailableErrorPermissionDenied(t *testing.T) {
	err := fmt.Errorf("resolve chat thread: %w", &pgconn.PgError{Code: "42501", Message: "permission denied for schema public"})
	assert.True(t, isChatStorageUnavailableError(err))

	err = fmt.Errorf("something else: %w", &pgconn.PgError{Code: "42501", Message: "permission denied for table users"})
	assert.False(t, isChatStorageUnavailableError(err))
}

func TestChatWSOutboundEventSerialization(t *testing.T) {
	event := chatWSOutboundEvent{
		Type: chatEventMessageCreated,
		Payload: chatMessageCreatedPayload{
			ThreadID: "thread-1",
			Message: ChatMessageResponse{
				ID:       "message-1",
				ThreadID: "thread-1",
				Body:     "hello",
				Sender: ChatSenderResponse{
					UserID:      "user-1",
					FullName:    "Sender",
					CompanyName: "Agency",
					Role:        "ForeignAgent",
				},
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	raw, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, chatEventMessageCreated, decoded["type"])

	payload, ok := decoded["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "thread-1", payload["thread_id"])

	message, ok := payload["message"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "hello", message["body"])
}

func decodeChatHandlerBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]string {
	t.Helper()

	body := make(map[string]string)
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	return body
}
