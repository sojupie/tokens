package publish

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/crawl"
	"github.com/sojupie/tokens/source/tokensnapshot"
)

func TestBuildRetainsTimelineAcrossPublishAndLoad(t *testing.T) {
	root := t.TempDir()
	historyFixture(t, root, "a", "2026-06-01T12:00:00Z", 40)
	historyFixture(t, root, "b", "2026-06-02T12:00:00Z", 48)
	current := historyFixture(t, root, "c", "2026-06-03T12:00:00Z", 56)
	historyFixture(t, root, "future", "2026-06-04T12:00:00Z", 64)
	old := filepath.Join(root, "old")
	if err := os.Mkdir(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "manifest.json"), []byte(`{"schemaVersion":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var messages []string
	config := BuildConfig{SnapshotDir: current, OutputDir: filepath.Join(t.TempDir(), "published"), Progress: func(s string) { messages = append(messages, s) }}
	publication, err := Build(config)
	if err != nil {
		t.Fatal(err)
	}
	if len(publication.Snapshots) != 3 || len(publication.Changes) != 2 {
		t.Fatalf("history = %d snapshots, %d changes", len(publication.Snapshots), len(publication.Changes))
	}
	if publication.Changes[0].SnapshotID != "c" || publication.Changes[1].SnapshotID != "b" {
		t.Fatalf("not newest first: %#v", publication.Changes)
	}
	var before tokensnapshot.ContextualValue
	if err := json.Unmarshal([]byte(publication.Changes[1].Before), &before); err != nil {
		t.Fatal(err)
	}
	if string(before.Value) != `{"length":{"unit":"DIPS","value":40}}` {
		t.Fatalf("earliest transition lost: %#v", publication.Changes[1])
	}
	if !strings.Contains(strings.Join(messages, "\n"), "unsupported schema 2") {
		t.Fatal("excluded history not reported")
	}
	if count := strings.Count(strings.Join(messages, "\n"), "Verifying and indexing history snapshot:"); count != 2 {
		t.Fatalf("historical loads = %d, want one per retained historical snapshot", count)
	}
	// The publication's compact path must produce the same timeline as the
	// complete CLI diagnostic comparison, including stable change IDs.
	for i, pair := range [][2]string{{"b", "c"}, {"a", "b"}} {
		diff, err := tokensnapshot.Diff(filepath.Join(root, pair[0]), filepath.Join(root, pair[1]))
		if err != nil {
			t.Fatal(err)
		}
		expected := changeViews(diff)
		if len(expected) != 1 {
			t.Fatalf("diagnostic comparison returned %d changes, want one", len(expected))
		}
		expected[0].DetectedAt = publication.Changes[i].DetectedAt
		if !reflect.DeepEqual(expected[0], publication.Changes[i]) {
			t.Fatalf("compact comparison differs from diagnostic comparison: %#v != %#v", publication.Changes[i], expected[0])
		}
	}
	again, err := Build(config)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again.Changes, publication.Changes) {
		t.Fatal("republish changed or duplicated history")
	}
	state, err := Load(filepath.Join(config.OutputDir, "tokens.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.Changes, publication.Changes) {
		t.Fatalf("history did not survive SQLite load: %#v", state.Changes)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "page-catalog.json"), []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(config); err == nil {
		t.Fatal("corrupt retained snapshot was silently skipped")
	}
	state, err = Load(filepath.Join(config.OutputDir, "tokens.sqlite"))
	if err != nil || len(state.Changes) != 2 {
		t.Fatal("failed publication destroyed previous history")
	}
}

func historyFixture(t *testing.T, root, id, date string, width int) string {
	t.Helper()
	inputs := t.TempDir()
	payload := []byte(fmt.Sprintf(`{"system":{"contextualReferenceTrees":{"width":{"contextualReferenceTree":[{"referenceTree":{"tokenName":"md.comp.snackbar.container.width"},"resolvedValue":{"length":{"unit":"DIPS","value":%d}}}]}}}}`, width))
	file, err := os.Create(filepath.Join(inputs, "payload.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	writer := gzip.NewWriter(file)
	if _, err := writer.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	page := "https://m3.material.io/components/snackbar/specs"
	payloadURL := "https://m3.material.io/_dsm/data/dsdb-m3/TOKEN_TABLE.fixture.json"
	catalog := filepath.Join(inputs, "catalog.json")
	if err := os.WriteFile(catalog, []byte(`{"schemaVersion":2,"source":"fixture","scope":"component specs","generatedAt":"2026-06-01T12:00:00Z","pages":[{"url":"https://m3.material.io/components/snackbar/specs","area":"components","subject":"snackbar","pageType":"specs"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	receipt := crawl.CrawlReceipt{SchemaVersion: 2, CrawlID: id, StartedAt: date, CompletedAt: date,
		Options: crawl.CrawlOptions{PageCatalogPath: catalog, CacheDir: inputs}, Pages: []string{page},
		Observations:   []crawl.Observation{{PageURL: page, Status: 200, SourcePayloadURLs: []string{payloadURL}}},
		SourcePayloads: []crawl.SourcePayloadRecord{{PageURLs: []string{page}, SourcePayloadURL: payloadURL, Hash: sha256Hex(payload), Bytes: len(payload), Artifact: "payload.json.gz"}},
	}
	body, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	run := filepath.Join(inputs, "receipt.json")
	if err := os.WriteFile(run, body, 0o644); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, id)
	if _, err := tokensnapshot.Build(tokensnapshot.BuildConfig{CrawlReceiptPath: run, SnapshotID: id, OutputDir: directory}); err != nil {
		t.Fatal(err)
	}
	return directory
}
