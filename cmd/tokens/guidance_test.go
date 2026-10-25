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
	// Move a complete record, commit retirement, then prove archive loss fails.
	write(fixture)
	json.Unmarshal(fixture, &doc)
	check = doc["checks"].([]any)[0].(map[string]any)
	event := map[string]any{}
	for k, v := range check["sourceHistory"].([]any)[0].(map[string]any) {
		event[k] = v
	}
	event["change"] = "removed"
	event["summary"] = "Upstream removed the requirement."
	check["sourceHistory"] = append(check["sourceHistory"].([]any), event)
	archive := map[string]any{"schemaVersion": 1, "designSystem": "material", "component": "example", "checks": []any{check}}
	archived, _ := json.Marshal(archive)
	archivePath := filepath.Join(filepath.Dir(path), "retired-checks.json")
	os.WriteFile(archivePath, archived, 0644)
	doc["checks"] = []any{}
	doc["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)["status"] = "pending"
	data, _ = json.Marshal(doc)
	write(data)
	if err = runGuidance([]string{"validate"}); err != nil {
		t.Fatalf("retired: %v", err)
	}
	git("add", path, archivePath)
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--quiet", "--no-gpg-sign", "-m", "Retire check")
	os.Remove(archivePath)
	if err = runGuidance([]string{"validate"}); err == nil || !strings.Contains(err.Error(), "was deleted") {
		t.Fatalf("lost archive: %v", err)
	}
	os.WriteFile(archivePath, archived, 0644)
	os.Remove(path)
	if err = runGuidance([]string{"validate"}); err == nil || !strings.Contains(err.Error(), "archive requires") {
		t.Fatalf("orphan archive: %v", err)
	}

}

func TestGuidanceCatalogSelection(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	catalog := `{"schemaVersion":2,"pages":[{"url":"https://m3.material.io/components/buttons/specs","area":"components","subject":"buttons","pageType":"specs"},{"url":"https://m3.material.io/styles/color/tokens","area":"styles","subject":"color","pageType":"tokens"}]}`
	if err := os.WriteFile("pages.json", []byte(catalog), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--page-catalog", "pages.json", "--list"},
		{"--page-catalog", "pages.json", "--component", "buttons", "--list"},
		{"--page-catalog", "pages.json", "--component", "buttons", "--url", "https://m3.material.io/styles/motion/overview", "--list"},
		{"--url", "https://m3.material.io/components/buttons/guidelines", "--list"},
	} {
		if err := runGuidance(args); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	if err := runGuidance([]string{"--page-catalog", "pages.json", "--component", "missing", "--list"}); err == nil || !strings.Contains(err.Error(), "absent") {
		t.Fatalf("unknown component: %v", err)
	}
	if err := runGuidance([]string{"--page-catalog", "missing.json", "--list"}); err == nil {
		t.Fatal("accepted missing catalog")
	}
}
