// Package motion records repeatable browser interactions for visual inspection.
package motion

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

//go:embed recorder.js
var recorderJS string

//go:embed viewer.html
var viewerHTML string

type Track struct {
	Optional bool   `json:"optional,omitempty"`
	Name     string `json:"name"`
	Selector string `json:"selector"`
	Property string `json:"property"`
	Unit     string `json:"unit"`
}
type Action struct {
	At       float64  `json:"at"`
	Type     string   `json:"type"`
	Selector string   `json:"selector,omitempty"`
	Key      string   `json:"key,omitempty"`
	X        *float64 `json:"x,omitempty"`
	Y        *float64 `json:"y,omitempty"`
}
type Recipe struct {
	URL      string   `json:"url"`
	Source   string   `json:"source,omitempty"`
	Width    int64    `json:"width"`
	Height   int64    `json:"height"`
	Duration float64  `json:"duration"`
	Ready    string   `json:"ready,omitempty"`
	Tracks   []Track  `json:"tracks"`
	Actions  []Action `json:"actions"`
}
type Frame struct {
	T    float64 `json:"t"`
	File string  `json:"file"`
}
type Presentation struct {
	OriginalFile   string  `json:"originalFile,omitempty"`
	OriginalURL    string  `json:"originalURL,omitempty"`
	OriginalSHA256 string  `json:"originalSha256,omitempty"`
	Kind           string  `json:"kind"`
	Frames         []Frame `json:"frames"`
	Dropped        int     `json:"dropped"`
	Error          string  `json:"error,omitempty"`
	Video          string  `json:"video,omitempty"`
}
type Report struct {
	SchemaVersion int             `json:"schemaVersion"`
	Complete      bool            `json:"complete"`
	Error         string          `json:"error,omitempty"`
	Recipe        Recipe          `json:"recipe"`
	Trace         json.RawMessage `json:"trace"`
	Presentation  Presentation    `json:"presentation"`
}

func Validate(r Recipe) error {
	if !strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://") {
		return errors.New("recipe URL must use HTTP or HTTPS")
	}
	if r.Width < 100 || r.Height < 100 || r.Width > 4096 || r.Height > 4096 {
		return errors.New("viewport dimensions must be 100–4096 CSS px")
	}
	if r.Duration < 100 || r.Duration > 60000 {
		return errors.New("duration must be 100–60000 ms")
	}
	if len(r.Tracks) == 0 || len(r.Tracks) > 32 {
		return errors.New("choose 1–32 tracks")
	}
	previous := float64(0)
	for _, a := range r.Actions {
		if a.At < previous || a.At > r.Duration {
			return errors.New("actions must be ordered and within duration")
		}
		previous = a.At
		switch a.Type {
		case "hover", "press", "release", "click", "key", "touchstart", "touchmove", "touchend":
		default:
			return fmt.Errorf("unknown action %q", a.Type)
		}
		if a.Type == "key" && a.Key == "" {
			return errors.New("key action requires key")
		}
		if a.Type != "key" && a.Type != "touchend" && a.Selector == "" && (a.X == nil || a.Y == nil) {
			return errors.New("pointer action requires a selector or both coordinates")
		}
	}
	return nil
}

// Capture uses its own browser profile. Screencast frame times and DOM events
// share the browser epoch; screencasting can perturb cadence and is not a phone
// display recording. No frame rate or conformance pass is inferred.
func Capture(ctx context.Context, r Recipe, chrome, output string) (dir string, err error) {
	if err = Validate(r); err != nil {
		return "", err
	}
	if err = os.MkdirAll(output, 0755); err != nil {
		return "", err
	}
	dir, err = os.MkdirTemp(output, time.Now().UTC().Format("20060102T150405-")+"*")
	if err != nil {
		return "", err
	}
	report := Report{SchemaVersion: 1, Recipe: r, Trace: json.RawMessage(`null`), Presentation: Presentation{Kind: "chromium-screencast", Frames: []Frame{}}}
	defer func() {
		report.Complete = err == nil
		if err != nil {
			report.Error = err.Error()
		}
		if e := writeReport(dir, report); err == nil {
			err = e
		}
	}()
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts, chromedp.WindowSize(int(r.Width), int(r.Height)), chromedp.Flag("hide-scrollbars", false))
	if chrome != "" {
		opts = append(opts, chromedp.ExecPath(chrome))
	}
	allocator, stop := chromedp.NewExecAllocator(ctx, opts...)
	defer stop()
	browser, closeBrowser := chromedp.NewContext(allocator)
	defer closeBrowser()
	browser, cancel := context.WithTimeout(browser, time.Duration(r.Duration)*time.Millisecond+45*time.Second)
	defer cancel()
	ready := r.Ready
	if ready == "" {
		ready = "document.readyState === 'complete'"
	}
	if err = chromedp.Run(browser, emulation.SetDeviceMetricsOverride(r.Width, r.Height, 1, false), chromedp.Navigate(r.URL), chromedp.Poll(ready, nil, chromedp.WithPollingTimeout(25*time.Second)), chromedp.Evaluate(`document.fonts.ready.then(()=>true)`, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }), chromedp.ScrollIntoView(r.Tracks[0].Selector, chromedp.ByQuery)); err != nil {
		return dir, err
	}
	if err = chromedp.Run(browser, chromedp.Evaluate(recorderJS, nil)); err != nil {
		return dir, err
	}
	config, _ := json.Marshal(r)
	if err = chromedp.Run(browser, chromedp.Evaluate(`window.__motionCapture=MotionInspection.instrument(`+string(config)+`)`, nil)); err != nil {
		return dir, err
	}
	var epoch float64
	if err = chromedp.Run(browser, chromedp.Evaluate(`performance.timeOrigin+performance.now()`, &epoch)); err != nil {
		return dir, err
	}
	var exactEpoch float64
	if err = os.Mkdir(filepath.Join(dir, "frames"), 0755); err != nil {
		return dir, err
	}
	frames := make(chan *page.EventScreencastFrame, 64)
	var mu sync.Mutex
	closed := false
	done := make(chan struct{})
	var frameError error
	listening, stopListening := context.WithCancel(browser)
	defer stopListening()
	chromedp.ListenTarget(listening, func(event any) {
		if f, ok := event.(*page.EventScreencastFrame); ok {
			mu.Lock()
			defer mu.Unlock()
			if closed {
				return
			}
			select {
			case frames <- f:
			default:
				report.Presentation.Dropped++
			}
		}
	})
	go func() {
		defer close(done)
		for f := range frames {
			_ = chromedp.Run(browser, page.ScreencastFrameAck(f.SessionID))
			if frameError != nil {
				continue
			}
			if f.Metadata.Timestamp == nil {
				frameError = errors.New("screencast frame has no timestamp")
				continue
			}
			data, e := base64.StdEncoding.DecodeString(f.Data)
			if e != nil {
				frameError = e
				continue
			}
			file := fmt.Sprintf("frames/%06d.jpg", len(report.Presentation.Frames))
			if e = os.WriteFile(filepath.Join(dir, file), data, 0644); e != nil {
				frameError = e
				continue
			}
			report.Presentation.Frames = append(report.Presentation.Frames, Frame{T: float64(f.Metadata.Timestamp.Time().UnixMicro())/1000 - epoch, File: file})
		}
	}()
	defer func() {
		stopListening()
		mu.Lock()
		closed = true
		close(frames)
		mu.Unlock()
		<-done
		if string(report.Trace) == "null" {
			_ = chromedp.Run(browser, chromedp.Evaluate(`window.__motionCapture.stop()`, &report.Trace))
		}
		var stopped struct {
			Epoch float64 `json:"epoch"`
		}
		if json.Unmarshal(report.Trace, &stopped) == nil && stopped.Epoch != 0 {
			exactEpoch = stopped.Epoch
		}
		if exactEpoch != 0 {
			for i := range report.Presentation.Frames {
				report.Presentation.Frames[i].T += epoch - exactEpoch
			}
		}
		if err == nil && frameError != nil {
			err = frameError
		}
		if report.Presentation.Dropped > 0 && err == nil {
			err = errors.New("screencast queue dropped frames")
		}
	}()
	if err = chromedp.Run(browser, page.StartScreencast().WithFormat(page.ScreencastFormatJpeg).WithQuality(95).WithEveryNthFrame(1)); err != nil {
		return dir, err
	}
	started := time.Now()
	for _, a := range r.Actions {
		if err = waitUntil(browser, started.Add(time.Duration(a.At*float64(time.Millisecond)))); err != nil {
			return dir, err
		}
		if err = perform(browser, a); err != nil {
			return dir, err
		}
	}
	if err = waitUntil(browser, started.Add(time.Duration(r.Duration*float64(time.Millisecond)))); err != nil {
		return dir, err
	}
	if err = chromedp.Run(browser, chromedp.Evaluate(`window.__motionCapture.stop()`, &report.Trace), page.StopScreencast()); err != nil {
		return dir, err
	}
	// Correct all presentation timestamps using the recorder's exact start epoch.
	var trace struct {
		Epoch         float64 `json:"epoch"`
		SamplingError string  `json:"samplingError"`
	}
	if e := json.Unmarshal(report.Trace, &trace); e != nil {
		return dir, e
	}
	// Store the offset now; the writer uses it after all queued frames are drained.
	exactEpoch = trace.Epoch
	if trace.SamplingError != "" {
		return dir, errors.New(trace.SamplingError)
	}
	return dir, nil
}

func waitUntil(ctx context.Context, until time.Time) error {
	timer := time.NewTimer(time.Until(until))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func perform(ctx context.Context, a Action) error {
	data, _ := json.Marshal(a)
	if err := chromedp.Run(ctx, chromedp.Evaluate(`window.__motionCapture.mark('recipe-action',`+string(data)+`)`, nil)); err != nil {
		return err
	}
	if a.Type == "key" {
		if a.Selector != "" {
			selector, _ := json.Marshal(a.Selector)
			if err := chromedp.Run(ctx, chromedp.Evaluate(`(()=>{const nodes=document.querySelectorAll(`+string(selector)+`);if(nodes.length!==1)throw Error('Key target must match one element');nodes[0].focus({preventScroll:true});if(document.activeElement!==nodes[0])throw Error('Key target did not receive focus')})()`, nil)); err != nil {
				return err
			}
		}
		key := map[string]string{"ArrowRight": kb.ArrowRight, "ArrowLeft": kb.ArrowLeft, "ArrowUp": kb.ArrowUp, "ArrowDown": kb.ArrowDown, "Enter": kb.Enter, "Escape": kb.Escape, "Tab": kb.Tab, "Home": kb.Home, "End": kb.End, "Space": " "}[a.Key]
		if key == "" {
			key = a.Key
		}
		return chromedp.Run(ctx, chromedp.KeyEvent(key))
	}
	if a.Type == "touchend" {
		return chromedp.Run(ctx, input.DispatchTouchEvent(input.TouchEnd, []*input.TouchPoint{}))
	}
	x, y := float64(0), float64(0)
	if a.Selector != "" {
		selector, _ := json.Marshal(a.Selector)
		var box struct{ X, Y float64 }
		if err := chromedp.Run(ctx, chromedp.Evaluate(`(()=>{const els=document.querySelectorAll(`+string(selector)+`);if(els.length!==1)throw Error('Action selector must match one element');const el=els[0],r=el.getBoundingClientRect();if(r.width===0||r.height===0||r.x+r.width/2<0||r.x+r.width/2>innerWidth||r.y+r.height/2<0||r.y+r.height/2>innerHeight)throw Error('Action target is not visible in viewport');if(!el.contains(document.elementFromPoint(r.x+r.width/2,r.y+r.height/2)))throw Error('Action target is covered by another element');return{x:r.x+r.width/2,y:r.y+r.height/2}})()`, &box)); err != nil {
			return err
		}
		x, y = box.X, box.Y
	} else {
		x, y = *a.X, *a.Y
	}
	switch a.Type {
	case "click":
		return chromedp.Run(ctx, chromedp.MouseClickXY(x, y))
	case "hover":
		return chromedp.Run(ctx, chromedp.MouseEvent(input.MouseMoved, x, y))
	case "press":
		return chromedp.Run(ctx, chromedp.MouseEvent(input.MousePressed, x, y, chromedp.ButtonLeft, chromedp.ClickCount(1)))
	case "release":
		return chromedp.Run(ctx, chromedp.MouseEvent(input.MouseReleased, x, y, chromedp.ButtonLeft, chromedp.ClickCount(1)))
	case "touchstart", "touchmove":
		typ := input.TouchStart
		if a.Type == "touchmove" {
			typ = input.TouchMove
		}
		return chromedp.Run(ctx, input.DispatchTouchEvent(typ, []*input.TouchPoint{{X: x, Y: y}}))
	}
	return errors.New("unhandled action")
}

func writeReport(dir string, r Report) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dir, "capture.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	html, err := renderViewer(data, []byte("null"))
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dir, "review.html"), []byte(html), 0644); err != nil {
		return err
	}
	lines := []string{"# Motion review", "", "Source: " + r.Recipe.Source, "", "Implementation: " + r.Recipe.URL, "", "[Open traces and browser frames](review.html)", "", "Assessment: not assessed.", "", "The capture contains DOM/CSS samples and Chromium screencast frames. Frame delivery may be irregular. It does not measure physical display presentation.", "", "## Questions", "", "- Do the measured channels describe the reference behavior?", "- Do the rendered frames agree with those channels?", "- Which differences remain unverified?", ""}
	if r.Presentation.Kind == "source-video" {
		lines = []string{"# Motion review", "", "Source: " + r.Recipe.Source, "", "Media: " + r.Presentation.OriginalURL, "", "Original SHA-256: `" + r.Presentation.OriginalSHA256 + "`", "", "[Open decoded source frames](review.html)", "", "Assessment: not assessed. Input/release time has not been identified.", "", "Frame zero is the first decoded frame; original frame intervals are preserved.", "", "## Questions", "", "Record the unresolved behavior and relevant source frame timestamps here.", ""}
	}
	var trace struct {
		Events []struct {
			T    float64 `json:"t"`
			Type string  `json:"type"`
		} `json:"events"`
		Animations []struct {
			Selector  string   `json:"selector"`
			StartTime *float64 `json:"startTime"`
			Timing    struct {
				Duration any    `json:"duration"`
				Easing   string `json:"easing"`
			} `json:"timing"`
		} `json:"animations"`
	}
	if json.Unmarshal(r.Trace, &trace) == nil && (len(trace.Events) > 0 || len(trace.Animations) > 0) {
		lines = append(lines, "## Recorded timing", "", "| Marker | Time (ms) |", "| --- | ---: |")
		for _, e := range trace.Events {
			lines = append(lines, fmt.Sprintf("| %s | %.3f |", e.Type, e.T))
		}
		if len(trace.Animations) > 0 {
			lines = append(lines, "", "| Implementation animation | Start (ms) | Duration (ms) | Easing |", "| --- | ---: | --- | --- |")
			for _, a := range trace.Animations {
				start := "pending"
				if a.StartTime != nil {
					start = fmt.Sprintf("%.3f", *a.StartTime)
				}
				lines = append(lines, fmt.Sprintf("| `%s` | %s | %v | %s |", a.Selector, start, a.Timing.Duration, a.Timing.Easing))
			}
		}
	}
	var issues struct {
		Items []struct {
			Track  string `json:"track"`
			Reason string `json:"reason"`
		} `json:"measurementIssues"`
	}
	if json.Unmarshal(r.Trace, &issues) == nil && len(issues.Items) > 0 {
		lines = append(lines, "", "## Measurement issues", "")
		for _, issue := range issues.Items {
			lines = append(lines, "- "+issue.Track+": "+issue.Reason)
		}
	}

	lines = append(lines, "")
	review, err := os.OpenFile(filepath.Join(dir, "review.md"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, writeErr := review.WriteString(strings.Join(lines, "\n"))
	return errors.Join(writeErr, review.Close())
}

// Encode preserves variable frame intervals. The JSON and JPEGs remain primary
// evidence; the video is a convenience export and never a 60fps claim.
func Encode(dir string) error {
	data, err := os.ReadFile(filepath.Join(dir, "capture.json"))
	if err != nil {
		return err
	}
	var r Report
	if err = json.Unmarshal(data, &r); err != nil {
		return err
	}
	frames := r.Presentation.Frames
	if len(frames) < 2 {
		return errors.New("need at least two browser frames")
	}
	var duration struct {
		Duration float64 `json:"duration"`
	}
	if err = json.Unmarshal(r.Trace, &duration); err != nil {
		return err
	}
	var list strings.Builder
	list.WriteString("ffconcat version 1.0\n")
	for i, f := range frames {
		if !validFramePath(f.File) {
			return errors.New("invalid captured frame path")
		}
		end := duration.Duration
		if i+1 < len(frames) {
			end = frames[i+1].T
		}
		if end <= f.T {
			continue
		}
		fmt.Fprintf(&list, "file '%s'\noption framerate 1000\nduration %.9f\n", f.File, (end-f.T)/1000)
	}
	fmt.Fprintf(&list, "file '%s'\noption framerate 1000\n", frames[len(frames)-1].File)
	if err = os.WriteFile(filepath.Join(dir, "frames.ffconcat"), []byte(list.String()), 0644); err != nil {
		return err
	}
	command := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", "frames.ffconcat", "-fps_mode", "vfr", "-enc_time_base", "1:1000", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2", "-movflags", "+faststart", "capture.mp4")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("encode: %w: %s", err, output)
	}
	r.Presentation.Video = "capture.mp4"
	return writeReport(dir, r)
}

func renderViewer(capture, reference []byte) (string, error) {
	for _, marker := range []string{"/*CAPTURE*/ null", "/*REFERENCE*/ null", "/*RECORDER*/"} {
		if strings.Count(viewerHTML, marker) != 1 {
			return "", fmt.Errorf("viewer placeholder %s is missing or duplicated", marker)
		}
	}
	return strings.NewReplacer("/*CAPTURE*/ null", string(capture), "/*REFERENCE*/ null", string(reference), "/*RECORDER*/", recorderJS).Replace(viewerHTML), nil
}
