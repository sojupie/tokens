package publish

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceOutputReplacesOnlyAfterStagingCompletes(t *testing.T) {
	parent := t.TempDir()
	outputPath := filepath.Join(parent, "published")
	stagingPath := filepath.Join(parent, "staging")
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stagingPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputPath, "publication.json"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stagingPath, "publication.json"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceOutput(stagingPath, outputPath); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(outputPath, "publication.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "new" {
		t.Fatalf("publication = %q, want new", body)
	}
}
