package tokensnapshot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/crawl"
)

func TestBuildIsDeterministicAndVerifiable(t *testing.T) {
	root := t.TempDir()
	catalogPath := filepath.Join(root, "catalog.json")
	runPath := filepath.Join(root, "run.json")
	catalog := []byte(`{"schemaVersion":1,"source":"fixture","scope":"/components/<component>/specs","generatedAt":"2026-08-11T00:00:00Z","pages":[{"url":"https://m3.material.io/components/buttons/specs","component":"buttons"}]}` + "\n")
	if err := os.WriteFile(catalogPath, catalog, 0o644); err != nil {
		t.Fatal(err)
	}
	run := crawl.CrawlReceipt{
		SchemaVersion: 1,
		CrawlID:       "20260811T000000.000000000Z",
		StartedAt:     "2026-08-11T00:00:00Z",
		CompletedAt:   "2026-08-11T00:01:00Z",
		Options:       crawl.CrawlOptions{PageCatalogPath: catalogPath},
		Pages:         []string{"https://m3.material.io/components/buttons/specs"},
		Observations: []crawl.Observation{{
			PageURL:       "https://m3.material.io/components/buttons/specs",
			Status:        200,
			TokenTableURL: "https://m3.material.io/components/buttons/TOKEN_TABLE.json",
		}},
		SourcePayloads: []crawl.SourcePayloadRecord{{
			PageURL:       "https://m3.material.io/components/buttons/specs",
			TokenTableURL: "https://m3.material.io/components/buttons/TOKEN_TABLE.json",
			Hash:          "fixture-hash",
			Artifact:      "source-payloads/fi/fixture.json.gz",
		}},
	}
	runBody, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runPath, runBody, 0o644); err != nil {
		t.Fatal(err)
	}
	cacheDir := filepath.Join(root, "cache")
	writeGzipArtifact(t, cacheDir, "source-payloads/fi/fixture.json.gz", `{"system":{"tokens":[{"name":"fixture-token","tokenName":"md.sys.color.primary"}]}}`)
	run.Options.CacheDir = cacheDir
	runBody, err = json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runPath, runBody, 0o644); err != nil {
		t.Fatal(err)
	}
	config := BuildConfig{CrawlReceiptPath: runPath, PageCatalogPath: catalogPath, SnapshotID: "m3-fixture-1"}
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
	for _, name := range []string{"manifest.json", "page-catalog.json", "generated-token-graph.json.gz", "source-payload-details.json", "page-results.json", "checksums.txt"} {
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
	if _, err := os.Stat(filepath.Join(first, "source-payloads", "fi", "fixture.json.gz")); err != nil {
		t.Fatalf("raw payload artifact was not bundled: %v", err)
	}
	graph, err := readGraph(first)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Records) == 0 || strings.HasPrefix(graph.Records[0].SourcePayloadID, "payload-0001") {
		t.Fatalf("payload provenance ID is not stable: %#v", graph.Records)
	}
	coverageBody, err := os.ReadFile(filepath.Join(first, "page-results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pageResults PageResults
	if err := json.Unmarshal(coverageBody, &pageResults); err != nil {
		t.Fatal(err)
	}
	if len(pageResults.Pages) != 1 || pageResults.Pages[0].PayloadsObserved != 1 || pageResults.Pages[0].PayloadsFetched != 1 {
		t.Fatalf("unexpected page results: %#v", pageResults.Pages)
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
		run  crawl.CrawlReceipt
		want string
	}{
		{
			name: "error-only",
			run: crawl.CrawlReceipt{
				SchemaVersion: 1,
				StartedAt:     "2026-08-11T00:00:00Z",
				Observations:  []crawl.Observation{{PageURL: "https://example.test/specs", Status: 0, Error: "context canceled"}},
			},
			want: "no successful pages",
		},
		{
			name: "no-payload",
			run: crawl.CrawlReceipt{
				SchemaVersion: 1,
				StartedAt:     "2026-08-11T00:00:00Z",
				Observations:  []crawl.Observation{{PageURL: "https://example.test/specs", Status: 200}},
			},
			want: "no successful payloads",
		},
		{
			name: "partial",
			run: crawl.CrawlReceipt{
				SchemaVersion: 1,
				StartedAt:     "2026-08-11T00:00:00Z",
				Observations: []crawl.Observation{
					{PageURL: "https://example.test/ok/specs", Status: 200},
					{PageURL: "https://example.test/failed/specs", Status: 0, Error: "timeout"},
				},
				SourcePayloads: []crawl.SourcePayloadRecord{{Hash: "fixture-hash", TokenTableURL: "https://example.test/TOKEN_TABLE.json"}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "partial" {
				test.run.Options.CacheDir = filepath.Join(root, "partial-cache")
				test.run.SourcePayloads[0].Artifact = "source-payloads/partial.json.gz"
				writeGzipArtifact(t, test.run.Options.CacheDir, test.run.SourcePayloads[0].Artifact, `{"system":{"tokens":[]}}`)
			}
			runPath := filepath.Join(root, test.name+".json")
			body, err := json.Marshal(test.run)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(runPath, body, 0o644); err != nil {
				t.Fatal(err)
			}
			_, err = Build(BuildConfig{
				CrawlReceiptPath: runPath,
				PageCatalogPath:  catalogPath,
				OutputDir:        filepath.Join(root, test.name+"-snapshot"),
				SnapshotID:       "m3-incomplete-" + test.name,
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

func writeGzipArtifact(t *testing.T, cacheDir, artifact, body string) {
	t.Helper()
	path := filepath.Join(cacheDir, filepath.FromSlash(artifact))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, compressed.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
