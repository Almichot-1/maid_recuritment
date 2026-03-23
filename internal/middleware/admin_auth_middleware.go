package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type adminContextKey string

const (
	adminIDContextKey   adminContextKey = "admin_id"
	adminRoleContextKey adminContextKey = "admin_role"
)

func AdminAuthMiddleware(authService *service.AdminAuthService) func(http.Handler) http.Handler {
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

			adminID, role, err := authService.ValidateToken(parts[1])
			if err != nil {
				switch {
				case errors.Is(err, service.ErrAdminInvalidToken):
					_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid admin token"})
				case errors.Is(err, service.ErrAdminAccountLocked):
					_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "admin account locked"})
				case errors.Is(err, service.ErrAdminInactive):
					_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "admin account inactive"})
				default:
					_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				}
				return
			}

			ctx := context.WithValue(r.Context(), adminIDContextKey, adminID)
			ctx = context.WithValue(ctx, adminRoleContextKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminIDFromContext(ctx context.Context) (string, bool) {
	adminID, ok := ctx.Value(adminIDContextKey).(string)
	return adminID, ok
}

func AdminRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(adminRoleContextKey).(string)
	return role, ok
}

func RequireAdminRole(allowedRoles ...domain.AdminRole) func(http.Handler) http.Handler {
	normalized := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		if strings.TrimSpace(string(role)) == "" {
			continue
		}
		normalized[strings.TrimSpace(string(role))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := AdminRoleFromContext(r.Context())
			if !ok || strings.TrimSpace(role) == "" {
				_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin role required"})
				return
			}
			if _, allowed := normalized[strings.TrimSpace(role)]; !allowed {
				_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient admin permissions"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
