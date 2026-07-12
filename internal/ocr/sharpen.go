package ocr

import (
	"image"
	"image/color"
	"math"
)

func gaussianBlur(src *image.Gray, sigma float64) *image.Gray {
	if src == nil {
		return nil
	}
	if sigma <= 0 {
		sigma = 1.0
	}

	ksize := 2*int(math.Ceil(2*sigma)) + 1
	kernel := make([]float64, ksize)
	var sum float64
	for i := 0; i < ksize; i++ {
		x := float64(i - ksize/2)
		kernel[i] = math.Exp(-x*x/(2*sigma*sigma)) / (math.Sqrt(2*math.Pi) * sigma)
		sum += kernel[i]
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	radius := ksize / 2

	temp := image.NewGray(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var val float64
			for kx := 0; kx < ksize; kx++ {
				sx := x + (kx - radius)
				if sx < 0 {
					sx = -sx - 1
				} else if sx >= width {
					sx = 2*width - sx - 1
				}
				val += float64(src.GrayAt(sx, y).Y) * kernel[kx]
			}
			temp.SetGray(x, y, color.Gray{Y: clampF(val, 0, 255)})
		}
	}

	dst := image.NewGray(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var val float64
			for ky := 0; ky < ksize; ky++ {
				sy := y + (ky - radius)
				if sy < 0 {
					sy = -sy - 1
				} else if sy >= height {
					sy = 2*height - sy - 1
				}
				val += float64(temp.GrayAt(x, sy).Y) * kernel[ky]
			}
			dst.SetGray(x, y, color.Gray{Y: clampF(val, 0, 255)})
		}
	}

	return dst
}

func unsharpMask(src *image.Gray, sigma float64, amount float64) *image.Gray {
	if src == nil {
		return nil
	}

	blurred := gaussianBlur(src, sigma)
	if blurred == nil {
		return nil
	}

	bounds := src.Bounds()
	dst := image.NewGray(bounds)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			orig := float64(src.GrayAt(x, y).Y)
			blur := float64(blurred.GrayAt(x, y).Y)
			sharpened := orig + amount*(orig-blur)
			dst.SetGray(x, y, color.Gray{Y: clampF(sharpened, 0, 255)})
		}
	}

	return dst
}

func clampF(val, min, max float64) uint8 {
	if val < min {
		return uint8(min)
	}
	if val > max {
		return uint8(max)
	}
	return uint8(math.Round(val))
}
