package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"maid-recruitment-tracking/internal/config"
	"maid-recruitment-tracking/internal/domain"
	"maid-recruitment-tracking/internal/ocr"
	"maid-recruitment-tracking/internal/repository"
)

var (
	ErrPassportOCRRequiresImage = errors.New("passport OCR requires a JPG or PNG image")
	ErrPassportDataNotFound     = errors.New("passport data not found")
	ErrPassportOCRParseFailed   = errors.New("passport OCR could not parse the image")
	ErrPassportOCRUnavailable   = errors.New("passport OCR is unavailable")
)

type PassportOCRService struct {
	candidateRepository domain.CandidateRepository
	passportRepository  domain.PassportDataRepository
	ocrProcessor        *ocr.OCRProcessor
	previewCacheMu      sync.Mutex
	previewCache        map[string]cachedPassportPreview
	previewCacheOrder   []string
}

type cachedPassportPreview struct {
	data     *domain.PassportData
	cachedAt time.Time
}

type PassportPreviewMetrics struct {
	ReadDuration  time.Duration
	OCRDuration   time.Duration
	TotalDuration time.Duration
	CacheHit      bool
}

const (
	passportPreviewCacheTTL        = 15 * time.Minute
	passportPreviewCacheMaxEntries = 128
)

func NewPassportOCRService(
	cfg *config.Config,
	candidateRepository domain.CandidateRepository,
	passportRepository domain.PassportDataRepository,
) (*PassportOCRService, error) {
	if candidateRepository == nil {
		return nil, fmt.Errorf("candidate repository is nil")
	}
	if passportRepository == nil {
		return nil, fmt.Errorf("passport repository is nil")
	}

	tesseractPath := ""
	ocrLanguage := "eng"
	if cfg != nil {
		tesseractPath = strings.TrimSpace(cfg.TesseractPath)
		if strings.TrimSpace(cfg.OCRLanguage) != "" {
			ocrLanguage = strings.TrimSpace(cfg.OCRLanguage)
		}
	}

	return &PassportOCRService{
		candidateRepository: candidateRepository,
		passportRepository:  passportRepository,
		ocrProcessor:        ocr.NewOCRProcessor(tesseractPath, ocrLanguage),
		previewCache:        make(map[string]cachedPassportPreview),
		previewCacheOrder:   make([]string, 0, passportPreviewCacheMaxEntries),
	}, nil
}

func (s *PassportOCRService) ParseAndStore(candidateID, requestedBy string, file io.Reader, fileName string) (*domain.PassportData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(requestedBy) == "" {
		return nil, ErrForbidden
	}
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("file name is required")
	}

	contentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, err
	}
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, ErrPassportOCRRequiresImage
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(requestedBy) {
		return nil, ErrForbidden
	}

	buffer, err := io.ReadAll(io.LimitReader(file, maxDocumentFileSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read passport image: %w", err)
	}
	if int64(len(buffer)) > maxDocumentFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	cacheKey := fingerprintPassportFile(buffer)
	if cached := s.getCachedPreview(cacheKey); cached != nil {
		passportData := clonePassportData(cached)
		passportData.CandidateID = candidateID
		if err := s.passportRepository.Upsert(passportData); err != nil {
			return nil, err
		}
		return s.passportRepository.GetByCandidateID(candidateID)
	}

	tempFile, err := os.CreateTemp("", "passport-ocr-*"+strings.ToLower(filepath.Ext(fileName)))
	if err != nil {
		return nil, fmt.Errorf("create temp passport image: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, bytes.NewReader(buffer)); err != nil {
		return nil, fmt.Errorf("write temp passport image: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp passport image: %w", err)
	}

	result, err := s.ocrProcessor.ExtractPassportData(tempPath)
	if err != nil {
		if isPassportOCRUnavailable(err) {
			return nil, fmt.Errorf("%w: %v", ErrPassportOCRUnavailable, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrPassportOCRParseFailed, err)
	}
	if strings.TrimSpace(result.PlaceOfBirth) == "" || result.DateOfIssue.IsZero() {
		fallbackPlace, fallbackIssueDate := s.extractPreviewFallbackFields(
			tempPath,
			result.DateOfBirth,
			strings.TrimSpace(result.PlaceOfBirth) == "",
			result.DateOfIssue.IsZero(),
		)
		if strings.TrimSpace(result.PlaceOfBirth) == "" && fallbackPlace != "" {
			result.PlaceOfBirth = fallbackPlace
		}
		if result.DateOfIssue.IsZero() && !fallbackIssueDate.IsZero() {
			result.DateOfIssue = fallbackIssueDate
		}
	}

	holderName := strings.TrimSpace(strings.TrimSpace(result.GivenNames + " " + result.Surname))
	passportData := &domain.PassportData{
		CandidateID:      candidateID,
		HolderName:       holderName,
		PassportNumber:   strings.TrimSpace(result.PassportNumber),
		CountryCode:      strings.TrimSpace(result.CountryCode),
		Nationality:      strings.TrimSpace(result.Nationality),
		DateOfBirth:      result.DateOfBirth.UTC(),
		PlaceOfBirth:     strings.TrimSpace(result.PlaceOfBirth),
		Gender:           strings.TrimSpace(result.Sex),
		ExpiryDate:       result.DateOfExpiry.UTC(),
		IssuingAuthority: strings.TrimSpace(result.IssuingAuthority),
		MRZLine1:         strings.TrimSpace(result.MRZLine1),
		MRZLine2:         strings.TrimSpace(result.MRZLine2),
		Confidence:       result.Confidence,
		ExtractedAt:      time.Now().UTC(),
	}
	if !result.DateOfIssue.IsZero() {
		issueDate := result.DateOfIssue.UTC()
		passportData.IssueDate = &issueDate
	}
	s.cachePreview(cacheKey, passportData)

	if err := s.passportRepository.Upsert(passportData); err != nil {
		return nil, err
	}

	return s.passportRepository.GetByCandidateID(candidateID)
}

func (s *PassportOCRService) ParsePreview(file io.Reader, fileName string) (*domain.PassportData, error) {
	passportData, _, err := s.ParsePreviewWithMetrics(file, fileName)
	return passportData, err
}

func (s *PassportOCRService) ParsePreviewWithMetrics(file io.Reader, fileName string) (*domain.PassportData, PassportPreviewMetrics, error) {
	var metrics PassportPreviewMetrics
	startedAt := time.Now()

	if file == nil {
		return nil, metrics, fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, metrics, fmt.Errorf("file name is required")
	}

	contentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, metrics, err
	}
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, metrics, ErrPassportOCRRequiresImage
	}

	readStartedAt := time.Now()
	buffer, err := readPassportImageBuffer(file)
	metrics.ReadDuration = time.Since(readStartedAt)
	if err != nil {
		metrics.TotalDuration = time.Since(startedAt)
		return nil, metrics, err
	}

	passportData, cacheHit, ocrDuration, err := s.parsePreviewBuffer(buffer, fileName)
	metrics.CacheHit = cacheHit
	metrics.OCRDuration = ocrDuration
	metrics.TotalDuration = time.Since(startedAt)
	if err != nil {
		return nil, metrics, err
	}

	return passportData, metrics, nil
}

func (s *PassportOCRService) parsePreviewBuffer(buffer []byte, fileName string) (*domain.PassportData, bool, time.Duration, error) {
	cacheKey := fingerprintPassportFile(buffer)
	if cached := s.getCachedPreview(cacheKey); cached != nil {
		return clonePassportData(cached), true, 0, nil
	}

	tempFile, err := os.CreateTemp("", "passport-ocr-preview-*"+strings.ToLower(filepath.Ext(fileName)))
	if err != nil {
		return nil, false, 0, fmt.Errorf("create temp passport image: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, bytes.NewReader(buffer)); err != nil {
		return nil, false, 0, fmt.Errorf("write temp passport image: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, false, 0, fmt.Errorf("close temp passport image: %w", err)
	}

	ocrStartedAt := time.Now()
	result, err := s.ocrProcessor.ExtractPassportPreviewData(tempPath)
	if err != nil {
		if isPassportOCRUnavailable(err) {
			return nil, false, time.Since(ocrStartedAt), fmt.Errorf("%w: %v", ErrPassportOCRUnavailable, err)
		}
		return nil, false, time.Since(ocrStartedAt), fmt.Errorf("%w: %v", ErrPassportOCRParseFailed, err)
	}
	if strings.TrimSpace(result.PlaceOfBirth) == "" || result.DateOfIssue.IsZero() {
		fallbackPlace, fallbackIssueDate := s.extractPreviewFallbackFields(
			tempPath,
			result.DateOfBirth,
			strings.TrimSpace(result.PlaceOfBirth) == "",
			result.DateOfIssue.IsZero(),
		)
		if strings.TrimSpace(result.PlaceOfBirth) == "" && fallbackPlace != "" {
			result.PlaceOfBirth = fallbackPlace
		}
		if result.DateOfIssue.IsZero() && !fallbackIssueDate.IsZero() {
			result.DateOfIssue = fallbackIssueDate
		}
	}

	holderName := strings.TrimSpace(strings.TrimSpace(result.GivenNames + " " + result.Surname))
	passportData := &domain.PassportData{
		HolderName:       holderName,
		PassportNumber:   strings.TrimSpace(result.PassportNumber),
		CountryCode:      strings.TrimSpace(result.CountryCode),
		Nationality:      strings.TrimSpace(result.Nationality),
		DateOfBirth:      result.DateOfBirth.UTC(),
		PlaceOfBirth:     strings.TrimSpace(result.PlaceOfBirth),
		Gender:           strings.TrimSpace(result.Sex),
		ExpiryDate:       result.DateOfExpiry.UTC(),
		IssuingAuthority: strings.TrimSpace(result.IssuingAuthority),
		MRZLine1:         strings.TrimSpace(result.MRZLine1),
		MRZLine2:         strings.TrimSpace(result.MRZLine2),
		Confidence:       result.Confidence,
		ExtractedAt:      time.Now().UTC(),
	}
	if !result.DateOfIssue.IsZero() {
		issueDate := result.DateOfIssue.UTC()
		passportData.IssueDate = &issueDate
	}
	s.cachePreview(cacheKey, passportData)

	return passportData, false, time.Since(ocrStartedAt), nil
}

func (s *PassportOCRService) StoreCachedPreview(candidateID, requestedBy, fileName string, buffer []byte) (*domain.PassportData, bool, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, false, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(requestedBy) == "" {
		return nil, false, ErrForbidden
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, false, fmt.Errorf("file name is required")
	}
	if len(buffer) == 0 {
		return nil, false, fmt.Errorf("file is required")
	}

	contentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, false, err
	}
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, false, ErrPassportOCRRequiresImage
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(requestedBy) {
		return nil, false, ErrForbidden
	}

	cached := s.getCachedPreview(fingerprintPassportFile(buffer))
	if cached == nil {
		return nil, false, nil
	}

	cached.CandidateID = candidateID
	if err := s.passportRepository.Upsert(cached); err != nil {
		return nil, false, err
	}

	return clonePassportData(cached), true, nil
}

func (s *PassportOCRService) GetByCandidateID(candidateID, requestedBy string) (*domain.PassportData, error) {
	if strings.TrimSpace(candidateID) == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	if strings.TrimSpace(requestedBy) == "" {
		return nil, ErrForbidden
	}

	candidate, err := s.candidateRepository.GetByID(candidateID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(candidate.CreatedBy) != strings.TrimSpace(requestedBy) {
		return nil, ErrForbidden
	}

	data, err := s.passportRepository.GetByCandidateID(candidateID)
	if err != nil {
		if errors.Is(err, repository.ErrPassportDataNotFound) {
			return nil, ErrPassportDataNotFound
		}
		return nil, err
	}
	return data, nil
}

func isPassportOCRUnavailable(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "executable file not found"):
		return true
	case strings.Contains(message, "system cannot find the file specified"):
		return true
	case strings.Contains(message, "tesseract is not recognized"):
		return true
	default:
		return false
	}
}

func fingerprintPassportFile(buffer []byte) string {
	sum := sha256.Sum256(buffer)
	return hex.EncodeToString(sum[:])
}

func (s *PassportOCRService) getCachedPreview(key string) *domain.PassportData {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	now := time.Now().UTC()
	s.previewCacheMu.Lock()
	defer s.previewCacheMu.Unlock()

	s.cleanupExpiredPreviewCacheLocked(now)

	cached, ok := s.previewCache[key]
	if !ok || cached.data == nil {
		s.removePreviewCacheKeyLocked(key)
		return nil
	}
	cached.cachedAt = now
	s.previewCache[key] = cached
	s.touchPreviewCacheKeyLocked(key)
	return clonePassportData(cached.data)
}

func (s *PassportOCRService) cachePreview(key string, data *domain.PassportData) {
	if strings.TrimSpace(key) == "" || data == nil {
		return
	}
	now := time.Now().UTC()

	s.previewCacheMu.Lock()
	defer s.previewCacheMu.Unlock()

	s.cleanupExpiredPreviewCacheLocked(now)
	s.previewCache[key] = cachedPassportPreview{
		data:     clonePassportData(data),
		cachedAt: now,
	}
	s.touchPreviewCacheKeyLocked(key)

	for len(s.previewCacheOrder) > passportPreviewCacheMaxEntries {
		s.removePreviewCacheKeyLocked(s.previewCacheOrder[0])
	}
}

func (s *PassportOCRService) cleanupExpiredPreviewCacheLocked(now time.Time) {
	for _, key := range append([]string(nil), s.previewCacheOrder...) {
		cached, ok := s.previewCache[key]
		if !ok {
			s.removePreviewCacheKeyLocked(key)
			continue
		}
		if now.Sub(cached.cachedAt) > passportPreviewCacheTTL {
			s.removePreviewCacheKeyLocked(key)
		}
	}
}

func (s *PassportOCRService) touchPreviewCacheKeyLocked(key string) {
	if strings.TrimSpace(key) == "" {
		return
	}
	for index, existingKey := range s.previewCacheOrder {
		if existingKey == key {
			s.previewCacheOrder = append(s.previewCacheOrder[:index], s.previewCacheOrder[index+1:]...)
			break
		}
	}
	s.previewCacheOrder = append(s.previewCacheOrder, key)
}

func (s *PassportOCRService) removePreviewCacheKeyLocked(key string) {
	delete(s.previewCache, key)
	for index, existingKey := range s.previewCacheOrder {
		if existingKey == key {
			s.previewCacheOrder = append(s.previewCacheOrder[:index], s.previewCacheOrder[index+1:]...)
			return
		}
	}
}

func readPassportImageBuffer(file io.Reader) ([]byte, error) {
	buffer, err := io.ReadAll(io.LimitReader(file, maxDocumentFileSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read passport image: %w", err)
	}
	if int64(len(buffer)) > maxDocumentFileSizeBytes {
		return nil, ErrFileTooLarge
	}
	return buffer, nil
}

func clonePassportData(data *domain.PassportData) *domain.PassportData {
	if data == nil {
		return nil
	}

	cloned := *data
	if data.IssueDate != nil {
		issueDate := data.IssueDate.UTC()
		cloned.IssueDate = &issueDate
	}
	return &cloned
}

func (s *PassportOCRService) extractPreviewFallbackFields(tempPath string, dateOfBirth time.Time, needPlaceOfBirth, needIssueDate bool) (string, time.Time) {
	if strings.TrimSpace(tempPath) == "" || (!needPlaceOfBirth && !needIssueDate) {
		return "", time.Time{}
	}

	resolveFromText := func(rawText string) (string, time.Time) {
		if strings.TrimSpace(rawText) == "" {
			return "", time.Time{}
		}

		placeOfBirth := ""
		issueDate := time.Time{}
		if needPlaceOfBirth && !dateOfBirth.IsZero() {
			placeOfBirth = extractPlaceOfBirthFromOCRText(rawText, dateOfBirth)
		}
		if needIssueDate {
			issueDate = extractIssueDateFromOCRText(rawText)
		}
		return placeOfBirth, issueDate
	}

	if rawText, err := s.ocrProcessor.ExtractFastText(tempPath); err == nil {
		placeOfBirth, issueDate := resolveFromText(rawText)
		if (!needPlaceOfBirth || placeOfBirth != "") && (!needIssueDate || !issueDate.IsZero()) {
			return placeOfBirth, issueDate
		}

		needPlaceOfBirth = needPlaceOfBirth && placeOfBirth == ""
		needIssueDate = needIssueDate && issueDate.IsZero()
		if !needPlaceOfBirth && !needIssueDate {
			return placeOfBirth, issueDate
		}

		if rawText, err := s.ocrProcessor.ExtractText(tempPath); err == nil {
			fullPlaceOfBirth, fullIssueDate := resolveFromText(rawText)
			if placeOfBirth == "" {
				placeOfBirth = fullPlaceOfBirth
			}
			if issueDate.IsZero() {
				issueDate = fullIssueDate
			}
		}
		return placeOfBirth, issueDate
	}

	if rawText, err := s.ocrProcessor.ExtractText(tempPath); err == nil {
		return resolveFromText(rawText)
	}

	return "", time.Time{}
}

func extractPlaceOfBirthFromOCRText(text string, dateOfBirth time.Time) string {
	if strings.TrimSpace(text) == "" || dateOfBirth.IsZero() {
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	month := strings.ToUpper(dateOfBirth.Format("Jan"))
	day := dateOfBirth.Day()
	datePattern := regexp.MustCompile(fmt.Sprintf(`(?i)\b%d\s*%s(?:['\s]*\d{1,2})?\b`, day, month))

	for index, rawLine := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(rawLine))
		if upperLine == "" {
			continue
		}

		if !datePattern.MatchString(upperLine) {
			continue
		}

		if candidate := cleanupPlaceOfBirthValue(datePattern.ReplaceAllString(rawLine, " ")); candidate != "" {
			return candidate
		}

		if index+1 < len(lines) {
			if candidate := cleanupPlaceOfBirthValue(lines[index+1]); candidate != "" {
				return candidate
			}
		}
	}

	return ""
}

func cleanupPlaceOfBirthValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	value = strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), unicode.IsSpace(r):
			return unicode.ToUpper(r)
		default:
			return ' '
		}
	}, value)

	parts := strings.Fields(value)
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "SEX", "F", "M", "MF", "FM", "PLACE", "BIRTH", "DATE", "OF", "DOB", "NATIONALITY":
			continue
		}
		if len(part) < 3 {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) == 0 {
		return ""
	}
	if len(filtered) > 3 {
		filtered = filtered[len(filtered)-3:]
	}
	return strings.Join(filtered, " ")
}

var (
	passportIssueISO      = regexp.MustCompile(`\b(\d{4})-(\d{2})-(\d{2})\b`)
	passportIssueSlash    = regexp.MustCompile(`\b(\d{2})[\./-](\d{2})[\./-](\d{4})\b`)
	passportIssueDMMMYY   = regexp.MustCompile(`\b(\d{1,2})\s*([A-Z]{3})\s*(\d{2})\b`)
	passportIssueDMMMYYYY = regexp.MustCompile(`\b(\d{1,2})\s*([A-Z]{3})\s*(\d{4})\b`)
)

func extractIssueDateFromOCRText(text string) time.Time {
	text = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(text), "\r", "\n"))
	if text == "" {
		return time.Time{}
	}

	candidates := make([]time.Time, 0, 8)
	for _, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		candidates = append(candidates, extractPassportIssueCandidates(line)...)
	}

	if len(candidates) == 0 {
		return time.Time{}
	}

	now := time.Now().UTC()
	cutoffPast := now.AddDate(-15, 0, 0)
	best := time.Time{}

	for _, candidate := range candidates {
		candidate = candidate.UTC()
		if candidate.IsZero() || candidate.After(now.AddDate(0, 0, 1)) || candidate.Before(cutoffPast) {
			continue
		}
		if best.IsZero() || candidate.After(best) {
			best = candidate
		}
	}

	return best
}

func extractPassportIssueCandidates(text string) []time.Time {
	candidates := make([]time.Time, 0, 4)

	for _, match := range passportIssueISO.FindAllStringSubmatch(text, -1) {
		year, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		if parsed := safePassportIssueDate(year, month, day); !parsed.IsZero() {
			candidates = append(candidates, parsed)
		}
	}

	for _, match := range passportIssueSlash.FindAllStringSubmatch(text, -1) {
		day, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		year, _ := strconv.Atoi(match[3])
		if parsed := safePassportIssueDate(year, month, day); !parsed.IsZero() {
			candidates = append(candidates, parsed)
		}
	}

	for _, match := range passportIssueDMMMYYYY.FindAllStringSubmatch(text, -1) {
		day, _ := strconv.Atoi(match[1])
		month := passportMonthFromMMM(match[2])
		year, _ := strconv.Atoi(match[3])
		if parsed := safePassportIssueDate(year, month, day); !parsed.IsZero() {
			candidates = append(candidates, parsed)
		}
	}

	for _, match := range passportIssueDMMMYY.FindAllStringSubmatch(text, -1) {
		day, _ := strconv.Atoi(match[1])
		month := passportMonthFromMMM(match[2])
		year, _ := strconv.Atoi(match[3])
		if year <= 99 {
			currentYY := time.Now().UTC().Year() % 100
			if year <= currentYY+1 {
				year += 2000
			} else {
				year += 1900
			}
		}
		if parsed := safePassportIssueDate(year, month, day); !parsed.IsZero() {
			candidates = append(candidates, parsed)
		}
	}

	return candidates
}

func passportMonthFromMMM(value string) int {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "JAN":
		return 1
	case "FEB":
		return 2
	case "MAR":
		return 3
	case "APR":
		return 4
	case "MAY":
		return 5
	case "JUN":
		return 6
	case "JUL":
		return 7
	case "AUG":
		return 8
	case "SEP":
		return 9
	case "OCT":
		return 10
	case "NOV":
		return 11
	case "DEC":
		return 12
	default:
		return 0
	}
}

func safePassportIssueDate(year, month, day int) time.Time {
	if year <= 0 || month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}
	}
	parsed := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if parsed.Year() != year || int(parsed.Month()) != month || parsed.Day() != day {
		return time.Time{}
	}
	return parsed
}
