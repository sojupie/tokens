package motion

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Compare creates a local view over two preserved runs. It does not resample,
// align, or rewrite their evidence. Serve their common output root over HTTP.
func Compare(capture, reference, output string) (string, error) {
	if err := os.MkdirAll(output, 0755); err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp(output, time.Now().UTC().Format("20060102T150405-")+"comparison-*")
	if err != nil {
		return "", err
	}
	load := func(root string) (Report, error) {
		var r Report
		data, e := os.ReadFile(filepath.Join(root, "capture.json"))
		if e != nil {
			return r, e
		}
		if e = json.Unmarshal(data, &r); e != nil {
			return r, e
		}
		for i := range r.Presentation.Frames {
			base, _ := filepath.Abs(dir)
			target, _ := filepath.Abs(filepath.Join(root, r.Presentation.Frames[i].File))
			file, e := filepath.Rel(base, target)
			if e != nil {
				return r, e
			}
			r.Presentation.Frames[i].File = filepath.ToSlash(file)
		}
		if r.Presentation.Video != "" {
			base, _ := filepath.Abs(dir)
			target, _ := filepath.Abs(filepath.Join(root, r.Presentation.Video))
			path, e := filepath.Rel(base, target)
			if e != nil {
				return r, e
			}
			r.Presentation.Video = filepath.ToSlash(path)
		}
		return r, nil
	}
	a, err := load(capture)
	if err != nil {
		return dir, err
	}
	b, err := load(reference)
	if err != nil {
		return dir, err
	}
	if data, e := os.ReadFile(filepath.Join(reference, "pixels.json")); e == nil {
		if !json.Valid(data) {
			return dir, fmt.Errorf("invalid reference pixel trace")
		}
		b.Trace = data
	}
	data, err := json.Marshal(a)
	if err != nil {
		return dir, err
	}
	ref, err := json.Marshal(b)
	if err != nil {
		return dir, err
	}
	html, err := renderViewer(data, ref)
	if err != nil {
		return dir, err
	}
	if err = os.WriteFile(filepath.Join(dir, "review.html"), []byte(html), 0644); err != nil {
		return dir, err
	}
	text := fmt.Sprintf("# Motion comparison\n\nSource: %s\n\nImplementation: %s\n\n[Open source frames, implementation frames and traces](review.html).\n\nAssessment: not assessed. Alignment has not been inferred. Choose an input or animation marker and the matching source frame in the viewer.\n\n## Assistance needed\n\nRecord the specific unresolved question here, with source section and frame timestamps.\n", b.Recipe.Source, a.Recipe.URL)
	return dir, os.WriteFile(filepath.Join(dir, "review.md"), []byte(text), 0644)
}
