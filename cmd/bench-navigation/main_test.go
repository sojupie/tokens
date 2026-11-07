package main

import "testing"

func TestPercentileUsesNearestRankOnUnorderedSamples(t *testing.T) {
	values := []float64{9, 1, 5, 3, 7}
	if percentile(values, .5) != 5 || percentile(values, .95) != 9 {
		t.Fatal(values)
	}
	if percentile([]float64{4}, .95) != 4 {
		t.Fatal("one-sample percentile")
	}
}
