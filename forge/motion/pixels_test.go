package motion

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestEdgeTrackingFindsMotionAndRejectsCompetingEdges(t *testing.T) {
	for _, x := range []int{12, 24, 36} {
		img := image.NewGray(image.Rect(0, 0, 60, 20))
		for j := 0; j < 20; j++ {
			for i := x; i < 60; i++ {
				img.SetGray(i, j, color.Gray{Y: 240})
			}
		}
		v, err := MeasureEdge(img, img.Bounds(), "x")
		if err != nil || v.Position == nil || math.Abs(*v.Position-float64(x)) > .6 {
			t.Fatalf("edge %d: %+v %v", x, v, err)
		}
	}
	img := image.NewGray(image.Rect(0, 0, 60, 20))
	for j := 0; j < 20; j++ {
		for i := 15; i < 40; i++ {
			img.SetGray(i, j, color.Gray{Y: 240})
		}
	}
	v, err := MeasureEdge(img, img.Bounds(), "x")
	if err != nil || v.Position != nil || v.Reason == "" {
		t.Fatalf("competing edges accepted: %+v", v)
	}
}
func TestFramePathsStayInsideCapture(t *testing.T) {
	for _, p := range []string{"../secret.jpg", "frames/../../x.jpg", "frames/000001.jpg'\nfile 'x"} {
		if validFramePath(p) {
			t.Fatal(p)
		}
	}
	if !validFramePath("frames/000001.jpg") {
		t.Fatal("valid path rejected")
	}
}
