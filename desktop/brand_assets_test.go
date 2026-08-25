package main

import (
	"bytes"
	"encoding/binary"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestBrandPNGAssetsHaveTransparentCorners(t *testing.T) {
	for _, name := range []string{
		"build/appicon.png",
		"frontend/src/assets/logo-symbol.png",
		"frontend/src/assets/logo-wordmark.png",
	} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		bounds := img.Bounds()
		corners := []image.Point{{bounds.Min.X, bounds.Min.Y}, {bounds.Max.X - 1, bounds.Min.Y}, {bounds.Min.X, bounds.Max.Y - 1}, {bounds.Max.X - 1, bounds.Max.Y - 1}}
		for _, point := range corners {
			_, _, _, alpha := img.At(point.X, point.Y).RGBA()
			if alpha != 0 {
				t.Fatalf("%s corner %v has alpha %d; expected transparent", name, point, alpha)
			}
		}
	}
}

func TestSymbolHasBalancedTransparentSafeArea(t *testing.T) {
	data, err := os.ReadFile("frontend/src/assets/logo-symbol.png")
	if err != nil {
		t.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	bounds := img.Bounds()
	alphaBounds := image.Rectangle{Min: image.Point{X: bounds.Max.X, Y: bounds.Max.Y}}
	alphaBounds.Max = image.Point{X: bounds.Min.X, Y: bounds.Min.Y}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha == 0 {
				continue
			}
			if x < alphaBounds.Min.X {
				alphaBounds.Min.X = x
			}
			if y < alphaBounds.Min.Y {
				alphaBounds.Min.Y = y
			}
			if x+1 > alphaBounds.Max.X {
				alphaBounds.Max.X = x + 1
			}
			if y+1 > alphaBounds.Max.Y {
				alphaBounds.Max.Y = y + 1
			}
		}
	}
	left := alphaBounds.Min.X - bounds.Min.X
	top := alphaBounds.Min.Y - bounds.Min.Y
	right := bounds.Max.X - alphaBounds.Max.X
	bottom := bounds.Max.Y - alphaBounds.Max.Y
	if left < 35 || top < 35 || right < 35 || bottom < 35 {
		t.Fatalf("symbol safe area is too small: left=%d top=%d right=%d bottom=%d", left, top, right, bottom)
	}
	if absInt(left-right) > 2 || absInt(top-bottom) > 2 {
		t.Fatalf("symbol safe area is unbalanced: left=%d top=%d right=%d bottom=%d", left, top, right, bottom)
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func TestWindowsBrandICOHasRequiredSizes(t *testing.T) {
	data, err := os.ReadFile(filepath.FromSlash("build/windows/icon.ico"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 6 || binary.LittleEndian.Uint16(data[0:2]) != 0 || binary.LittleEndian.Uint16(data[2:4]) != 1 {
		t.Fatalf("invalid ICO header")
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	want := map[int]bool{16: false, 20: false, 24: false, 32: false, 48: false, 64: false, 128: false, 256: false}
	for i := 0; i < count; i++ {
		at := 6 + i*16
		if at+16 > len(data) {
			t.Fatalf("truncated ICO directory")
		}
		width := int(data[at])
		if width == 0 {
			width = 256
		}
		if _, ok := want[width]; ok {
			want[width] = true
		}
		length := int(binary.LittleEndian.Uint32(data[at+8 : at+12]))
		offset := int(binary.LittleEndian.Uint32(data[at+12 : at+16]))
		if offset < 0 || length <= 0 || offset+length > len(data) {
			t.Fatalf("invalid ICO frame for %dpx", width)
		}
	}
	for size, present := range want {
		if !present {
			t.Errorf("ICO is missing %dpx frame", size)
		}
	}
}
