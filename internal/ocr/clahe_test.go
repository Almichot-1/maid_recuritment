package ocr

import (
	"image"
	"image/color"
	"testing"
)

func TestApplyCLAHE_NilInput(t *testing.T) {
	result := applyCLAHE(nil, 3.0, 8)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestApplyCLAHE_LowContrast(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return uint8(100 + (x+y)%21)
	})
	result := applyCLAHE(img, 3.0, 8)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	min, max := uint8(255), uint8(0)
	for _, v := range result.Pix {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max-min <= 20 {
		t.Fatalf("expected contrast enhancement (wider range), got min=%d max=%d", min, max)
	}
}

func TestApplyCLAHE_Identity(t *testing.T) {
	img := createGrayImage(100, 100, func(x, y int) uint8 {
		return uint8((x + y*7) % 256)
	})
	result := applyCLAHE(img, 3.0, 8)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	min, max := uint8(255), uint8(0)
	for _, v := range result.Pix {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	rangeSize := int(max) - int(min)
	if rangeSize < 250 {
		t.Fatalf("expected near-full range preserved, got [%d, %d] (range=%d)", min, max, rangeSize)
	}
}

func TestApplyCLAHE_Uniform(t *testing.T) {
	img := createGrayImage(50, 50, func(x, y int) uint8 {
		return 128
	})
	result := applyCLAHE(img, 3.0, 8)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	for _, v := range result.Pix {
		if v != result.Pix[0] {
			t.Fatalf("expected uniform output, got varying values")
		}
	}
}

func TestApplyCLAHE_Edge(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetGray(x, y, color.Gray{Y: uint8(x*50 + y*10)})
		}
	}

	result := applyCLAHE(img, 3.0, 8)
	if result == nil {
		t.Fatal("expected non-nil result for 4x4 image")
	}
	if result.Bounds().Dx() != 4 || result.Bounds().Dy() != 4 {
		t.Fatalf("expected 4x4 output, got %dx%d", result.Bounds().Dx(), result.Bounds().Dy())
	}
}
