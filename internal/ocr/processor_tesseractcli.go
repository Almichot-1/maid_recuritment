package ocr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var ErrInvalidMRZOutput = errors.New("OCR produced invalid MRZ output")

func mergeOCRTextLines(texts ...string) string {
	seen := make(map[string]struct{}, 256)
	out := make([]string, 0, 128)
	for _, text := range texts {
		text = strings.ReplaceAll(text, "\r", "")
		for _, raw := range strings.Split(text, "\n") {
			line := strings.TrimSpace(raw)
			if line == "" {
				continue
			}
			key := strings.ToUpper(line)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func (p *OCRProcessor) runTesseractText(imagePath, lang string, extraArgs []string) (string, error) {
	return p.runTesseractTextWithTimeout(imagePath, lang, extraArgs, 20*time.Second)
}

func (p *OCRProcessor) runTesseractTextWithTimeout(imagePath, lang string, extraArgs []string, timeout time.Duration) (string, error) {
	args := []string{imagePath, "stdout"}
	args = append(args, extraArgs...)
	args = append(args, "-l", lang)

	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.tesseractExe(), args...)
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env,
		"OMP_THREAD_LIMIT=1",
		"OMP_NUM_THREADS=1",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("tesseract CLI timed out after %s for language %s", timeout.Round(time.Second), lang)
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("tesseract CLI failed: %s", message)
	}

	return stdout.String(), nil
}

func (p *OCRProcessor) tesseractExe() string {
	if strings.TrimSpace(p.tesseractPath) != "" {
		return p.tesseractPath
	}
	if resolved, err := exec.LookPath("tesseract"); err == nil && strings.TrimSpace(resolved) != "" {
		return resolved
	}
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\\Program Files\\Tesseract-OCR\\tesseract.exe`,
			`C:\\Program Files (x86)\\Tesseract-OCR\\tesseract.exe`,
		}
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil && info != nil && !info.IsDir() {
				return candidate
			}
		}
	}
	return "tesseract"
}
