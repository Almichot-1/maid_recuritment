package ocr

import (
	"image"
	"image/color"
	"testing"
)

func createGrayImage(width, height int, generator func(x, y int) uint8) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetGray(x, y, color.Gray{Y: generator(x, y)})
		}
	}
	return img
}

func TestBlurScore_SharpCheckerboard(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	score := BlurScore(img)
	if score < 100 {
		t.Fatalf("expected sharp image score >= 100, got %f", score)
	}
}

func TestBlurScore_Uniform(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return 128
	})
	score := BlurScore(img)
	if score > 1 {
		t.Fatalf("expected uniform image score near 0, got %f", score)
	}
}

func TestBlurScore_BlurryGradient(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return uint8(float64(x+y) / 200.0 * 255.0)
	})
	score := BlurScore(img)
	if score > 50 {
		t.Fatalf("expected gradient score < 50, got %f", score)
	}
}

func TestBlurScore_NilInput(t *testing.T) {
	score := BlurScore(nil)
	if score != 0 {
		t.Fatalf("expected 0 for nil input, got %f", score)
	}
}

func TestBlurScore_SmallImage(t *testing.T) {
	img := createGrayImage(2, 2, func(x, y int) uint8 {
		return 128
	})
	score := BlurScore(img)
	if score != 0 {
		t.Fatalf("expected 0 for too-small image, got %f", score)
	}
}

func TestBlurScore_LargeImage(t *testing.T) {
	img := createGrayImage(1000, 1000, func(x, y int) uint8 {
		if (x/10+y/10)%2 == 0 {
			return 0
		}
		return 255
	})
	score := BlurScore(img)
	if score < 1000 {
		t.Fatalf("expected large sharp image score >= 1000, got %f", score)
	}
}

func TestAssessQuality_Sharp(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	report := AssessQuality(img)
	if !report.Pass {
		t.Fatalf("expected sharp image to pass, got: %s", report.Message)
	}
	if report.BlurScore < 100 {
		t.Fatalf("expected high blur score, got %f", report.BlurScore)
	}
}

func TestAssessQuality_NilInput(t *testing.T) {
	report := AssessQuality(nil)
	if report.Pass {
		t.Fatal("expected nil input to fail")
	}
}

func TestAssessQuality_Uniform(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return 128
	})
	report := AssessQuality(img)
	if report.BlurScore > 1 {
		t.Fatalf("expected low blur score for uniform, got %f", report.BlurScore)
	}
	if report.Pass {
		t.Fatal("expected pass=false for uniform image (score < 0.01)")
	}
}
