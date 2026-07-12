package ocr

import (
	"image"
	"math"
)

func deskew(src *image.Gray) *image.Gray {
	if src == nil {
		return nil
	}
	angle := detectSkewAngle(src)
	if angle == 0 {
		return src
	}
	return rotateImage(src, angle)
}

func detectSkewAngle(src *image.Gray) float64 {
	if src == nil {
		return 0
	}
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w < 20 || h < 20 {
		return 0
	}
	edge := sobelEdgeMagnitude(src)
	bestTheta := houghTransform(edge, 91)
	if math.Abs(bestTheta) <= 1.0 {
		return 0
	}
	return bestTheta
}

func sobelEdgeMagnitude(src *image.Gray) *image.Gray {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	dst := image.NewGray(b)
	if w < 3 || h < 3 {
		return dst
	}
	pixels := src.Pix
	sStride := src.Stride
	dPixels := dst.Pix
	dStride := dst.Stride
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			gx := -float64(pixels[(y-1)*sStride+(x-1)]) +
				float64(pixels[(y-1)*sStride+(x+1)]) +
				-2*float64(pixels[y*sStride+(x-1)]) +
				2*float64(pixels[y*sStride+(x+1)]) +
				-float64(pixels[(y+1)*sStride+(x-1)]) +
				float64(pixels[(y+1)*sStride+(x+1)])
			gy := -float64(pixels[(y-1)*sStride+(x-1)]) +
				-2*float64(pixels[(y-1)*sStride+x]) +
				-float64(pixels[(y-1)*sStride+(x+1)]) +
				float64(pixels[(y+1)*sStride+(x-1)]) +
				2*float64(pixels[(y+1)*sStride+x]) +
				float64(pixels[(y+1)*sStride+(x+1)])
			mag := math.Sqrt(gx*gx + gy*gy)
			if mag > 128 {
				dPixels[y*dStride+x] = 255
			}
		}
	}
	return dst
}

func houghTransform(edgeImg *image.Gray, thetaSteps int) float64 {
	b := edgeImg.Bounds()
	w := b.Dx()
	h := b.Dy()
	diag := int(math.Ceil(math.Sqrt(float64(w*w + h*h))))
	halfSteps := (thetaSteps - 1) / 2

	acc := make([][]int, thetaSteps)
	for i := range acc {
		acc[i] = make([]int, 2*diag+1)
	}

	pixels := edgeImg.Pix
	stride := edgeImg.Stride

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if pixels[y*stride+x] == 0 {
				continue
			}
			for ti := 0; ti < thetaSteps; ti++ {
				theta := float64(ti-halfSteps) * (math.Pi / 180.0)
				rho := -float64(x)*math.Sin(theta) + float64(y)*math.Cos(theta)
				ri := int(math.Round(rho)) + diag
				if ri >= 0 && ri < 2*diag+1 {
					acc[ti][ri]++
				}
			}
		}
	}

	bestVotes := 0
	bestTheta := 0.0
	for ti := 0; ti < thetaSteps; ti++ {
		for ri := 0; ri < len(acc[ti]); ri++ {
			if acc[ti][ri] > bestVotes {
				bestVotes = acc[ti][ri]
				bestTheta = float64(ti - halfSteps)
			}
		}
	}

	return bestTheta
}

func rotateImage(src *image.Gray, angleDeg float64) *image.Gray {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w == 0 || h == 0 {
		return src
	}

	angleRad := angleDeg * (math.Pi / 180.0)
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	size := w
	if h > size {
		size = h
	}

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	outCx := float64(size) / 2.0

	dst := image.NewGray(image.Rect(0, 0, size, size))
	for i := range dst.Pix {
		dst.Pix[i] = 255
	}

	for oy := 0; oy < size; oy++ {
		for ox := 0; ox < size; ox++ {
			dx := float64(ox) - outCx
			dy := float64(oy) - outCx
			sx := cx + dx*cosA - dy*sinA
			sy := cy + dx*sinA + dy*cosA

			if sx < 0 || sx >= float64(w) || sy < 0 || sy >= float64(h) {
				continue
			}

			val := bilinearInterpolate(src, sx, sy)
			dst.Pix[oy*dst.Stride+ox] = val
		}
	}

	return dst
}

func bilinearInterpolate(src *image.Gray, x, y float64) uint8 {
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	stride := src.Stride
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	if xi < 0 || xi >= w-1 || yi < 0 || yi >= h-1 {
		xi = max(0, min(xi, w-1))
		yi = max(0, min(yi, h-1))
		return src.Pix[yi*stride+xi]
	}

	xFrac := x - float64(xi)
	yFrac := y - float64(yi)

	p00 := float64(src.Pix[yi*stride+xi])
	p10 := float64(src.Pix[yi*stride+xi+1])
	p01 := float64(src.Pix[(yi+1)*stride+xi])
	p11 := float64(src.Pix[(yi+1)*stride+xi+1])

	top := p00 + (p10-p00)*xFrac
	bottom := p01 + (p11-p01)*xFrac
	val := top + (bottom-top)*yFrac

	return uint8(math.Round(val))
}
