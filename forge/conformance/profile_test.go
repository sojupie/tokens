package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"runtime"
	"strconv"
	"testing"
)

func TestCPUProfileUsesOnlyConfiguredProcesses(t *testing.T) {
	handler, err := cpuProfile("")
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest("GET", "/profile/cpu", nil))
	if response.Code != 501 {
		t.Fatal("unconfigured counters enabled")
	}
	if _, err := cpuProfile("../1"); err == nil {
		t.Fatal("invalid PID accepted")
	}
	if runtime.GOOS != "linux" {
		return
	}
	pid := strconv.Itoa(os.Getpid())
	handler, err = cpuProfile(pid)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	handler(response, httptest.NewRequest("GET", "/profile/cpu", nil))
	var counters map[string]float64
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &counters) != nil || len(counters) != 1 {
		t.Fatal(response)
	}
	if _, ok := counters[pid]; !ok {
		t.Fatal("missing selected process")
	}
}
