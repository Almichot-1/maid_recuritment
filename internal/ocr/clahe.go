package ocr

import (
	"image"
	"math"
)

// applyCLAHE performs Contrast Limited Adaptive Histogram Equalization
// on a grayscale image. This enhances local contrast while limiting
// noise amplification.
// clipLimit: maximum histogram count (typical 2.0-4.0, good default 3.0)
// tileSize: size of the grid tiles (must be > 0, good default 8)
// Returns the contrast-enhanced *image.Gray.
func applyCLAHE(src *image.Gray, clipLimit float64, tileSize int) *image.Gray {
	if src == nil {
		return nil
	}

	if clipLimit <= 0 || tileSize < 2 {
		clipLimit = 3.0
		tileSize = 8
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return image.NewGray(bounds)
	}

	numTilesX := (width + tileSize - 1) / tileSize
	numTilesY := (height + tileSize - 1) / tileSize

	tiles := make([][][256]uint8, numTilesY)
	for ty := 0; ty < numTilesY; ty++ {
		tiles[ty] = make([][256]uint8, numTilesX)
		y0 := ty * tileSize
		y1 := y0 + tileSize
		if y1 > height {
			y1 = height
		}
		for tx := 0; tx < numTilesX; tx++ {
			x0 := tx * tileSize
			x1 := x0 + tileSize
			if x1 > width {
				x1 = width
			}

			pixels := make([]uint8, 0, (y1-y0)*(x1-x0))
			for y := y0; y < y1; y++ {
				rowStart := y*src.Stride + x0
				pixels = append(pixels, src.Pix[rowStart:rowStart+(x1-x0)]...)
			}

			hist := buildHistogram(pixels)
			clip := int(clipLimit * float64(len(pixels)) / 256.0)
			if clip < 1 {
				clip = 1
			}
			hist = clipHistogram(hist, clip)
			tiles[ty][tx] = cdfFromHistogram(hist)
		}
	}

	dst := image.NewGray(bounds)
	dstStride := dst.Stride

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := src.Pix[y*src.Stride+x]
			dstIdx := y*dstStride + x

			tx := x / tileSize
			ty := y / tileSize
			if tx >= numTilesX {
				tx = numTilesX - 1
			}
			if ty >= numTilesY {
				ty = numTilesY - 1
			}

			xLocal := x - tx*tileSize
			yLocal := y - ty*tileSize
			weightX := float64(xLocal) / float64(tileSize)
			weightY := float64(yLocal) / float64(tileSize)

			tx0, tx1 := tx, tx
			if tx+1 < numTilesX {
				tx1 = tx + 1
			}
			ty0, ty1 := ty, ty
			if ty+1 < numTilesY {
				ty1 = ty + 1
			}

			if tx0 == tx1 && ty0 == ty1 {
				dst.Pix[dstIdx] = tiles[ty][tx][val]
				continue
			}

			if ty0 == ty1 {
				m0 := float64(tiles[ty][tx0][val])
				m1 := float64(tiles[ty][tx1][val])
				dst.Pix[dstIdx] = uint8(math.Round(m0*(1-weightX) + m1*weightX))
				continue
			}

			if tx0 == tx1 {
				m0 := float64(tiles[ty0][tx][val])
				m1 := float64(tiles[ty1][tx][val])
				dst.Pix[dstIdx] = uint8(math.Round(m0*(1-weightY) + m1*weightY))
				continue
			}

			mappedTL := float64(tiles[ty0][tx0][val])
			mappedTR := float64(tiles[ty0][tx1][val])
			mappedBL := float64(tiles[ty1][tx0][val])
			mappedBR := float64(tiles[ty1][tx1][val])

			top := mappedTL*(1-weightX) + mappedTR*weightX
			bottom := mappedBL*(1-weightX) + mappedBR*weightX
			dst.Pix[dstIdx] = uint8(math.Round(top*(1-weightY) + bottom*weightY))
		}
	}

	return dst
}

func buildHistogram(pixels []uint8) [256]int {
	var hist [256]int
	for _, p := range pixels {
		hist[p]++
	}
	return hist
}

func clipHistogram(hist [256]int, clipLimit int) [256]int {
	if clipLimit <= 0 {
		return hist
	}

	totalExcess := 0
	for i := 0; i < 256; i++ {
		if hist[i] > clipLimit {
			totalExcess += hist[i] - clipLimit
			hist[i] = clipLimit
		}
	}

	if totalExcess == 0 {
		return hist
	}

	redistribute := totalExcess / 256
	remainder := totalExcess % 256

	for i := 0; i < 256; i++ {
		hist[i] += redistribute
	}

	for i := 0; i < remainder; i++ {
		hist[i]++
	}

	return hist
}

func cdfFromHistogram(hist [256]int) [256]uint8 {
	var cdf [256]uint8
	total := 0

	for i := 0; i < 256; i++ {
		total += hist[i]
	}

	if total == 0 {
		return cdf
	}

	cumSum := 0
	for i := 0; i < 256; i++ {
		cumSum += hist[i]
		cdf[i] = uint8(cumSum * 255 / total)
	}

	return cdf
}
