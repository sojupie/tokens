package publish

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerationActivationPinsReads(t *testing.T) {
	root := t.TempDir()
	snapshots := filepath.Join(root, "snapshots")
	first := historyFixture(t, snapshots, "a", "2026-09-01T00:00:00Z", 40)
	second := historyFixture(t, snapshots, "b", "2026-09-02T00:00:00Z", 48)
	for _, item := range []struct{ snapshot, name string }{{first, "one"}, {second, "two"}} {
		if _, err := Build(BuildConfig{SnapshotDir: item.snapshot, OutputDir: filepath.Join(root, item.name)}); err != nil {
			t.Fatal(err)
		}
	}
	probe := filepath.Join(root, "probe")
	if err := os.Symlink(filepath.Join(root, "one"), probe); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	current := filepath.Join(root, "current")
	if err := Activate(current, filepath.Join(root, "one")); err != nil {
		t.Fatal(err)
	}
	pinned, err := ResolveGeneration(current)
	if err != nil {
		t.Fatal(err)
	}
	if err := Activate(current, filepath.Join(root, "two")); err != nil {
		t.Fatal(err)
	}
	state, err := Load(filepath.Join(pinned, "tokens.sqlite"))
	if err != nil || state.PublicationID != "a" {
		t.Fatalf("pinned generation: %v %v", state, err)
	}
	latest, err := ResolveGeneration(current)
	if err != nil || latest == pinned {
		t.Fatalf("activation did not switch: %s %v", latest, err)
	}
	if err := Activate(current, filepath.Join(root, "missing")); err == nil {
		t.Fatal("accepted invalid generation")
	}
	still, _ := ResolveGeneration(current)
	if still != latest {
		t.Fatal("failed activation changed current")
	}
	if err := Activate(filepath.Join(root, "one"), filepath.Join(root, "two")); err == nil {
		t.Fatal("replaced real directory")
	}
}

func TestGenerationDigestDetectsEdits(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "download.json")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := GenerationDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err := GenerationDigest(root)
	if err != nil || before == after {
		t.Fatalf("digest did not detect edit: %v", err)
	}
}
