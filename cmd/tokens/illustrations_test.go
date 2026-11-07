package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestThumbnailJPEGDimensions(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 720, 450))
	for y := 0; y < 450; y++ {
		for x := 0; x < 720; x++ {
			source.SetRGBA(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	for _, size := range []struct{ width, height, quality int }{{720, 450, 80}, {360, 225, 75}} {
		data, err := thumbnailJPEG(source, size.width, size.height, size.quality)
		if err != nil {
			t.Fatal(err)
		}
		config, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width != size.width || config.Height != size.height {
			t.Fatalf("invalid thumbnail: %+v %v", config, err)
		}
	}
}
