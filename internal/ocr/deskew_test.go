package ocr

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestDetectSkewAngle_NilInput(t *testing.T) {
	if angle := detectSkewAngle(nil); angle != 0 {
		t.Fatalf("expected 0 for nil input, got %f", angle)
	}
}

func TestDetectSkewAngle_Straight(t *testing.T) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if y%20 < 3 {
			return 0
		}
		return 255
	})
	angle := detectSkewAngle(img)
	if math.Abs(angle) > 1.0 {
		t.Fatalf("expected angle near 0 for horizontal lines, got %f", angle)
	}
}

func TestDetectSkewAngle_SlightlyTilted(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 200, 200))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	targetAngle := 3.0
	tanA := math.Tan(targetAngle * math.Pi / 180)
	for row := 0; row < 8; row++ {
		yOffset := row * 25
		for x := 0; x < 200; x++ {
			baseY := int(math.Round(tanA * float64(x))) + yOffset
			for dy := 0; dy < 3; dy++ {
				yy := baseY + dy
				if yy >= 0 && yy < 200 {
					img.SetGray(x, yy, color.Gray{Y: 0})
				}
			}
		}
	}
	angle := detectSkewAngle(img)
	t.Logf("detected angle: %f", angle)
	if math.Abs(angle-targetAngle) > 1.0 {
		t.Fatalf("expected angle ~%f, got %f", targetAngle, angle)
	}
}

func TestDeskew_NilInput(t *testing.T) {
	if result := deskew(nil); result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestDeskew_StraightImage(t *testing.T) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if y%20 < 2 {
			return 0
		}
		return 255
	})
	angle := detectSkewAngle(img)
	t.Logf("detected skew angle for straight image: %f", angle)
	result := deskew(img)
	if result != img {
		t.Fatal("expected same image pointer for straight image")
	}
}

func TestDeskew_SlightlyRotated(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 200, 200))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	tanA := math.Tan(3.0 * math.Pi / 180)
	for row := 0; row < 8; row++ {
		yOffset := row * 25
		for x := 0; x < 200; x++ {
			baseY := int(math.Round(tanA * float64(x))) + yOffset
			for dy := 0; dy < 3; dy++ {
				yy := baseY + dy
				if yy >= 0 && yy < 200 {
					img.SetGray(x, yy, color.Gray{Y: 0})
				}
			}
		}
	}

	deskewed := deskew(img)
	if deskewed == nil {
		t.Fatal("deskew returned nil")
	}
	if deskewed == img {
		t.Fatal("deskew should return a new image for rotated input")
	}

	newAngle := detectSkewAngle(deskewed)
	t.Logf("original angle: ~3.0, new angle: %f", newAngle)
	if math.Abs(newAngle) >= 1.0 {
		t.Fatalf("expected deskewed angle < 1.0, got %f", newAngle)
	}
}

func TestDeskew_PreservesContent(t *testing.T) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if y%15 < 2 {
			return 0
		}
		return 255
	})

	originalPix := make([]uint8, len(img.Pix))
	copy(originalPix, img.Pix)

	angle := detectSkewAngle(img)
	t.Logf("detected skew angle for preserves test: %f", angle)
	result := deskew(img)

	if result != img {
		t.Fatal("expected same image for straight content")
	}
	for i := range img.Pix {
		if img.Pix[i] != originalPix[i] {
			t.Fatalf("pixel data changed at index %d: got %d, expected %d", i, img.Pix[i], originalPix[i])
		}
	}
}
