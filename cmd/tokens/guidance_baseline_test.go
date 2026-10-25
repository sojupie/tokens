package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuidanceBaselinePreservesIndexAndWorktree(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "--quiet")
	os.WriteFile("source.txt", []byte("base"), 0644)
	git("add", "source.txt")
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--quiet", "--no-gpg-sign", "-m", "Initial")
	os.WriteFile("source.txt", []byte("staged"), 0644)
	git("add", "source.txt")
	os.WriteFile("source.txt", []byte("current"), 0644)
	os.WriteFile("new.js", []byte("untracked source"), 0644)
	output := filepath.Join(t.TempDir(), "snapshots")
	if err := captureGuidanceBaseline([]string{"--output", output}); err != nil {
		t.Fatal(err)
	}
	if got := git("show", ":source.txt"); got != "staged" {
		t.Fatalf("real index changed: %s", got)
	}
	b, _ := os.ReadFile("source.txt")
	if string(b) != "current" {
		t.Fatal("worktree changed")
	}
	raw, err := os.ReadFile("forge/material/guidance/baselines.json")
	if err != nil {
		t.Fatal(err)
	}
	var r guidanceBaselines
	if err = json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Baselines) != 1 {
		t.Fatal("wrong baseline count")
	}
	for id, entry := range r.Baselines {
		if git("show", id+":source.txt") != "current" || git("show", id+":new.js") != "untracked source" {
			t.Fatal("baseline missed current files")
		}
		if _, err = os.Stat(entry.Archive); err != nil {
			t.Fatal(err)
		}
	}
}
