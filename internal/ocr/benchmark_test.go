package ocr

import (
	"image/color"
	"math/rand"
	"testing"
)

func benchMakeResult(confidence float64, line1, line2 string) *MRZPassResult {
	parsed, _ := NewMRZParser().ParseMRZ(line1, line2)
	return &MRZPassResult{
		Parsed:     parsed,
		Confidence: confidence,
	}
}

func BenchmarkSauvolaThreshold(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sauvolaThresholdDefault(img)
	}
}

func BenchmarkApplyCLAHE(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		return uint8(float64(x+y) * 255.0 / 400.0)
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyCLAHE(img, 3.0, 8)
	}
}

func BenchmarkMedianDenoise(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	rng := rand.New(rand.NewSource(42))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			if rng.Intn(10) == 0 {
				img.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		medianDenoise(img, 3)
	}
}

func BenchmarkBlurScore(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BlurScore(img)
	}
}

func BenchmarkDeskew(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if y%20 < 3 {
			return 0
		}
		return 255
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deskew(img)
	}
}

func BenchmarkUnsharpMask(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsharpMask(img, 1.5, 0.8)
	}
}

func BenchmarkGaussianBlur(b *testing.B) {
	img := createGrayImage(200, 200, func(x, y int) uint8 {
		if (x/5+y/5)%2 == 0 {
			return 0
		}
		return 255
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gaussianBlur(img, 1.5)
	}
}

func BenchmarkCalculateCheckDigit(b *testing.B) {
	s := "L898902C<3UTO6908061F9406236ZE184226B<<<<<14"
	for i := 0; i < b.N; i++ {
		CalculateCheckDigit(s)
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	a := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	s := "baaaaaaaaaabaaaaaaaaaabaaaaaaaaaabaaaaaaaaaaaaaa"
	for i := 0; i < b.N; i++ {
		levenshteinDistance(a, s)
	}
}

func BenchmarkParseMRZ(b *testing.B) {
	line1 := "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	line2 := "L898902C<3UTO6908061F9406236ZE184226B<<<<<14"
	parser := NewMRZParser()
	for i := 0; i < b.N; i++ {
		parser.ParseMRZ(line1, line2)
	}
}

func BenchmarkFuseMRZResults(b *testing.B) {
	line1 := "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<"
	line2 := "L898902C<3UTO6908061F9406236ZE184226B<<<<<14"
	r1 := benchMakeResult(0.85, line1, line2)
	r2 := benchMakeResult(0.90, line1, line2)
	r3 := benchMakeResult(0.80, line1, line2)
	results := []*MRZPassResult{r1, r2, r3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fuseMRZResults(results)
	}
}
