package middleware

import (
	"net/http"

	"maid-recruitment-tracking/pkg/utils"
)

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			if _, exists := allowed[role]; !exists {
				_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
