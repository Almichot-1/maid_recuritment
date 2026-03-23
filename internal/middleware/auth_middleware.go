package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type contextKey string

const (
	userIDContextKey contextKey = "user_id"
	roleContextKey   contextKey = "role"
)

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
				return
			}

			userID, role, err := authService.ValidateToken(parts[1])
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
