package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestAcceptanceRequiresChecksAndNoFailures(t *testing.T) {
	for _, result := range []string{`{}`, `{"checked":0,"failures":[]}`, `{"passed":1,"error":"broken"}`, `{"checked":1,"failureCount":81}`, `{"checked":1,"failures":["clipped"]}`} {
		if accept([]byte(result)) == nil {
			t.Errorf("accepted %s", result)
		}
	}
	for _, result := range []string{`{"passed":1}`, `{"propertiesChecked":5,"failures":[]}`, `{"checked":1,"failureCount":0}`} {
		if err := accept([]byte(result)); err != nil {
			t.Errorf("rejected %s: %v", result, err)
		}
	}
}

func TestUnpackRejectsTraversalAndLinks(t *testing.T) {
	for _, entry := range []struct {
		name string
		kind byte
	}{{"package/../../outside", tar.TypeReg}, {"package/..\\..\\outside", tar.TypeReg}, {"/absolute", tar.TypeReg}, {"package/link", tar.TypeSymlink}, {"package/link", tar.TypeLink}} {
		t.Run(entry.name, func(t *testing.T) {
			dir := t.TempDir()
			file, err := os.Create(filepath.Join(dir, "input.tgz"))
			if err != nil {
				t.Fatal(err)
			}
			zip := gzip.NewWriter(file)
			writer := tar.NewWriter(zip)
			if err = writer.WriteHeader(&tar.Header{Name: entry.name, Typeflag: entry.kind, Linkname: "outside", Mode: 0644}); err != nil {
				t.Fatal(err)
			}
			for _, closer := range []interface{ Close() error }{writer, zip, file} {
				if err = closer.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err = unpack(filepath.Join(dir, "input.tgz"), filepath.Join(dir, "out")); err == nil {
				t.Fatal("unsafe package entry was accepted")
			}
		})
	}
}

func TestCapturedRequiresMatchingPublication(t *testing.T) {
	if captured("", "") == nil {
		t.Fatal("missing configuration accepted")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "publication.json"), []byte(`{"publicationId":"first"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"publicationId":"second"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if captured(dir, dir) == nil {
		t.Fatal("mismatched publication accepted")
	}
}

func TestCapturedSummaryRejectsMissingOrSkippedTests(t *testing.T) {
	for _, result := range []string{"", "# pass 0\n# skipped 0\n", "# pass 3\n# skipped 1\n", "# pass 4\n# skipped 1\n"} {
		if capturedSummary(result) == nil {
			t.Errorf("accepted incomplete captured checks: %q", result)
		}
	}
	if err := capturedSummary("# pass 4\n# skipped 0\n"); err != nil {
		t.Fatal(err)
	}
}
