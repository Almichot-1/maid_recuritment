package ocr

import (
	"image/color"
	"math"
	"testing"
)

func TestGaussianBlur_NilInput(t *testing.T) {
	result := gaussianBlur(nil, 1.5)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestGaussianBlur_Uniform(t *testing.T) {
	img := createGrayImage(50, 50, func(x, y int) uint8 {
		return 128
	})

	blurred := gaussianBlur(img, 1.5)
	if blurred == nil {
		t.Fatal("blurred image should not be nil")
	}

	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if blurred.GrayAt(x, y).Y != 128 {
				t.Fatalf("expected 128 at (%d,%d), got %d", x, y, blurred.GrayAt(x, y).Y)
			}
		}
	}
}

func TestGaussianBlur_Impulse(t *testing.T) {
	img := createGrayImage(11, 11, func(x, y int) uint8 {
		return 0
	})
	img.SetGray(5, 5, color.Gray{Y: 255})

	blurred := gaussianBlur(img, 1.0)
	if blurred == nil {
		t.Fatal("blurred image should not be nil")
	}

	center := blurred.GrayAt(5, 5).Y
	if center >= 255 {
		t.Fatalf("expected center < 255 after blur, got %d", center)
	}
	if center < 1 {
		t.Fatalf("expected center > 0 after blur, got %d", center)
	}

	neighbor := blurred.GrayAt(4, 5).Y
	if neighbor < 1 {
		t.Fatalf("expected neighbor > 0 after blur, got %d", neighbor)
	}
	far := blurred.GrayAt(0, 0).Y
	if far > center {
		t.Fatalf("expected far pixel value (%d) < center (%d)", far, center)
	}

	if neighbor < far {
		t.Fatalf("expected neighbor (%d) >= far (%d)", neighbor, far)
	}
}

func TestGaussianBlur_SigmaZero(t *testing.T) {
	img := createGrayImage(10, 10, func(x, y int) uint8 {
		return uint8(x * 25)
	})

	blurred := gaussianBlur(img, 0)
	if blurred == nil {
		t.Fatal("blurred image should not be nil")
	}
}

func TestUnsharpMask_NilInput(t *testing.T) {
	result := unsharpMask(nil, 1.5, 0.8)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestUnsharpMask_EnhancesEdges(t *testing.T) {
	width, height := 50, 50
	img := createGrayImage(width, height, func(x, y int) uint8 {
		if x < 20 {
			return 0
		} else if x > 30 {
			return 255
		}
		return uint8((x - 20) * 25)
	})

	sharpened := unsharpMask(img, 1.5, 0.8)
	if sharpened == nil {
		t.Fatal("sharpened image should not be nil")
	}

	var maxOriginalGrad, maxSharpenedGrad float64
	for y := 0; y < height; y++ {
		for x := 1; x < width; x++ {
			og := math.Abs(float64(img.GrayAt(x, y).Y) - float64(img.GrayAt(x-1, y).Y))
			sg := math.Abs(float64(sharpened.GrayAt(x, y).Y) - float64(sharpened.GrayAt(x-1, y).Y))
			if og > maxOriginalGrad {
				maxOriginalGrad = og
			}
			if sg > maxSharpenedGrad {
				maxSharpenedGrad = sg
			}
		}
	}

	if maxSharpenedGrad <= maxOriginalGrad {
		t.Fatalf("expected steeper gradient after unsharp mask (original max: %.1f, sharpened max: %.1f)", maxOriginalGrad, maxSharpenedGrad)
	}
}

func TestUnsharpMask_Uniform(t *testing.T) {
	img := createGrayImage(30, 30, func(x, y int) uint8 {
		return 100
	})

	sharpened := unsharpMask(img, 1.5, 0.8)
	if sharpened == nil {
		t.Fatal("sharpened image should not be nil")
	}

	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			if sharpened.GrayAt(x, y).Y != 100 {
				t.Fatalf("expected 100 at (%d,%d), got %d", x, y, sharpened.GrayAt(x, y).Y)
			}
		}
	}
}
