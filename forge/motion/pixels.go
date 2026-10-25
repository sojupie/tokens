package motion

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
)

type EdgeReading struct {
	Position  *float64 `json:"position"`
	Contrast  float64  `json:"contrast"`
	PeakRatio float64  `json:"peakRatio"`
	Reason    string   `json:"reason,omitempty"`
}

// MeasureEdge measures an isolated vertical or horizontal brightness boundary.
// Ambiguous competing boundaries produce no position. PeakRatio is a separation
// measure, not a probability or proof that the intended object was selected.
func MeasureEdge(img image.Image, band image.Rectangle, axis string) (EdgeReading, error) {
	result := EdgeReading{}
	if !band.In(img.Bounds()) || band.Dx() < 3 || band.Dy() < 3 {
		return result, errors.New("band must fit inside the image and be at least 3×3")
	}
	if axis != "x" && axis != "y" {
		return result, errors.New("axis must be x or y")
	}
	n, across := band.Dx(), band.Dy()
	if axis == "y" {
		n, across = across, n
	}
	profile := make([]float64, n)
	for i := range n {
		for j := range across {
			x, y := band.Min.X+i, band.Min.Y+j
			if axis == "y" {
				x, y = band.Min.X+j, band.Min.Y+i
			}
			r, g, b, _ := img.At(x, y).RGBA()
			profile[i] += (.2126*float64(r) + .7152*float64(g) + .0722*float64(b)) / 257 / float64(across)
		}
	}
	strengths := make([]float64, n)
	best, index := 0.0, 0
	for i := 1; i < n-1; i++ {
		strengths[i] = math.Abs(profile[i+1] - profile[i-1])
		if strengths[i] > best {
			best, index = strengths[i], i
		}
	}
	second := 0.0
	for i, v := range strengths {
		if math.Abs(float64(i-index)) > 4 {
			second = math.Max(second, v)
		}
	}
	result.Contrast = best
	result.PeakRatio = best / math.Max(second, 1)
	if best < 8 {
		result.Reason = "insufficient contrast"
		return result, nil
	}
	if result.PeakRatio < 1.5 {
		result.Reason = "competing edges; narrow the band or mark the landmark manually"
		return result, nil
	}
	origin := band.Min.X
	if axis == "y" {
		origin = band.Min.Y
	}
	position := float64(origin+index) + .5
	result.Position = &position
	return result, nil
}

func TrackEdge(dir string, band image.Rectangle, axis string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "capture.json"))
	if err != nil {
		return "", err
	}
	var r Report
	if err = json.Unmarshal(data, &r); err != nil {
		return "", err
	}
	frames := []Sample{}
	readings := []EdgeReading{}
	missing := 0
	for _, f := range r.Presentation.Frames {
		if !validFramePath(f.File) {
			return "", errors.New("invalid captured frame path")
		}
		file, e := os.Open(filepath.Join(dir, f.File))
		if e != nil {
			return "", e
		}
		img, _, e := image.Decode(file)
		file.Close()
		if e != nil {
			return "", e
		}
		value, e := MeasureEdge(img, band, axis)
		if e != nil {
			return "", e
		}
		if value.Position == nil {
			missing++
		}
		readings = append(readings, value)
		frames = append(frames, Sample{T: f.T, Values: map[string]*float64{"edge." + axis: value.Position}})
	}
	out := filepath.Join(dir, "pixels.json")
	data, err = json.MarshalIndent(map[string]any{"schemaVersion": 1, "kind": "image-edge", "unit": "recording px", "band": []int{band.Min.X, band.Min.Y, band.Dx(), band.Dy()}, "axis": axis, "frames": frames, "readings": readings}, "", "  ")
	if err != nil {
		return "", err
	}
	if err = os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return "", err
	}
	review, err := os.OpenFile(filepath.Join(dir, "review.md"), os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(review, "\n## Pixel edge measurement\n\nBand: %v. Axis: %s. Units: recording pixels. Ambiguous/undetected frames: %d of %d.\n\n[Measured trace](pixels.json). Confirm the band contains the intended boundary before comparing it.\n", band, axis, missing, len(frames))
	return out, errors.Join(err, review.Close())
}

func validFramePath(name string) bool {
	if len(name) != len("frames/000000.jpg") || name[:7] != "frames/" || name[len(name)-4:] != ".jpg" {
		return false
	}
	for _, c := range name[7:13] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
