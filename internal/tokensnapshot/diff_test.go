package tokensnapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sojupie/tokens/internal/crawl"
)

func TestDiffReportsAddedAndRemovedPageResults(t *testing.T) {
	diff := diffPageResults(
		[]PageResult{{URL: "https://example.test/removed", Status: 200}},
		[]PageResult{{URL: "https://example.test/added", Status: 200}},
	)
	if len(diff.Added) != 1 || diff.Added[0].URL != "https://example.test/added" {
		t.Fatalf("added page result missing: %#v", diff)
	}
	if len(diff.Removed) != 1 || diff.Removed[0].URL != "https://example.test/removed" {
		t.Fatalf("removed page result missing: %#v", diff)
	}
}

func TestDiffReportsPayloadChanges(t *testing.T) {
	root := t.TempDir()
	catalogPath := filepath.Join(root, "catalog.json")
	catalog := []byte(`{"schemaVersion":2,"source":"fixture","scope":"component specs","generatedAt":"2026-08-11T00:00:00Z","pages":[{"url":"https://m3.material.io/components/example/specs","area":"components","subject":"example","pageType":"specs"}]}` + "\n")
	if err := os.WriteFile(catalogPath, catalog, 0o644); err != nil {
		t.Fatal(err)
	}

	writeRun := func(path, hash string, status int) {
		cacheDir := filepath.Join(root, filepath.Base(path)+"-cache")
		artifact := "source-payloads/fixture.json.gz"
		payloadBody := `{"system":{"tokens":[{"name":"` + hash + `","tokenName":"md.sys.color.primary"}]}}`
		writeGzipArtifact(t, cacheDir, artifact, payloadBody)
		run := crawl.CrawlReceipt{
			SchemaVersion: 2,
			CrawlID:       filepath.Base(path),
			StartedAt:     "2026-08-11T00:00:00Z",
			Options:       crawl.CrawlOptions{PageCatalogPath: catalogPath, CacheDir: cacheDir},
			Pages:         []string{"https://example.test/specs"},
			Observations: []crawl.Observation{{
				PageURL:           "https://m3.material.io/components/example/specs",
				Status:            status,
				SourcePayloadURLs: []string{"https://example.test/TOKEN_TABLE.json"},
			}},
			SourcePayloads: []crawl.SourcePayloadRecord{{
				PageURLs:         []string{"https://m3.material.io/components/example/specs"},
				SourcePayloadURL: "https://example.test/TOKEN_TABLE.json",
				Hash:             sha256Hex([]byte(payloadBody)),
				Artifact:         artifact,
				Bytes:            len(payloadBody),
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

func TestDiffMatchesPayloadsAcrossReleaseDirectories(t *testing.T) {
	fromURL := "https://m3.material.io/_dsm/data/dsdb-m3/2026-08-05_09-00-19/TOKEN_TABLE.01c5310367774eb8.json"
	toURL := "https://m3.material.io/_dsm/data/dsdb-m3/2026-09-02_06-10-10/TOKEN_TABLE.01c5310367774eb8.json"
	diff := diffSourcePayloads(
		[]GraphPayload{{SourcePayloadURL: fromURL, Hash: "old"}},
		[]GraphPayload{{SourcePayloadURL: toURL, Hash: "new"}},
	)
	if len(diff.Added) != 0 || len(diff.Removed) != 0 || len(diff.Changed) != 1 {
		t.Fatalf("release-directory rollover was not treated as a payload change: %#v", diff)
	}
	change := diff.Changed[0]
	if change.Key != "m3.material.io/TOKEN_TABLE.01c5310367774eb8.json" {
		t.Fatalf("unexpected stable payload key: %#v", change)
	}
	if change.BeforeSourcePayloadURL != fromURL || change.AfterSourcePayloadURL != toURL {
		t.Fatalf("payload URL history was not preserved: %#v", change)
	}
}

func TestSourcePayloadIDMatchesAcrossReleaseDirectories(t *testing.T) {
	from := crawl.SourcePayloadRecord{
		SourcePayloadURL: "https://m3.material.io/_dsm/data/dsdb-m3/2026-08-05_09-00-19/ELEVATION.20543ce18892f7d9.json",
		Hash:             "old",
	}
	to := crawl.SourcePayloadRecord{
		SourcePayloadURL: "https://m3.material.io/_dsm/data/dsdb-m3/2026-09-02_06-10-10/ELEVATION.20543ce18892f7d9.json",
		Hash:             "new",
	}
	if sourcePayloadID(from) != sourcePayloadID(to) {
		t.Fatalf("payload ID changed across release directories: %q != %q", sourcePayloadID(from), sourcePayloadID(to))
	}
}

func TestNormalizeHistoricalRecordPayloadIDs(t *testing.T) {
	records := []ExtractedRecord{{SourcePayloadID: "payload-old"}}
	normalized := normalizeRecordPayloadIDs(records, map[string]string{"payload-old": "m3.material.io/TOKEN_TABLE.json"})
	if normalized[0].SourcePayloadID != "m3.material.io/TOKEN_TABLE.json" {
		t.Fatalf("historical payload ID was not normalized: %#v", normalized)
	}
	if records[0].SourcePayloadID != "payload-old" {
		t.Fatalf("normalization mutated input records: %#v", records)
	}
}

func TestNormalizeHistoricalLifecycleNoticePayloadIDs(t *testing.T) {
	notices := []LifecycleNotice{{
		ID:              "notice-old",
		SourcePayloadID: "payload-old",
		SourceRecordID:  "record-1",
		RecordKind:      "token",
		SourceField:     "description",
		Message:         "Deprecated token",
	}}
	normalized := normalizeLifecycleNoticePayloadIDs(notices, map[string]string{"payload-old": "m3.material.io/TOKEN_TABLE.json"})
	if normalized[0].SourcePayloadID != "m3.material.io/TOKEN_TABLE.json" {
		t.Fatalf("historical notice payload ID was not normalized: %#v", normalized)
	}
	if normalized[0].ID != lifecycleNoticeID("m3.material.io/TOKEN_TABLE.json", "token", "record-1", "description", "Deprecated token") {
		t.Fatalf("historical notice ID was not regenerated: %#v", normalized)
	}
	if notices[0].ID != "notice-old" {
		t.Fatalf("normalization mutated input notices: %#v", notices)
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

func TestDiffIgnoresRecordReordering(t *testing.T) {
	from := []ExtractedRecord{{SourcePayloadID: "payload-a", Kind: "token", ID: "same", Ordinal: 0, Raw: []byte(`{"value":1}`)}}
	to := []ExtractedRecord{{SourcePayloadID: "payload-a", Kind: "token", ID: "same", Ordinal: 37, Raw: []byte(`{"value":1}`)}}
	diff := diffRecords(from, to)
	if len(diff.Added) != 0 || len(diff.Removed) != 0 || len(diff.Changed) != 0 {
		t.Fatalf("record reordering was reported as a change: %#v", diff)
	}
}

func TestSemanticDiffIgnoresContextualVersionIDs(t *testing.T) {
	oldRecord := ExtractedRecord{
		SourcePayloadKind: "elevation",
		SourceSubject:     "elevation",
		Kind:              "contextualReferenceTree",
		ID:                "token-1",
		Raw:               []byte(`{"contextualReferenceTree":[{"referenceTree":{"value":{"version":"old"},"tokenName":"md.sys.elevation.level4"},"resolvedValue":{"elevation":{"value":8,"unit":"DIPS"}}}]}`),
	}
	newRecord := oldRecord
	newRecord.Raw = []byte(`{"contextualReferenceTree":[{"referenceTree":{"value":{"version":"new"},"tokenName":"md.sys.elevation.level4"},"resolvedValue":{"elevation":{"value":8,"unit":"DIPS"}}}]}`)
	diff := diffSemanticRecords([]ExtractedRecord{oldRecord}, []ExtractedRecord{newRecord})
	if len(diff.Added) != 0 || len(diff.Removed) != 0 || len(diff.Changed) != 0 {
		t.Fatalf("version-only change was treated as semantic: %#v", diff)
	}
}

func TestSemanticDiffReportsResolvedValueChange(t *testing.T) {
	oldRecord := ExtractedRecord{
		SourcePayloadKind: "elevation",
		SourceSubject:     "elevation",
		Kind:              "contextualReferenceTree",
		ID:                "token-1",
		Raw:               []byte(`{"contextualReferenceTree":[{"referenceTree":{"tokenName":"md.sys.elevation.level4"},"resolvedValue":{"elevation":{"value":8,"unit":"DIPS"}}}]}`),
	}
	newRecord := oldRecord
	newRecord.Raw = []byte(`{"contextualReferenceTree":[{"referenceTree":{"tokenName":"md.sys.elevation.level4"},"resolvedValue":{"elevation":{"value":12,"unit":"DIPS"}}}]}`)
	diff := diffSemanticRecords([]ExtractedRecord{oldRecord}, []ExtractedRecord{newRecord})
	if len(diff.Changed) != 1 || diff.Changed[0].Name != "md.sys.elevation.level4" {
		t.Fatalf("resolved value change was not reported: %#v", diff)
	}
}

func TestSemanticDiffIgnoresRepeatedComponentsInStylePayloads(t *testing.T) {
	record := ExtractedRecord{
		SourcePayloadKind: "elevation",
		SourceSubject:     "elevation",
		Kind:              "component",
		DisplayName:       "Date pickers",
		Raw:               []byte(`{"displayName":"Date pickers","state":"ACTIVE"}`),
	}
	diff := diffSemanticRecords(nil, []ExtractedRecord{record})
	if len(diff.Added) != 0 {
		t.Fatalf("shared component copy was reported as a component change: %#v", diff)
	}
}
