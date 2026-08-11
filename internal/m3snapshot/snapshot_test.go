package m3snapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/m3crawl"
)

func TestBuildIsDeterministicAndVerifiable(t *testing.T) {
	root := t.TempDir()
	catalogPath := filepath.Join(root, "catalog.json")
	runPath := filepath.Join(root, "run.json")
	catalog := []byte(`{"schemaVersion":1,"source":"fixture","scope":"/components/<component>/specs","generatedAt":"2026-08-11T00:00:00Z","pages":[{"url":"https://m3.material.io/components/buttons/specs","component":"buttons"}]}` + "\n")
	if err := os.WriteFile(catalogPath, catalog, 0o644); err != nil {
		t.Fatal(err)
	}
	run := m3crawl.RunManifest{
		SchemaVersion: 1,
		RunID:         "20260811T000000.000000000Z",
		StartedAt:     "2026-08-11T00:00:00Z",
		CompletedAt:   "2026-08-11T00:01:00Z",
		Options:       m3crawl.ManifestOptions{CatalogPath: catalogPath},
		Pages:         []string{"https://m3.material.io/components/buttons/specs"},
		Observations: []m3crawl.Observation{{
			PageURL:       "https://m3.material.io/components/buttons/specs",
			Status:        200,
			TokenTableURL: "https://m3.material.io/components/buttons/TOKEN_TABLE.json",
		}},
		Payloads: []m3crawl.PayloadRecord{{
			PageURL:       "https://m3.material.io/components/buttons/specs",
			TokenTableURL: "https://m3.material.io/components/buttons/TOKEN_TABLE.json",
			Hash:          "fixture-hash",
			Artifact:      "payloads/fi/fixture.json.gz",
		}},
	}
	runBody, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runPath, runBody, 0o644); err != nil {
		t.Fatal(err)
	}
	config := BuildConfig{RunManifestPath: runPath, CatalogPath: catalogPath, SnapshotID: "m3-fixture-1"}
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	config.OutputDir = first
	if _, err := Build(config); err != nil {
		t.Fatal(err)
	}
	config.OutputDir = second
	if _, err := Build(config); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"manifest.json", "catalog.json", "material.graph.json.gz", "provenance.json", "coverage.json", "checksums.txt"} {
		left, err := os.ReadFile(filepath.Join(first, name))
		if err != nil {
			t.Fatal(err)
		}
		right, err := os.ReadFile(filepath.Join(second, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(left) != string(right) {
			t.Fatalf("%s differs between deterministic builds", name)
		}
	}
	if _, err := Verify(first); err != nil {
		t.Fatal(err)
	}
	coverageBody, err := os.ReadFile(filepath.Join(first, "coverage.json"))
	if err != nil {
		t.Fatal(err)
	}
	var coverage Coverage
	if err := json.Unmarshal(coverageBody, &coverage); err != nil {
		t.Fatal(err)
	}
	if len(coverage.Pages) != 1 || coverage.Pages[0].PayloadsObserved != 1 || coverage.Pages[0].PayloadsFetched != 1 {
		t.Fatalf("unexpected coverage: %#v", coverage.Pages)
	}
}

func TestBuildRejectsIncompleteCapture(t *testing.T) {
	root := t.TempDir()
	catalogPath := filepath.Join(root, "catalog.json")
	if err := os.WriteFile(catalogPath, []byte(`{"schemaVersion":1,"pages":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		run  m3crawl.RunManifest
		want string
	}{
		{
			name: "error-only",
			run: m3crawl.RunManifest{
				SchemaVersion: 1,
				StartedAt:     "2026-08-11T00:00:00Z",
				Observations:  []m3crawl.Observation{{PageURL: "https://example.test/specs", Status: 0, Error: "context canceled"}},
			},
			want: "no successful pages",
		},
		{
			name: "no-payload",
			run: m3crawl.RunManifest{
				SchemaVersion: 1,
				StartedAt:     "2026-08-11T00:00:00Z",
				Observations:  []m3crawl.Observation{{PageURL: "https://example.test/specs", Status: 200}},
			},
			want: "no successful payloads",
		},
		{
			name: "partial",
			run: m3crawl.RunManifest{
				SchemaVersion: 1,
				StartedAt:     "2026-08-11T00:00:00Z",
				Observations: []m3crawl.Observation{
					{PageURL: "https://example.test/ok/specs", Status: 200},
					{PageURL: "https://example.test/failed/specs", Status: 0, Error: "timeout"},
				},
				Payloads: []m3crawl.PayloadRecord{{Hash: "fixture-hash", TokenTableURL: "https://example.test/TOKEN_TABLE.json"}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runPath := filepath.Join(root, test.name+".json")
			body, err := json.Marshal(test.run)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(runPath, body, 0o644); err != nil {
				t.Fatal(err)
			}
			_, err = Build(BuildConfig{
				RunManifestPath: runPath,
				CatalogPath:     catalogPath,
				OutputDir:       filepath.Join(root, test.name+"-snapshot"),
				SnapshotID:      "m3-incomplete-" + test.name,
			})
			if test.want == "" {
				if err != nil {
					t.Fatalf("Build error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Build error = %v, want %q", err, test.want)
			}
		})
	}
}
