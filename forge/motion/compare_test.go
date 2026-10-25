package motion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComparisonReferencesOriginalFramesWithoutRewritingRuns(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	os.Mkdir(a, 0755)
	os.Mkdir(b, 0755)
	report := Report{SchemaVersion: 1, Complete: true, Recipe: Recipe{URL: "http://localhost"}, Trace: json.RawMessage(`{"frames":[]}`), Presentation: Presentation{Frames: []Frame{{T: 0, File: "frames/000000.jpg"}}}}
	data, _ := json.Marshal(report)
	os.WriteFile(filepath.Join(a, "capture.json"), data, 0600)
	os.WriteFile(filepath.Join(b, "capture.json"), data, 0600)
	dir, err := Compare(a, b, root)
	if err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(dir, "review.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "../a/frames/000000.jpg") || !strings.Contains(string(html), "../b/frames/000000.jpg") {
		t.Fatal("comparison lost original frame references")
	}
	original, _ := os.ReadFile(filepath.Join(a, "capture.json"))
	if string(original) != string(data) {
		t.Fatal("comparison rewrote source capture")
	}
}
