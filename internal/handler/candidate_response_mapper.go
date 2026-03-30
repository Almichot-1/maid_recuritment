package handler

import (
	"encoding/json"
	"time"

	"maid-recruitment-tracking/internal/domain"
)

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
		responseDocuments = append(responseDocuments, CandidateDocumentResponse{
			ID:           document.ID,
			DocumentType: string(document.DocumentType),
			FileURL:      document.FileURL,
			FileName:     document.FileName,
			FileSize:     document.FileSize,
			UploadedAt:   document.UploadedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
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

func formatOptionalDate(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}
	formatted := value.UTC().Format("2006-01-02")
	return &formatted
}
