package publish

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteDownloadsIsDeterministicAndSubjectScoped(t *testing.T) {
	publication := Publication{
		PublicationID:   "publication-1",
		CurrentSnapshot: SnapshotView{ID: "snapshot-1"},
		Subjects: []SubjectView{
			{Key: SubjectKey{Kind: SubjectStyle, Slug: "spacing"}, Name: "Spacing"},
			{Key: SubjectKey{Kind: SubjectStyle, Slug: "motion/overview"}, Name: "Motion"},
			{Key: SubjectKey{Kind: SubjectComponent, Slug: "button"}, Name: "Button"},
		},
	}
	first := t.TempDir()
	second := t.TempDir()
	if err := writeDownloads(first, publication); err != nil {
		t.Fatal(err)
	}
	if err := writeDownloads(second, publication); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		filepath.Join("downloads", "tokens.json"),
		filepath.Join("downloads", "component", "button.json"),
		filepath.Join("downloads", "style", "spacing.json"),
		filepath.Join("downloads", "style", "motion", "overview.json"),
	}
	for _, relativePath := range paths {
		left, err := os.ReadFile(filepath.Join(first, relativePath))
		if err != nil {
			t.Fatal(err)
		}
		right, err := os.ReadFile(filepath.Join(second, relativePath))
		if err != nil {
			t.Fatal(err)
		}
		if string(left) != string(right) {
			t.Fatalf("download %s differs between builds", relativePath)
		}
	}
	allTokens, err := os.ReadFile(filepath.Join(first, "downloads", "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	buttonIndex := strings.Index(string(allTokens), `"slug": "button"`)
	spacingIndex := strings.Index(string(allTokens), `"slug": "spacing"`)
	if buttonIndex == -1 || spacingIndex == -1 || buttonIndex > spacingIndex {
		t.Fatalf("subjects are not deterministically ordered: %s", allTokens)
	}
}

func TestWriteDownloadsRejectsPathSeparators(t *testing.T) {
	directory := t.TempDir()
	err := writeDownloads(directory, Publication{
		CurrentSnapshot: SnapshotView{ID: "snapshot-1"},
		Subjects:        []SubjectView{{Key: SubjectKey{Kind: SubjectComponent, Slug: "../escape"}}},
	})
	if err == nil {
		t.Fatal("writeDownloads() error = nil, want invalid slug error")
	}
	if _, statErr := os.Stat(filepath.Join(directory, "downloads", "tokens.json")); !os.IsNotExist(statErr) {
		t.Fatalf("unexpected normalized download after validation error: %v", statErr)
	}
}
