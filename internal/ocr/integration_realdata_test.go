package ocr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_RealPassportImages_Preprocess(t *testing.T) {
	files := findTestImages(t)
	if len(files) == 0 {
		t.Skip("no test images found in testdata/")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			info, err := os.Stat(file)
			if err != nil {
				t.Fatalf("cannot stat file: %v", err)
			}

			img, _, err := decodeImageFile(file)
			if err != nil {
				t.Fatalf("cannot decode image: %v", err)
			}

			bounds := img.Bounds()
			if bounds.Dx() < 100 || bounds.Dy() < 100 {
				t.Fatalf("image too small: %dx%d", bounds.Dx(), bounds.Dy())
			}

			grayImg := toGray(img)
			report := AssessQuality(grayImg)
			t.Logf("quality: pass=%v score=%.1f msg=%s", report.Pass, report.BlurScore, report.Message)

			mrzPath, cleanup, err := prepareMRZImage(file)
			if err != nil {
				t.Fatalf("prepareMRZImage failed: %v", err)
			}
			defer cleanup()

			procImg, _, err := decodeImageFile(mrzPath)
			if err != nil {
				t.Fatalf("cannot decode processed image: %v", err)
			}
			t.Logf("original: %dx%d (%d bytes), mrz: %dx%d",
				bounds.Dx(), bounds.Dy(), info.Size(),
				procImg.Bounds().Dx(), procImg.Bounds().Dy())

			visPath, visCleanup, err := prepareVisualZoneImage(file)
			if err != nil {
				t.Logf("visual zone prep (non-fatal): %v", err)
			} else {
				defer visCleanup()
				visImg, _, err := decodeImageFile(visPath)
				if err == nil {
					t.Logf("visual zone: %dx%d", visImg.Bounds().Dx(), visImg.Bounds().Dy())
				}
			}
		})
	}
}

func TestIntegration_RealPassportImages_OCRExtract(t *testing.T) {
	files := findTestImages(t)
	if len(files) == 0 {
		t.Skip("no test images found in testdata/")
	}

	if !tesseractAvailable() {
		t.Skip("Tesseract not available on this system")
	}

	processor := NewOCRProcessor("", "eng")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			mrz1, mrz2, conf, err := processor.ExtractMRZ(file)
			if err != nil {
				t.Skipf("ExtractMRZ failed (non-fatal): %v", err)
			}
			t.Logf("MRZ lines extracted (conf=%.2f)", conf)
			t.Logf("  line1: %q", mrz1)
			t.Logf("  line2: %q", mrz2)

			if len(mrz1) > 0 && len(mrz2) > 0 {
				parsed, parseErr := processor.parser.ParseMRZ(mrz1, mrz2)
				if parseErr != nil {
					t.Logf("ParseMRZ failed: %v", parseErr)
				} else if parsed != nil {
					t.Logf("  document: %s, country: %s", parsed.DocumentType, parsed.IssuingCountry)
					t.Logf("  surname: %s, given: %v", parsed.Surname, parsed.GivenNames)
					t.Logf("  passport: %s, nationality: %s, sex: %s", parsed.PassportNumber, parsed.Nationality, parsed.Sex)
					t.Logf("  dob: %s, expiry: %s", parsed.DateOfBirth.Format("2006-01-02"), parsed.DateOfExpiry.Format("2006-01-02"))
					t.Logf("  valid: %v, conf: %.2f, corrections: %d", parsed.IsValid, parsed.Confidence, parsed.CorrectionCount)
				}
			}

			vz, err := processor.ExtractVisualZone(file)
			if err != nil {
				t.Logf("ExtractVisualZone: %v", err)
			} else if vz != nil {
				t.Logf("  place_of_birth: %q, authority: %q", vz.PlaceOfBirth, vz.Authority)
				if !vz.DateOfIssue.IsZero() {
					t.Logf("  date_of_issue: %s", vz.DateOfIssue.Format("2006-01-02"))
				}
			}

			data, err := processor.ExtractPassportData(file)
			if err != nil {
				t.Logf("ExtractPassportData: %v", err)
			} else if data != nil {
				t.Logf("  final confidence: %.2f", data.Confidence)
			}
		})
	}
}

func TestIntegration_AssessQualityOnRealImages(t *testing.T) {
	files := findTestImages(t)
	if len(files) == 0 {
		t.Skip("no test images found in testdata/")
	}

	var sharpCount, blurryCount int
	for _, file := range files {
		img, _, err := decodeImageFile(file)
		if err != nil {
			continue
		}
		gray := toGray(img)
		report := AssessQuality(gray)
		if report.Pass {
			sharpCount++
		} else {
			blurryCount++
		}
		t.Logf("%s: score=%.1f pass=%v", filepath.Base(file), report.BlurScore, report.Pass)
	}
	t.Logf("results: %d sharp, %d blurry", sharpCount, blurryCount)
}

func tesseractAvailable() bool {
	proc := &OCRProcessor{tesseractPath: "", lang: "eng"}
	exe := proc.tesseractExe()
	if exe == "" || exe == "tesseract" {
		_, err := os.Stat(exe)
		return err == nil
	}
	_, err := os.Stat(exe)
	return err == nil
}

func findTestImages(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Logf("cannot read testdata/: %v", err)
		return nil
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".jpeg") && !strings.HasSuffix(name, ".jpg") && !strings.HasSuffix(name, ".png") {
			continue
		}
		files = append(files, filepath.Join("testdata", entry.Name()))
	}
	return files
}
