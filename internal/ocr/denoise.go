package ocr

import (
	"image"
	"image/color"
	"sort"
)

// medianDenoise applies a median filter to remove salt-and-pepper noise
// while preserving edges. Each pixel is replaced by the median value
// within its kernelSize × kernelSize neighborhood.
// kernelSize: the filter kernel size (must be odd, e.g. 3 or 5; good default 3)
// Returns the denoised *image.Gray.
func medianDenoise(src *image.Gray, kernelSize int) *image.Gray {
	if src == nil {
		return nil
	}

	if kernelSize == 1 {
		return src
	}
	if kernelSize < 3 || kernelSize%2 == 0 {
		kernelSize = 3
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	radius := kernelSize / 2

	dst := image.NewGray(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			values := make([]uint8, 0, kernelSize*kernelSize)
			for ky := -radius; ky <= radius; ky++ {
				for kx := -radius; kx <= radius; kx++ {
					px := mirrorCoord(x+kx, width)
					py := mirrorCoord(y+ky, height)
					values = append(values, src.GrayAt(px, py).Y)
				}
			}
			dst.SetGray(x, y, color.Gray{Y: median(values)})
		}
	}

	return dst
}

func mirrorCoord(coord, limit int) int {
	if coord < 0 {
		return -coord - 1
	}
	if coord >= limit {
		return 2*limit - coord - 1
	}
	return coord
}

func median(values []uint8) uint8 {
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values[len(values)/2]
}
