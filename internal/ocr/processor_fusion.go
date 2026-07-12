package ocr

import (
	"strings"
	"time"
)

type FusedMRZResult struct {
	DocumentType     string
	IssuingCountry   string
	Surname          string
	GivenNames       []string
	PassportNumber   string
	Nationality      string
	DateOfBirth      time.Time
	DateOfExpiry     time.Time
	Sex              string
	OptionalData     string
	Line1            string
	Line2            string
	Confidence       float64
	FieldConfidence  map[string]float64
	SuccessfulPasses int
	TotalPasses      int
}

func fuseMRZResults(results []*MRZPassResult) *FusedMRZResult {
	valid := validPassResults(results)
	fused := &FusedMRZResult{
		FieldConfidence:  make(map[string]float64),
		TotalPasses:      len(results),
		SuccessfulPasses: len(valid),
	}
	if len(valid) == 0 {
		return fused
	}

	fused.DocumentType = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.DocumentType })
	fused.IssuingCountry = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.IssuingCountry })
	fused.Surname = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.Surname })
	fused.PassportNumber = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.PassportNumber })
	fused.Nationality = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.Nationality })
	fused.Sex = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.Sex })
	fused.OptionalData = majorityVoteString(valid, func(r *MRZPassResult) string { return r.Parsed.OptionalData })

	givenJoined := majorityVoteString(valid, func(r *MRZPassResult) string { return strings.Join(r.Parsed.GivenNames, " ") })
	if givenJoined != "" {
		fused.GivenNames = strings.Fields(givenJoined)
	}

	fused.DateOfBirth = majorityVoteDate(valid, func(r *MRZPassResult) time.Time { return r.Parsed.DateOfBirth })
	fused.DateOfExpiry = majorityVoteDate(valid, func(r *MRZPassResult) time.Time { return r.Parsed.DateOfExpiry })

	best := valid[0]
	for _, r := range valid[1:] {
		if r.Confidence > best.Confidence {
			best = r
		}
	}
	fused.Line1 = best.Line1
	fused.Line2 = best.Line2

	sum := 0.0
	for _, r := range valid {
		sum += r.Confidence
	}
	fused.Confidence = sum / float64(len(valid))

	fused.FieldConfidence["document_type"] = fieldAgreement(valid, func(r *MRZPassResult) string { return r.Parsed.DocumentType }, fused.DocumentType)
	fused.FieldConfidence["issuing_country"] = fieldAgreement(valid, func(r *MRZPassResult) string { return r.Parsed.IssuingCountry }, fused.IssuingCountry)
	fused.FieldConfidence["surname"] = fieldAgreement(valid, func(r *MRZPassResult) string { return r.Parsed.Surname }, fused.Surname)
	fused.FieldConfidence["passport_number"] = fieldAgreement(valid, func(r *MRZPassResult) string { return r.Parsed.PassportNumber }, fused.PassportNumber)
	fused.FieldConfidence["nationality"] = fieldAgreement(valid, func(r *MRZPassResult) string { return r.Parsed.Nationality }, fused.Nationality)
	fused.FieldConfidence["sex"] = fieldAgreement(valid, func(r *MRZPassResult) string { return r.Parsed.Sex }, fused.Sex)
	fused.FieldConfidence["date_of_birth"] = dateAgreement(valid, func(r *MRZPassResult) time.Time { return r.Parsed.DateOfBirth }, fused.DateOfBirth)
	fused.FieldConfidence["date_of_expiry"] = dateAgreement(valid, func(r *MRZPassResult) time.Time { return r.Parsed.DateOfExpiry }, fused.DateOfExpiry)

	return fused
}

func validPassResults(results []*MRZPassResult) []*MRZPassResult {
	out := make([]*MRZPassResult, 0, len(results))
	for _, r := range results {
		if r.Err == nil && r.Parsed != nil {
			out = append(out, r)
		}
	}
	return out
}

func majorityVoteString(results []*MRZPassResult, extract func(*MRZPassResult) string) string {
	counts := make(map[string]int)
	for _, r := range results {
		counts[extract(r)]++
	}
	best := ""
	bestCount := 0
	for value, count := range counts {
		if count > bestCount || (count == bestCount && value > best) {
			bestCount = count
			best = value
		}
	}
	return best
}

func majorityVoteDate(results []*MRZPassResult, extract func(*MRZPassResult) time.Time) time.Time {
	best := time.Time{}
	bestCount := 0
	for i, r := range results {
		d := extract(r)
		if d.IsZero() {
			continue
		}
		count := 1
		for j, r2 := range results {
			if i != j && extract(r2).Equal(d) {
				count++
			}
		}
		if count > bestCount || (count == bestCount && d.After(best)) {
			bestCount = count
			best = d
		}
	}
	return best
}

func fieldAgreement(results []*MRZPassResult, extract func(*MRZPassResult) string, target string) float64 {
	if len(results) == 0 {
		return 0
	}
	count := 0
	for _, r := range results {
		if extract(r) == target {
			count++
		}
	}
	return float64(count) / float64(len(results))
}

func dateAgreement(results []*MRZPassResult, extract func(*MRZPassResult) time.Time, target time.Time) float64 {
	if len(results) == 0 {
		return 0
	}
	count := 0
	for _, r := range results {
		if extract(r).Equal(target) {
			count++
		}
	}
	return float64(count) / float64(len(results))
}
