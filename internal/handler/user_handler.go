package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/pkg/utils"
)

type UpdateProfileRequest struct {
	FullName    string `json:"full_name" validate:"required,min=2"`
	CompanyName string `json:"company_name" validate:"required,min=2"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UserHandler struct {
	userRepository domain.UserRepository
	inputValidator *validator.Validate
}

func NewUserHandler(userRepository domain.UserRepository) *UserHandler {
	return &UserHandler{
		userRepository: userRepository,
		inputValidator: validator.New(),
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

	_ = utils.WriteJSON(w, http.StatusOK, map[string]AuthUserView{
		"user": mapUserToAuthUserView(user),
	})
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

func (h *UserHandler) writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
	case errors.Is(err, repository.ErrInvalidPassword), errors.Is(err, repository.ErrInvalidEmail), errors.Is(err, repository.ErrDuplicateEmail):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
