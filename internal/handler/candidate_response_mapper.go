package handler

import (
	"encoding/json"
	"strings"
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
		PlaceOfBirth:        candidate.PlaceOfBirth,
		PassportNumber:      candidate.PassportNumber,
		IssueDate:           formatOptionalDate(candidate.IssueDate),
		ExpiryDate:          formatOptionalDate(candidate.ExpiryDate),
		Gender:              candidate.Gender,
		IssuingAuthority:    candidate.IssuingAuthority,
		ExperienceAbroad:    decodeExperienceAbroad(candidate.ExperienceAbroad),
		Religion:            candidate.Religion,
		MaritalStatus:   candidate.MaritalStatus,
		ChildrenCount:       candidate.ChildrenCount,
		EducationLevel:      candidate.EducationLevel,
		ExperienceYears:     candidate.ExperienceYears,
		CountryOfExperience: candidate.CountryOfExperience,
		Languages:           decodeLanguages(candidate.Languages),
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

func decodeLanguages(value json.RawMessage) json.RawMessage {
	if len(value) == 0 || string(value) == "null" {
		entries := []domain.LanguageEntry{}
		result, _ := json.Marshal(entries)
		return result
	}

	var entries []domain.LanguageEntry
	if err := json.Unmarshal(value, &entries); err == nil {
		result, _ := json.Marshal(entries)
		return result
	}

	var legacy []string
	if err := json.Unmarshal(value, &legacy); err != nil {
		return json.RawMessage("[]")
	}
	entries = make([]domain.LanguageEntry, 0, len(legacy))
	for _, s := range legacy {
		lang := strings.TrimSpace(s)
		if lang == "" {
			continue
		}
		entries = append(entries, domain.LanguageEntry{Language: lang, Proficiency: "Basic"})
	}
	result, _ := json.Marshal(entries)
	return result
}

func decodeExperienceAbroad(value json.RawMessage) json.RawMessage {
	if len(value) == 0 || string(value) == "null" {
		return json.RawMessage("[]")
	}

	var entries []domain.ExperienceEntry
	if err := json.Unmarshal(value, &entries); err == nil {
		normalized := make([]domain.ExperienceEntry, 0, len(entries))
		for _, e := range entries {
			country := strings.TrimSpace(e.Country)
			if country == "" {
				continue
			}
			years := e.Years
			if years < 0 {
				years = 0
			}
			normalized = append(normalized, domain.ExperienceEntry{Country: country, Years: years})
		}
		data, _ := json.Marshal(normalized)
		return data
	}

	var legacy string
	if err := json.Unmarshal(value, &legacy); err != nil {
		return json.RawMessage("[]")
	}
	legacy = strings.TrimSpace(legacy)
	if legacy == "" {
		return json.RawMessage("[]")
	}
	entries = []domain.ExperienceEntry{{Country: legacy, Years: 0}}
	data, _ := json.Marshal(entries)
	return data
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
