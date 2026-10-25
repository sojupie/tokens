package motion

import (
	"encoding/json"
	"errors"
	"math"
	"os"
)

type Sample struct {
	T      float64             `json:"t"`
	Values map[string]*float64 `json:"values"`
}
type Analysis struct {
	Track            string    `json:"track"`
	Start            float64   `json:"start"`
	End              float64   `json:"end"`
	Duration         float64   `json:"duration"`
	From             float64   `json:"from"`
	To               float64   `json:"to"`
	Samples          int       `json:"samples"`
	Overshoot        float64   `json:"overshoot"`
	Undershoot       float64   `json:"undershoot"`
	MaxSampleGap     float64   `json:"maxSampleGap"`
	Easing           []float64 `json:"easing,omitempty"`
	RMSProgressError *float64  `json:"rmsProgressError,omitempty"`
	MaxProgressError *float64  `json:"maxProgressError,omitempty"`
}

func Progress(x float64, c []float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	lo, hi := 0.0, 1.0
	for range 40 {
		t := (lo + hi) / 2
		v := 3*(1-t)*(1-t)*t*c[0] + 3*(1-t)*t*t*c[2] + t*t*t
		if v < x {
			lo = t
		} else {
			hi = t
		}
	}
	t := (lo + hi) / 2
	return 3*(1-t)*(1-t)*t*c[1] + 3*(1-t)*t*t*c[3] + t*t*t
}

func Analyze(file, track string, start, end, from, to float64, easing []float64) (Analysis, error) {
	result := Analysis{Track: track, Start: start, End: end, Duration: end - start, From: from, To: to, Easing: easing}
	for _, v := range []float64{start, end, from, to} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return result, errors.New("endpoints must be finite")
		}
	}
	if end <= start || from == to {
		return result, errors.New("choose increasing time bounds and distinct values")
	}
	if len(easing) != 0 {
		if len(easing) != 4 {
			return result, errors.New("cubic-bezier needs four numbers")
		}
		for _, v := range easing {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return result, errors.New("curve values must be finite")
			}
		}
		if easing[0] < 0 || easing[0] > 1 || easing[2] < 0 || easing[2] > 1 {
			return result, errors.New("cubic-bezier x values must be between 0 and 1")
		}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return result, err
	}
	var envelope struct {
		Trace json.RawMessage `json:"trace"`
	}
	if err = json.Unmarshal(data, &envelope); err != nil {
		return result, err
	}
	if len(envelope.Trace) > 0 && string(envelope.Trace) != "null" {
		data = envelope.Trace
	}
	var trace struct {
		Frames []Sample `json:"frames"`
	}
	if err = json.Unmarshal(data, &trace); err != nil {
		return result, err
	}
	previous := -math.MaxFloat64
	last := start
	sum, maxError := 0.0, 0.0
	covered := false
	for _, f := range trace.Frames {
		if f.T < previous {
			return result, errors.New("frame timestamps are not ordered")
		}
		previous = f.T
		if f.T > end {
			covered = true
		}
		if f.T < start || f.T > end {
			continue
		}
		value := f.Values[track]
		if value == nil {
			continue
		}
		p := (*value - from) / (to - from)
		result.Overshoot = math.Max(result.Overshoot, p-1)
		result.Undershoot = math.Max(result.Undershoot, -p)
		result.MaxSampleGap = math.Max(result.MaxSampleGap, f.T-last)
		last = f.T
		result.Samples++
		if len(easing) > 0 {
			delta := p - Progress((f.T-start)/(end-start), easing)
			sum += delta * delta
			maxError = math.Max(maxError, math.Abs(delta))
		}
	}
	result.MaxSampleGap = math.Max(result.MaxSampleGap, end-last)
	if result.Samples < 2 {
		return result, errors.New("need at least two finite samples in the selected window")
	}
	if len(trace.Frames) == 0 || trace.Frames[0].T > start || (!covered && trace.Frames[len(trace.Frames)-1].T < end) {
		return result, errors.New("selected time bounds extend beyond sampled data")
	}
	if len(easing) > 0 {
		rms := math.Sqrt(sum / float64(result.Samples))
		result.RMSProgressError = &rms
		result.MaxProgressError = &maxError
	}
	return result, nil
}
