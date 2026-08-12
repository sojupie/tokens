package tokensnapshot

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPayloadPreservesGenericRecords(t *testing.T) {
	cacheDir := t.TempDir()
	artifact := "source-payloads/aa/example.json.gz"
	path := filepath.Join(cacheDir, filepath.FromSlash(artifact))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(`{"system":{"custom":{"futureField":true},"tokens":[{"name":"token-id","tokenName":"md.sys.color.primary","tokenValueType":"COLOR"}],"values":[{"name":"value-id","tokenName":"md.sys.color.primary","undefined":true}],"contextualReferenceTrees":{"token-id":{"resolvedValue":{"undefined":true}}}}}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, compressed.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := extractPayload(cacheDir, m3Payload{Artifact: artifact, Hash: "payload-hash"}, "payload-0001")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 4 {
		t.Fatalf("got %d records, want 4: %#v", len(records), records)
	}
	var token ExtractedRecord
	var system ExtractedRecord
	for _, record := range records {
		if record.Kind == "token" {
			token = record
		}
		if record.Kind == "system" {
			system = record
		}
	}
	if token.TokenName != "md.sys.color.primary" || token.SourcePayloadID != "payload-0001" {
		t.Fatalf("unexpected token record: %#v", token)
	}
	if !bytes.Contains(system.Raw, []byte(`"custom"`)) {
		t.Fatalf("system record lost unknown fields: %#v", system)
	}
}

func TestLifecycleNoticesPreserveScopeAndEvidence(t *testing.T) {
	records := []ExtractedRecord{
		{
			ID:              "component-1",
			Kind:            "component",
			SourcePayloadID: "payload-1",
			Raw:             []byte(`{"name":"segmented-buttons","description":"No longer recommended; use connected button group."}`),
		},
		{
			ID:              "token-1",
			Kind:            "token",
			SourcePayloadID: "payload-1",
			Raw:             []byte(`{"tokenName":"md.comp.fab.enabled","deprecationMessage":{"message":"Token is deprecated."}}`),
		},
	}
	notices := lifecycleNotices(records)
	if len(notices) != 2 {
		t.Fatalf("got %d lifecycle notices, want 2: %#v", len(notices), notices)
	}
	for _, notice := range notices {
		if notice.ID == "" || notice.Message == "" || notice.SourcePayloadID != "payload-1" || notice.SourceRecordID == "" {
			t.Fatalf("notice lost evidence: %#v", notice)
		}
		if notice.Scope == "component" && notice.State != "not-recommended" {
			t.Fatalf("unexpected component notice: %#v", notice)
		}
		if notice.Scope == "token" && notice.State != "deprecated" {
			t.Fatalf("unexpected token notice: %#v", notice)
		}
	}
}
