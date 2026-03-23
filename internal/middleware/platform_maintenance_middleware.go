package middleware

import (
	"net/http"
	"strings"

	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

func PlatformMaintenance(settingsService *service.PlatformSettingsService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if settingsService == nil {
				next.ServeHTTP(w, r)
				return
			}

			settings, err := settingsService.Get()
			if err != nil || settings == nil || !settings.MaintenanceMode {
				next.ServeHTTP(w, r)
				return
			}

			message := strings.TrimSpace(settings.MaintenanceMessage)
			if message == "" {
				message = "The platform is currently under scheduled maintenance. Please try again later."
			}

			_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error":               "platform under maintenance",
				"maintenance_message": message,
			})
		})
	}
}
