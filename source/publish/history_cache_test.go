package publish

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestHistoryCacheReusesAndRepairs(t *testing.T) {
	root := t.TempDir()
	historyFixture(t, root, "a", "2026-09-01T00:00:00Z", 40)
	current := historyFixture(t, root, "b", "2026-09-02T00:00:00Z", 48)
	var messages []string
	config := BuildConfig{SnapshotDir: current, OutputDir: filepath.Join(t.TempDir(), "published"), HistoryCacheDir: filepath.Join(t.TempDir(), "cache"), Progress: func(s string) { messages = append(messages, s) }}
	first, err := Build(config)
	if err != nil {
		t.Fatal(err)
	}
	messages = nil
	second, err := Build(config)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Changes, second.Changes) {
		t.Fatal("warm cache changed timeline")
	}
	joined := strings.Join(messages, "\n")
	if !strings.Contains(joined, "Reusing history index: a") || !strings.Contains(joined, "Reusing history comparison: a -> b") {
		t.Fatal(joined)
	}
	entries, err := os.ReadDir(config.HistoryCacheDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if err := os.WriteFile(filepath.Join(config.HistoryCacheDir, entry.Name()), []byte("broken"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	third, err := Build(config)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Changes, third.Changes) {
		t.Fatal("cache repair changed timeline")
	}
	if err := os.WriteFile(filepath.Join(root, "a", "page-catalog.json"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(config); err == nil {
		t.Fatal("warm cache accepted damaged snapshot")
	}
}
