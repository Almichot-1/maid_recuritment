package ocr

import (
	"testing"
)

func TestSauvolaThreshold_SharpText(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	result := sauvolaThresholdDefault(img)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	hasBlack := false
	hasWhite := false
	for _, v := range result.Pix {
		if v == 0 {
			hasBlack = true
		}
		if v == 255 {
			hasWhite = true
		}
	}
	if !hasBlack {
		t.Fatal("expected black pixels for sharp text")
	}
	if !hasWhite {
		t.Fatal("expected white pixels for sharp text")
	}
}

func TestSauvolaThreshold_Uniform(t *testing.T) {
	img := createGrayImage(50, 50, func(x, y int) uint8 {
		return 128
	})
	result := sauvolaThresholdDefault(img)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	for _, v := range result.Pix {
		if v != 255 {
			t.Fatalf("expected all white (255) for uniform image, got %d", v)
		}
	}
}

func TestSauvolaThreshold_NilInput(t *testing.T) {
	result := sauvolaThresholdDefault(nil)
	if result != nil {
		t.Fatal("expected nil for nil input via default")
	}
	result = sauvolaThreshold(nil, 31, 0.34, 128)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestSauvolaThreshold_SmallImage(t *testing.T) {
	img := createGrayImage(2, 2, func(x, y int) uint8 {
		return 128
	})
	result := sauvolaThresholdDefault(img)
	if result == nil {
		t.Fatal("expected non-nil result for small image")
	}
}

func TestSauvolaThreshold_Gradient(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return uint8(float64(x) / 100.0 * 255.0)
	})
	result := sauvolaThresholdDefault(img)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	hasBlack := false
	hasWhite := false
	for _, v := range result.Pix {
		if v == 0 {
			hasBlack = true
		}
		if v == 255 {
			hasWhite = true
		}
	}
	if !hasBlack {
		t.Fatal("expected black pixels for dark region")
	}
	if !hasWhite {
		t.Fatal("expected white pixels for light region")
	}
}

func TestSauvolaThreshold_UnevenLighting(t *testing.T) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		base := uint8(50 + (x+y)*200/(200+200))
		if (x/10+y/10)%2 == 0 {
			return uint8(max(0, int(base)-60))
		}
		return uint8(min(255, int(base)+60))
	})
	result := sauvolaThresholdDefault(img)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	hasBlack := false
	hasWhite := false
	for _, v := range result.Pix {
		if v == 0 {
			hasBlack = true
		}
		if v == 255 {
			hasWhite = true
		}
	}
	if !hasBlack {
		t.Fatal("expected black pixels in uneven lighting output")
	}
	if !hasWhite {
		t.Fatal("expected white pixels in uneven lighting output")
	}
}
