package service

import (
	"bytes"
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
	"time"

	"github.com/jung-kurt/gofpdf"

	"maid-recruitment-tracking/internal/domain"
)

var (
	ErrMissingRequiredDocuments = errors.New("missing required documents")
	ErrPDFGenerationFailed      = errors.New("pdf generation failed")
)

const maxRemoteAssetBytes int64 = 55 << 20

var remoteAssetHTTPClient = &http.Client{Timeout: 15 * time.Second}

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateCandidateCV(candidate *domain.Candidate, documents []*domain.Document, branding CandidateCVBranding, passportData *domain.PassportData) ([]byte, error) {
	if candidate == nil {
		return nil, fmt.Errorf("candidate is nil")
	}

	photoDoc, passportDoc, _ := pickCandidateDocuments(documents)
	if photoDoc == nil || passportDoc == nil || strings.TrimSpace(photoDoc.FileURL) == "" || strings.TrimSpace(passportDoc.FileURL) == "" {
		return nil, ErrMissingRequiredDocuments
	}

	languages := parseJSONList(candidate.Languages)
	skills := parseJSONList(candidate.Skills)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Candidate CV", false)
	pdf.SetAuthor("Maid Recruitment Platform", false)
	pdf.SetMargins(15, 15, 15)
	pdf.SetFont("Arial", "", 12)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	s.drawBrandingHeader(pdf, branding)
	s.drawHeader(pdf, branding)

	photoBytes, photoContentType, err := fetchImage(photoDoc.FileURL)
	if err != nil {
		return nil, fmt.Errorf("fetch photo: %w", err)
	}
	s.drawApplicationProfileSheet(pdf, candidate, branding, passportData, photoBytes, photoContentType, languages, skills)

	appendPassportSection(pdf, passportDoc)
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
	pdf.SetDrawColor(196, 196, 196)
	pdf.Rect(12, 12, 186, 18, "D")
	pdf.SetXY(12, 18)
	pdf.SetTextColor(92, 130, 54)
	pdf.SetFont("Arial", "B", 19)
	pdf.CellFormat(186, 6, "Application Form", "", 1, "C", false, 0, "")
	pdf.SetY(34)
}

func (s *PDFService) drawHeader(pdf *gofpdf.Fpdf, branding CandidateCVBranding) {
	headerY := pdf.GetY()
	pdf.SetFillColor(244, 244, 244)
	pdf.SetDrawColor(196, 196, 196)
	pdf.Rect(12, headerY, 186, 28, "DF")

	companyName := compactValue(strings.TrimSpace(branding.CompanyName), "Maid Recruitment Agency")

	pdf.SetXY(16, headerY+4)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(50, 4.5, companyName, "", 1, "L", false, 0, "")
	pdf.SetX(16)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "", 7.2)
	pdf.MultiCell(50, 3.7, "Ethiopia, Addis Ababa\nPrepared for employer review\nEmail delivery and CV package ready", "", "L", false)

	logoX := 82.0
	logoY := headerY + 2
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(214, 214, 214)
	pdf.RoundedRect(82, logoY, 46, 24, 12, "1234", "DF")
	if logoBytes, logoContentType, err := decodeLogoDataURL(branding.LogoDataURL); err == nil && len(logoBytes) > 0 {
		_ = addImageFromBytesFit(pdf, logoBytes, logoContentType, "agency_logo_header", logoX+7, logoY+2.5, 32, 19)
	} else {
		pdf.SetTextColor(230, 147, 19)
		pdf.SetFont("Arial", "B", 22)
		pdf.SetXY(82, headerY+8)
		pdf.CellFormat(46, 7, "A", "", 1, "C", false, 0, "")
	}

	pdf.SetXY(140, headerY+4)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(50, 4.5, "Foreign Employment File", "", 1, "R", false, 0, "")
	pdf.SetX(140)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "", 7.2)
	pdf.MultiCell(50, 3.7, "Housemaid placement profile\nPassport and body photo attached\nReady for review and approval", "", "R", false)
	pdf.SetY(headerY + 31)
}

func (s *PDFService) drawApplicationProfileSheet(
	pdf *gofpdf.Fpdf,
	candidate *domain.Candidate,
	branding CandidateCVBranding,
	passportData *domain.PassportData,
	photoBytes []byte,
	photoContentType string,
	languages []string,
	skills []string,
) {
	startY := pdf.GetY()
	pdf.SetDrawColor(191, 191, 191)
	pdf.Rect(12, startY, 186, 223, "D")

	pdf.SetFillColor(230, 230, 230)
	pdf.Rect(15, startY+3, 180, 7, "F")
	pdf.SetXY(15, startY+4.4)
	pdf.SetTextColor(92, 104, 77)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(180, 3.5, "Registration Details", "", 1, "C", false, 0, "")

	regY := startY + 13
	s.drawFormInfoRow(pdf, 15, regY, 55, 22, 7, [][2]string{
		{"Date", currentPDFDate()},
		{"Contract Period", "2YRS"},
		{"Bio. Met. ID", "-"},
	})
	s.drawFormInfoRow(pdf, 75, regY, 55, 22, 7, [][2]string{
		{"Applied for", "HOUSEMAID"},
		{"Monthly Sal", "1000"},
		{"Ref. No", shortReference(candidate.ID)},
	})
	s.drawFormInfoRow(pdf, 135, regY, 60, 25, 7, [][2]string{
		{"Applied Cont.", "K S A"},
		{"B.Met Rg./s.", "-"},
		{"Experience", formatExperienceLevel(candidate.ExperienceYears)},
	})

	nameY := regY + 24
	s.drawNamedBand(pdf, 15, nameY, 180, strings.ToUpper(compactValue(candidateProfileName(candidate, passportData), "CANDIDATE PROFILE")))

	leftX, leftW := 15.0, 60.0
	midX, midW := 80.0, 46.0
	rightX, rightW := 131.0, 64.0
	detailY := nameY + 10

	ageValue := manualCandidateAgeValue(candidate)
	applicantRows := []pdfFormRow{
		{Label: "Nationality", Value: manualCandidateUpper(candidate.Nationality, "N/A")},
		{Label: "Date Of Birth", Value: manualCandidateDate(candidate.DateOfBirth)},
		{Label: "Age", Value: ageValue},
		{Label: "Place Of Birth", Value: manualCandidateUpper(candidate.PlaceOfBirth, "N/A")},
		{Label: "Religion", Value: compactValue(strings.ToUpper(strings.TrimSpace(candidate.Religion)), "N/A")},
		{Label: "Marital Status", Value: compactValue(strings.ToUpper(strings.TrimSpace(candidate.MaritalStatus)), "N/A")},
		{Label: "Children", Value: formatChildrenCount(candidate.ChildrenCount)},
		{Label: "Education Level", Value: compactValue(strings.ToUpper(strings.TrimSpace(candidate.EducationLevel)), "N/A")},
	}
	s.drawFormTable(pdf, leftX, detailY, leftW, "Applicants Detail", applicantRows, 26)

	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(191, 191, 191)
	pdf.Rect(midX, detailY, midW, 82, "D")
	_ = addImageFromBytesFit(pdf, photoBytes, photoContentType, "candidate_sheet_photo", midX+2, detailY+2, midW-4, 78)

	passportRows := []pdfFormRow{
		{Label: "Passport No.", Value: compactValue(strings.ToUpper(passportValue(passportData, func(data *domain.PassportData) string { return data.PassportNumber })), "NOT SCANNED")},
		{Label: "Issue Date", Value: passportDatePointerValue(passportData, func(data *domain.PassportData) *time.Time { return data.IssueDate })},
		{Label: "Expiry Date", Value: passportDateValue(passportData, func(data *domain.PassportData) time.Time { return data.ExpiryDate })},
		{Label: "Remaining Year", Value: passportRemainingYears(passportData)},
	}
	s.drawFormTable(pdf, rightX, detailY, rightW, "Passport Detail", passportRows, 27)

	langRows := []pdfFormRow{
		{Label: "Arabic", Value: boolText(listContainsFold(languages, "Arabic"))},
		{Label: "English", Value: boolText(listContainsFold(languages, "English"))},
		{Label: "French", Value: boolText(listContainsFold(languages, "French"))},
	}
	s.drawFormTable(pdf, rightX, detailY+40, rightW, "Language Known", langRows, 24)

	experienceRows := []pdfFormRow{
		{Label: "Baby Sitting", Value: boolText(listContainsAnyFold(skills, "Childcare"))},
		{Label: "Cleaning", Value: boolText(listContainsAnyFold(skills, "Cleaning"))},
		{Label: "Ironing", Value: boolText(listContainsAnyFold(skills, "Ironing", "Laundry"))},
		{Label: "Cooking", Value: boolText(listContainsAnyFold(skills, "Cooking"))},
		{Label: "Home Teaching", Value: boolText(listContainsAnyFold(skills, "Teaching"))},
		{Label: "Personal Care", Value: boolText(listContainsAnyFold(skills, "Elderly Care", "First Aid", "Pet Care"))},
	}
	bottomY := detailY + 86
	s.drawFormTable(pdf, rightX, bottomY, rightW, "Work Experience", experienceRows, 30)

	expY := bottomY
	s.drawExperiencedAbroadSection(pdf, 15, expY, 111, [][2]string{
		{"NO", "-"},
		{"NO", "-"},
		{"NO", "-"},
	})

	remarkY := expY + 46
	s.drawRemarkSection(pdf, 15, remarkY, 111, "Remark", "")
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
	sectionGap := 10.0
	sectionWidth := (width - sectionGap) / 2
	leftX := x
	rightX := x + sectionWidth + sectionGap

	s.drawExperienceColumn(pdf, leftX, y, sectionWidth, "Experienced Abroad", "Country", extractExperienceColumn(rows, 0))
	s.drawExperienceColumn(pdf, rightX, y, sectionWidth, "Experience Abroad Duration", "How Long", extractExperienceColumn(rows, 1))
}

func (s *PDFService) drawExperienceColumn(pdf *gofpdf.Fpdf, x, y, width float64, title, label string, values []string) {
	titleHeight := 8.0
	rowHeight := 12.0

	pdf.SetFillColor(245, 245, 245)
	pdf.SetDrawColor(191, 191, 191)
	pdf.Rect(x, y, width, titleHeight, "DF")
	pdf.SetXY(x, y+2)
	pdf.SetTextColor(233, 122, 45)
	pdf.SetFont("Arial", "B", 8.4)
	pdf.CellFormat(width, 4, title, "", 1, "C", false, 0, "")

	currentY := y + titleHeight + 4
	for _, value := range values {
		s.drawFormKeyValueRow(pdf, x, currentY, width, width*0.54, rowHeight, label, value, false)
		currentY += rowHeight + 4
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
	pdf.MultiCell(
		112,
		5.5,
		fmt.Sprintf(
			"%s brings %s years of domestic work experience, with strengths in %s and communication in %s.",
			candidateProfileName(candidate, nil),
			formatIntPointer(candidate.ExperienceYears),
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
	pdf.MultiCell(
		172,
		5.5,
		fmt.Sprintf(
			"%s is being presented by %s with %s years of experience, practical strengths in %s, and spoken languages including %s. The latest photo appears on this profile, and the passport copy is included at the end for final review.",
			candidateProfileName(candidate, nil),
			agencyName,
			formatIntPointer(candidate.ExperienceYears),
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

func appendPassportSection(pdf *gofpdf.Fpdf, passportDoc *domain.Document) {
	pdf.AddPage()
	passportBytes, passportContentType, err := fetchRemoteAsset(passportDoc.FileURL)
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
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(209, 213, 219)
	pdf.RoundedRect(15, 12, 180, 16, 1.5, "1234", "DF")
	pdf.SetXY(15, 17)
	pdf.SetTextColor(55, 65, 81)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(180, 6, title, "", 1, "C", false, 0, "")
}

func drawAttachmentImageFrame(pdf *gofpdf.Fpdf, fileBytes []byte, contentType, fileName, imageID string) {
	pdf.SetDrawColor(107, 114, 128)
	pdf.Rect(16, 24, 178, 236, "D")

	if normalized := normalizeImageContentType(contentType); normalized != "" {
		_ = addImageFromBytesFit(pdf, fileBytes, normalized, imageID, 20, 28, 170, 228)
		return
	}

	pdf.SetXY(32, 124)
	pdf.SetTextColor(31, 41, 55)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(146, 8, "Original document uploaded as PDF", "", 1, "C", false, 0, "")
	pdf.SetX(36)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(75, 85, 99)
	pdf.MultiCell(138, 6, fmt.Sprintf("The original file \"%s\" is stored in the platform. Open it from the candidate page when you need the exact scanned PDF pages.", compactValue(fileName, "document")), "", "C", false)
}

func drawFullBodyPhotoFrame(pdf *gofpdf.Fpdf, fileBytes []byte, contentType, fileName, imageID string) {
	pdf.SetDrawColor(107, 114, 128)
	pdf.Rect(34, 26, 132, 220, "D")

	if normalized := normalizeImageContentType(contentType); normalized != "" {
		_ = addImageFromBytesFit(pdf, fileBytes, normalized, imageID, 40, 30, 120, 212)
		return
	}

	pdf.SetXY(46, 132)
	pdf.SetTextColor(31, 41, 55)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(108, 8, "Original photo stored in the platform", "", 1, "C", false, 0, "")
	pdf.SetX(46)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(75, 85, 99)
	pdf.MultiCell(108, 6, fmt.Sprintf("Open \"%s\" from the candidate page when you need the exact uploaded file.", compactValue(fileName, "photo")), "", "C", false)
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
	now := time.Now().UTC()
	expiry := passportData.ExpiryDate.UTC()
	if expiry.Before(now) {
		return "EXPIRED"
	}
	years := expiry.Year() - now.Year()
	anniversary := time.Date(now.Year(), expiry.Month(), expiry.Day(), 0, 0, 0, 0, time.UTC)
	if now.After(anniversary) {
		years--
	}
	if years < 0 {
		years = 0
	}
	if years == 0 {
		months := int(expiry.Sub(now).Hours() / 24 / 30)
		if months < 1 {
			return "<1 YR"
		}
		return fmt.Sprintf("%d MOS", months)
	}
	return fmt.Sprintf("%d YRS", years)
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

func fetchRemoteAsset(fileURL string) ([]byte, string, error) {
	response, err := remoteAssetHTTPClient.Get(fileURL)
	if err != nil {
		return nil, "", fmt.Errorf("download image: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download image: unexpected status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxRemoteAssetBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("read image body: %w", err)
	}
	if int64(len(body)) > maxRemoteAssetBytes {
		return nil, "", fmt.Errorf("download image: file too large")
	}

	contentType := strings.ToLower(strings.TrimSpace(strings.SplitN(response.Header.Get("Content-Type"), ";", 2)[0]))
	if contentType == "" {
		contentType = inferContentTypeFromURL(fileURL)
	}
	return body, contentType, nil
}

func fetchImage(fileURL string) ([]byte, string, error) {
	body, contentType, err := fetchRemoteAsset(fileURL)
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
