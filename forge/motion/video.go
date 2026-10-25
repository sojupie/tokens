package motion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// ImportVideo retains the original and extracts each decoded frame using its
// presentation timestamp. It does not invent input/release timestamps.
func ImportVideo(ctx context.Context, file, source, mediaURL, output string) (string, error) {
	absolute, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	file = absolute
	command := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "frame=best_effort_timestamp_time", "-of", "json", file)
	raw, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("read video timestamps: %w", err)
	}
	var probe struct {
		Frames []struct {
			Time string `json:"best_effort_timestamp_time"`
		} `json:"frames"`
	}
	if err = json.Unmarshal(raw, &probe); err != nil {
		return "", err
	}
	if len(probe.Frames) == 0 || len(probe.Frames) > 3600 {
		return "", errors.New("import a clip containing 1–3600 frames")
	}
	times := make([]float64, len(probe.Frames))
	for i, f := range probe.Frames {
		v, e := strconv.ParseFloat(f.Time, 64)
		if e != nil {
			return "", fmt.Errorf("frame %d has no timestamp", i)
		}
		times[i] = v * 1000
		if i > 0 && times[i] < times[i-1] {
			return "", errors.New("video timestamps are not ordered")
		}
	}
	if err = os.MkdirAll(output, 0755); err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp(output, time.Now().UTC().Format("20060102T150405-")+"video-*")
	if err != nil {
		return "", err
	}
	if err = os.Mkdir(filepath.Join(dir, "frames"), 0755); err != nil {
		return dir, err
	}
	in, err := os.Open(file)
	if err != nil {
		return dir, err
	}
	defer in.Close()
	out, err := os.Create(filepath.Join(dir, "source"+filepath.Ext(file)))
	if err != nil {
		return dir, err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(out, hash), in)
	closeErr := out.Close()
	if err = errors.Join(copyErr, closeErr); err != nil {
		return dir, err
	}
	command = exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", file, "-map", "0:v:0", "-fps_mode", "passthrough", "-q:v", "2", "-start_number", "0", filepath.Join(dir, "frames", "%06d.jpg"))
	if log, e := command.CombinedOutput(); e != nil {
		return dir, fmt.Errorf("extract frames: %w: %s", e, log)
	}
	r := Report{SchemaVersion: 1, Complete: true, Recipe: Recipe{URL: "Source video", Source: source}, Trace: json.RawMessage(`null`), Presentation: Presentation{Kind: "source-video", Video: "source" + filepath.Ext(file), Frames: []Frame{}}}
	for i, t := range times {
		name := fmt.Sprintf("frames/%06d.jpg", i)
		if _, err = os.Stat(filepath.Join(dir, name)); err != nil {
			return dir, err
		}
		r.Presentation.Frames = append(r.Presentation.Frames, Frame{T: t - times[0], File: name})
	}
	r.Presentation.OriginalFile = "source" + filepath.Ext(file)
	r.Presentation.OriginalURL = mediaURL
	r.Presentation.OriginalSHA256 = hex.EncodeToString(hash.Sum(nil))
	if err = writeReport(dir, r); err != nil {
		return dir, err
	}
	return dir, nil
}
