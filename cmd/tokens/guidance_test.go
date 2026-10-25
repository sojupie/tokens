package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuidanceValidationCLI(t *testing.T) {
	fixture, err := os.ReadFile("../../source/guidance/testdata/curation-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	t.Chdir(root)
	git := func(args ...string) string {
		t.Helper()
		data, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, data)
		}
		return strings.TrimSpace(string(data))
	}
	git("init", "--quiet")
	path := "forge/material/guidance/example/checks.json"
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	write := func(data []byte) {
		t.Helper()
		if err = os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(fixture)
	git("add", path)
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--quiet", "--no-gpg-sign", "-m", "Initial guidance")
	if err = runGuidance([]string{"validate"}); err != nil {
		t.Fatal(err)
	}
	if err = runGuidance([]string{"validate", "--file", filepath.Join(root, path)}); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err = json.Unmarshal(fixture, &doc); err != nil {
		t.Fatal(err)
	}
	check := doc["checks"].([]any)[0].(map[string]any)
	check["fixes"] = []any{map[string]any{"commit": strings.Repeat("b", 40), "sourceRevision": 1, "summary": "Fix traversal."}}
	data, _ := json.Marshal(doc)
	write(data)
	if err = runGuidance([]string{"validate"}); err == nil || !strings.Contains(err.Error(), "commit does not exist") {
		t.Fatalf("unknown commit: %v", err)
	}
	check["fixes"].([]any)[0].(map[string]any)["commit"] = git("rev-parse", "HEAD")
	data, _ = json.Marshal(doc)
	write(data)
	if err = runGuidance([]string{"validate"}); err != nil {
		t.Fatalf("existing commit: %v", err)
	}
	write(fixture)
	if err = os.WriteFile("checks.json", fixture, 0644); err != nil {
		t.Fatal(err)
	}
	if err = runGuidance([]string{"validate", "--file", "checks.json"}); err == nil || !strings.Contains(err.Error(), "belong at") {
		t.Fatalf("noncanonical file: %v", err)
	}
	doc["checks"] = []any{}
	doc["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)["status"] = "pending"
	data, _ = json.Marshal(doc)
	write(data)
	if err = runGuidance([]string{"validate"}); err == nil || !strings.Contains(err.Error(), "was deleted") {
		t.Fatalf("deleted check: %v", err)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err = runGuidance([]string{"validate"}); err == nil || !strings.Contains(err.Error(), "guidance file was deleted") {
		t.Fatalf("deleted component file: %v", err)
	}
}
