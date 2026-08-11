package tokensnapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sojupie/m3-source/internal/crawl"
)

func TestDiffReportsPayloadChanges(t *testing.T) {
	root := t.TempDir()
	catalogPath := filepath.Join(root, "catalog.json")
	catalog := []byte(`{"schemaVersion":1,"pages":[{"url":"https://example.test/specs","component":"example"}]}\n`)
	if err := os.WriteFile(catalogPath, catalog, 0o644); err != nil {
		t.Fatal(err)
	}

	writeRun := func(path, hash string, status int) {
		cacheDir := filepath.Join(root, filepath.Base(path)+"-cache")
		artifact := "source-payloads/fixture.json.gz"
		writeGzipArtifact(t, cacheDir, artifact, `{"system":{"tokens":[{"name":"`+hash+`","tokenName":"md.sys.color.primary"}]}}`)
		run := crawl.CrawlReceipt{
			SchemaVersion: 1,
			CrawlID:       filepath.Base(path),
			StartedAt:     "2026-08-11T00:00:00Z",
			Options:       crawl.CrawlOptions{PageCatalogPath: catalogPath, CacheDir: cacheDir},
			Pages:         []string{"https://example.test/specs"},
			Observations: []crawl.Observation{{
				PageURL:       "https://example.test/specs",
				Status:        status,
				TokenTableURL: "https://example.test/TOKEN_TABLE.json",
			}},
			SourcePayloads: []crawl.SourcePayloadRecord{{
				PageURL:       "https://example.test/specs",
				TokenTableURL: "https://example.test/TOKEN_TABLE.json",
				Hash:          hash,
				Artifact:      artifact,
				Bytes:         len(hash),
			}},
		}
		body, err := json.Marshal(run)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	fromRun := filepath.Join(root, "from-run.json")
	toRun := filepath.Join(root, "to-run.json")
	writeRun(fromRun, "hash-a", 200)
	writeRun(toRun, "hash-b", 304)
	fromDir := filepath.Join(root, "from")
	toDir := filepath.Join(root, "to")
	if _, err := Build(BuildConfig{CrawlReceiptPath: fromRun, PageCatalogPath: catalogPath, OutputDir: fromDir, SnapshotID: "m3-from"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(BuildConfig{CrawlReceiptPath: toRun, PageCatalogPath: catalogPath, OutputDir: toDir, SnapshotID: "m3-to"}); err != nil {
		t.Fatal(err)
	}

	diff, err := Diff(fromDir, toDir)
	if err != nil {
		t.Fatal(err)
	}
	if !diff.Changed || len(diff.SourcePayloads.Changed) != 1 || len(diff.PageResults.Changed) != 1 {
		t.Fatalf("unexpected diff: %#v", diff)
	}
	if diff.PageCatalogChanged || len(diff.Pages.Added) != 0 || len(diff.Pages.Removed) != 0 {
		t.Fatalf("unexpected structural diff: %#v", diff)
	}
}

func TestDiffPreservesDuplicateRecords(t *testing.T) {
	diff := diffRecords(nil, []ExtractedRecord{
		{SourcePayloadID: "payload-a", Kind: "token", ID: "same", Ordinal: 0, Raw: []byte(`{"value":1}`)},
		{SourcePayloadID: "payload-a", Kind: "token", ID: "same", Ordinal: 1, Raw: []byte(`{"value":2}`)},
	})
	if len(diff.Added) != 2 {
		t.Fatalf("duplicate records were collapsed: %#v", diff)
	}
}
