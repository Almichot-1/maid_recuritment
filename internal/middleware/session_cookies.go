package middleware

import (
	"net/http"
	"strings"
	"time"
)

const (
	UserSessionCookieName  = "auth_session"
	AdminSessionCookieName = "admin_auth_session"

	UserSessionMaxAgeSeconds  = 24 * 60 * 60
	AdminSessionMaxAgeSeconds = 60 * 60
)

func ResolveBearerToken(r *http.Request, cookieName string) string {
	if r == nil {
		return ""
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token := strings.TrimSpace(parts[1])
			if token != "" {
				return token
			}
		}
	}

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}

func SetSessionCookie(w http.ResponseWriter, r *http.Request, cookieName, token string, maxAgeSeconds int) {
	if strings.TrimSpace(token) == "" {
		return
	}

	sameSite, secure := sessionCookieSettings(r)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   maxAgeSeconds,
		Expires:  time.Now().Add(time.Duration(maxAgeSeconds) * time.Second),
	})
}

func ClearSessionCookie(w http.ResponseWriter, r *http.Request, cookieName string) {
	sameSite, secure := sessionCookieSettings(r)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func sessionCookieSettings(r *http.Request) (http.SameSite, bool) {
	if isLocalRequest(r) {
		return http.SameSiteLaxMode, false
	}

	return http.SameSiteNoneMode, true
}

func isLocalRequest(r *http.Request) bool {
	if r == nil {
		return true
	}

	values := []string{
		strings.TrimSpace(r.Host),
		strings.TrimSpace(r.Header.Get("Origin")),
		strings.TrimSpace(r.Header.Get("Referer")),
	}

	for _, value := range values {
		lower := strings.ToLower(value)
		if lower == "" {
			continue
		}
		if strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") {
			return true
		}
	}

	return false
}
