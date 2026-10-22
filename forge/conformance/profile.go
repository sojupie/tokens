package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Optional Linux process counters for local renderer comparisons. The caller
// selects exact PIDs; this endpoint is never part of the public Catalog server.
func cpuProfile(pids string) (http.HandlerFunc, error) {
	ids := strings.FieldsFunc(pids, func(r rune) bool { return r == ',' })
	if len(ids) == 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "CPU counters not configured", http.StatusNotImplemented)
		}, nil
	}
	for _, id := range ids {
		if pid, err := strconv.Atoi(id); err != nil || pid <= 0 {
			return nil, fmt.Errorf("invalid profile PID")
		}
	}
	output, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return nil, err
	}
	ticks, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil || ticks <= 0 {
		return nil, fmt.Errorf("invalid clock tick frequency")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		counters := map[string]float64{}
		for _, id := range ids {
			data, err := os.ReadFile("/proc/" + id + "/stat")
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			fields := strings.Fields(string(data)[strings.LastIndex(string(data), ")")+1:])
			if len(fields) < 13 {
				http.Error(w, "short process stat", http.StatusServiceUnavailable)
				return
			}
			user, e1 := strconv.ParseFloat(fields[11], 64)
			system, e2 := strconv.ParseFloat(fields[12], 64)
			if e1 != nil || e2 != nil {
				http.Error(w, "invalid process counters", http.StatusServiceUnavailable)
				return
			}
			counters[id] = (user + system) * 1000 / ticks
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(counters)
	}, nil
}
