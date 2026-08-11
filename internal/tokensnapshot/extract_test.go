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
