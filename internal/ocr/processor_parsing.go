package ocr

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func extractMRZLinesFromText(text string) (string, string) {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ToUpper(text)
	text = strings.ReplaceAll(text, "Â«", "<")
	text = strings.ReplaceAll(text, "â€¹", "<")
	text = strings.ReplaceAll(text, "â€º", "<")

	candidates := make([]string, 0, 32)
	for _, raw := range strings.Split(text, "\n") {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		s = strings.ReplaceAll(s, " ", "")
		b := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '<' {
				b = append(b, c)
			}
		}
		s = string(b)
		if len(s) < 30 {
			continue
		}
		if idx := strings.Index(s, "P<"); idx > 0 {
			s = s[idx:]
		}
		if len(s) == 44 {
			candidates = append(candidates, s)
			continue
		}
		if len(s) > 44 {
			for i := 0; i+44 <= len(s); i++ {
				cand := s[i : i+44]
				if strings.Count(cand, "<") < 5 {
					continue
				}
				candidates = append(candidates, cand)
			}
			continue
		}
		if len(s) >= 40 && len(s) < 44 && strings.HasPrefix(s, "P<") {
			candidates = append(candidates, s+strings.Repeat("<", 44-len(s)))
		}
	}

	if len(candidates) == 0 {
		return splitMRZText(text)
	}

	parser := NewMRZParser()
	bestL1, bestL2 := "", ""
	bestErrors := 1 << 30

	for _, l1 := range candidates {
		if len(l1) != 44 || l1[0] != 'P' {
			continue
		}
		for _, l2 := range candidates {
			if len(l2) != 44 || l2[0] == 'P' {
				continue
			}
			data, err := parser.ParseMRZ(l1, l2)
			if err != nil {
				continue
			}
			if data.IsValid {
				return l1, l2
			}
			if len(data.ValidationErrors) < bestErrors {
				bestErrors = len(data.ValidationErrors)
				bestL1, bestL2 = l1, l2
			}
		}
	}

	if bestL1 != "" && bestL2 != "" {
		return bestL1, bestL2
	}
	if len(candidates) >= 2 {
		return candidates[0], candidates[1]
	}
	return "", ""
}

func splitMRZText(text string) (string, string) {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "\r", "")
	lines := strings.Split(text, "\n")

	out := make([]string, 0, 2)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, strings.ToUpper(line))
	}
	if len(out) < 2 {
		return "", ""
	}
	return out[0], out[1]
}

func extractVisualZoneFields(text string, overallConf float64) *VisualZoneData {
	vz := &VisualZoneData{}
	vz.PlaceOfBirth, vz.PlaceOfBirthConf = findLabelValue(text, []string{"PLACE OF BIRTH", "BIRTH PLACE", "PLACE OF BIR", "PLACE OF BIRTHH"}, overallConf)
	if vz.PlaceOfBirth == "" {
		if pb := findPlaceOfBirthFallback(text); pb != "" {
			vz.PlaceOfBirth = pb
			vz.PlaceOfBirthConf = overallConf
		}
	}
	vz.Authority, vz.AuthorityConf = findLabelValue(text, []string{"ISSUING AUTHORITY", "AUTHORITY", "ISSUING AUTH"}, overallConf)
	if vz.Authority == "" {
		if a := findAuthorityFallback(text); a != "" {
			vz.Authority = a
			vz.AuthorityConf = overallConf
		}
	}
	d, dc := findDateOfIssue(text, overallConf)
	vz.DateOfIssue = d
	vz.DateOfIssueConf = dc
	return vz
}

func findLabelValue(text string, labels []string, conf float64) (string, float64) {
	upper := strings.ToUpper(text)
	origLines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	upperLines := strings.Split(strings.ReplaceAll(upper, "\r", ""), "\n")

	labelsUpper := make([]string, 0, len(labels))
	labelsNorm := make([]string, 0, len(labels))
	for _, label := range labels {
		value := strings.TrimSpace(strings.ToUpper(label))
		if value == "" {
			continue
		}
		labelsUpper = append(labelsUpper, value)
		labelsNorm = append(labelsNorm, normalizeLettersDigits(value))
	}

	for i := 0; i < len(upperLines); i++ {
		lUpper := strings.TrimSpace(upperLines[i])
		lOrig := strings.TrimSpace(origLines[i])
		if lUpper == "" {
			continue
		}

		for li := 0; li < len(labelsUpper); li++ {
			label := labelsUpper[li]
			labelNorm := labelsNorm[li]

			if idx := strings.Index(lUpper, label); idx >= 0 {
				vUpper := strings.TrimSpace(strings.TrimPrefix(lUpper[idx:], label))
				vUpper = strings.TrimLeft(vUpper, ":- ")
				vUpper = strings.TrimSpace(vUpper)
				if vUpper != "" {
					return cleanupVisualValue(vUpper), conf
				}
				if next := nextNonEmptyLine(origLines, i+1, 3); next != "" {
					return cleanupVisualValue(next), conf
				}
				continue
			}

			lNorm := normalizeLettersDigits(lUpper)
			if labelNorm != "" && strings.Contains(lNorm, labelNorm) {
				if parts := strings.SplitN(lOrig, ":", 2); len(parts) == 2 {
					if v := strings.TrimSpace(parts[1]); v != "" {
						return cleanupVisualValue(v), conf
					}
				}
				if parts := strings.SplitN(lOrig, "-", 2); len(parts) == 2 {
					if v := strings.TrimSpace(parts[1]); v != "" {
						return cleanupVisualValue(v), conf
					}
				}
				if next := nextNonEmptyLine(origLines, i+1, 3); next != "" {
					return cleanupVisualValue(next), conf
				}
			}
		}
	}
	return "", 0
}

func cleanupVisualValue(value string) string {
	value = strings.ReplaceAll(value, "<", " ")
	value = strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, value)
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return ""
	}
	keep := make([]string, 0, len(parts))
	for _, part := range parts {
		up := strings.ToUpper(part)
		if up == "OF" || up == "IN" || up == "AND" || up == "FOR" || up == "THE" {
			keep = append(keep, up)
			continue
		}
		if len(up) < 3 {
			continue
		}
		keep = append(keep, up)
	}
	if len(keep) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(keep, " "))
}

func findDateOfIssue(text string, conf float64) (time.Time, float64) {
	upper := strings.ToUpper(strings.ReplaceAll(text, "\r", ""))
	linesUpper := strings.Split(upper, "\n")
	linesOrig := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")

	for i := 0; i < len(linesUpper); i++ {
		line := strings.TrimSpace(linesUpper[i])
		if line == "" {
			continue
		}
		lNorm := normalizeLettersDigits(line)
		if !strings.Contains(line, "DATE OF ISSUE") && !strings.Contains(line, "ISSUE DATE") && !strings.Contains(lNorm, "DATEOFISSUE") && !strings.Contains(lNorm, "ISSUEDATE") {
			continue
		}
		for j := i; j < len(linesOrig) && j <= i+3; j++ {
			cand := strings.TrimSpace(linesOrig[j])
			if cand == "" {
				continue
			}
			if dates := extractDates(cand); len(dates) > 0 {
				return dates[0].UTC(), conf
			}
		}
	}

	allDates := make([]time.Time, 0, 8)
	for _, line := range linesOrig {
		for _, d := range extractDates(line) {
			allDates = append(allDates, d.UTC())
		}
	}
	if d := pickLikelyIssueDate(allDates); !d.IsZero() {
		return d, conf
	}
	return time.Time{}, 0
}

func findAuthorityFallback(text string) string {
	upper := strings.ToUpper(strings.ReplaceAll(text, "\r", ""))
	lines := strings.Split(upper, "\n")

	keywords := []string{"DEPARTMENT", "IMMIGRATION", "NATIONALITY", "AFFAIRS", "MINISTRY", "GOVERNMENT", "FEDERAL", "AUTHORITY"}
	best := ""
	bestScore := -1
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "P<") {
			continue
		}
		score := 0
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				score += 3
			}
		}
		if len(line) >= 20 {
			score++
		}
		if len(line) >= 35 {
			score++
		}
		if score > bestScore {
			bestScore = score
			best = line
		}
	}
	best = cleanupVisualValue(best)
	if bestScore < 3 {
		return ""
	}
	return best
}

func findPlaceOfBirthFallback(text string) string {
	upper := strings.ToUpper(strings.ReplaceAll(text, "\r", ""))
	linesUpper := strings.Split(upper, "\n")
	linesOrig := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")

	for i := 0; i < len(linesUpper) && i < len(linesOrig); i++ {
		lU := strings.TrimSpace(linesUpper[i])
		if lU == "" {
			continue
		}
		lNorm := normalizeLettersDigits(lU)
		if (strings.Contains(lNorm, "PLACE") || strings.Contains(lNorm, "PIACE")) && (strings.Contains(lNorm, "BIR") || strings.Contains(lNorm, "BIRT")) {
			orig := strings.TrimSpace(linesOrig[i])
			if parts := strings.SplitN(orig, ":", 2); len(parts) == 2 {
				if v := strings.TrimSpace(parts[1]); v != "" {
					return cleanupVisualValue(v)
				}
			}
			if parts := strings.SplitN(orig, "-", 2); len(parts) == 2 {
				if v := strings.TrimSpace(parts[1]); v != "" {
					return cleanupVisualValue(v)
				}
			}
			if next := nextNonEmptyLine(linesOrig, i+1, 4); next != "" {
				return cleanupVisualValue(next)
			}
		}
	}
	return ""
}

func findPlaceOfBirthNearBirthDate(text string, dateOfBirth time.Time) string {
	if strings.TrimSpace(text) == "" || dateOfBirth.IsZero() {
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	month := strings.ToUpper(dateOfBirth.Format("Jan"))
	day := dateOfBirth.Day()
	datePattern := regexp.MustCompile(`(?i)\b` + strconv.Itoa(day) + `\s*` + month + `(?:['\s]*\d{1,2})?\b`)

	for index, rawLine := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(rawLine))
		if upperLine == "" || !datePattern.MatchString(upperLine) {
			continue
		}

		if candidate := cleanupPlaceOfBirthCandidate(datePattern.ReplaceAllString(rawLine, " ")); candidate != "" {
			return candidate
		}

		if index+1 < len(lines) {
			if candidate := cleanupPlaceOfBirthCandidate(lines[index+1]); candidate != "" {
				return candidate
			}
		}
	}

	return ""
}

func cleanupPlaceOfBirthCandidate(value string) string {
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

func normalizeLettersDigits(value string) string {
	value = strings.ToUpper(value)
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func nextNonEmptyLine(lines []string, startIndex, maxLookahead int) string {
	if startIndex < 0 {
		startIndex = 0
	}
	end := startIndex + maxLookahead
	if end > len(lines) {
		end = len(lines)
	}
	for i := startIndex; i < end; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(line), "P<") {
			continue
		}
		return line
	}
	return ""
}

var (
	reISO      = regexp.MustCompile(`\b(\d{4})-(\d{2})-(\d{2})\b`)
	reSlash    = regexp.MustCompile(`\b(\d{2})[\./](\d{2})[\./](\d{4})\b`)
	reDMMMYY   = regexp.MustCompile(`\b(\d{2})\s*([A-Z]{3})\s*(\d{2})\b`)
	reDMMMYYYY = regexp.MustCompile(`\b(\d{2})\s*([A-Z]{3})\s*(\d{4})\b`)
)

func extractDates(line string) []time.Time {
	line = strings.ToUpper(strings.TrimSpace(line))
	if line == "" {
		return nil
	}
	line = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, line)

	var out []time.Time
	for _, m := range reISO.FindAllStringSubmatch(line, -1) {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if t := safeDate(y, mo, d); !t.IsZero() {
			out = append(out, t)
		}
	}
	for _, m := range reSlash.FindAllStringSubmatch(line, -1) {
		d, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		y, _ := strconv.Atoi(m[3])
		if t := safeDate(y, mo, d); !t.IsZero() {
			out = append(out, t)
		}
	}
	for _, m := range reDMMMYYYY.FindAllStringSubmatch(line, -1) {
		d, _ := strconv.Atoi(m[1])
		mo := monthFromMMM(m[2])
		y, _ := strconv.Atoi(m[3])
		if t := safeDate(y, mo, d); !t.IsZero() {
			out = append(out, t)
		}
	}
	for _, m := range reDMMMYY.FindAllStringSubmatch(line, -1) {
		d, _ := strconv.Atoi(m[1])
		mo := monthFromMMM(m[2])
		yy, _ := strconv.Atoi(m[3])
		y := expand2DigitYear(yy)
		if t := safeDate(y, mo, d); !t.IsZero() {
			out = append(out, t)
		}
	}
	return out
}

func monthFromMMM(value string) int {
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

func expand2DigitYear(yy int) int {
	if yy < 0 || yy > 99 {
		return 0
	}
	nowYY := time.Now().UTC().Year() % 100
	if yy <= nowYY+1 {
		return 2000 + yy
	}
	return 1900 + yy
}

func safeDate(y, m, d int) time.Time {
	if y <= 0 || m < 1 || m > 12 || d < 1 || d > 31 {
		return time.Time{}
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return time.Time{}
	}
	return t
}

func pickLikelyIssueDate(dates []time.Time) time.Time {
	if len(dates) == 0 {
		return time.Time{}
	}
	now := time.Now().UTC()
	cutoffPast := now.AddDate(-15, 0, 0)

	best := time.Time{}
	for _, d := range dates {
		d = d.UTC()
		if d.IsZero() || d.After(now.AddDate(0, 0, 1)) || d.Before(cutoffPast) {
			continue
		}
		if best.IsZero() || d.After(best) {
			best = d
		}
	}
	return best
}
