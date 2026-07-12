package service

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"

	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrMissingRequiredDocuments = errors.New("missing required documents")
	ErrPDFGenerationFailed      = errors.New("pdf generation failed")
)

const maxRemoteAssetBytes int64 = 5 << 20

//go:embed assets/*.png
var assetsFS embed.FS

type PDFService struct {
	storageService StorageService
}

func NewPDFService(storageService StorageService) *PDFService {
	return &PDFService{storageService: storageService}
}

type assetCacheEntry struct {
	data      []byte
	contentType string
	expiresAt time.Time
}

var (
	assetCacheMu sync.RWMutex
	assetCache   = make(map[string]*assetCacheEntry)
	assetCacheTTL = 5 * time.Minute
)

func getCachedAsset(url string) ([]byte, string, bool) {
	assetCacheMu.RLock()
	entry, ok := assetCache[url]
	assetCacheMu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, "", false
	}
	return entry.data, entry.contentType, true
}

func setCachedAsset(url string, data []byte, contentType string) {
	assetCacheMu.Lock()
	assetCache[url] = &assetCacheEntry{
		data:        data,
		contentType: contentType,
		expiresAt:   time.Now().Add(assetCacheTTL),
	}
	assetCacheMu.Unlock()
}

func (s *PDFService) GenerateCandidateCV(candidate *domain.Candidate, documents []*domain.Document, branding CandidateCVBranding, passportData *domain.PassportData) ([]byte, error) {
	if candidate == nil {
		return nil, fmt.Errorf("candidate is nil")
	}

	photoDoc, passportDoc, _ := pickCandidateDocuments(documents)
	if photoDoc == nil || passportDoc == nil || strings.TrimSpace(photoDoc.FileURL) == "" || strings.TrimSpace(passportDoc.FileURL) == "" {
		return nil, ErrMissingRequiredDocuments
	}

	langEntries := parseLanguageEntries(candidate.Languages)
	skills := parseSkills(candidate.Skills)
	experienceEntries := parseExperienceEntries(candidate.ExperienceAbroad)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Candidate CV", false)
	pdf.SetAuthor("Maid Recruitment Platform", false)
	pdf.SetMargins(15, 15, 15)
	pdf.SetFont("Arial", "", 12)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	// s.drawBrandingHeader(pdf, branding)
	// s.drawHeader(pdf, branding)

	photoBytes, photoContentType, err := s.fetchImage(photoDoc.FileURL)
	if err != nil {
		return nil, fmt.Errorf("fetch photo: %w", err)
	}
	s.drawApplicationProfileSheet(pdf, candidate, branding, passportData, photoBytes, photoContentType, langEntries, skills, experienceEntries)

	s.appendPassportSection(pdf, passportDoc)
	appendFullBodyPhotoSection(pdf, photoDoc, photoBytes, photoContentType)

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPDFGenerationFailed, err)
	}

	return buffer.Bytes(), nil
}

func (s *PDFService) appendPassportDetailsSection(pdf *gofpdf.Fpdf, passportData *domain.PassportData) {
	if passportData == nil {
		return
	}

	pdf.AddPage()
	drawAttachmentHeader(pdf, "Passport Details")

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(209, 213, 219)
	pdf.RoundedRect(15, 36, 180, 96, 2.5, "1234", "DF")

	fields := []struct {
		label string
		value string
		alert bool
	}{
		{label: "Holder Name", value: compactValue(passportData.HolderName, "Not available")},
		{label: "Passport No", value: compactValue(passportData.PassportNumber, "Not available")},
		{label: "Nationality", value: compactValue(passportData.Nationality, "Not available")},
		{label: "Date of Birth", value: formatPassportDate(passportData.DateOfBirth)},
		{label: "Place of Birth", value: compactValue(passportData.PlaceOfBirth, "Not available")},
		{label: "Gender", value: compactValue(passportData.Gender, "Not available")},
		{label: "Issue Date", value: formatPassportDatePointer(passportData.IssueDate)},
		{label: "Expiry Date", value: formatPassportDate(passportData.ExpiryDate), alert: isPassportExpiringSoon(passportData.ExpiryDate)},
	}

	startX := 22.0
	startY := 45.0
	columnGap := 86.0
	rowGap := 18.0
	for index, field := range fields {
		x := startX
		if index%2 == 1 {
			x += columnGap
		}
		y := startY + (float64(index/2) * rowGap)

		pdf.SetXY(x, y)
		pdf.SetTextColor(107, 114, 128)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(70, 4, field.label, "", 1, "L", false, 0, "")
		pdf.SetXY(x, y+5)
		if field.alert {
			pdf.SetTextColor(220, 38, 38)
		} else {
			pdf.SetTextColor(31, 41, 55)
		}
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(72, 5, field.value, "", "L", false)
	}

	pdf.SetFillColor(255, 247, 237)
	pdf.SetDrawColor(251, 191, 36)
	pdf.RoundedRect(15, 138, 180, 46, 2.5, "1234", "DF")
	pdf.SetXY(20, 145)
	pdf.SetTextColor(55, 65, 81)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(170, 5, "Machine Readable Zone", "", 1, "L", false, 0, "")
	pdf.SetX(20)
	pdf.SetTextColor(75, 85, 99)
	pdf.SetFont("Courier", "", 11)
	pdf.MultiCell(170, 6, compactValue(passportData.MRZLine1, "Not available"), "", "L", false)
	pdf.SetX(20)
	pdf.MultiCell(170, 6, compactValue(passportData.MRZLine2, "Not available"), "", "L", false)
}

func (s *PDFService) drawBrandingHeader(pdf *gofpdf.Fpdf, branding CandidateCVBranding) {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetXY(12, 14)
	pdf.SetTextColor(30, 58, 138)
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(186, 8, "CANDIDATE APPLICATION PROFILE", "", 1, "C", false, 0, "")
	pdf.SetY(26)
}

func (s *PDFService) drawHeader(pdf *gofpdf.Fpdf, branding CandidateCVBranding) {
	headerY := pdf.GetY()
	
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.Rect(12, headerY, 186, 24, "DF")

	companyName := compactValue(strings.TrimSpace(branding.CompanyName), "Maid Recruitment Agency")
	foreignName := compactValue(strings.TrimSpace(branding.ForeignAgencyName), "Foreign Employment File")

	pdf.SetXY(16, headerY+8)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(60, 8, companyName, "", 0, "L", false, 0, "")

	logoWidth := 34.0
	logoHeight := 20.0
	logoX := 105.0 - (logoWidth / 2)
	logoY := headerY + 2
	
	if logoBytes, logoContentType, err := decodeLogoDataURL(branding.LogoDataURL); err == nil && len(logoBytes) > 0 {
		_ = addImageFromBytesFit(pdf, logoBytes, logoContentType, "agency_logo_header", logoX, logoY, logoWidth, logoHeight)
	}

	pdf.SetXY(134, headerY+8)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(60, 8, foreignName, "", 1, "R", false, 0, "")

	pdf.SetY(headerY + 28)
}

func drawAssetIcon(pdf *gofpdf.Fpdf, iconName string, x, y, w, h float64) {
	data, err := assetsFS.ReadFile("assets/" + iconName)
	if err == nil && len(data) > 0 {
		_ = addImageFromBytesFit(pdf, data, "image/png", iconName, x, y, w, h)
	}
}

func (s *PDFService) drawApplicationProfileSheet(
	pdf *gofpdf.Fpdf,
	candidate *domain.Candidate,
	branding CandidateCVBranding,
	passportData *domain.PassportData,
	photoBytes []byte,
	photoContentType string,
	langEntries []domain.LanguageEntry,
	skills []domain.Skill,
	experienceEntries []domain.ExperienceEntry,
) {
	// Constants for Layout
	leftX := 12.0
	leftWidth := 65.0
	rightX := 82.0
	rightWidth := 115.0

	// ----- FULL-WIDTH BRANDING HEADER -----
	headerY := 7.0
	fullW := rightX + rightWidth - leftX

	pdf.SetFillColor(cardBgR, cardBgG, cardBgB)
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.RoundedRect(leftX, headerY, fullW, 14, 3, "1234", "DF")

	ethiopianPartner := compactValue(strings.TrimSpace(branding.CompanyName), "Ethiopian Recruitment Agency")
	foreignEmployer := compactValue(strings.TrimSpace(branding.ForeignAgencyName), "Foreign Employer Agency")

	pdf.SetTextColor(navyR, navyG, navyB)
	drawFitPDFText(pdf, leftX+4, headerY+3, 70, 8, ethiopianPartner, "Arial", "B", 9, 6, "L")

	if logoBytes, logoContentType, err := decodeLogoDataURL(branding.LogoDataURL); err == nil && len(logoBytes) > 0 {
		_ = addImageFromBytesFit(pdf, logoBytes, logoContentType, "branding_logo", leftX+fullW/2-17, headerY+2, 34, 10)
	}

	pdf.SetTextColor(navyR, navyG, navyB)
	drawFitPDFText(pdf, leftX+fullW-4-70, headerY+3, 70, 8, foreignEmployer, "Arial", "B", 9, 6, "R")

	pdf.SetDrawColor(blueR, blueG, blueB)
	pdf.SetLineWidth(0.4)
	pdf.Line(leftX, headerY+16, leftX+fullW, headerY+16)
	pdf.SetLineWidth(0.2)

	pdf.SetXY(leftX, headerY+18)
	pdf.SetTextColor(navyR, navyG, navyB)
	pdf.SetFont("Arial", "B", 15)
	pdf.CellFormat(fullW, 8, "CANDIDATE APPLICATION PROFILE", "", 1, "C", false, 0, "")

	contentTop := headerY + 31

	// ----- LEFT COLUMN -----
	photoY := contentTop
	photoHeight := 74.0

	// Photo Card
	drawCard(pdf, leftX, photoY, leftWidth, photoHeight)

	if err := addImageFromBytesFit(pdf, photoBytes, photoContentType, "candidate_sheet_photo", leftX+2, photoY+2, leftWidth-4, photoHeight-4); err != nil {
		drawPhotoPlaceholder(pdf, leftX+2, photoY+2, leftWidth-4, photoHeight-4)
	}

	// Name Banner (below photo)
	names := strings.Fields(strings.ToUpper(compactValue(candidateProfileName(candidate, passportData), "CANDIDATE PROFILE")))
	firstName := ""
	lastName := ""
	if len(names) > 0 {
		firstName = names[0]
		if len(names) > 1 {
			lastName = strings.Join(names[1:], " ")
		}
	}

	bannerH := 14.0
	if lastName != "" {
		bannerH = 20.0
	}
	bannerY := photoY + photoHeight + 1

	pdf.SetFillColor(navyR, navyG, navyB)
	pdf.Rect(leftX, bannerY, leftWidth, bannerH, "F")

	pdf.SetFillColor(blueR, blueG, blueB)
	pdf.Rect(leftX, bannerY, 3, bannerH, "F")

	nameW := leftWidth - 6
	nameCenterX := leftX + 4
	pdf.SetTextColor(255, 255, 255)

	if lastName != "" {
		drawFitPDFText(pdf, nameCenterX, bannerY+3, nameW, 7, firstName, "Arial", "B", 8, 6, "C")
		drawFitPDFText(pdf, nameCenterX, bannerY+11, nameW, 7, lastName, "Arial", "B", 8, 6, "C")
	} else {
		drawFitPDFText(pdf, nameCenterX, bannerY+3, nameW, 9, firstName, "Arial", "B", 9, 7, "C")
	}

	// Metric Chips
	chipsY := bannerY + bannerH + 3
	chipH := 10.0
	chipW := (leftWidth - 4) / 3

	chipData := [][2]string{
		{"AGE", manualCandidateAgeValue(candidate)},
		{"EXP", formatExperienceLevel(candidate.ExperienceYears)},
		{"RELIGION", compactValue(strings.ToUpper(strings.TrimSpace(candidate.Religion)), "-")},
	}

	for i, chip := range chipData {
		cx := leftX + 2 + float64(i)*(chipW+1)
		pdf.SetFillColor(headerBgR, headerBgG, headerBgB)
		pdf.SetDrawColor(borderR, borderG, borderB)
		pdf.RoundedRect(cx, chipsY, chipW, chipH, 3, "1234", "DF")

		pdf.SetXY(cx+1, chipsY+1)
		pdf.SetTextColor(mutedR, mutedG, mutedB)
		pdf.SetFont("Arial", "B", 6)
		pdf.CellFormat(chipW-2, 3.5, chip[0], "", 1, "C", false, 0, "")

		pdf.SetX(cx + 1)
		pdf.SetTextColor(darkR, darkG, darkB)
		pdf.SetFont("Arial", "B", 7.5)
		pdf.CellFormat(chipW-2, 4, chip[1], "", 1, "C", false, 0, "")
	}

	appRows := [][2]string{
		{"Applied For", "HOUSEMAID"},
		{"Ref. No", shortReference(candidate.ID)},
		{"Contract Period", "2 YRS"},
		{"Monthly Salary", manualCandidateUpper(candidate.SalaryOffered, "N/A")},
		{"Applied Contract", manualCandidateUpper(candidate.CountryApplied, "N/A")},
		{"Experience", formatExperienceLevel(candidate.ExperienceYears)},
		{"Date", currentPDFDate()},
	}

	// Application Details Card
	appY := chipsY + chipH + 2
	drawCard(pdf, leftX, appY, leftWidth, 10+float64(len(appRows))*7.5)
	appRowH := 7.5

	pdf.SetFillColor(navyR, navyG, navyB)
	pdf.RoundedRect(leftX, appY, leftWidth, 9, cardRadius, "12", "F")
	pdf.SetXY(leftX, appY+1.5)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(leftWidth, 5, "APPLICATION DETAILS", "", 1, "C", false, 0, "")

	appRowY := appY + 11
	labelW := 26.0
	valW := leftWidth - labelW - 4
	for i, r := range appRows {
		pdf.SetXY(leftX+4, appRowY)
		pdf.SetTextColor(mutedR, mutedG, mutedB)
		pdf.SetFont("Arial", "", 7)
		pdf.CellFormat(labelW, appRowH, r[0], "", 0, "L", false, 0, "")

		pdf.SetTextColor(darkR, darkG, darkB)
		drawFitPDFText(pdf, leftX+4+labelW, appRowY+0.5, valW, appRowH-1, compactValue(r[1], "N/A"), "Arial", "B", 7.5, 5.5, "R")

		if i < len(appRows)-1 {
			pdf.SetDrawColor(borderR, borderG, borderB)
			pdf.SetLineWidth(0.2)
			pdf.Line(leftX+4, appRowY+appRowH-0.5, leftX+leftWidth-4, appRowY+appRowH-0.5)
			pdf.SetLineWidth(0.2)
		}

		appRowY += appRowH
	}

	// ----- RIGHT COLUMN CONTENT -----
	currentRY := contentTop

	// 1. PERSONAL INFORMATION
	ageValue := manualCandidateAgeValue(candidate)
	infoLabels := []string{"Nationality", "Date of Birth", "Age", "Place of Birth", "Gender", "Religion", "Marital Status", "Children", "Education Level"}
	infoValues := []string{
		manualCandidateUpper(candidate.Nationality, "N/A"),
		manualCandidateDate(candidate.DateOfBirth),
		ageValue,
		manualCandidateUpper(candidate.PlaceOfBirth, "N/A"),
		manualCandidateUpper(candidate.Gender, "N/A"),
		compactValue(strings.ToUpper(strings.TrimSpace(candidate.Religion)), "N/A"),
		compactValue(strings.ToUpper(strings.TrimSpace(candidate.MaritalStatus)), "N/A"),
		formatChildrenCount(candidate.ChildrenCount),
		compactValue(strings.ToUpper(strings.TrimSpace(candidate.EducationLevel)), "N/A"),
	}
	infoIcons := []string{"flag.png", "calendar.png", "user.png", "marker.png", "female.png", "church.png", "hearts.png", "children.png", "education.png"}

	infoRows := (len(infoLabels) + 1) / 2
	infoCardH := 9.0 + float64(infoRows)*7.0 + 3.0
	s.drawProfileSectionBox(pdf, rightX, currentRY, rightWidth, infoCardH, "PERSONAL INFORMATION", "user.png")

	rowH := 7.0
	startY := currentRY + 11
	lColX := rightX + 4
	rColX := rightX + rightWidth/2 + 2
	colW := rightWidth/2 - 6

	for i := 0; i < infoRows; i++ {
		drawAssetIcon(pdf, infoIcons[i], lColX, startY+float64(i)*rowH+1, 4, 4)
		pdf.SetXY(lColX+5, startY+float64(i)*rowH-0.5)
		pdf.SetTextColor(mutedR, mutedG, mutedB)
		pdf.SetFont("Arial", "", 6.5)
		pdf.CellFormat(colW, 3, infoLabels[i], "", 1, "L", false, 0, "")
		pdf.SetXY(lColX+5, startY+float64(i)*rowH+2.5)
		pdf.SetTextColor(darkR, darkG, darkB)
		drawFitPDFText(pdf, lColX+5, startY+float64(i)*rowH+2.5, colW-2, 4, compactValue(infoValues[i], "N/A"), "Arial", "B", 7.5, 5.5, "L")

		if ri := i + infoRows; ri < len(infoLabels) {
			drawAssetIcon(pdf, infoIcons[ri], rColX, startY+float64(i)*rowH+1, 4, 4)
			pdf.SetXY(rColX+5, startY+float64(i)*rowH-0.5)
			pdf.SetTextColor(mutedR, mutedG, mutedB)
			pdf.SetFont("Arial", "", 6.5)
			pdf.CellFormat(colW, 3, infoLabels[ri], "", 1, "L", false, 0, "")
			pdf.SetXY(rColX+5, startY+float64(i)*rowH+2.5)
			pdf.SetTextColor(darkR, darkG, darkB)
			drawFitPDFText(pdf, rColX+5, startY+float64(i)*rowH+2.5, colW-2, 4, compactValue(infoValues[ri], "N/A"), "Arial", "B", 7.5, 5.5, "L")
		}
	}
	currentRY += infoCardH + 6


	// 2. PASSPORT DETAILS
	passNo := passportValue(passportData, func(data *domain.PassportData) string { return data.PassportNumber })
	if passNo == "" && strings.TrimSpace(candidate.PassportNumber) != "" {
		passNo = strings.ToUpper(strings.TrimSpace(candidate.PassportNumber))
	}
	issueDateVal := passportDatePointerValue(passportData, func(data *domain.PassportData) *time.Time { return data.IssueDate })
	if issueDateVal == "N/A" {
		issueDateVal = manualCandidateDate(candidate.IssueDate)
	}
	expiryDateVal := passportDateValue(passportData, func(data *domain.PassportData) time.Time { return data.ExpiryDate })
	if expiryDateVal == "N/A" {
		expiryDateVal = manualCandidateDate(candidate.ExpiryDate)
	}
	remainingVal := passportRemainingYears(passportData)
	if remainingVal == "N/A" && candidate.ExpiryDate != nil && !candidate.ExpiryDate.IsZero() {
		remainingVal = computeRemainingYears(*candidate.ExpiryDate)
	}

	passCardH := 9.0 + 2*10.0 + 4.0
	s.drawProfileSectionBox(pdf, rightX, currentRY, rightWidth, passCardH, "PASSPORT DETAILS", "passport.png")

	passRowY := currentRY + 10
	passRowH := 10.0
	passLbl := func(label string, value string, x float64, y float64, w float64) {
		pdf.SetXY(x, y)
		pdf.SetTextColor(mutedR, mutedG, mutedB)
		pdf.SetFont("Arial", "", 7)
		pdf.CellFormat(w, 3.5, label, "", 1, "L", false, 0, "")
		pdf.SetXY(x, y+3.5)
		pdf.SetTextColor(darkR, darkG, darkB)
		pdf.SetFont("Arial", "B", 8.5)
		pdf.CellFormat(w, 4.5, compactValue(value, "N/A"), "", 1, "L", false, 0, "")
	}

	pL := rightX + 5
	pW := rightWidth/2 - 8
	passLbl("Passport No.", strings.ToUpper(passNo), pL, passRowY, pW)
	passLbl("Issue Date", issueDateVal, pL, passRowY+passRowH, pW)

	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.SetLineWidth(0.2)
	pdf.Line(rightX+rightWidth/2-1, currentRY+11, rightX+rightWidth/2-1, currentRY+passCardH-3)

	pR := rightX + rightWidth/2 + 4
	passLbl("Expiry Date", expiryDateVal, pR, passRowY, pW)
	passLbl("Remaining Years", remainingVal, pR, passRowY+passRowH, pW)

	currentRY += passCardH + 6

	// 3. WORK EXPERIENCE (DOMESTIC)
	skillIconMapping := map[string]string{
		"CLEANING":      "broom.png",
		"COOKING":       "pot.png",
		"IRONING":       "iron.png",
		"BABY SITTING":  "baby.png",
		"CHILDCARE":     "children.png",
		"HOME TEACHING": "book.png",
		"ELDERLY CARE":  "hearts.png",
		"LAUNDRY":       "broom.png",
		"DRIVING":       "marker.png",
		"FIRST AID":     "cross.png",
		"PET CARE":      "hearts.png",
		"PERSONAL CARE": "hearts.png",
	}

	var activeSkills []domain.Skill
	for _, s := range skills {
		if s.Value {
			activeSkills = append(activeSkills, s)
		}
	}
	if len(activeSkills) == 0 {
		activeSkills = skills
	}

	skillCount := len(activeSkills)
	rows := (skillCount + 1) / 2
	if rows < 1 {
		rows = 1
	}
	boxHeight := 10.0 + float64(rows)*7.0
	colWidth := 55.0

	s.drawProfileSectionBox(pdf, rightX, currentRY, rightWidth, boxHeight, "WORK EXPERIENCE (DOMESTIC)", "briefcase.png")

	for i, skill := range activeSkills {
		col := i % 2
		row := i / 2
		x := rightX + 6 + (float64(col) * colWidth)
		y := currentRY + 11 + (float64(row) * 7)

		if iconName, ok := skillIconMapping[strings.ToUpper(strings.TrimSpace(skill.Name))]; ok {
			drawAssetIcon(pdf, iconName, x, y, 4, 4)
		}

		pdf.SetXY(x+5, y+0.5)
		pdf.SetTextColor(darkR, darkG, darkB)
		pdf.SetFont("Arial", "B", 7.5)
		pdf.CellFormat(25, 4, skill.Name, "", 0, "L", false, 0, "")

		drawBadgePill(pdf, x+28, y+0.25, 22, 5, "YES", emeraldR, emeraldG, emeraldB)
	}

	currentRY += boxHeight + 6

	// 4. EXPERIENCED ABROAD
	var expRows [][2]string
	for _, e := range experienceEntries {
		c := strings.TrimSpace(e.Country)
		if c == "" {
			continue
		}
		yearsLabel := fmt.Sprintf("%d YEARS", e.Years)
		if e.Years <= 0 {
			yearsLabel = "-"
		}
		expRows = append(expRows, [2]string{strings.ToUpper(c), yearsLabel})
	}
	if len(expRows) == 0 && candidate.ExperienceYears != nil && *candidate.ExperienceYears > 0 && strings.TrimSpace(candidate.CountryOfExperience) != "" {
		expRows = append(expRows, [2]string{
			manualCandidateUpper(candidate.CountryOfExperience, "-"),
			fmt.Sprintf("%d YEARS", *candidate.ExperienceYears),
		})
	}
	
	// Minimum box height: 24 if empty (for "FIRST TIME" label), else header + rows
	expBoxH := 24.0
	if len(expRows) > 0 {
		expBoxH = 9.0 + float64(len(expRows))*7.0 + 6.0
	}
	s.drawProfileSectionBox(pdf, rightX, currentRY, rightWidth, expBoxH, "EXPERIENCED ABROAD", "globe.png")

	if len(expRows) == 0 {
		drawBadgePill(pdf, rightX+rightWidth/2-18, currentRY+13, 36, 7, "FIRST TIME", blueR, blueG, blueB)
	} else {
		yPos := currentRY + 11
		for i, r := range expRows {
			pdf.SetXY(rightX+4, yPos+0.5)
			pdf.SetTextColor(mutedR, mutedG, mutedB)
			pdf.SetFont("Arial", "", 6.5)
			pdf.CellFormat(8, 3, fmt.Sprintf("%d.", i+1), "", 0, "C", false, 0, "")
			pdf.SetTextColor(darkR, darkG, darkB)
			drawFitPDFText(pdf, rightX+14, yPos+0.5, rightWidth-55, 6, r[0], "Arial", "B", 7.5, 5.5, "L")

			pdf.SetXY(rightX+rightWidth-32, yPos+1)
			pdf.SetTextColor(emeraldR, emeraldG, emeraldB)
			pdf.SetFont("Arial", "B", 7)
			pdf.CellFormat(28, 5, r[1], "", 1, "R", false, 0, "")

			if i < len(expRows)-1 {
				pdf.SetDrawColor(borderR, borderG, borderB)
				pdf.SetLineWidth(0.2)
				pdf.Line(rightX+4, yPos+7, rightX+rightWidth-4, yPos+7)
				pdf.SetLineWidth(0.2)
			}
			yPos += 7
		}
	}

	currentRY += expBoxH + 6

	// 5. LANGUAGE KNOWN
	langBoxH := 24.0
	if len(langEntries) > 0 {
		langBoxH = 9.0 + float64(len(langEntries))*7.0 + 6.0
	}
	s.drawProfileSectionBox(pdf, rightX, currentRY, rightWidth, langBoxH, "LANGUAGE KNOWN", "chat.png")

	if len(langEntries) == 0 {
		drawBadgePill(pdf, rightX+rightWidth/2-18, currentRY+13, 36, 7, "FIRST TIME", blueR, blueG, blueB)
	} else {
		levelColor := func(level string) (r, g, b int) {
			switch strings.ToUpper(level) {
			case "FLUENT", "ADVANCED", "NATIVE":
				return emeraldR, emeraldG, emeraldB
			case "INTERMEDIATE":
				return amberR, amberG, amberB
			default:
				return blueR, blueG, blueB
			}
		}

		yPos := currentRY + 11
		for i, r := range langEntries {
			lang := strings.TrimSpace(r.Language)
			if lang == "" {
				continue
			}
			level := strings.ToUpper(strings.TrimSpace(r.Proficiency))
			if level == "" {
				level = "BASIC"
			}

			pdf.SetXY(rightX+4, yPos+0.5)
			pdf.SetTextColor(mutedR, mutedG, mutedB)
			pdf.SetFont("Arial", "", 6.5)
			pdf.CellFormat(8, 3, fmt.Sprintf("%d.", i+1), "", 0, "C", false, 0, "")
			pdf.SetTextColor(darkR, darkG, darkB)
			drawFitPDFText(pdf, rightX+14, yPos+0.5, rightWidth-58, 6, strings.ToUpper(lang), "Arial", "B", 7.5, 5.5, "L")

			lr, lg, lb := levelColor(level)
			drawBadgePill(pdf, rightX+rightWidth-36, yPos+0.5, 32, 5.5, level, lr, lg, lb)

			if i < len(langEntries)-1 {
				pdf.SetDrawColor(borderR, borderG, borderB)
				pdf.SetLineWidth(0.2)
				pdf.Line(rightX+4, yPos+7, rightX+rightWidth-4, yPos+7)
				pdf.SetLineWidth(0.2)
			}
			yPos += 7
		}
	}
	drawPageFooter(pdf)
}

// Helper to draw section box with header
func (s *PDFService) drawProfileSectionBox(pdf *gofpdf.Fpdf, x float64, y float64, w float64, h float64, title string, iconName string) {
	pdf.SetFillColor(cardBgR, cardBgG, cardBgB)
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.RoundedRect(x, y, w, h, cardRadius, "1234", "DF")

	pdf.SetFillColor(headerBgR, headerBgG, headerBgB)
	pdf.RoundedRect(x, y, w, 9, cardRadius, "12", "F")

	textX := x + 4
	if iconName != "" {
		drawAssetIcon(pdf, iconName, x+3, y+2, 5, 5)
		textX += 7
	}

	pdf.SetXY(textX, y+1.5)
	pdf.SetTextColor(darkR, darkG, darkB)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(w-8, 6, title, "", 1, "L", false, 0, "")
}

func (s *PDFService) drawCompactField(pdf *gofpdf.Fpdf, x, y, width, height float64, label, value string) {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(209, 213, 219)
	pdf.Rect(x, y, width, height, "D")
	pdf.SetXY(x+2, y+2)
	pdf.SetTextColor(107, 114, 128)
	pdf.SetFont("Arial", "B", 7.5)
	pdf.CellFormat(width-4, 3.5, label, "", 1, "L", false, 0, "")
	pdf.SetX(x + 2)
	pdf.SetTextColor(31, 41, 55)
	pdf.SetFont("Arial", "B", 9.5)
	pdf.CellFormat(width-4, 4.5, value, "", 1, "L", false, 0, "")
}

func (s *PDFService) drawWrappedField(pdf *gofpdf.Fpdf, x, y, width, height float64, title string, lines []string) {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(209, 213, 219)
	pdf.Rect(x, y, width, height, "D")
	pdf.SetXY(x+2, y+2)
	pdf.SetTextColor(234, 88, 12)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(width-4, 4, title, "", 1, "L", false, 0, "")
	pdf.SetX(x + 2)
	pdf.SetTextColor(55, 65, 81)
	pdf.SetFont("Arial", "", 8.4)
	for _, line := range lines {
		pdf.SetX(x + 2)
		pdf.MultiCell(width-4, 4.4, line, "", "L", false)
	}
}

func (s *PDFService) drawWideField(pdf *gofpdf.Fpdf, x, y, width, height float64, title, value string) {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(209, 213, 219)
	pdf.Rect(x, y, width, height, "D")
	pdf.SetXY(x+3, y+3)
	pdf.SetTextColor(234, 88, 12)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(width-6, 4, title, "", 1, "L", false, 0, "")
	pdf.SetX(x + 3)
	pdf.SetTextColor(55, 65, 81)
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(width-6, 5, value, "", "L", false)
}

type pdfFormRow struct {
	Label string
	Value string
}

func (s *PDFService) drawFormInfoRow(pdf *gofpdf.Fpdf, x, y, width, labelWidth, rowHeight float64, rows [][2]string) {
	for index, row := range rows {
		rowY := y + (float64(index) * rowHeight)
		s.drawFormKeyValueRow(pdf, x, rowY, width, labelWidth, rowHeight, row[0], row[1], false)
	}
}

func (s *PDFService) drawNamedBand(pdf *gofpdf.Fpdf, x, y, width float64, value string) {
	s.drawFormKeyValueRow(pdf, x, y, width, 35, 8.5, "Full Name", compactValue(value, "N/A"), true)
}

func (s *PDFService) drawFormTable(pdf *gofpdf.Fpdf, x, y, width float64, title string, rows []pdfFormRow, labelWidth float64) {
	titleHeight := 8.0
	rowHeight := 8.0

	pdf.SetFillColor(245, 245, 245)
	pdf.SetDrawColor(191, 191, 191)
	pdf.Rect(x, y, width, titleHeight, "DF")
	pdf.SetXY(x, y+2)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "B", 8.2)
	pdf.CellFormat(width, 4, title, "", 1, "C", false, 0, "")

	currentY := y + titleHeight
	for _, row := range rows {
		s.drawFormKeyValueRow(pdf, x, currentY, width, labelWidth, rowHeight, row.Label, row.Value, false)
		currentY += rowHeight
	}
}

func (s *PDFService) drawFormKeyValueRow(pdf *gofpdf.Fpdf, x, y, width, labelWidth, height float64, label, value string, highlightValue bool) {
	pdf.SetDrawColor(191, 191, 191)
	pdf.SetFillColor(246, 246, 246)
	pdf.Rect(x, y, labelWidth, height, "DF")
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(x+labelWidth, y, width-labelWidth, height, "DF")

	pdf.SetTextColor(92, 104, 77)
	drawFitPDFText(pdf, x+2, y, labelWidth-4, height, compactValue(label, "-"), "Arial", "B", 6.6, 4.8, "L")

	if highlightValue {
		pdf.SetTextColor(33, 33, 33)
		drawFitPDFText(pdf, x+labelWidth+2, y, width-labelWidth-4, height, compactValue(value, "-"), "Arial", "B", 13, 8.6, "C")
	} else {
		pdf.SetTextColor(70, 70, 70)
		drawFitPDFText(pdf, x+labelWidth+2, y, width-labelWidth-4, height, compactValue(value, "-"), "Arial", "B", 8, 5.3, "C")
	}
}

func (s *PDFService) drawSplitHistorySection(pdf *gofpdf.Fpdf, x, y, width float64, title string, columns []string, rows [][2]string) {
	pdf.SetFillColor(245, 245, 245)
	pdf.SetDrawColor(191, 191, 191)
	pdf.Rect(x, y, width, 8, "DF")
	pdf.SetXY(x, y+2)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "B", 8.2)
	pdf.CellFormat(width, 4, title, "", 1, "C", false, 0, "")

	columnWidth := width / 2
	rowY := y + 8
	s.drawFormKeyValueRow(pdf, x, rowY, columnWidth-1, 21, 7.2, columns[0], "", false)
	s.drawFormKeyValueRow(pdf, x+columnWidth+1, rowY, columnWidth-1, 21, 7.2, columns[1], "", false)
	rowY += 7.2
	for _, row := range rows {
		s.drawFormKeyValueRow(pdf, x, rowY, columnWidth-1, 21, 7.2, row[0], "", false)
		s.drawFormKeyValueRow(pdf, x+columnWidth+1, rowY, columnWidth-1, 21, 7.2, row[1], "", false)
		rowY += 7.2
	}
}

func (s *PDFService) drawExperiencedAbroadSection(pdf *gofpdf.Fpdf, x, y, width float64, rows [][2]string) {
	sectionGap := 6.0
	sectionWidth := (width - sectionGap) / 2
	leftX := x
	rightX := x + sectionWidth + sectionGap

	s.drawExperienceColumn(pdf, leftX, y, sectionWidth, "Experienced Abroad", "Country", extractExperienceColumn(rows, 0))
	s.drawExperienceColumn(pdf, rightX, y, sectionWidth, "Experience Abroad Duration", "How Long", extractExperienceColumn(rows, 1))
}

func (s *PDFService) drawExperienceColumn(pdf *gofpdf.Fpdf, x, y, width float64, title, label string, values []string) {
	titleHeight := 6.8
	rowHeight := 9.6

	pdf.SetFillColor(245, 245, 245)
	pdf.SetDrawColor(191, 191, 191)
	pdf.Rect(x, y, width, titleHeight, "DF")
	pdf.SetXY(x, y+1.6)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "B", 7.4)
	pdf.CellFormat(width, 4, title, "", 1, "C", false, 0, "")

	currentY := y + titleHeight + 2.8
	for _, value := range values {
		s.drawFormKeyValueRow(pdf, x, currentY, width, width*0.56, rowHeight, label, value, false)
		currentY += rowHeight + 2.6
	}
}

func extractExperienceColumn(rows [][2]string, index int) []string {
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		if index >= 0 && index < len(row) {
			values = append(values, row[index])
			continue
		}
		values = append(values, "-")
	}
	return values
}

func (s *PDFService) drawRemarkSection(pdf *gofpdf.Fpdf, x, y, width float64, title, remark string) {
	s.drawFormKeyValueRow(pdf, x, y, width, 35, 7.2, title, "", false)
	pdf.SetDrawColor(191, 191, 191)
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(x, y+7.2, width, 7.4, "DF")
	pdf.Rect(x, y+14.6, width, 7.4, "DF")
	pdf.Rect(x, y+22.0, width, 7.4, "DF")
	if strings.TrimSpace(remark) != "" {
		pdf.SetXY(x+2.5, y+9.1)
		pdf.SetTextColor(85, 85, 85)
		pdf.SetFont("Arial", "", 7.8)
		pdf.MultiCell(width-5, 4, remark, "", "L", false)
	}
}

func (s *PDFService) drawSectionHeading(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFillColor(3, 105, 161)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 12)
	pdf.RoundedRect(15, pdf.GetY(), 180, 9, 3, "1234", "F")
	pdf.SetX(18)
	pdf.CellFormat(174, 9, title, "", 1, "L", false, 0, "")
	pdf.SetY(pdf.GetY() + 2)
}

func (s *PDFService) drawHeroCard(
	pdf *gofpdf.Fpdf,
	candidate *domain.Candidate,
	branding CandidateCVBranding,
	photoBytes []byte,
	photoContentType string,
	languages []string,
	skills []string,
) {
	cardY := pdf.GetY()
	cardHeight := 84.0
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.RoundedRect(15, cardY, 180, cardHeight, 5, "1234", "DF")
	pdf.SetFillColor(14, 165, 233)
	pdf.RoundedRect(15, cardY, 180, 14, 5, "12", "F")
	pdf.SetXY(20, cardY+4)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(110, 5, "Candidate Overview", "", 1, "L", false, 0, "")

	photoFrameX := 144.0
	photoFrameY := cardY + 18
	photoFrameWidth := 46.0
	photoFrameHeight := 58.0
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(125, 211, 252)
	pdf.RoundedRect(photoFrameX, photoFrameY, photoFrameWidth, photoFrameHeight, 4, "1234", "DF")
	_ = addImageFromBytesFit(pdf, photoBytes, photoContentType, "candidate_photo", photoFrameX+2, photoFrameY+2, photoFrameWidth-4, photoFrameHeight-4)

	pdf.SetXY(20, cardY+20)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Arial", "B", 19)
	pdf.CellFormat(116, 8, candidateProfileName(candidate, nil), "", 1, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetFont("Arial", "", 9.5)
	agencyName := strings.TrimSpace(branding.CompanyName)
	if agencyName == "" {
		agencyName = "Recruitment Agency"
	}
	pdf.CellFormat(116, 5, fmt.Sprintf("Prepared by %s", agencyName), "", 1, "L", false, 0, "")

	s.drawHeroMetricChip(pdf, 20, cardY+37, 34, "Age", formatIntPointer(candidate.Age))
	s.drawHeroMetricChip(pdf, 58, cardY+37, 40, "Experience", fmt.Sprintf("%s years", formatIntPointer(candidate.ExperienceYears)))
	s.drawHeroMetricChip(pdf, 102, cardY+37, 34, "Languages", fmt.Sprintf("%d", len(languages)))

	pdf.SetXY(20, cardY+51)
	pdf.SetTextColor(51, 65, 85)
	pdf.SetFont("Arial", "", 10)
	
	experienceText := fmt.Sprintf("%s years", formatIntPointer(candidate.ExperienceYears))
	if strings.TrimSpace(candidate.CountryOfExperience) != "" {
		experienceText = fmt.Sprintf("%s years in %s", formatIntPointer(candidate.ExperienceYears), strings.ToUpper(candidate.CountryOfExperience))
	}
	
	pdf.MultiCell(
		112,
		5.5,
		fmt.Sprintf(
			"%s brings %s of domestic work experience, with strengths in %s and communication in %s.",
			candidateProfileName(candidate, nil),
			experienceText,
			formatListForPDF(skills),
			formatListForPDF(languages),
		),
		"",
		"L",
		false,
	)

	pdf.SetY(cardY + cardHeight + 5)
}

func (s *PDFService) drawHeroMetricChip(pdf *gofpdf.Fpdf, x, y, width float64, label, value string) {
	pdf.SetFillColor(224, 242, 254)
	pdf.SetDrawColor(125, 211, 252)
	pdf.RoundedRect(x, y, width, 12, 3, "1234", "DF")
	pdf.SetXY(x+2, y+2)
	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetTextColor(3, 105, 161)
	pdf.CellFormat(width-4, 3.5, strings.ToUpper(label), "", 1, "L", false, 0, "")
	pdf.SetX(x + 2)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(width-4, 4.5, value, "", 1, "L", false, 0, "")
}

func (s *PDFService) drawMetricCards(pdf *gofpdf.Fpdf, metrics [3][2]string) {
	cardY := pdf.GetY()
	cardWidth := 56.0
	positions := []float64{15, 77, 139}

	for index, metric := range metrics {
		x := positions[index]
		pdf.SetFillColor(241, 245, 249)
		pdf.SetDrawColor(203, 213, 225)
		pdf.RoundedRect(x, cardY, cardWidth, 18, 4, "1234", "DF")
		pdf.SetXY(x+3, cardY+3)
		pdf.SetFont("Arial", "B", 8)
		pdf.SetTextColor(3, 105, 161)
		pdf.CellFormat(cardWidth-6, 4, strings.ToUpper(metric[0]), "", 1, "L", false, 0, "")
		pdf.SetX(x + 3)
		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(15, 23, 42)
		pdf.MultiCell(cardWidth-6, 4.5, metric[1], "", "L", false)
	}

	pdf.SetY(cardY + 22)
}

func (s *PDFService) drawListPanels(pdf *gofpdf.Fpdf, languages, skills []string) {
	cardY := pdf.GetY()
	s.drawListPanel(pdf, 15, cardY, 86, "Languages", languages, [3]int{219, 234, 254}, [3]int{59, 130, 246}, "Communication languages captured for employer review.")
	s.drawListPanel(pdf, 109, cardY, 86, "Service Strengths", skills, [3]int{254, 243, 199}, [3]int{217, 119, 6}, "Core household strengths included in the candidate profile.")
	pdf.SetY(cardY + 43)
}

func (s *PDFService) drawListPanel(pdf *gofpdf.Fpdf, x, y, width float64, title string, values []string, fill [3]int, accent [3]int, emptyMessage string) {
	pdf.SetFillColor(fill[0], fill[1], fill[2])
	pdf.SetDrawColor(accent[0], accent[1], accent[2])
	pdf.RoundedRect(x, y, width, 40, 4, "1234", "DF")
	pdf.SetXY(x+4, y+4)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(accent[0], accent[1], accent[2])
	pdf.CellFormat(width-8, 5, title, "", 1, "L", false, 0, "")
	pdf.SetX(x + 4)
	pdf.SetFont("Arial", "", 9.5)
	pdf.SetTextColor(30, 41, 59)
	if len(values) == 0 {
		pdf.MultiCell(width-8, 5, emptyMessage, "", "L", false)
		return
	}
	pdf.MultiCell(width-8, 5, formatListForPDF(values), "", "L", false)
}

func (s *PDFService) drawSummaryBlock(pdf *gofpdf.Fpdf, candidate *domain.Candidate, branding CandidateCVBranding, languages, skills []string) {
	pdf.SetFillColor(250, 245, 255)
	pdf.SetDrawColor(192, 132, 252)
	pdf.RoundedRect(15, pdf.GetY(), 180, 28, 4, "1234", "DF")
	pdf.SetXY(19, pdf.GetY()+4)
	pdf.SetTextColor(88, 28, 135)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(172, 5, "Profile Overview", "", 1, "L", false, 0, "")
	pdf.SetX(19)
	pdf.SetTextColor(51, 65, 85)
	pdf.SetFont("Arial", "", 10)
	agencyName := strings.TrimSpace(branding.CompanyName)
	if agencyName == "" {
		agencyName = "the recruitment agency"
	}
	
	experienceText := fmt.Sprintf("%s years", formatIntPointer(candidate.ExperienceYears))
	if strings.TrimSpace(candidate.CountryOfExperience) != "" {
		experienceText = fmt.Sprintf("%s years in %s", formatIntPointer(candidate.ExperienceYears), candidate.CountryOfExperience)
	}
	
	pdf.MultiCell(
		172,
		5.5,
		fmt.Sprintf(
			"%s is being presented by %s with %s of experience, practical strengths in %s, and spoken languages including %s. The latest photo appears on this profile, and the passport copy is included at the end for final review.",
			candidateProfileName(candidate, nil),
			agencyName,
			experienceText,
			formatListForPDF(skills),
			formatListForPDF(languages),
		),
		"",
		"L",
		false,
	)
	pdf.SetY(pdf.GetY() + 2)
}

func pickCandidateDocuments(documents []*domain.Document) (photoDoc, passportDoc, videoDoc *domain.Document) {
	for _, document := range documents {
		if document == nil {
			continue
		}
		switch document.DocumentType {
		case domain.Photo:
			if photoDoc == nil {
				photoDoc = document
			}
		case domain.Passport:
			if passportDoc == nil {
				passportDoc = document
			}
		case domain.Video:
			if videoDoc == nil {
				videoDoc = document
			}
		}
	}
	return photoDoc, passportDoc, videoDoc
}

func parseLanguageEntries(raw json.RawMessage) []domain.LanguageEntry {
	if len(raw) == 0 || string(raw) == "null" {
		return []domain.LanguageEntry{}
	}

	var entries []domain.LanguageEntry
	if err := json.Unmarshal(raw, &entries); err == nil {
		return entries
	}

	var legacy []string
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return []domain.LanguageEntry{}
	}
	entries = make([]domain.LanguageEntry, 0, len(legacy))
	for _, s := range legacy {
		lang := strings.TrimSpace(s)
		if lang == "" {
			continue
		}
		entries = append(entries, domain.LanguageEntry{Language: lang, Proficiency: "Basic"})
	}
	return entries
}

func parseExperienceEntries(raw json.RawMessage) []domain.ExperienceEntry {
	if len(raw) == 0 || string(raw) == "null" {
		return []domain.ExperienceEntry{}
	}

	var entries []domain.ExperienceEntry
	if err := json.Unmarshal(raw, &entries); err == nil {
		return entries
	}

	var legacy string
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return []domain.ExperienceEntry{}
	}
	legacy = strings.TrimSpace(legacy)
	if legacy == "" {
		return []domain.ExperienceEntry{}
	}
	return []domain.ExperienceEntry{{Country: legacy, Years: 0}}
}

func parseJSONList(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	if len(values) == 0 {
		return []string{}
	}
	return values
}

func parseSkills(raw json.RawMessage) []domain.Skill {
	if len(raw) == 0 || string(raw) == "null" {
		return []domain.Skill{}
	}
	var skills []domain.Skill
	if err := json.Unmarshal(raw, &skills); err == nil {
		return skills
	}
	// Fallback to legacy string array
	var legacy []string
	if err := json.Unmarshal(raw, &legacy); err == nil {
		for _, s := range legacy {
			skills = append(skills, domain.Skill{Name: s, Value: true})
		}
	}
	return skills
}

func formatListForPDF(values []string) string {
	if len(values) == 0 {
		return "Not provided"
	}
	return strings.Join(values, ", ")
}

func formatIntPointer(value *int) string {
	if value == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", *value)
}

func (s *PDFService) appendPassportSection(pdf *gofpdf.Fpdf, passportDoc *domain.Document) {
	pdf.AddPage()
	passportBytes, passportContentType, err := s.fetchRemoteAsset(passportDoc.FileURL)
	if err != nil {
		return
	}
	drawAttachmentImageFrame(pdf, passportBytes, passportContentType, strings.TrimSpace(passportDoc.FileName), "passport_attachment")
}

func appendFullBodyPhotoSection(pdf *gofpdf.Fpdf, photoDoc *domain.Document, photoBytes []byte, photoContentType string) {
	pdf.AddPage()
	drawFullBodyPhotoFrame(pdf, photoBytes, photoContentType, compactValue(strings.TrimSpace(photoDoc.FileName), "photo"), "full_body_attachment")
}

func drawAttachmentHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFillColor(navyR, navyG, navyB)
	pdf.SetDrawColor(navyR, navyG, navyB)
	pdf.RoundedRect(15, 7, 180, 14, cardRadius, "1234", "DF")
	pdf.SetXY(15, 9.5)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 6, title, "", 1, "C", false, 0, "")
}

func drawAttachmentImageFrame(pdf *gofpdf.Fpdf, fileBytes []byte, contentType, fileName, imageID string) {
	pdf.SetFillColor(cardBgR, cardBgG, cardBgB)
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.RoundedRect(15, 24, 180, 250, cardRadius, "1234", "DF")

	if normalized := normalizeImageContentType(contentType); normalized != "" {
		_ = addImageFromBytesFit(pdf, fileBytes, normalized, imageID, 19, 28, 172, 242)
		return
	}

	pdf.SetY(130)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(180, 8, "Original document uploaded as PDF", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.MultiCell(156, 5, fmt.Sprintf("The original file \"%s\" is stored in the platform. Open it from the candidate page when you need the exact scanned PDF pages.", compactValue(fileName, "document")), "", "C", false)

	drawPageFooter(pdf)
}

func drawFullBodyPhotoFrame(pdf *gofpdf.Fpdf, fileBytes []byte, contentType, fileName, imageID string) {
	pdf.SetFillColor(cardBgR, cardBgG, cardBgB)
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.RoundedRect(40, 24, 120, 250, cardRadius, "1234", "DF")

	if normalized := normalizeImageContentType(contentType); normalized != "" {
		_ = addImageFromBytesFit(pdf, fileBytes, normalized, imageID, 44, 28, 112, 242)
		return
	}

	pdf.SetY(130)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(180, 8, "Original photo stored in the platform", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.MultiCell(136, 5, fmt.Sprintf("Open \"%s\" from the candidate page when you need the exact uploaded file.", compactValue(fileName, "photo")), "", "C", false)

	drawPageFooter(pdf)
}

func compactValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "Not provided" {
		return fallback
	}
	return trimmed
}

func candidateProfileName(candidate *domain.Candidate, passportData *domain.PassportData) string {
	if passportData != nil {
		if holderName := strings.TrimSpace(passportData.HolderName); holderName != "" {
			return holderName
		}
	}
	if candidate == nil {
		return ""
	}
	return strings.TrimSpace(candidate.FullName)
}

func manualCandidateUpper(value, fallback string) string {
	return compactValue(strings.ToUpper(strings.TrimSpace(value)), fallback)
}

func manualCandidateDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "N/A"
	}
	return value.UTC().Format("02-Jan-06")
}

func manualCandidateAgeValue(candidate *domain.Candidate) string {
	if candidate == nil || candidate.Age == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d YRS", *candidate.Age)
}

func formatChildrenCount(value *int) string {
	if value == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", *value)
}

func drawFitPDFText(pdf *gofpdf.Fpdf, x, y, width, height float64, value, family, style string, baseSize, minSize float64, align string) {
	display := compactValue(value, "-")
	fontSize := baseSize
	pdf.SetFont(family, style, fontSize)
	for fontSize > minSize && pdf.GetStringWidth(display) > width {
		fontSize -= 0.2
		pdf.SetFont(family, style, fontSize)
	}

	if pdf.GetStringWidth(display) > width {
		display = ellipsizePDFText(pdf, display, width)
	}

	lineHeight := fontSize * 0.42
	if lineHeight < 3.2 {
		lineHeight = 3.2
	}
	yOffset := y + ((height - lineHeight) / 2)
	if yOffset < y {
		yOffset = y
	}

	pdf.SetXY(x, yOffset)
	pdf.CellFormat(width, lineHeight, display, "", 1, align, false, 0, "")
}

func ellipsizePDFText(pdf *gofpdf.Fpdf, value string, width float64) string {
	if pdf.GetStringWidth(value) <= width {
		return value
	}

	runes := []rune(value)
	ellipsis := "..."
	for len(runes) > 0 {
		candidate := string(runes) + ellipsis
		if pdf.GetStringWidth(candidate) <= width {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}

	return ellipsis
}

func passportValue(passportData *domain.PassportData, getter func(*domain.PassportData) string) string {
	if passportData == nil || getter == nil {
		return ""
	}
	return strings.TrimSpace(getter(passportData))
}

func passportDateValue(passportData *domain.PassportData, getter func(*domain.PassportData) time.Time) string {
	if passportData == nil || getter == nil {
		return "N/A"
	}
	value := getter(passportData)
	if value.IsZero() {
		return "N/A"
	}
	return value.UTC().Format("02-Jan-06")
}

func passportDatePointerValue(passportData *domain.PassportData, getter func(*domain.PassportData) *time.Time) string {
	if passportData == nil || getter == nil {
		return "N/A"
	}
	value := getter(passportData)
	if value == nil || value.IsZero() {
		return "N/A"
	}
	return value.UTC().Format("02-Jan-06")
}

func passportRemainingYears(passportData *domain.PassportData) string {
	if passportData == nil || passportData.ExpiryDate.IsZero() {
		return "N/A"
	}
	return remainingYearsText(passportData.ExpiryDate)
}

func computeRemainingYears(expiry time.Time) string {
	if expiry.IsZero() {
		return "N/A"
	}
	return remainingYearsText(expiry)
}

func remainingYearsText(expiry time.Time) string {
	now := time.Now().UTC()
	expiry = expiry.UTC()
	if expiry.Before(now) {
		return "EXPIRED"
	}
	daysRemaining := int(expiry.Sub(now).Hours() / 24)
	if daysRemaining < 30 {
		if daysRemaining <= 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", daysRemaining)
	}

	monthsRemaining := int(expiry.Sub(now).Hours() / 24 / 30)
	if monthsRemaining < 12 {
		if monthsRemaining <= 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", monthsRemaining)
	}

	yearsRemaining := monthsRemaining / 12
	remainingMonths := monthsRemaining % 12
	if yearsRemaining <= 1 && remainingMonths == 0 {
		return "1 year"
	}
	if remainingMonths == 0 {
		return fmt.Sprintf("%d years", yearsRemaining)
	}
	if yearsRemaining <= 1 {
		if remainingMonths == 1 {
			return "1 year 1 month"
		}
		return fmt.Sprintf("1 year %d months", remainingMonths)
	}
	if remainingMonths == 1 {
		return fmt.Sprintf("%d years 1 month", yearsRemaining)
	}
	return fmt.Sprintf("%d years %d months", yearsRemaining, remainingMonths)
}

func shortReference(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= 8 {
		return strings.ToUpper(trimmed)
	}
	return strings.ToUpper(trimmed[:8])
}

func boolText(value bool) string {
	if value {
		return "YES"
	}
	return "NO"
}

func listContainsFold(values []string, target string) bool {
	target = strings.TrimSpace(strings.ToLower(target))
	if target == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(strings.ToLower(value)) == target {
			return true
		}
	}
	return false
}

func listContainsAnyFold(values []string, targets ...string) bool {
	for _, target := range targets {
		if listContainsFold(values, target) {
			return true
		}
	}
	return false
}

func formatCandidateStatusForPDF(status domain.CandidateStatus) string {
	trimmed := strings.TrimSpace(string(status))
	if trimmed == "" {
		return "Draft"
	}
	parts := strings.Split(trimmed, "_")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func currentPDFDate() string {
	return time.Now().Format("2/1/2006")
}

func formatExperienceLevel(value *int) string {
	if value == nil || *value == 0 {
		return "FIRST TIME"
	}
	return fmt.Sprintf("%d YEARS", *value)
}

func formatPassportDate(value time.Time) string {
	if value.IsZero() {
		return "Not available"
	}
	return value.UTC().Format("02 Jan 2006")
}

func formatPassportDatePointer(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "Not available"
	}
	return value.UTC().Format("02 Jan 2006")
}

func isPassportExpiringSoon(expiryDate time.Time) bool {
	if expiryDate.IsZero() {
		return false
	}
	return time.Now().UTC().AddDate(0, 6, 0).After(expiryDate.UTC())
}

func decodeLogoDataURL(dataURL string) ([]byte, string, error) {
	trimmed := strings.TrimSpace(dataURL)
	if trimmed == "" {
		return nil, "", nil
	}

	// Handle regular HTTP/HTTPS URLs by fetching the image remotely.
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		resp, err := http.Get(trimmed)
		if err != nil {
			return nil, "", fmt.Errorf("fetch logo url: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("fetch logo url: status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if err != nil {
			return nil, "", fmt.Errorf("read logo response: %w", err)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/png"
		}

		return body, contentType, nil
	}

	parts := strings.SplitN(trimmed, ",", 2)
	if len(parts) != 2 || !strings.Contains(parts[0], ";base64") {
		return nil, "", fmt.Errorf("invalid data url")
	}

	contentType := ""
	switch {
	case strings.Contains(parts[0], "image/png"):
		contentType = "image/png"
	case strings.Contains(parts[0], "image/jpeg"), strings.Contains(parts[0], "image/jpg"):
		contentType = "image/jpeg"
	default:
		return nil, "", fmt.Errorf("unsupported logo type")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", err
	}

	return decoded, contentType, nil
}

func (s *PDFService) fetchRemoteAsset(fileURL string) ([]byte, string, error) {
	if data, ct, ok := getCachedAsset(fileURL); ok {
		return data, ct, nil
	}

	reader, contentType, err := s.storageService.Open(fileURL)
	if err != nil {
		return nil, "", fmt.Errorf("open object from storage: %w", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(io.LimitReader(reader, maxRemoteAssetBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("read object body: %w", err)
	}
	if int64(len(body)) > maxRemoteAssetBytes {
		return nil, "", fmt.Errorf("download image: file too large")
	}

	setCachedAsset(fileURL, body, contentType)
	return body, contentType, nil
}

func (s *PDFService) fetchImage(fileURL string) ([]byte, string, error) {
	body, contentType, err := s.fetchRemoteAsset(fileURL)
	if err != nil {
		return nil, "", err
	}

	normalized := normalizeImageContentType(contentType)
	if normalized == "" {
		return nil, "", ErrMissingRequiredDocuments
	}

	return body, normalized, nil
}

func inferContentTypeFromURL(fileURL string) string {
	lowerURL := strings.ToLower(strings.TrimSpace(fileURL))
	switch {
	case strings.Contains(lowerURL, ".png"):
		return "image/png"
	case strings.Contains(lowerURL, ".jpg"), strings.Contains(lowerURL, ".jpeg"):
		return "image/jpeg"
	case strings.Contains(lowerURL, ".pdf"):
		return "application/pdf"
	default:
		return ""
	}
}

func normalizeImageContentType(contentType string) string {
	switch {
	case strings.HasPrefix(contentType, "image/png"):
		return "image/png"
	case strings.HasPrefix(contentType, "image/jpeg"), strings.HasPrefix(contentType, "image/jpg"):
		return "image/jpeg"
	default:
		return ""
	}
}

func addImageFromBytes(pdf *gofpdf.Fpdf, imageBytes []byte, contentType, imageID string, x, y, width, height float64) error {
	imageType := "JPG"
	if strings.EqualFold(contentType, "image/png") {
		imageType = "PNG"
	}

	options := gofpdf.ImageOptions{
		ImageType: imageType,
		ReadDpi:   true,
	}
	pdf.RegisterImageOptionsReader(imageID, options, bytes.NewReader(imageBytes))
	pdf.ImageOptions(imageID, x, y, width, height, false, options, 0, "")
	return nil
}

func addImageFromBytesFit(pdf *gofpdf.Fpdf, imageBytes []byte, contentType, imageID string, x, y, maxWidth, maxHeight float64) error {
	config, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil || config.Width == 0 || config.Height == 0 {
		return addImageFromBytes(pdf, imageBytes, contentType, imageID, x, y, maxWidth, maxHeight)
	}

	widthScale := maxWidth / float64(config.Width)
	heightScale := maxHeight / float64(config.Height)
	scale := widthScale
	if heightScale < scale {
		scale = heightScale
	}
	if scale <= 0 {
		scale = 1
	}

	width := float64(config.Width) * scale
	height := float64(config.Height) * scale
	offsetX := x + (maxWidth-width)/2
	offsetY := y + (maxHeight-height)/2
	return addImageFromBytes(pdf, imageBytes, contentType, imageID, offsetX, offsetY, width, height)
}

// =============================================================================
// REDESIGN UTILITY CONSTANTS
// =============================================================================

var (
	navyR, navyG, navyB             = 27, 42, 74    // #1B2A4A
	blueR, blueG, blueB             = 37, 99, 235   // #2563EB
	cardHeaderR, cardHeaderG, cardHeaderB = 30, 58, 138 // #1E3A8A
	emeraldR, emeraldG, emeraldB    = 5, 150, 105   // #059669
	amberR, amberG, amberB          = 217, 119, 6   // #D97706
	alertR, alertG, alertB          = 220, 38, 38   // #DC2626
	pageBgR, pageBgG, pageBgB       = 248, 250, 252 // #F8FAFC
	cardBgR, cardBgG, cardBgB       = 255, 255, 255 // #FFFFFF
	borderR, borderG, borderB       = 226, 232, 240 // #E2E8F0
	mutedR, mutedG, mutedB          = 100, 116, 139 // #64748B
	darkR, darkG, darkB             = 30, 41, 59    // #1E293B
	headerBgR, headerBgG, headerBgB = 241, 245, 249 // #F1F5F9
	placeholderR, placeholderG, placeholderB = 241, 245, 249 // #F1F5F9
	cardRadius                      = 4.0
)

// =============================================================================
// REDESIGN DRAW HELPERS
// =============================================================================

func drawCard(pdf *gofpdf.Fpdf, x, y, w, h float64) {
	pdf.SetFillColor(cardBgR, cardBgG, cardBgB)
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.RoundedRect(x, y, w, h, cardRadius, "1234", "DF")
}

func drawCardHeader(pdf *gofpdf.Fpdf, x, y, w, h float64, iconName, title string, bgR, bgG, bgB, textR, textG, textB int) {
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.RoundedRect(x, y, w, h, cardRadius, "12", "F")

	textX := x + 4
	if iconName != "" {
		drawAssetIcon(pdf, iconName, x+3, y+(h-5)/2, 5, 5)
		textX += 7
	}

	pdf.SetXY(textX, y+(h-8)/2)
	pdf.SetTextColor(textR, textG, textB)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(w-(textX-x)-4, 8, title, "", 1, "L", false, 0, "")
}

func drawLabelValue(pdf *gofpdf.Fpdf, x, y, labelW, valW, rowH float64, label, value string, valBaseSize, valMinSize float64) {
	pdf.SetXY(x, y)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.SetFont("Arial", "", 7)
	pdf.CellFormat(labelW, rowH, compactValue(label, "-"), "", 0, "L", false, 0, "")

	pdf.SetTextColor(darkR, darkG, darkB)
	drawFitPDFText(pdf, x+labelW, y, valW, rowH, compactValue(value, "N/A"), "Arial", "B", valBaseSize, valMinSize, "L")
}

func drawPhotoPlaceholder(pdf *gofpdf.Fpdf, x, y, w, h float64) {
	pdf.SetFillColor(placeholderR, placeholderG, placeholderB)
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.RoundedRect(x, y, w, h, cardRadius, "1234", "DF")

	pdf.SetXY(x, y+h/2-4)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(w, 8, "PHOTO", "", 1, "C", false, 0, "")
}

func drawBadgePill(pdf *gofpdf.Fpdf, x, y, w, h float64, text string, bgR, bgG, bgB int) {
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.RoundedRect(x, y, w, h, h/2, "1234", "F")
	pdf.SetXY(x, y)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 7)
	pdf.CellFormat(w, h, text, "", 1, "C", false, 0, "")
}

func drawPageFooter(pdf *gofpdf.Fpdf) {
	footerY := 285.0
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, footerY, 195, footerY)
	pdf.SetLineWidth(0.2)

	pdf.SetTextColor(148, 163, 184)
	pdf.SetFont("Arial", "", 6.5)

	pdf.SetXY(15, footerY+2)
	pdf.CellFormat(50, 4, "Maid Recruitment Platform", "", 0, "L", false, 0, "")

	pdf.SetX(90)
	pdf.CellFormat(30, 4, "Generated: "+time.Now().UTC().Format("2/1/2006"), "", 0, "C", false, 0, "")

	pdf.SetX(165)
	pdf.CellFormat(30, 4, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "R", false, 0, "")
}

func drawSectionDivider(pdf *gofpdf.Fpdf, x, y, w float64) {
	pdf.SetDrawColor(borderR, borderG, borderB)
	pdf.SetLineWidth(0.3)
	pdf.Line(x, y, x+w, y)
	pdf.SetLineWidth(0.2)
}
