package ocr

import "image"

func preprocessPipeline(src *image.Gray, mrzMode bool) *image.Gray {
	if src == nil {
		return nil
	}

	result := medianDenoise(src, medianKernelSize)
	result = applyCLAHE(result, claheClipLimit, claheTileSize)
	result = deskew(result)
	result = unsharpMask(result, unsharpSigma, unsharpAmount)

	if mrzMode {
		result = sauvolaThresholdDefault(result)
	}

	return result
}
