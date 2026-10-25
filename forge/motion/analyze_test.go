package motion

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalysisPreservesOvershootAndRequiresMeasuredWindow(t *testing.T) {
	p := filepath.Join(t.TempDir(), "trace.json")
	raw := []byte(`{"frames":[{"t":0,"values":{"x":100}},{"t":50,"values":{"x":-20}},{"t":100,"values":{"x":0}}]}`)
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(p, "x", 0, 100, 100, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Overshoot-.2) > 1e-10 || r.RMSProgressError != nil {
		t.Fatalf("bad signed analysis: %+v", r)
	}
	if _, err = Analyze(p, "x", 0, 150, 100, 0, nil); err == nil {
		t.Fatal("accepted unsampled end")
	}
	if _, err = Analyze(p, "missing", 0, 100, 0, 1, nil); err == nil {
		t.Fatal("accepted missing channel")
	}
}
func TestSuppliedCurveAndDuration(t *testing.T) {
	curve := []float64{.2, 0, 0, 1}
	frames := []Sample{}
	for i := range 21 {
		v := 8 + 20*Progress(float64(i)/20, curve)
		frames = append(frames, Sample{T: float64(i) * 25, Values: map[string]*float64{"x": &v}})
	}
	data, _ := json.Marshal(map[string]any{"trace": map[string]any{"frames": frames}})
	p := filepath.Join(t.TempDir(), "capture.json")
	os.WriteFile(p, data, 0600)
	r, err := Analyze(p, "x", 0, 500, 8, 28, curve)
	if err != nil {
		t.Fatal(err)
	}
	if r.Duration != 500 || *r.RMSProgressError > 1e-10 {
		t.Fatalf("bad comparison: %+v", r)
	}
}
func TestRecipeRejectsUnorderedOrUnknownActions(t *testing.T) {
	r := Recipe{URL: "http://localhost", Width: 800, Height: 600, Duration: 500, Tracks: []Track{{Name: "x", Selector: "button", Property: "x"}}, Actions: []Action{{At: 300, Type: "key", Key: "Enter"}, {At: 100, Type: "key", Key: "Enter"}}}
	if Validate(r) == nil {
		t.Fatal("unordered actions accepted")
	}
	r.Actions = []Action{{At: 100, Type: "invented"}}
	if Validate(r) == nil {
		t.Fatal("unknown action accepted")
	}
}

func TestRefreshingViewerPreservesReviewerNotes(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "review.md")
	if err := os.WriteFile(p, []byte("Reviewer decision: needs source comparison.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := Report{SchemaVersion: 1, Trace: json.RawMessage(`null`)}
	if err := writeReport(dir, r); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Reviewer decision: needs source comparison.\n" {
		t.Fatal("encoding overwrote reviewer notes")
	}
}
