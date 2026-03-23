package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrPairingNotFound            = errors.New("pairing not found")
	ErrPairingAlreadyExists       = errors.New("pairing already exists")
	ErrPairingRequired            = errors.New("pairing is required")
	ErrNoActivePairings           = errors.New("no active pairings")
	ErrPairingAccessDenied        = errors.New("pairing access denied")
	ErrPairingNotActive           = errors.New("pairing is not active")
	ErrInvalidPairingParticipants = errors.New("invalid pairing participants")
	ErrCandidateAlreadyShared     = errors.New("candidate already shared to pairing")
	ErrCandidateShareNotFound     = errors.New("candidate share not found")
)

type PairingService struct {
	userRepository      domain.UserRepository
	pairingRepository   domain.AgencyPairingRepository
	shareRepository     domain.CandidatePairShareRepository
	selectionRepository domain.SelectionRepository
	auditRepository     domain.AuditLogRepository
}

func NewPairingService(
	userRepository domain.UserRepository,
	pairingRepository domain.AgencyPairingRepository,
	shareRepository domain.CandidatePairShareRepository,
	selectionRepository domain.SelectionRepository,
	auditRepository domain.AuditLogRepository,
) (*PairingService, error) {
	if userRepository == nil {
		return nil, fmt.Errorf("user repository is nil")
	}
	if pairingRepository == nil {
		return nil, fmt.Errorf("pairing repository is nil")
	}
	if shareRepository == nil {
		return nil, fmt.Errorf("share repository is nil")
	}
	if selectionRepository == nil {
		return nil, fmt.Errorf("selection repository is nil")
	}
	if auditRepository == nil {
		return nil, fmt.Errorf("audit repository is nil")
	}

	return &PairingService{
		userRepository:      userRepository,
		pairingRepository:   pairingRepository,
		shareRepository:     shareRepository,
		selectionRepository: selectionRepository,
		auditRepository:     auditRepository,
	}, nil
}

func (s *PairingService) ListAdminPairings(filters domain.AgencyPairingFilters) ([]*domain.AgencyPairing, error) {
	return s.pairingRepository.List(filters)
}

func (s *PairingService) ListAgencyPairings(agencyID string) ([]*domain.AgencyPairing, error) {
	return s.pairingRepository.List(domain.AgencyPairingFilters{UserID: strings.TrimSpace(agencyID)})
}

func (s *PairingService) ListActivePairingsForUser(userID string) ([]*domain.AgencyPairing, error) {
	status := domain.AgencyPairingActive
	return s.pairingRepository.List(domain.AgencyPairingFilters{
		UserID: strings.TrimSpace(userID),
		Status: &status,
	})
}

func (s *PairingService) ResolveActivePairing(userID, role, pairingID string) (*domain.AgencyPairing, error) {
	pairings, err := s.ListActivePairingsForUser(userID)
	if err != nil {
		return nil, err
	}
	if len(pairings) == 0 {
		return nil, ErrNoActivePairings
	}

	trimmedPairingID := strings.TrimSpace(pairingID)
	if trimmedPairingID == "" {
		if len(pairings) == 1 {
			return pairings[0], nil
		}
		return nil, ErrPairingRequired
	}

	for _, pairing := range pairings {
		if pairing == nil {
			continue
		}
		if strings.TrimSpace(pairing.ID) != trimmedPairingID {
			continue
		}
		if !pairingIncludesUser(pairing, strings.TrimSpace(userID), strings.TrimSpace(role)) {
			return nil, ErrPairingAccessDenied
		}
		return pairing, nil
	}

	return nil, ErrPairingAccessDenied
}

func (s *PairingService) CreatePairing(adminID, ethiopianUserID, foreignUserID, notes, ipAddress string) (*domain.AgencyPairing, error) {
	ethiopianUser, err := s.userRepository.GetByID(strings.TrimSpace(ethiopianUserID))
	if err != nil {
		return nil, err
	}
	foreignUser, err := s.userRepository.GetByID(strings.TrimSpace(foreignUserID))
	if err != nil {
		return nil, err
	}
	if ethiopianUser.Role != domain.EthiopianAgent || foreignUser.Role != domain.ForeignAgent {
		return nil, ErrInvalidPairingParticipants
	}
	if ethiopianUser.AccountStatus != domain.AccountStatusActive || foreignUser.AccountStatus != domain.AccountStatusActive {
		return nil, ErrInvalidPairingParticipants
	}

	now := time.Now().UTC()
	adminID = strings.TrimSpace(adminID)
	notes = strings.TrimSpace(notes)
	pairing := &domain.AgencyPairing{
		EthiopianUserID:   ethiopianUser.ID,
		ForeignUserID:     foreignUser.ID,
		Status:            domain.AgencyPairingActive,
		ApprovedByAdminID: optionalString(adminID),
		ApprovedAt:        &now,
		Notes:             optionalString(notes),
	}

	if err := s.pairingRepository.Create(pairing); err != nil {
		if errors.Is(err, repository.ErrDuplicateAgencyPairing) {
			return nil, ErrPairingAlreadyExists
		}
		return nil, err
	}

	if err := s.createAudit(adminID, "create_pairing", "pairing", pairing.ID, map[string]any{
		"ethiopian_user_id": pairing.EthiopianUserID,
		"foreign_user_id":   pairing.ForeignUserID,
		"status":            pairing.Status,
		"notes":             notes,
	}, ipAddress); err != nil {
		return nil, err
	}

	return pairing, nil
}

func (s *PairingService) UpdatePairingStatus(adminID, pairingID string, status domain.AgencyPairingStatus, notes, ipAddress string) (*domain.AgencyPairing, error) {
	pairing, err := s.pairingRepository.GetByID(strings.TrimSpace(pairingID))
	if err != nil {
		if errors.Is(err, repository.ErrAgencyPairingNotFound) {
			return nil, ErrPairingNotFound
		}
		return nil, err
	}

	now := time.Now().UTC()
	pairing.Status = status
	trimmedNotes := strings.TrimSpace(notes)
	pairing.Notes = optionalString(trimmedNotes)
	if status == domain.AgencyPairingEnded {
		pairing.EndedAt = &now
	} else {
		pairing.EndedAt = nil
	}

	if err := s.pairingRepository.Update(pairing); err != nil {
		if errors.Is(err, repository.ErrAgencyPairingNotFound) {
			return nil, ErrPairingNotFound
		}
		if errors.Is(err, repository.ErrDuplicateAgencyPairing) {
			return nil, ErrPairingAlreadyExists
		}
		return nil, err
	}

	if err := s.createAudit(strings.TrimSpace(adminID), "update_pairing_status", "pairing", pairing.ID, map[string]any{
		"status": status,
		"notes":  trimmedNotes,
	}, ipAddress); err != nil {
		return nil, err
	}

	return pairing, nil
}

func (s *PairingService) ShareCandidate(candidateID, pairingID, sharedBy string) error {
	pairing, err := s.pairingRepository.GetByID(strings.TrimSpace(pairingID))
	if err != nil {
		if errors.Is(err, repository.ErrAgencyPairingNotFound) {
			return ErrPairingNotFound
		}
		return err
	}
	if pairing.Status != domain.AgencyPairingActive {
		return ErrPairingNotActive
	}
	if strings.TrimSpace(sharedBy) != strings.TrimSpace(pairing.EthiopianUserID) {
		return ErrPairingAccessDenied
	}

	share := &domain.CandidatePairShare{
		PairingID:      pairing.ID,
		CandidateID:    strings.TrimSpace(candidateID),
		SharedByUserID: strings.TrimSpace(sharedBy),
		IsActive:       true,
		SharedAt:       time.Now().UTC(),
	}

	if err := s.shareRepository.Create(share); err != nil {
		if errors.Is(err, repository.ErrDuplicateCandidatePairShare) {
			return ErrCandidateAlreadyShared
		}
		return err
	}

	return nil
}

func (s *PairingService) UnshareCandidate(candidateID, pairingID, sharedBy string) error {
	pairing, err := s.pairingRepository.GetByID(strings.TrimSpace(pairingID))
	if err != nil {
		if errors.Is(err, repository.ErrAgencyPairingNotFound) {
			return ErrPairingNotFound
		}
		return err
	}
	if strings.TrimSpace(sharedBy) != strings.TrimSpace(pairing.EthiopianUserID) {
		return ErrPairingAccessDenied
	}

	if err := s.shareRepository.Deactivate(strings.TrimSpace(pairingID), strings.TrimSpace(candidateID), time.Now().UTC()); err != nil {
		if errors.Is(err, repository.ErrCandidatePairShareNotFound) {
			return ErrCandidateShareNotFound
		}
		return err
	}

	return nil
}

func (s *PairingService) IsCandidateSharedWithPairing(candidateID, pairingID string) (bool, error) {
	_, err := s.shareRepository.GetActiveByPairingAndCandidate(strings.TrimSpace(pairingID), strings.TrimSpace(candidateID))
	if err != nil {
		if errors.Is(err, repository.ErrCandidatePairShareNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *PairingService) ListCandidateShares(candidateID string) ([]*domain.CandidatePairShare, error) {
	return s.shareRepository.ListByCandidateID(strings.TrimSpace(candidateID), true)
}

func (s *PairingService) CanUserAccessCandidate(candidate *domain.Candidate, userID, role, pairingID string) (bool, error) {
	if candidate == nil {
		return false, nil
	}

	trimmedUserID := strings.TrimSpace(userID)
	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		return strings.TrimSpace(candidate.CreatedBy) == trimmedUserID, nil
	case string(domain.ForeignAgent):
		pairing, err := s.ResolveActivePairing(trimmedUserID, role, pairingID)
		if err != nil {
			return false, err
		}
		if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(pairing.EthiopianUserID) {
			return false, nil
		}
		shared, err := s.IsCandidateSharedWithPairing(candidate.ID, pairing.ID)
		if err != nil {
			return false, err
		}
		if shared {
			return true, nil
		}

		selection, err := s.selectionRepository.GetByCandidateIDAndPairingID(candidate.ID, pairing.ID)
		if err != nil {
			if errors.Is(err, repository.ErrSelectionNotFound) {
				return false, nil
			}
			return false, err
		}
		return strings.TrimSpace(selection.SelectedBy) == trimmedUserID, nil
	default:
		return false, nil
	}
}

func (s *PairingService) createAudit(adminID, action, targetType, targetID string, details map[string]any, ipAddress string) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	targetID = strings.TrimSpace(targetID)
	return s.auditRepository.Create(&domain.AuditLog{
		AdminID:    strings.TrimSpace(adminID),
		Action:     strings.TrimSpace(action),
		TargetType: strings.TrimSpace(targetType),
		TargetID:   &targetID,
		Details:    payload,
		IPAddress:  strings.TrimSpace(ipAddress),
	})
}

func pairingIncludesUser(pairing *domain.AgencyPairing, userID, role string) bool {
	if pairing == nil {
		return false
	}

	switch strings.TrimSpace(role) {
	case string(domain.EthiopianAgent):
		return strings.TrimSpace(pairing.EthiopianUserID) == strings.TrimSpace(userID)
	case string(domain.ForeignAgent):
		return strings.TrimSpace(pairing.ForeignUserID) == strings.TrimSpace(userID)
	default:
		return false
	}
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	return &trimmed
}
