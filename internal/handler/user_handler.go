package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

const maxAvatarFileSizeBytes = 2 * 1024 * 1024

type UpdateProfileRequest struct {
	FullName    string `json:"full_name" validate:"required,min=2"`
	CompanyName string `json:"company_name" validate:"required,min=2"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UpdateSharingPreferencesRequest struct {
	AutoShareCandidates     bool    `json:"auto_share_candidates"`
	DefaultForeignPairingID *string `json:"default_foreign_pairing_id"`
}

type UserSessionResponse struct {
	ID          string `json:"id"`
	DeviceLabel string `json:"device_label"`
	BrowserName string `json:"browser_name"`
	OSName      string `json:"os_name"`
	IPAddress   string `json:"ip_address,omitempty"`
	LastSeenAt  string `json:"last_seen_at"`
	ExpiresAt   string `json:"expires_at"`
}

type UserSessionsListResponse struct {
	Sessions         []UserSessionResponse `json:"sessions"`
	CurrentSessionID string                `json:"current_session_id,omitempty"`
}

type UserHandler struct {
	userRepository    domain.UserRepository
	sessionRepository domain.UserSessionRepository
	storageService    service.StorageService
	pairingService    *service.PairingService
	inputValidator    *validator.Validate
}

func NewUserHandler(userRepository domain.UserRepository, sessionRepository domain.UserSessionRepository, storageService service.StorageService, pairingService *service.PairingService) *UserHandler {
	return &UserHandler{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		storageService:    storageService,
		pairingService:    pairingService,
		inputValidator:    validator.New(),
	}
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	user, err := h.userRepository.GetByID(strings.TrimSpace(userID))
	if err != nil {
		h.writeUserError(w, err)
		return
	}

	user.FullName = strings.TrimSpace(req.FullName)
	user.CompanyName = strings.TrimSpace(req.CompanyName)

	if err := h.userRepository.Update(user); err != nil {
		h.writeUserError(w, err)
		return
	}

	sessionID, _ := middleware.SessionIDFromContext(r.Context())
	_ = utils.WriteJSON(w, http.StatusOK, map[string]AuthUserView{
		"user": mapUserToAuthUserView(user, sessionID),
	})
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if h.storageService == nil {
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "profile photo upload is not configured"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.userRepository.GetByID(strings.TrimSpace(userID))
	if err != nil {
		h.writeUserError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarFileSizeBytes+1024)
	if err := r.ParseMultipartForm(maxAvatarFileSizeBytes + 1024); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			_ = utils.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "profile photo must be smaller than 2 MB"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	bufferedFile, contentType, err := service.ValidateAndBufferUploadForProfile(file, fileHeader.Filename)
	if err != nil {
		h.writeUserError(w, err)
		return
	}

	bufferedBytes, err := io.ReadAll(bufferedFile)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read profile photo"})
		return
	}

	oldAvatarURL := strings.TrimSpace(user.AvatarURL)
	newAvatarURL, err := h.storageService.Upload(bytes.NewReader(bufferedBytes), fileHeader.Filename, contentType)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to upload profile photo"})
		return
	}

	user.AvatarURL = newAvatarURL
	if err := h.userRepository.Update(user); err != nil {
		_ = h.storageService.Delete(newAvatarURL)
		h.writeUserError(w, err)
		return
	}
	if oldAvatarURL != "" && oldAvatarURL != newAvatarURL {
		_ = h.storageService.Delete(oldAvatarURL)
	}

	sessionID, _ := middleware.SessionIDFromContext(r.Context())
	_ = utils.WriteJSON(w, http.StatusOK, map[string]AuthUserView{
		"user": mapUserToAuthUserView(user, sessionID),
	})
}

func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.userRepository.GetByID(strings.TrimSpace(userID))
	if err != nil {
		h.writeUserError(w, err)
		return
	}

	oldAvatarURL := strings.TrimSpace(user.AvatarURL)
	user.AvatarURL = ""
	if err := h.userRepository.Update(user); err != nil {
		h.writeUserError(w, err)
		return
	}
	if oldAvatarURL != "" && h.storageService != nil {
		_ = h.storageService.Delete(oldAvatarURL)
	}

	sessionID, _ := middleware.SessionIDFromContext(r.Context())
	_ = utils.WriteJSON(w, http.StatusOK, map[string]AuthUserView{
		"user": mapUserToAuthUserView(user, sessionID),
	})
}

func (h *UserHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	currentSessionID, _ := middleware.SessionIDFromContext(r.Context())
	if h.sessionRepository == nil {
		_ = utils.WriteJSON(w, http.StatusOK, UserSessionsListResponse{
			Sessions:         []UserSessionResponse{},
			CurrentSessionID: currentSessionID,
		})
		return
	}

	sessions, err := h.sessionRepository.ListActiveByUserID(strings.TrimSpace(userID))
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load active sessions"})
		return
	}

	response := make([]UserSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, mapUserSessionResponse(session))
	}

	_ = utils.WriteJSON(w, http.StatusOK, UserSessionsListResponse{
		Sessions:         response,
		CurrentSessionID: currentSessionID,
	})
}

func (h *UserHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.sessionRepository == nil {
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "session management is not configured"})
		return
	}

	sessionID := chi.URLParam(r, "id")
	if strings.TrimSpace(sessionID) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "session id is required"})
		return
	}

	if err := h.sessionRepository.RevokeByID(strings.TrimSpace(userID), sessionID); err != nil {
		if errors.Is(err, repository.ErrUserSessionNotFound) {
			_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke session"})
		return
	}

	currentSessionID, _ := middleware.SessionIDFromContext(r.Context())
	if strings.TrimSpace(currentSessionID) == strings.TrimSpace(sessionID) {
		middleware.ClearSessionCookie(w, r, middleware.UserSessionCookieName)
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "session revoked"})
}

func (h *UserHandler) LogoutAllSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.sessionRepository == nil {
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "session management is not configured"})
		return
	}

	if err := h.sessionRepository.RevokeAllByUserID(strings.TrimSpace(userID), ""); err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to clear active sessions"})
		return
	}

	middleware.ClearSessionCookie(w, r, middleware.UserSessionCookieName)
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "all sessions cleared"})
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}

	user, err := h.userRepository.GetByID(strings.TrimSpace(userID))
	if err != nil {
		h.writeUserError(w, err)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)) != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "current password is incorrect"})
		return
	}

	user.PasswordHash = req.NewPassword
	if err := h.userRepository.Update(user); err != nil {
		h.writeUserError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

func (h *UserHandler) UpdateSharingPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.userRepository.GetByID(strings.TrimSpace(userID))
	if err != nil {
		h.writeUserError(w, err)
		return
	}
	if user.Role != domain.EthiopianAgent {
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var req UpdateSharingPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	defaultPairingID, err := h.validateDefaultForeignPairing(user.ID, req.DefaultForeignPairingID)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user.AutoShareCandidates = req.AutoShareCandidates
	user.DefaultForeignPairingID = defaultPairingID

	if err := h.userRepository.Update(user); err != nil {
		h.writeUserError(w, err)
		return
	}

	sessionID, _ := middleware.SessionIDFromContext(r.Context())
	_ = utils.WriteJSON(w, http.StatusOK, map[string]AuthUserView{
		"user": mapUserToAuthUserView(user, sessionID),
	})
}

func (h *UserHandler) writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
	case errors.Is(err, repository.ErrInvalidPassword), errors.Is(err, repository.ErrInvalidEmail), errors.Is(err, repository.ErrDuplicateEmail):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
	case errors.Is(err, service.ErrInvalidFileType), errors.Is(err, service.ErrUnsupportedContentType):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Please upload a JPG, PNG, or WEBP image."})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func mapUserSessionResponse(session *domain.UserSession) UserSessionResponse {
	browserName, osName := parseSessionUserAgent(session.UserAgent)
	return UserSessionResponse{
		ID:          session.ID,
		DeviceLabel: buildDeviceLabel(browserName, osName),
		BrowserName: browserName,
		OSName:      osName,
		IPAddress:   strings.TrimSpace(session.IPAddress),
		LastSeenAt:  session.LastSeenAt.UTC().Format(time.RFC3339),
		ExpiresAt:   session.ExpiresAt.UTC().Format(time.RFC3339),
	}
}

func parseSessionUserAgent(userAgent string) (string, string) {
	normalized := strings.ToLower(strings.TrimSpace(userAgent))

	browserName := "Browser"
	switch {
	case strings.Contains(normalized, "edg/"):
		browserName = "Microsoft Edge"
	case strings.Contains(normalized, "chrome/"):
		browserName = "Chrome"
	case strings.Contains(normalized, "firefox/"):
		browserName = "Firefox"
	case strings.Contains(normalized, "safari/") && !strings.Contains(normalized, "chrome/"):
		browserName = "Safari"
	}

	osName := "Unknown OS"
	switch {
	case strings.Contains(normalized, "windows"):
		osName = "Windows"
	case strings.Contains(normalized, "android"):
		osName = "Android"
	case strings.Contains(normalized, "iphone"), strings.Contains(normalized, "ipad"), strings.Contains(normalized, "ios"):
		osName = "iOS"
	case strings.Contains(normalized, "mac os"), strings.Contains(normalized, "macintosh"):
		osName = "macOS"
	case strings.Contains(normalized, "linux"):
		osName = "Linux"
	}

	return browserName, osName
}

func buildDeviceLabel(browserName, osName string) string {
	switch {
	case browserName != "" && osName != "":
		return browserName + " on " + osName
	case browserName != "":
		return browserName
	case osName != "":
		return osName
	default:
		return "Active session"
	}
}

func (h *UserHandler) validateDefaultForeignPairing(userID string, pairingID *string) (*string, error) {
	if pairingID == nil || strings.TrimSpace(*pairingID) == "" {
		return nil, nil
	}
	if h.pairingService == nil {
		return nil, errors.New("sharing preferences are not configured")
	}

	activePairings, err := h.pairingService.ListActivePairingsForUser(strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	for _, pairing := range activePairings {
		if pairing != nil && strings.TrimSpace(pairing.ID) == strings.TrimSpace(*pairingID) {
			trimmed := strings.TrimSpace(*pairingID)
			return &trimmed, nil
		}
	}

	return nil, errors.New("default foreign partner must be one of your active pairings")
}
