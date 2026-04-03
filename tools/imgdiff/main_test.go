package main

import (
	"image"
	"image/color"
	"testing"
)

func TestDiffPercent_IdenticalImages(t *testing.T) {
	img := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	pct := diffPercent(img, img)
	if pct != 0.0 {
		t.Errorf("expected 0%% diff for identical images, got %.4f%%", pct)
	}
}

func TestDiffPercent_CompletelyDifferent(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	pct := diffPercent(a, b)
	if pct != 100.0 {
		t.Errorf("expected 100%% diff for completely different images, got %.4f%%", pct)
	}
}

func TestDiffPercent_PartialDiff(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Change 1 pixel out of 100 → 1%
	b.Set(0, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	pct := diffPercent(a, b)
	if pct != 1.0 {
		t.Errorf("expected 1%% diff, got %.4f%%", pct)
	}
}

func TestDiffPercent_DifferentDimensions(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(20, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	pct := diffPercent(a, b)
	if pct != 100.0 {
		t.Errorf("expected 100%% for different dimensions, got %.4f%%", pct)
	}
}

func TestDiffPercent_EmptyImages(t *testing.T) {
	a := solidImage(0, 0, color.RGBA{})
	b := solidImage(0, 0, color.RGBA{})
	pct := diffPercent(a, b)
	if pct != 0.0 {
		t.Errorf("expected 0%% for empty images, got %.4f%%", pct)
	}
}

func TestDiffPercent_EmptyVsNonEmpty(t *testing.T) {
	empty := solidImage(0, 0, color.RGBA{})
	filled := solidImage(10, 10, color.RGBA{R: 255, A: 255})
	pct := diffPercent(empty, filled)
	if pct != 100.0 {
		t.Errorf("expected 100%% for empty vs non-empty, got %.4f%%", pct)
	}
}

func TestToNRGBA_Passthrough(t *testing.T) {
	// An *image.NRGBA should be returned as-is without copying.
	orig := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	got := toNRGBA(orig)
	if got != orig {
		t.Error("expected toNRGBA to return the same *image.NRGBA pointer")
	}
}

func TestToNRGBA_ConvertsRGBA(t *testing.T) {
	src := solidImage(2, 2, color.RGBA{R: 100, G: 200, B: 50, A: 255})
	got := toNRGBA(src)
	// Verify dimensions are preserved
	if got.Bounds() != src.Bounds() {
		t.Errorf("bounds mismatch: %v vs %v", got.Bounds(), src.Bounds())
	}
	// Spot-check a pixel
	r, g, b, a := got.At(0, 0).RGBA()
	er, eg, eb, ea := src.At(0, 0).RGBA()
	if r != er || g != eg || b != eb || a != ea {
		t.Error("pixel colour mismatch after NRGBA conversion")
	}
}

// solidImage creates an RGBA image filled with a single colour.
func solidImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}
