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

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateCandidateCV(candidate *domain.Candidate, documents []*domain.Document, branding CandidateCVBranding) ([]byte, error) {
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
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-8.5)
		pdf.SetFont("Arial", "", 8.5)
		pdf.SetTextColor(100, 116, 139)
		pdf.CellFormat(0, 4, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	s.drawBrandingHeader(pdf, branding)
	s.drawHeader(pdf, branding)

	photoBytes, photoContentType, err := fetchImage(photoDoc.FileURL)
	if err != nil {
		return nil, fmt.Errorf("fetch photo: %w", err)
	}
	s.drawApplicationProfileSheet(pdf, candidate, branding, photoBytes, photoContentType, languages, skills)

	appendPassportSection(pdf, passportDoc)
	appendFullBodyPhotoSection(pdf, photoDoc, photoBytes, photoContentType)

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPDFGenerationFailed, err)
	}

	return buffer.Bytes(), nil
}

func (s *PDFService) drawBrandingHeader(pdf *gofpdf.Fpdf, branding CandidateCVBranding) {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(209, 213, 219)
	pdf.RoundedRect(15, 12, 180, 18, 1.5, "1234", "DF")
	pdf.SetXY(15, 18)
	pdf.SetTextColor(101, 163, 13)
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(180, 6, "Application Form", "", 1, "C", false, 0, "")
	pdf.SetY(35)
}

func (s *PDFService) drawHeader(pdf *gofpdf.Fpdf, branding CandidateCVBranding) {
	headerY := pdf.GetY()
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(209, 213, 219)
	pdf.RoundedRect(15, headerY, 180, 34, 1.5, "1234", "DF")

	pdf.SetXY(18, headerY+5)
	pdf.SetTextColor(234, 88, 12)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(48, 5, compactValue(strings.TrimSpace(branding.CompanyName), "Agency Profile"), "", 1, "L", false, 0, "")
	pdf.SetX(18)
	pdf.SetTextColor(75, 85, 99)
	pdf.SetFont("Arial", "", 9)
	companyName := strings.TrimSpace(branding.CompanyName)
	if companyName == "" {
		companyName = "Agency Recruitment Profile"
	}
	pdf.MultiCell(48, 4.5, fmt.Sprintf("%s\nPrepared candidate review package", companyName), "", "L", false)

	logoX := 77.0
	logoY := headerY + 1.5
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(229, 231, 235)
	pdf.RoundedRect(77, headerY+1.5, 56, 31, 15, "1234", "DF")
	if logoBytes, logoContentType, err := decodeLogoDataURL(branding.LogoDataURL); err == nil && len(logoBytes) > 0 {
		_ = addImageFromBytesFit(pdf, logoBytes, logoContentType, "agency_logo_header", logoX+8, logoY+4, 40, 23)
	} else {
		pdf.SetFillColor(14, 165, 233)
		pdf.RoundedRect(93, headerY+6, 24, 21, 4, "1234", "F")
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 12)
		pdf.SetXY(93, headerY+13)
		pdf.CellFormat(24, 6, "AG", "", 0, "C", false, 0, "")
	}

	pdf.SetXY(143, headerY+5)
	pdf.SetTextColor(234, 88, 12)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(46, 5, "Placement Ready", "", 1, "R", false, 0, "")
	pdf.SetX(143)
	pdf.SetTextColor(75, 85, 99)
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(46, 4.5, "Candidate review file with passport and full-body photo attached after page one.", "", "R", false)
	pdf.SetY(headerY + 40)
}

func (s *PDFService) drawApplicationProfileSheet(
	pdf *gofpdf.Fpdf,
	candidate *domain.Candidate,
	branding CandidateCVBranding,
	photoBytes []byte,
	photoContentType string,
	languages []string,
	skills []string,
) {
	startY := pdf.GetY()
	pdf.SetDrawColor(209, 213, 219)
	pdf.RoundedRect(15, startY, 180, 182, 1.5, "1234", "D")

	pdf.SetFillColor(243, 244, 246)
	pdf.Rect(15, startY, 180, 10, "F")
	pdf.SetXY(15, startY+3)
	pdf.SetTextColor(75, 85, 99)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(180, 4, "Registration Details", "", 1, "C", false, 0, "")

	topGridY := startY + 14
	s.drawCompactField(pdf, 18, topGridY, 54, 12, "Date", currentPDFDate())
	s.drawCompactField(pdf, 78, topGridY, 54, 12, "Applied For", "HOUSEMAID")
	s.drawCompactField(pdf, 138, topGridY, 54, 12, "Applied Country", "JORDAN")
	s.drawCompactField(pdf, 18, topGridY+16, 54, 12, "Experience", formatExperienceLevel(candidate.ExperienceYears))
	s.drawCompactField(pdf, 78, topGridY+16, 54, 12, "Languages", fmt.Sprintf("%d tracked", len(languages)))
	s.drawCompactField(pdf, 138, topGridY+16, 54, 12, "Status", formatCandidateStatusForPDF(candidate.Status))

	nameY := topGridY + 36
	pdf.SetFillColor(255, 247, 237)
	pdf.SetDrawColor(251, 191, 36)
	pdf.RoundedRect(18, nameY, 174, 10, 1.5, "1234", "DF")
	pdf.SetXY(20, nameY+2.5)
	pdf.SetTextColor(55, 65, 81)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(170, 5, strings.ToUpper(candidate.FullName), "", 1, "C", false, 0, "")

	contentY := nameY + 14
	s.drawWrappedField(pdf, 18, contentY, 50, 64, "Applicant Details", []string{
		fmt.Sprintf("Age: %s", formatIntPointer(candidate.Age)),
		fmt.Sprintf("Experience: %s years", formatIntPointer(candidate.ExperienceYears)),
		fmt.Sprintf("Languages: %s", compactValue(formatListForPDF(languages), "Not provided")),
		fmt.Sprintf("Agency: %s", compactValue(strings.TrimSpace(branding.CompanyName), "Agency profile")),
	})

	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(156, 163, 175)
	pdf.Rect(72, contentY, 68, 92, "D")
	_ = addImageFromBytesFit(pdf, photoBytes, photoContentType, "candidate_sheet_photo", 74, contentY+2, 64, 88)

	s.drawWrappedField(pdf, 144, contentY, 48, 64, "Document Status", []string{
		"Passport: Attached",
		"Full body photo: Attached",
		fmt.Sprintf("Skills: %s", compactValue(formatListForPDF(skills), "Not provided")),
		fmt.Sprintf("Applied for: %s", "Housemaid"),
	})

	s.drawWrappedField(pdf, 18, contentY+70, 50, 24, "Profile Notes", []string{
		"Prepared for employer review.",
		"Latest uploaded logo is used in this header.",
	})

	s.drawWrappedField(pdf, 144, contentY+70, 48, 24, "Quick Summary", []string{
		fmt.Sprintf("Languages: %s", compactValue(formatListForPDF(languages), "Not provided")),
		fmt.Sprintf("Skills: %s", compactValue(formatListForPDF(skills), "Not provided")),
	})

	bottomY := contentY + 100
	s.drawWideField(pdf, 18, bottomY, 82, 30, "Language Known", compactValue(formatListForPDF(languages), "Not provided"))
	s.drawWideField(pdf, 110, bottomY, 82, 30, "Work Experience", compactValue(formatListForPDF(skills), "Not provided"))
	s.drawWideField(pdf, 18, bottomY+34, 174, 18, "Remark", "Candidate profile prepared for employer review.")
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
	pdf.CellFormat(116, 8, candidate.FullName, "", 1, "L", false, 0, "")

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
			candidate.FullName,
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
			candidate.FullName,
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
	drawAttachmentHeader(pdf, "Passport Copy")
	passportBytes, passportContentType, err := fetchRemoteAsset(passportDoc.FileURL)
	if err != nil {
		return
	}
	drawAttachmentImageFrame(pdf, passportBytes, passportContentType, strings.TrimSpace(passportDoc.FileName), "passport_attachment")
}

func appendFullBodyPhotoSection(pdf *gofpdf.Fpdf, photoDoc *domain.Document, photoBytes []byte, photoContentType string) {
	pdf.AddPage()
	title := "Full Body Photo"
	if photoDoc != nil && strings.TrimSpace(photoDoc.FileName) != "" {
		title = "Full Body Photo"
	}
	drawAttachmentHeader(pdf, title)
	drawAttachmentImageFrame(pdf, photoBytes, photoContentType, compactValue(strings.TrimSpace(photoDoc.FileName), "photo"), "full_body_attachment")
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
	pdf.Rect(24, 40, 162, 235, "D")

	if normalized := normalizeImageContentType(contentType); normalized != "" {
		_ = addImageFromBytesFit(pdf, fileBytes, normalized, imageID, 28, 44, 154, 227)
		return
	}

	pdf.SetXY(34, 132)
	pdf.SetTextColor(31, 41, 55)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(142, 8, "Original document uploaded as PDF", "", 1, "C", false, 0, "")
	pdf.SetX(38)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(75, 85, 99)
	pdf.MultiCell(134, 6, fmt.Sprintf("The original file \"%s\" is stored in the platform. Open it from the candidate page when you need the exact scanned PDF pages.", compactValue(fileName, "document")), "", "C", false)
}

func compactValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "Not provided" {
		return fallback
	}
	return trimmed
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
	response, err := http.Get(fileURL)
	if err != nil {
		return nil, "", fmt.Errorf("download image: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download image: unexpected status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read image body: %w", err)
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
