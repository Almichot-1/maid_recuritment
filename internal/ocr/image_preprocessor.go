package ocr

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

import (
	_ "image/jpeg"
	_ "image/png"
)

func prepareMRZImage(imagePath string) (string, func(), error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return imagePath, func() {}, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return imagePath, func() {}, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return imagePath, func() {}, nil
	}

	startY := bounds.Min.Y + int(float64(height)*0.58)
	if startY >= bounds.Max.Y {
		return imagePath, func() {}, nil
	}

	sourceRect := image.Rect(bounds.Min.X, startY, bounds.Max.X, bounds.Max.Y)
	cropped := image.NewRGBA(image.Rect(0, 0, sourceRect.Dx(), sourceRect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), img, sourceRect.Min, draw.Src)

	tempFile, err := os.CreateTemp("", "passport-mrz-*.png")
	if err != nil {
		return imagePath, func() {}, err
	}

	tempPath := tempFile.Name()
	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	if err := png.Encode(tempFile, cropped); err != nil {
		_ = tempFile.Close()
		cleanup()
		return imagePath, func() {}, err
	}
	if err := tempFile.Close(); err != nil {
		cleanup()
		return imagePath, func() {}, err
	}

	return tempPath, cleanup, nil
}
