package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/sojupie/tokens/forge/motion"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func runMotion(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: tokens motion capture|import|track|analyze|compare|encode [flags]")
	}
	flags := flag.NewFlagSet("tokens motion "+args[0], flag.ContinueOnError)
	switch args[0] {
	case "capture":
		recipe := flags.String("recipe", "", "JSON recipe with URL, tracks, timed actions and duration in milliseconds")
		chrome := flags.String("chrome-path", "", "Chromium executable; default searches the executable path")
		output := flags.String("output", "snapshots/motion", "parent directory for a new capture")
		video := flags.Bool("video", false, "also encode variable-frame-rate MP4 with ffmpeg")
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		if *recipe == "" || flags.NArg() != 0 {
			return errors.New("supply --recipe FILE")
		}
		data, err := os.ReadFile(*recipe)
		if err != nil {
			return err
		}
		var config motion.Recipe
		if err = json.Unmarshal(data, &config); err != nil {
			return err
		}
		dir, err := motion.Capture(context.Background(), config, *chrome, *output)
		if dir != "" {
			fmt.Println(dir + "/review.md")
		}
		if err != nil {
			return err
		}
		if *video {
			return motion.Encode(dir)
		}
		return nil
	case "compare":
		capture := flags.String("capture", "", "implementation capture directory")
		reference := flags.String("reference", "", "source capture or video-import directory")
		output := flags.String("output", "snapshots/motion", "common parent directory for the comparison")
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		if *capture == "" || *reference == "" || flags.NArg() != 0 {
			return errors.New("supply --capture DIR --reference DIR")
		}
		dir, err := motion.Compare(*capture, *reference, *output)
		if err == nil {
			fmt.Println(dir + "/review.md")
		}
		return err
	case "import":
		file := flags.String("input", "", "local source video")
		source := flags.String("source", "", "source section URL")
		mediaURL := flags.String("media-url", "", "original media URL, if known")
		output := flags.String("output", "snapshots/motion", "parent directory for a new import")
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		if *file == "" || flags.NArg() != 0 {
			return errors.New("supply --input FILE")
		}
		dir, err := motion.ImportVideo(context.Background(), *file, *source, *mediaURL, *output)
		if dir != "" {
			fmt.Println(dir + "/review.md")
		}
		return err
	case "track":
		capture := flags.String("capture", "", "capture directory with decoded frames")
		band := flags.String("band", "", "isolated edge band: x,y,width,height in recording pixels")
		axis := flags.String("axis", "x", "x for a vertical edge; y for a horizontal edge")
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		parts := strings.Split(*band, ",")
		if *capture == "" || len(parts) != 4 || flags.NArg() != 0 {
			return errors.New("supply --capture DIR --band x,y,width,height")
		}
		var box [4]int
		for i, p := range parts {
			v, e := strconv.Atoi(strings.TrimSpace(p))
			if e != nil {
				return e
			}
			box[i] = v
		}
		if box[0] < 0 || box[1] < 0 || box[2] < 3 || box[3] < 3 || box[2] > 8192 || box[3] > 8192 {
			return errors.New("invalid pixel band")
		}
		out, err := motion.TrackEdge(*capture, image.Rect(box[0], box[1], box[0]+box[2], box[1]+box[3]), *axis)
		if err == nil {
			fmt.Println(out)
		}
		return err
	case "analyze":
		file := flags.String("input", "", "capture.json or exported landmark trace")
		track := flags.String("track", "", "numeric channel name")
		start := flags.Float64("start", 0, "transition start in recorded milliseconds")
		end := flags.Float64("end", 0, "transition end in recorded milliseconds")
		from := flags.Float64("from", 0, "transition start value")
		to := flags.Float64("to", 0, "transition target value")
		bezier := flags.String("bezier", "", "optional x1,y1,x2,y2 reference; no curve is assumed")
		output := flags.String("output", "", "optional analysis JSON file; also adds its measurements to sibling review.md")
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		if *file == "" || *track == "" || flags.NArg() != 0 {
			return errors.New("supply --input FILE --track NAME and explicit transition endpoints")
		}
		var curve []float64
		if *bezier != "" {
			for _, part := range strings.Split(*bezier, ",") {
				v, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
				if err != nil {
					return err
				}
				curve = append(curve, v)
			}
		}
		result, err := motion.Analyze(*file, *track, *start, *end, *from, *to, curve)
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		if *output != "" {
			if err = os.WriteFile(*output, append(data, '\n'), 0644); err != nil {
				return err
			}
			review := filepath.Join(filepath.Dir(*output), "review.md")
			if f, err := os.OpenFile(review, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				_, writeErr := fmt.Fprintf(f, "\n## Measurement: %s\n\nWindow: %.2f–%.2f ms (%.2f ms). Values: %.4f to %.4f. Samples: %d. Maximum sample gap: %.2f ms.\n\nSigned progress beyond target: %.4f. Before start: %.4f.\n\nAnalysis: [%s](%s). No conformance result assigned.\n", result.Track, result.Start, result.End, result.Duration, result.From, result.To, result.Samples, result.MaxSampleGap, result.Overshoot, result.Undershoot, filepath.Base(*output), filepath.Base(*output))
				closeErr := f.Close()
				return errors.Join(writeErr, closeErr)
			}
		}
		return nil
	case "encode":
		capture := flags.String("capture", "", "capture directory")
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		if *capture == "" || flags.NArg() != 0 {
			return errors.New("supply --capture DIR")
		}
		return motion.Encode(*capture)
	default:
		return fmt.Errorf("unknown motion command %q", args[0])
	}
}
