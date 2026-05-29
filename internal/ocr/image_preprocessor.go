package ocr

import (
	"image"
	stddraw "image/draw"
	"image/png"
	"math"
	"os"

	xdraw "golang.org/x/image/draw"
)

import (
	_ "image/jpeg"
	_ "image/png"
)

const (
	mrzCropStartRatio        = 0.78
	visualZoneCropEndRatio   = 0.78
	mrzMinTargetWidth        = 1200
	mrzMaxTargetWidth        = 1600
	visualZoneMaxTargetWidth = 1400
)

func prepareMRZImage(imagePath string) (string, func(), error) {
	img, bounds, err := decodeImageFile(imagePath)
	if err != nil {
		return imagePath, func() {}, err
	}

	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return imagePath, func() {}, nil
	}

	startY := bounds.Min.Y + int(math.Round(float64(height)*mrzCropStartRatio))
	if startY >= bounds.Max.Y {
		return imagePath, func() {}, nil
	}

	sourceRect := image.Rect(bounds.Min.X, startY, bounds.Max.X, bounds.Max.Y)
	cropped := cropToRGBA(img, sourceRect)
	gray := toGray(cropped)
	targetWidth := clamp(width, mrzMinTargetWidth, mrzMaxTargetWidth)
	gray = resizeGray(gray, targetWidth)
	gray = binarizeGray(gray)

	return writeTempPNG(gray, "passport-mrz-*.png")
}

func prepareVisualZoneImage(imagePath string) (string, func(), error) {
	img, bounds, err := decodeImageFile(imagePath)
	if err != nil {
		return imagePath, func() {}, err
	}

	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return imagePath, func() {}, nil
	}

	endY := bounds.Min.Y + int(math.Round(float64(height)*visualZoneCropEndRatio))
	if endY <= bounds.Min.Y || endY > bounds.Max.Y {
		endY = bounds.Max.Y
	}

	sourceRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, endY)
	cropped := cropToRGBA(img, sourceRect)
	gray := toGray(cropped)
	if gray.Bounds().Dx() > visualZoneMaxTargetWidth {
		gray = resizeGray(gray, visualZoneMaxTargetWidth)
	}

	return writeTempPNG(gray, "passport-visual-*.png")
}

func decodeImageFile(imagePath string) (image.Image, image.Rectangle, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	return img, img.Bounds(), nil
}

func cropToRGBA(src image.Image, rect image.Rectangle) *image.RGBA {
	rect = rect.Intersect(src.Bounds())
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	stddraw.Draw(dst, dst.Bounds(), src, rect.Min, stddraw.Src)
	return dst
}

func toGray(src image.Image) *image.Gray {
	bounds := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	stddraw.Draw(dst, dst.Bounds(), src, bounds.Min, stddraw.Src)
	return dst
}

func resizeGray(src *image.Gray, targetWidth int) *image.Gray {
	if src == nil {
		return nil
	}

	currentWidth := src.Bounds().Dx()
	currentHeight := src.Bounds().Dy()
	if currentWidth == 0 || currentHeight == 0 || targetWidth <= 0 || currentWidth == targetWidth {
		return src
	}

	scale := float64(targetWidth) / float64(currentWidth)
	targetHeight := int(math.Round(float64(currentHeight) * scale))
	if targetHeight < 1 {
		targetHeight = 1
	}

	dst := image.NewGray(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func binarizeGray(src *image.Gray) *image.Gray {
	if src == nil {
		return nil
	}

	threshold := otsuThreshold(src)
	dst := image.NewGray(src.Bounds())
	for i, value := range src.Pix {
		if value > threshold {
			dst.Pix[i] = 255
			continue
		}
		dst.Pix[i] = 0
	}
	return dst
}

func otsuThreshold(src *image.Gray) uint8 {
	var histogram [256]int
	for _, value := range src.Pix {
		histogram[value]++
	}

	total := len(src.Pix)
	if total == 0 {
		return 127
	}

	sum := 0.0
	for i, count := range histogram {
		sum += float64(i * count)
	}

	sumBackground := 0.0
	weightBackground := 0
	maxVariance := -1.0
	threshold := 127

	for i, count := range histogram {
		weightBackground += count
		if weightBackground == 0 {
			continue
		}

		weightForeground := total - weightBackground
		if weightForeground == 0 {
			break
		}

		sumBackground += float64(i * count)
		meanBackground := sumBackground / float64(weightBackground)
		meanForeground := (sum - sumBackground) / float64(weightForeground)
		variance := float64(weightBackground) * float64(weightForeground) * math.Pow(meanBackground-meanForeground, 2)
		if variance > maxVariance {
			maxVariance = variance
			threshold = i
		}
	}

	if threshold < 32 {
		return 32
	}
	if threshold > 223 {
		return 223
	}
	return uint8(threshold)
}

func writeTempPNG(img image.Image, pattern string) (string, func(), error) {
	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", func() {}, err
	}

	tempPath := tempFile.Name()
	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	if err := png.Encode(tempFile, img); err != nil {
		_ = tempFile.Close()
		cleanup()
		return "", func() {}, err
	}
	if err := tempFile.Close(); err != nil {
		cleanup()
		return "", func() {}, err
	}

	return tempPath, cleanup, nil
}

func clamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
