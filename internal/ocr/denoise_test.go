package ocr

import (
	"image/color"
	"math/rand"
	"testing"
)

func TestMedianDenoise_NilInput(t *testing.T) {
	result := medianDenoise(nil, 3)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestMedianDenoise_SaltPepper(t *testing.T) {
	img := createGrayImage(50, 50, func(x, y int) uint8 {
		return 255
	})

	rng := rand.New(rand.NewSource(42))
	noisyPixels := 0
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if rng.Intn(10) == 0 {
				img.SetGray(x, y, color.Gray{Y: 0})
				noisyPixels++
			}
		}
	}

	denoised := medianDenoise(img, 3)
	if denoised == nil {
		t.Fatal("denoised image should not be nil")
	}

	blackCount := 0
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if denoised.GrayAt(x, y).Y == 0 {
				blackCount++
			}
		}
	}

	if blackCount >= noisyPixels {
		t.Fatalf("expected fewer black pixels after denoising (%d), got %d", noisyPixels, blackCount)
	}
}

func TestMedianDenoise_Identity(t *testing.T) {
	img := createGrayImage(20, 20, func(x, y int) uint8 {
		return 128
	})

	denoised := medianDenoise(img, 3)
	if denoised == nil {
		t.Fatal("denoised image should not be nil")
	}

	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if denoised.GrayAt(x, y).Y != 128 {
				t.Fatalf("expected pixel value 128 at (%d,%d), got %d", x, y, denoised.GrayAt(x, y).Y)
			}
		}
	}
}

func TestMedianDenoise_EdgesPreserved(t *testing.T) {
	img := createGrayImage(30, 30, func(x, y int) uint8 {
		if x < 15 {
			return 0
		}
		return 255
	})

	denoised := medianDenoise(img, 3)
	if denoised == nil {
		t.Fatal("denoised image should not be nil")
	}

	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			expected := uint8(0)
			if x >= 15 {
				expected = 255
			}
			if denoised.GrayAt(x, y).Y != expected {
				t.Fatalf("expected %d at (%d,%d), got %d", expected, x, y, denoised.GrayAt(x, y).Y)
			}
		}
	}
}

func TestMedianDenoise_SmallKernel(t *testing.T) {
	img := createGrayImage(10, 10, func(x, y int) uint8 {
		return uint8((x + y) % 256)
	})

	denoised := medianDenoise(img, 1)
	if denoised == nil {
		t.Fatal("denoised image should not be nil")
	}

	if denoised.Bounds() != img.Bounds() {
		t.Fatal("bounds should match for kernelSize=1")
	}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			expected := img.GrayAt(x, y).Y
			if denoised.GrayAt(x, y).Y != expected {
				t.Fatalf("expected %d at (%d,%d) for kernelSize=1, got %d", expected, x, y, denoised.GrayAt(x, y).Y)
			}
		}
	}
}

func TestMedianDenoise_LargeKernel(t *testing.T) {
	img := createGrayImage(30, 30, func(x, y int) uint8 {
		return uint8((x * y) % 256)
	})

	denoised := medianDenoise(img, 7)
	if denoised == nil {
		t.Fatal("denoised image should not be nil")
	}

	if denoised.Bounds() != img.Bounds() {
		t.Fatal("bounds should match after denoising")
	}
}
