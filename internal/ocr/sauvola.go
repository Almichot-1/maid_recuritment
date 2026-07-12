package ocr

import (
	"image"
	"math"
)

func sauvolaThreshold(src *image.Gray, windowSize int, k float64, R float64) *image.Gray {
	if src == nil {
		return nil
	}
	if windowSize < 3 || windowSize%2 == 0 {
		windowSize = 31
	}

	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w == 0 || h == 0 {
		return image.NewGray(bounds)
	}

	stride := w + 1
	integral := make([]int, stride*(h+1))
	integralSq := make([]int, stride*(h+1))

	for y := 1; y <= h; y++ {
		for x := 1; x <= w; x++ {
			val := int(src.Pix[(y-1)*src.Stride+(x-1)])
			idx := y*stride + x
			integral[idx] = val + integral[(y-1)*stride+x] + integral[y*stride+(x-1)] - integral[(y-1)*stride+(x-1)]
			integralSq[idx] = val*val + integralSq[(y-1)*stride+x] + integralSq[y*stride+(x-1)] - integralSq[(y-1)*stride+(x-1)]
		}
	}

	half := windowSize / 2
	dst := image.NewGray(bounds)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			x1 := x - half
			y1 := y - half
			x2 := x + half
			y2 := y + half

			if x1 < 0 {
				x1 = 0
			}
			if y1 < 0 {
				y1 = 0
			}
			if x2 >= w {
				x2 = w - 1
			}
			if y2 >= h {
				y2 = h - 1
			}

			count := (x2 - x1 + 1) * (y2 - y1 + 1)

			sum := integral[(y2+1)*stride+(x2+1)] - integral[y1*stride+(x2+1)] - integral[(y2+1)*stride+x1] + integral[y1*stride+x1]
			sumSq := integralSq[(y2+1)*stride+(x2+1)] - integralSq[y1*stride+(x2+1)] - integralSq[(y2+1)*stride+x1] + integralSq[y1*stride+x1]

			mean := float64(sum) / float64(count)
			variance := float64(sumSq)/float64(count) - mean*mean
			if variance < 0 {
				variance = 0
			}
			stdDev := math.Sqrt(variance)

			threshold := mean * (1 + k*(stdDev/R-1))

			idx := y*src.Stride + x
			if float64(src.Pix[idx]) < threshold {
				dst.Pix[y*dst.Stride+x] = 0
			} else {
				dst.Pix[y*dst.Stride+x] = 255
			}
		}
	}

	return dst
}

func sauvolaThresholdDefault(src *image.Gray) *image.Gray {
	return sauvolaThreshold(src, 31, 0.34, 128)
}
