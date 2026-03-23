package middleware

import (
	"context"
	"net/http"
	"strings"
)

const pairingIDContextKey contextKey = "pairing_id"

func PairingContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pairingID := strings.TrimSpace(r.Header.Get("X-Pairing-ID"))
		if pairingID == "" {
			pairingID = strings.TrimSpace(r.URL.Query().Get("pairing_id"))
		}

		ctx := context.WithValue(r.Context(), pairingIDContextKey, pairingID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func PairingIDFromContext(ctx context.Context) (string, bool) {
	pairingID, ok := ctx.Value(pairingIDContextKey).(string)
	return strings.TrimSpace(pairingID), ok
}
