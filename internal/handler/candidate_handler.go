package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/middleware"
	"maid-recruitment-tracking/internal/repository"
	"maid-recruitment-tracking/internal/service"
	"maid-recruitment-tracking/pkg/utils"
)

type CreateCandidateRequest struct {
	FullName        string   `json:"full_name" validate:"required"`
	Nationality     string   `json:"nationality"`
	DateOfBirth     string   `json:"date_of_birth"`
	Age             *int     `json:"age" validate:"omitempty,min=18,max=65"`
	PlaceOfBirth    string   `json:"place_of_birth"`
	Religion        string   `json:"religion"`
	MaritalStatus   string   `json:"marital_status"`
	ChildrenCount   *int     `json:"children_count" validate:"omitempty,min=0"`
	EducationLevel  string   `json:"education_level"`
	ExperienceYears *int     `json:"experience_years" validate:"omitempty,min=0,max=30"`
	Languages       []string `json:"languages"`
	Skills          []string `json:"skills"`
}

type UpdateCandidateRequest struct {
	FullName        string   `json:"full_name" validate:"required"`
	Nationality     string   `json:"nationality"`
	DateOfBirth     string   `json:"date_of_birth"`
	Age             *int     `json:"age" validate:"omitempty,min=18,max=65"`
	PlaceOfBirth    string   `json:"place_of_birth"`
	Religion        string   `json:"religion"`
	MaritalStatus   string   `json:"marital_status"`
	ChildrenCount   *int     `json:"children_count" validate:"omitempty,min=0"`
	EducationLevel  string   `json:"education_level"`
	ExperienceYears *int     `json:"experience_years" validate:"omitempty,min=0,max=30"`
	Languages       []string `json:"languages"`
	Skills          []string `json:"skills"`
}

type ListCandidatesQuery struct {
	Status        []string
	Search        string
	MinAge        *int
	MaxAge        *int
	MinExperience *int
	MaxExperience *int
	Languages     []string
	SharedOnly    bool
	Page          int
	PageSize      int
}

type CandidateCreatedByInfo struct {
	ID string `json:"id"`
}

type CandidateDocumentResponse struct {
	ID           string `json:"id"`
	DocumentType string `json:"document_type"`
	FileURL      string `json:"file_url"`
	FileName     string `json:"file_name,omitempty"`
	FileSize     int64  `json:"file_size"`
	UploadedAt   string `json:"uploaded_at"`
}

type CandidateResponse struct {
	ID              string                      `json:"id"`
	CreatedBy       CandidateCreatedByInfo      `json:"created_by"`
	FullName        string                      `json:"full_name"`
	Nationality     string                      `json:"nationality,omitempty"`
	DateOfBirth     *string                     `json:"date_of_birth,omitempty"`
	Age             *int                        `json:"age,omitempty"`
	PlaceOfBirth    string                      `json:"place_of_birth,omitempty"`
	Religion        string                      `json:"religion,omitempty"`
	MaritalStatus   string                      `json:"marital_status,omitempty"`
	ChildrenCount   *int                        `json:"children_count,omitempty"`
	EducationLevel  string                      `json:"education_level,omitempty"`
	ExperienceYears *int                        `json:"experience_years,omitempty"`
	Languages       []string                    `json:"languages"`
	Skills          []string                    `json:"skills"`
	Status          string                      `json:"status"`
	LockedBy        *string                     `json:"locked_by,omitempty"`
	LockedAt        *string                     `json:"locked_at,omitempty"`
	LockExpiresAt   *string                     `json:"lock_expires_at,omitempty"`
	CVPDFURL        string                      `json:"cv_pdf_url,omitempty"`
	Documents       []CandidateDocumentResponse `json:"documents"`
	CreatedAt       string                      `json:"created_at"`
	UpdatedAt       string                      `json:"updated_at"`
}

type CandidateListResponse struct {
	Candidates []CandidateResponse `json:"candidates"`
	Meta       CandidateListMeta   `json:"meta"`
}

type CandidateListMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Count    int `json:"count"`
}

type GenerateCVResponse struct {
	CVPDFURL string `json:"cv_pdf_url"`
}

type PassportDataResponse struct {
	ID               string  `json:"id"`
	CandidateID      string  `json:"candidate_id"`
	HolderName       string  `json:"holder_name"`
	PassportNumber   string  `json:"passport_number"`
	CountryCode      string  `json:"country_code,omitempty"`
	Nationality      string  `json:"nationality"`
	DateOfBirth      string  `json:"date_of_birth"`
	PlaceOfBirth     string  `json:"place_of_birth,omitempty"`
	Gender           string  `json:"gender"`
	IssueDate        *string `json:"issue_date,omitempty"`
	ExpiryDate       string  `json:"expiry_date"`
	IssuingAuthority string  `json:"issuing_authority,omitempty"`
	MRZLine1         string  `json:"mrz_line_1"`
	MRZLine2         string  `json:"mrz_line_2"`
	Confidence       float64 `json:"confidence"`
	ExtractedAt      string  `json:"extracted_at"`
}

type GenerateCVRequest struct {
	BrandingLogoDataURL string `json:"branding_logo_data_url"`
	CompanyName         string `json:"company_name"`
}

type CandidateHandler struct {
	candidateService    *service.CandidateService
	passportOCRService  *service.PassportOCRService
	candidateRepository *repository.GormCandidateRepository
	selectionRepository domain.SelectionRepository
	pairingService      *service.PairingService
	inputValidator      *validator.Validate
}

func NewCandidateHandler(candidateService *service.CandidateService, passportOCRService *service.PassportOCRService, candidateRepository *repository.GormCandidateRepository, selectionRepository domain.SelectionRepository, pairingService *service.PairingService) *CandidateHandler {
	return &CandidateHandler{
		candidateService:    candidateService,
		passportOCRService:  passportOCRService,
		candidateRepository: candidateRepository,
		selectionRepository: selectionRepository,
		pairingService:      pairingService,
		inputValidator:      validator.New(),
	}
}

func (h *CandidateHandler) CreateCandidate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}
	dateOfBirth, err := parseOptionalDateString(req.DateOfBirth)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date of birth"})
		return
	}

	candidate, err := h.candidateService.CreateCandidate(userID, service.CandidateInput{
		FullName:        req.FullName,
		Nationality:     req.Nationality,
		DateOfBirth:     dateOfBirth,
		Age:             req.Age,
		PlaceOfBirth:    req.PlaceOfBirth,
		Religion:        req.Religion,
		MaritalStatus:   req.MaritalStatus,
		ChildrenCount:   req.ChildrenCount,
		EducationLevel:  req.EducationLevel,
		ExperienceYears: req.ExperienceYears,
		Languages:       req.Languages,
		Skills:          req.Skills,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, map[string]CandidateResponse{
		"candidate": mapCandidateResponse(candidate, nil),
	})
}

func (h *CandidateHandler) UpdateCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req UpdateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.inputValidator.Struct(req); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
		return
	}
	dateOfBirth, err := parseOptionalDateString(req.DateOfBirth)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date of birth"})
		return
	}

	err = h.candidateService.UpdateCandidate(id, userID, service.CandidateInput{
		FullName:        req.FullName,
		Nationality:     req.Nationality,
		DateOfBirth:     dateOfBirth,
		Age:             req.Age,
		PlaceOfBirth:    req.PlaceOfBirth,
		Religion:        req.Religion,
		MaritalStatus:   req.MaritalStatus,
		ChildrenCount:   req.ChildrenCount,
		EducationLevel:  req.EducationLevel,
		ExperienceYears: req.ExperienceYears,
		Languages:       req.Languages,
		Skills:          req.Skills,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "candidate updated"})
}

func (h *CandidateHandler) GetCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	candidate, documents, err := h.candidateService.GetCandidate(id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	canView, err := h.canViewCandidate(candidate, userID, role, pairingID)
	if err != nil {
		h.writePairingAccessError(w, err)
		return
	}
	if !canView {
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
		return
	}
	sanitizeCandidateForViewer(candidate, userID, role)

	_ = utils.WriteJSON(w, http.StatusOK, map[string]CandidateResponse{
		"candidate": mapCandidateResponse(candidate, documents),
	})
}

func (h *CandidateHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok || strings.TrimSpace(role) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	pairingID, _ := middleware.PairingIDFromContext(r.Context())

	query, err := parseListCandidatesQuery(r)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
		return
	}

	statusFilters, err := parseStatusFilters(query.Status)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status filter"})
		return
	}

	filters := domain.CandidateFilters{
		Statuses:      statusFilters,
		Search:        query.Search,
		MinAge:        query.MinAge,
		MaxAge:        query.MaxAge,
		MinExperience: query.MinExperience,
		MaxExperience: query.MaxExperience,
		Languages:     query.Languages,
		SharedOnly:    query.SharedOnly,
		Page:          query.Page,
		PageSize:      query.PageSize,
	}
	scopedFilters, err := applyRoleScopedCandidateFilters(role, userID, filters)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	if strings.TrimSpace(role) == string(domain.EthiopianAgent) && query.SharedOnly {
		if h.pairingService == nil {
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "workspace access is not configured"})
			return
		}
		pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingID)
		if err != nil {
			h.writePairingAccessError(w, err)
			return
		}
		scopedFilters.PairingID = pairing.ID
		scopedFilters.SharedOnly = true
	}
	if strings.TrimSpace(role) == string(domain.ForeignAgent) {
		if h.pairingService == nil {
			_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "workspace access is not configured"})
			return
		}
		pairing, err := h.pairingService.ResolveActivePairing(userID, role, pairingID)
		if err != nil {
			h.writePairingAccessError(w, err)
			return
		}
		scopedFilters.PairingID = pairing.ID
		scopedFilters.SharedOnly = true
		scopedFilters.Statuses = []domain.CandidateStatus{domain.CandidateStatusAvailable}
	}

	candidates, err := h.candidateService.ListCandidatesForWorkspace(role, userID, pairingID, scopedFilters)
	if err != nil {
		if h.tryWritePairingServiceError(w, err) {
			return
		}
		h.writeServiceError(w, err)
		return
	}

	total, err := h.candidateRepository.Count(scopedFilters)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	responses := make([]CandidateResponse, 0, len(candidates))
	for _, candidate := range candidates {
		sanitizeCandidateForViewer(candidate, userID, role)
		responses = append(responses, mapCandidateResponse(candidate, nil))
	}

	_ = utils.WriteJSON(w, http.StatusOK, CandidateListResponse{
		Candidates: responses,
		Meta: CandidateListMeta{
			Page:     query.Page,
			PageSize: query.PageSize,
			Count:    int(total),
		},
	})
}

func (h *CandidateHandler) PublishCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.candidateService.PublishCandidate(id, userID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "candidate published"})
}

func (h *CandidateHandler) UploadCandidateDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 52<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			_ = utils.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "file is too large"})
			return
		}
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	documentType := strings.TrimSpace(r.FormValue("document_type"))
	if documentType == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "document type is required"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	document, err := h.candidateService.UploadCandidateDocument(id, userID, service.UploadCandidateDocumentInput{
		DocumentType: documentType,
		File:         file,
		FileName:     fileHeader.Filename,
		FileSize:     fileHeader.Size,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, map[string]CandidateDocumentResponse{
		"document": mapDocumentResponse(document),
	})
}

func (h *CandidateHandler) GenerateCV(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req GenerateCVRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
	}

	if err := h.candidateService.GenerateCV(id, userID, service.CandidateCVBranding{
		CompanyName: strings.TrimSpace(req.CompanyName),
		LogoDataURL: strings.TrimSpace(req.BrandingLogoDataURL),
	}); err != nil {
		h.writeServiceError(w, err)
		return
	}

	candidate, _, err := h.candidateService.GetCandidate(id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, GenerateCVResponse{CVPDFURL: candidate.CVPDFURL})
}

func (h *CandidateHandler) ParsePassport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.passportOCRService == nil {
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "passport OCR is not configured"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	passportData, err := h.passportOCRService.ParseAndStore(id, userID, file, fileHeader.Filename)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	_ = h.candidateService.ApplyPassportAutofill(id, userID, passportData)

	_ = utils.WriteJSON(w, http.StatusOK, map[string]PassportDataResponse{
		"passport": mapPassportDataResponse(passportData),
	})
}

func (h *CandidateHandler) ParsePassportPreview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.passportOCRService == nil {
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "passport OCR is not configured"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	passportData, err := h.passportOCRService.ParsePreview(file, fileHeader.Filename)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]PassportDataResponse{
		"passport": mapPassportDataResponse(passportData),
	})
}

func (h *CandidateHandler) GetPassport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.passportOCRService == nil {
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "passport OCR is not configured"})
		return
	}

	passportData, err := h.passportOCRService.GetByCandidateID(id, userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]PassportDataResponse{
		"passport": mapPassportDataResponse(passportData),
	})
}

func (h *CandidateHandler) DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "candidate id is required"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		_ = utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.candidateService.DeleteCandidate(id, userID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "candidate deleted"})
}

func (h *CandidateHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrCandidateNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "candidate not found"})
	case errors.Is(err, service.ErrForbidden):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, service.ErrCandidateLocked):
		_ = utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": "candidate is locked"})
	case errors.Is(err, service.ErrMissingRequiredDocuments):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing required documents"})
	case errors.Is(err, service.ErrPDFGenerationFailed):
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "pdf generation failed"})
	case errors.Is(err, service.ErrFileTooLarge):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file size exceeds the allowed limit"})
	case errors.Is(err, service.ErrInvalidFileType):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported file type for this upload"})
	case errors.Is(err, service.ErrInvalidCandidateDocumentType):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document type"})
	case errors.Is(err, service.ErrPassportOCRRequiresImage):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "passport OCR requires a JPG or PNG image"})
	case errors.Is(err, service.ErrPassportOCRParseFailed):
		log.Printf("passport OCR parse failed: %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "We could not read that passport image. Use a clear JPG or PNG photo of the passport page."})
	case errors.Is(err, service.ErrPassportOCRUnavailable):
		log.Printf("passport OCR unavailable: %v", err)
		_ = utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "passport OCR is not available right now"})
	case errors.Is(err, service.ErrInvalidCandidateInput), errors.Is(err, service.ErrInvalidCandidateUpdateState), errors.Is(err, service.ErrInvalidCandidateDeleteState), errors.Is(err, repository.ErrInvalidStatusTransition):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "validation failed"})
	case errors.Is(err, service.ErrPassportDataNotFound), errors.Is(err, repository.ErrPassportDataNotFound):
		_ = utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "passport data not found"})
	default:
		_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func parseListCandidatesQuery(r *http.Request) (ListCandidatesQuery, error) {
	query := r.URL.Query()

	minAge, err := parseOptionalInt(query.Get("min_age"))
	if err != nil {
		return ListCandidatesQuery{}, err
	}
	maxAge, err := parseOptionalInt(query.Get("max_age"))
	if err != nil {
		return ListCandidatesQuery{}, err
	}
	minExperience, err := parseOptionalInt(query.Get("min_experience"))
	if err != nil {
		return ListCandidatesQuery{}, err
	}
	maxExperience, err := parseOptionalInt(query.Get("max_experience"))
	if err != nil {
		return ListCandidatesQuery{}, err
	}
	page, err := parseIntWithDefault(query.Get("page"), 1)
	if err != nil {
		return ListCandidatesQuery{}, err
	}
	pageSize, err := parseIntWithDefault(query.Get("page_size"), 20)
	if err != nil {
		return ListCandidatesQuery{}, err
	}

	languages := splitCSVQueryValues(query["languages"])
	if len(languages) == 0 {
		languages = splitCSVQueryValues(query["language"])
	}

	return ListCandidatesQuery{
		Status:        splitCSVQueryValues(query["status"]),
		Search:        strings.TrimSpace(query.Get("search")),
		MinAge:        minAge,
		MaxAge:        maxAge,
		MinExperience: minExperience,
		MaxExperience: maxExperience,
		Languages:     languages,
		SharedOnly:    strings.EqualFold(strings.TrimSpace(query.Get("shared_only")), "true"),
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

func parseStatusFilters(values []string) ([]domain.CandidateStatus, error) {
	parsed := make([]domain.CandidateStatus, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		switch trimmed {
		case "":
			continue
		case string(domain.CandidateStatusDraft), string(domain.CandidateStatusAvailable), string(domain.CandidateStatusLocked),
			string(domain.CandidateStatusUnderReview), string(domain.CandidateStatusApproved), string(domain.CandidateStatusRejected),
			string(domain.CandidateStatusInProgress), string(domain.CandidateStatusCompleted):
			parsed = append(parsed, domain.CandidateStatus(trimmed))
		default:
			return nil, errors.New("invalid status")
		}
	}
	return parsed, nil
}

func parseOptionalInt(value string) (*int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseIntWithDefault(value string, fallback int) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return fallback, nil
	}
	return parsed, nil
}

func splitCSVQueryValues(values []string) []string {
	parsed := make([]string, 0)
	for _, value := range values {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			parsed = append(parsed, trimmed)
		}
	}
	return parsed
}

func mapCandidateResponse(candidate *domain.Candidate, documents []*domain.Document) CandidateResponse {
	createdAt := candidate.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	updatedAt := candidate.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")

	var lockedAt *string
	if candidate.LockedAt != nil {
		formatted := candidate.LockedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		lockedAt = &formatted
	}

	var lockExpiresAt *string
	if candidate.LockExpiresAt != nil {
		formatted := candidate.LockExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		lockExpiresAt = &formatted
	}

	responseDocuments := make([]CandidateDocumentResponse, 0, len(documents)+len(candidate.Documents))
	for _, document := range documents {
		responseDocuments = append(responseDocuments, mapDocumentResponse(document))
	}
	if len(responseDocuments) == 0 {
		for _, document := range candidate.Documents {
			responseDocuments = append(responseDocuments, CandidateDocumentResponse{
				ID:           document.ID,
				DocumentType: document.DocumentType,
				FileURL:      document.FileURL,
				FileName:     document.FileName,
				FileSize:     dereferenceInt64(document.FileSize),
				UploadedAt:   document.UploadedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	return CandidateResponse{
		ID:              candidate.ID,
		CreatedBy:       CandidateCreatedByInfo{ID: candidate.CreatedBy},
		FullName:        candidate.FullName,
		Nationality:     candidate.Nationality,
		DateOfBirth:     formatOptionalDate(candidate.DateOfBirth),
		Age:             candidate.Age,
		PlaceOfBirth:    candidate.PlaceOfBirth,
		Religion:        candidate.Religion,
		MaritalStatus:   candidate.MaritalStatus,
		ChildrenCount:   candidate.ChildrenCount,
		EducationLevel:  candidate.EducationLevel,
		ExperienceYears: candidate.ExperienceYears,
		Languages:       decodeStringSlice(candidate.Languages),
		Skills:          decodeStringSlice(candidate.Skills),
		Status:          string(candidate.Status),
		LockedBy:        candidate.LockedBy,
		LockedAt:        lockedAt,
		LockExpiresAt:   lockExpiresAt,
		CVPDFURL:        candidate.CVPDFURL,
		Documents:       responseDocuments,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

func mapDocumentResponse(document *domain.Document) CandidateDocumentResponse {
	return CandidateDocumentResponse{
		ID:           document.ID,
		DocumentType: string(document.DocumentType),
		FileURL:      document.FileURL,
		FileName:     document.FileName,
		FileSize:     document.FileSize,
		UploadedAt:   document.UploadedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func decodeStringSlice(value json.RawMessage) []string {
	if len(value) == 0 {
		return []string{}
	}

	var parsed []string
	if err := json.Unmarshal(value, &parsed); err != nil {
		return []string{}
	}

	return parsed
}

func parseOptionalDateString(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}
	normalized := parsed.UTC()
	return &normalized, nil
}

func formatOptionalDate(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}
	formatted := value.UTC().Format("2006-01-02")
	return &formatted
}

func applyRoleScopedCandidateFilters(role, userID string, filters domain.CandidateFilters) (domain.CandidateFilters, error) {
	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		filters.CreatedBy = strings.TrimSpace(userID)
	case string(domain.ForeignAgent):
		// Pairing scope is applied later once the active workspace is resolved.
	default:
		return domain.CandidateFilters{}, service.ErrForbidden
	}

	return filters, nil
}

func (h *CandidateHandler) canViewCandidate(candidate *domain.Candidate, userID, role, pairingID string) (bool, error) {
	if candidate == nil {
		return false, nil
	}

	if h.pairingService != nil {
		return h.pairingService.CanUserAccessCandidate(candidate, userID, role, pairingID)
	}

	trimmedUserID := strings.TrimSpace(userID)
	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		return strings.TrimSpace(candidate.CreatedBy) == trimmedUserID, nil
	default:
		return false, nil
	}
}

func (h *CandidateHandler) tryWritePairingServiceError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, service.ErrPairingRequired):
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "select a partner workspace to continue"})
		return true
	case errors.Is(err, service.ErrNoActivePairings):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "no active partner workspaces"})
		return true
	case errors.Is(err, service.ErrPairingAccessDenied):
		_ = utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return true
	default:
		return false
	}
}

func (h *CandidateHandler) writePairingAccessError(w http.ResponseWriter, err error) {
	if h.tryWritePairingServiceError(w, err) {
		return
	}
	_ = utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func sanitizeCandidateForViewer(candidate *domain.Candidate, userID, role string) {
	if candidate == nil {
		return
	}
	if strings.TrimSpace(role) == string(domain.ForeignAgent) {
		candidate.CreatedBy = ""
		if candidate.LockedBy != nil && strings.TrimSpace(*candidate.LockedBy) != strings.TrimSpace(userID) {
			candidate.LockedBy = nil
		}
	}
}

func dereferenceInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func mapPassportDataResponse(passportData *domain.PassportData) PassportDataResponse {
	var issueDate *string
	if passportData.IssueDate != nil && !passportData.IssueDate.IsZero() {
		formatted := passportData.IssueDate.UTC().Format("2006-01-02T15:04:05Z07:00")
		issueDate = &formatted
	}

	return PassportDataResponse{
		ID:               passportData.ID,
		CandidateID:      passportData.CandidateID,
		HolderName:       passportData.HolderName,
		PassportNumber:   passportData.PassportNumber,
		CountryCode:      passportData.CountryCode,
		Nationality:      passportData.Nationality,
		DateOfBirth:      passportData.DateOfBirth.UTC().Format("2006-01-02T15:04:05Z07:00"),
		PlaceOfBirth:     passportData.PlaceOfBirth,
		Gender:           passportData.Gender,
		IssueDate:        issueDate,
		ExpiryDate:       passportData.ExpiryDate.UTC().Format("2006-01-02T15:04:05Z07:00"),
		IssuingAuthority: passportData.IssuingAuthority,
		MRZLine1:         passportData.MRZLine1,
		MRZLine2:         passportData.MRZLine2,
		Confidence:       passportData.Confidence,
		ExtractedAt:      passportData.ExtractedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
