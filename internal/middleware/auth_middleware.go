package middleware

import (
	"context"
	"errors"
	"net/http"

	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type contextKey string

const (
	userIDContextKey  contextKey = "user_id"
	roleContextKey    contextKey = "role"
	sessionContextKey contextKey = "session_id"
)

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ResolveBearerToken(r, UserSessionCookieName)
			if token == "" {
				_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authentication credentials"})
				return
			}

			userID, role, sessionID, err := authService.ValidateTokenWithSession(token)
			if err != nil {
				if errors.Is(err, service.ErrInvalidToken) {
					_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
					return
				}
				_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			ctx = context.WithValue(ctx, roleContextKey, role)
			ctx = context.WithValue(ctx, sessionContextKey, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleContextKey).(string)
	return role, ok
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionContextKey).(string)
	return sessionID, ok
}
