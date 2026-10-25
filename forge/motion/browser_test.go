package motion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Enable explicitly: TOKENS_TEST_CHROME_PATH=/path/to/chrome go test ./forge/motion -run TestBrowserCapture -v
func TestBrowserCapture(t *testing.T) {
	chrome := os.Getenv("TOKENS_TEST_CHROME_PATH")
	if chrome == "" {
		t.Skip("set TOKENS_TEST_CHROME_PATH for browser capture")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<!doctype html><style>#item{width:100px;height:50px;background:black;border-radius:50%;transition:transform 100ms linear}.moving{transform:translateX(30px)}</style><button onclick="item.classList.add('moving')">Move</button><div id="item"></div>`))
	}))
	defer server.Close()
	recipe := Recipe{URL: server.URL, Width: 700, Height: 500, Duration: 500, Tracks: []Track{{Name: "x", Selector: "#item", Property: "translateX"}, {Name: "radius", Selector: "#item", Property: "borderTopLeftRadius"}}, Actions: []Action{{At: 100, Type: "click", Selector: "button"}}}
	dir, err := Capture(context.Background(), recipe, chrome, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "capture.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r Report
	if err = json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	var trace struct {
		Frames []Sample `json:"frames"`
		Issues []struct {
			Track string `json:"track"`
		} `json:"measurementIssues"`
		Events []struct {
			Type string `json:"type"`
		} `json:"events"`
		Animations []struct {
			Start *float64 `json:"startTime"`
		} `json:"animations"`
	}
	if err = json.Unmarshal(r.Trace, &trace); err != nil {
		t.Fatal(err)
	}
	if !r.Complete || len(r.Presentation.Frames) < 3 || len(trace.Frames) < 5 {
		t.Fatalf("incomplete capture: %+v", r.Presentation)
	}
	if *trace.Frames[0].Values["x"] != 0 || *trace.Frames[len(trace.Frames)-1].Values["x"] != 30 {
		t.Fatal("input did not drive real transition")
	}
	for i, f := range r.Presentation.Frames {
		if i > 0 && f.T < r.Presentation.Frames[i-1].T {
			t.Fatal("unordered frame timestamps")
		}
		if _, err = os.Stat(filepath.Join(dir, f.File)); err != nil {
			t.Fatal(err)
		}
	}
	if len(trace.Issues) != 1 || trace.Issues[0].Track != "radius" {
		t.Fatal("unsupported measurement was not explained")
	}
	if len(trace.Animations) != 1 || trace.Animations[0].Start == nil {
		t.Fatal("animation timing origin was lost")
	}
	recipe.Actions[0].Selector = "#absent"
	failed, err := Capture(context.Background(), recipe, chrome, t.TempDir())
	if err == nil {
		t.Fatal("bad action accepted")
	}
	data, err = os.ReadFile(filepath.Join(failed, "capture.json"))
	if err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(data, &r)
	if r.Complete || string(r.Trace) == "null" {
		t.Fatal("failed capture did not preserve partial evidence")
	}
}
