package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCrawlReceiptAcceptsCurrentSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crawl-receipt.json")
	if err := os.WriteFile(path, []byte(`{"schemaVersion":2}`), 0o644); err != nil {
		t.Fatal(err)
	}

	receipt, err := readCrawlReceipt(path)
	if err != nil {
		t.Fatalf("readCrawlReceipt() error = %v", err)
	}
	if receipt.SchemaVersion != 2 {
		t.Fatalf("schema version = %d, want 2", receipt.SchemaVersion)
	}
}
