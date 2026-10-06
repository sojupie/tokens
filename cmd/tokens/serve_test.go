package main

import "testing"

func TestRunServeRejectsPositionalArguments(t *testing.T) {
	if err := runServe([]string{"publication"}); err == nil {
		t.Fatal("runServe() error = nil, want positional argument error")
	}
}
