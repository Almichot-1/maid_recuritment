package service

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/repository"
)

func TestAuthHelperFunctions(t *testing.T) {
	require.NoError(t, validateAuthEmail("a@b.com"))
	require.ErrorIs(t, validateAuthEmail("bad"), ErrInvalidCredentials)
	require.ErrorIs(t, validateAuthEmail(""), ErrInvalidCredentials)

	require.NoError(t, validateAuthPassword("12345678"))
	require.ErrorIs(t, validateAuthPassword("123"), ErrInvalidCredentials)

	role, err := parseUserRole(string(domain.EthiopianAgent))
	require.NoError(t, err)
	assert.Equal(t, domain.EthiopianAgent, role)
	_, err = parseUserRole("bad")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestCandidateHelpersAndUpdateSuccess(t *testing.T) {
	require.ErrorIs(t, validateCandidateInput(CandidateInput{FullName: ""}), ErrInvalidCandidateInput)
	age := 10
	require.ErrorIs(t, validateCandidateInput(CandidateInput{FullName: "A", Age: &age}), ErrInvalidCandidateInput)
	exp := 31
	require.ErrorIs(t, validateCandidateInput(CandidateInput{FullName: "A", ExperienceYears: &exp}), ErrInvalidCandidateInput)

	assert.False(t, isLockedByAnotherUser(nil, "u1"))
	owner := "u1"
	candidate := &domain.Candidate{LockedBy: &owner}
	assert.False(t, isLockedByAnotherUser(candidate, "u1"))
	assert.True(t, isLockedByAnotherUser(candidate, "u2"))

	items, err := marshalStringSlice([]string{" a ", "", "b"})
	require.NoError(t, err)
	assert.Equal(t, `["a","b"]`, string(items))

	repo := &candidateRepoBehaviorMock{}
	repo.getByID = func(id string) (*domain.Candidate, error) {
		return &domain.Candidate{ID: id, CreatedBy: "owner", Status: domain.CandidateStatusDraft, Languages: []byte("[]"), Skills: []byte("[]")}, nil
	}
	updated := false
	repo.updateFn = func(candidate *domain.Candidate) error {
		updated = true
		return nil
	}
	svc, err := NewCandidateService(repo, &documentRepositoryMock{}, &storageServiceMock{}, &PDFService{})
	require.NoError(t, err)
	require.NoError(t, svc.UpdateCandidate("id", "owner", CandidateInput{FullName: "Updated", Languages: []string{"en"}, Skills: []string{"cook"}}))
	assert.True(t, updated)
}

func TestApprovalRejectEdgeBranches(t *testing.T) {
	svc, db := setupApprovalService(t)
	seedApprovalScenario(t, db, domain.SelectionPending, time.Now().UTC().Add(2*time.Hour))

	err := svc.RejectSelection("sel-1", "other-user", "")
	require.ErrorIs(t, err, ErrNotAuthorized)

	require.NoError(t, db.Create(&domain.Approval{ID: "apr", SelectionID: "sel-1", UserID: "owner-1", Decision: domain.ApprovalApproved, DecidedAt: time.Now().UTC()}).Error)
	err = svc.RejectSelection("sel-1", "owner-1", "")
	require.ErrorIs(t, err, ErrAlreadyDecided)
}

func TestNotificationConstructorsAndEdges(t *testing.T) {
	_, err := NewNotificationService(&config.Config{}, nil, &notificationEmailMock{}, &notificationUserRepoMock{}, &notificationCandidateRepoMock{}, &notificationSelectionRepoMock{})
	require.Error(t, err)
	_, err = NewNotificationService(&config.Config{}, &notificationRepoMock{}, nil, &notificationUserRepoMock{}, &notificationCandidateRepoMock{}, &notificationSelectionRepoMock{})
	require.Error(t, err)
	_, err = NewNotificationService(&config.Config{}, &notificationRepoMock{}, &notificationEmailMock{}, nil, &notificationCandidateRepoMock{}, &notificationSelectionRepoMock{})
	require.Error(t, err)
	_, err = NewNotificationService(&config.Config{}, &notificationRepoMock{}, &notificationEmailMock{}, &notificationUserRepoMock{}, nil, &notificationSelectionRepoMock{})
	require.Error(t, err)
	_, err = NewNotificationService(&config.Config{}, &notificationRepoMock{}, &notificationEmailMock{}, &notificationUserRepoMock{}, &notificationCandidateRepoMock{}, nil)
	require.Error(t, err)

	userRepo := &notificationUserRepoMock{users: map[string]*domain.User{"owner": {ID: "owner", FullName: "Owner"}, "selector": {ID: "selector", FullName: "Selector"}}}
	candidateRepo := &notificationCandidateRepoMock{byID: map[string]*domain.Candidate{"cand": {ID: "cand", CreatedBy: "owner", FullName: "Cand"}}}
	selectionRepo := &notificationSelectionRepoMock{byID: map[string]*domain.Selection{"sel": {ID: "sel", CandidateID: "cand", SelectedBy: "selector"}}}
	repo := &notificationRepoErrMock{}

	svc, err := NewNotificationService(&config.Config{}, repo, &notificationEmailMock{}, userRepo, candidateRepo, selectionRepo)
	require.NoError(t, err)

	require.NoError(t, svc.NotifyRejection("sel", "selector"))
	require.NoError(t, svc.NotifyExpiry("sel"))
	assert.True(t, len(repo.created) >= 4)
}

func TestStatusStepUpdate_ErrorMapping(t *testing.T) {
	service := &StatusStepService{
		statusStepRepository: &statusStepRepoBehaviorMock{
			getByCandidateIDFn: func(candidateID string) ([]*domain.StatusStep, error) {
				return []*domain.StatusStep{{ID: "s1", CandidateID: candidateID, StepName: domain.MedicalTest, StepStatus: domain.Pending}}, nil
			},
			updateFn: func(step *domain.StatusStep) error { return repository.ErrStatusStepNotFound },
		},
		candidateRepository: &statusStepCandidateRepoMock{getByIDFn: func(id string) (*domain.Candidate, error) {
			return &domain.Candidate{ID: id, CreatedBy: "owner"}, nil
		}},
		selectionRepository: &statusStepSelectionRepoMock{},
		notificationService: &notificationSenderMock{foreignByID: map[string]bool{}},
	}

	err := service.UpdateStep("cand", domain.MedicalTest, "owner", domain.InProgress, "")
	require.ErrorIs(t, err, ErrStepNotFound)
}

func TestSMTPEmailService_SendSMTPError(t *testing.T) {
	service, err := NewSMTPEmailService(&config.Config{SMTPHost: "127.0.0.1", SMTPPort: "1", SMTPUser: "u", SMTPPass: "p"})
	require.NoError(t, err)
	err = service.Send("to@example.com", "s", "b")
	require.Error(t, err)
}

func TestNewS3StorageService_SuccessPath(t *testing.T) {
	service, err := NewS3StorageService(&config.Config{S3Bucket: "bucket", AWSRegion: "us-east-1", AWSAccessKey: "key", AWSSecretKey: "secret", S3Endpoint: "http://localhost:9000"})
	require.NoError(t, err)
	require.NotNil(t, service)
}

func TestSelectionService_ConstructorValidationMore(t *testing.T) {
	_, err := NewSelectionService(&selectionRepositoryMock{}, &candidateRepositoryMock{}, nil)
	require.Error(t, err)
}

func TestApprovalService_ConstructorValidationMore(t *testing.T) {
	_, err := NewApprovalService(&approvalRepositoryMock{}, nil, &candidateRepositoryMock{}, &StatusStepService{}, &notificationSenderMock{foreignByID: map[string]bool{}})
	require.Error(t, err)
	_, err = NewApprovalService(&approvalRepositoryMock{}, &selectionRepositoryMock{}, nil, &StatusStepService{}, &notificationSenderMock{foreignByID: map[string]bool{}})
	require.Error(t, err)
	_, err = NewApprovalService(&approvalRepositoryMock{}, &selectionRepositoryMock{}, &candidateRepositoryMock{}, nil, &notificationSenderMock{foreignByID: map[string]bool{}})
	require.Error(t, err)
	_, err = NewApprovalService(&approvalRepositoryMock{}, &selectionRepositoryMock{}, &candidateRepositoryMock{}, &StatusStepService{}, nil)
	require.Error(t, err)
}

func TestDocumentService_StorageUploadError(t *testing.T) {
	svc, err := NewDocumentService(&documentRepoMock{}, &documentStorageMock{uploadFn: func(fileName, contentType string) (string, error) {
		return "", errors.New("upload failed")
	}})
	require.NoError(t, err)
	_, err = svc.UploadDocument("cand", string(domain.Passport), bytes.NewReader(validPDFBytes()), "passport.pdf", int64(len(validPDFBytes())))
	require.Error(t, err)
}
