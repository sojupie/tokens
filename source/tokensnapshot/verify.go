// Verification distinguishes byte integrity from full document validation; integrity-only checks are safe only for previously validated cache entries.
package tokensnapshot

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"strings"

	"github.com/sojupie/tokens/source/crawl"
)

func Verify(directory string) (Manifest, error) {
	manifest, _, err := VerifyAndLoad(directory)
	return manifest, err
}

// VerifyAndLoad returns the graph already decoded during full verification.
func VerifyAndLoad(directory string) (Manifest, Graph, error) {
	var graph Graph
	manifest, err := verify(directory, true, &graph)
	return manifest, graph, err
}

// VerifyIntegrity checks bytes against the manifest without decoding documents.
// Use only when reusing a previously fully validated, versioned cache entry.
func VerifyIntegrity(directory string) (Manifest, error) {
	return verify(directory, false, nil)
}

func verify(directory string, full bool, loaded *Graph) (Manifest, error) {
	manifestBody, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBody, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	if manifest.SchemaVersion != 2 && manifest.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("unsupported snapshot schema version %d", manifest.SchemaVersion)
	}
	seenPaths := make(map[string]bool, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		path, err := snapshotArtifactPath(artifact.Path)
		if err != nil || path != filepath.ToSlash(artifact.Path) {
			return Manifest{}, fmt.Errorf("invalid artifact path %q", artifact.Path)
		}
		if seenPaths[path] {
			return Manifest{}, fmt.Errorf("duplicate artifact path %q", artifact.Path)
		}
		seenPaths[path] = true
		file, err := os.Open(filepath.Join(directory, filepath.FromSlash(path)))
		if err != nil {
			return Manifest{}, fmt.Errorf("read artifact %s: %w", artifact.Path, err)
		}
		hash := sha256.New()
		size, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return Manifest{}, copyErr
		}
		if closeErr != nil {
			return Manifest{}, closeErr
		}
		if got := hex.EncodeToString(hash.Sum(nil)); got != artifact.Sha256 {
			return Manifest{}, fmt.Errorf("artifact %s hash mismatch: got %s, want %s", artifact.Path, got, artifact.Sha256)
		}
		if size != int64(artifact.Bytes) {
			return Manifest{}, fmt.Errorf("artifact %s byte count mismatch", artifact.Path)
		}
		if full && artifact.ContentEncoding == "gzip" {
			if _, _, err := gzipDigest(filepath.Join(directory, filepath.FromSlash(path))); err != nil {
				return Manifest{}, fmt.Errorf("read gzip artifact %s: %w", artifact.Path, err)
			}
		}
	}
	checksums, err := os.ReadFile(filepath.Join(directory, "checksums.txt"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read checksums: %w", err)
	}
	want := checksumFile(manifestBody, manifest.Artifacts)
	if string(checksums) != want {
		return Manifest{}, errors.New("checksums.txt does not match snapshot artifacts")
	}
	if !full {
		return manifest, nil
	}
	if manifest.SchemaVersion == SchemaVersion {
		if err := validateJSONSchema("manifest.schema.json", manifestBody); err != nil {
			return Manifest{}, err
		}
		documents := make(map[string][]byte)
		for _, name := range []string{"source-payload-details.json", "page-results.json", "lifecycle-notices.json"} {
			body, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				return Manifest{}, fmt.Errorf("read required document %s: %w", name, err)
			}
			documents[name] = body
		}
		for name, schema := range map[string]string{
			"source-payload-details.json": "source-payload-details.schema.json",
			"page-results.json":           "page-results.schema.json",
			"lifecycle-notices.json":      "lifecycle-notices.schema.json",
		} {
			if err := validateJSONSchema(schema, documents[name]); err != nil {
				return Manifest{}, err
			}
		}
	}
	pageCatalogBody, err := os.ReadFile(filepath.Join(directory, "page-catalog.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read page catalog: %w", err)
	}
	if err := crawl.ValidatePageCatalog(pageCatalogBody); err != nil {
		return Manifest{}, fmt.Errorf("validate page catalog: %w", err)
	}
	graph, err := readGraphValidated(directory, manifest.SchemaVersion == SchemaVersion)
	if err != nil {
		return Manifest{}, fmt.Errorf("read generated graph: %w", err)
	}
	for _, payload := range graph.SourcePayloads {
		if payload.Hash == "" || payload.Error != "" {
			continue
		}
		path, err := snapshotArtifactPath(payload.Artifact)
		if err != nil {
			return Manifest{}, err
		}
		hash, size, err := gzipDigest(filepath.Join(directory, filepath.FromSlash(path)))
		if err != nil {
			return Manifest{}, fmt.Errorf("read source payload %s: %w", payload.SourcePayloadURL, err)
		}
		if hash != payload.Hash || size != int64(payload.Bytes) {
			return Manifest{}, fmt.Errorf("source payload %s hash or byte count mismatch", payload.SourcePayloadURL)
		}
	}
	if loaded != nil {
		*loaded = graph
	}
	return manifest, nil
}

func gzipDigest(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return "", 0, err
	}
	hash := sha256.New()
	size, readErr := io.Copy(hash, reader)
	closeErr := reader.Close()
	if readErr != nil {
		return "", 0, readErr
	}
	if closeErr != nil {
		return "", 0, closeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func snapshotArtifactPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("successful payload has no artifact path")
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if filepath.IsAbs(filepath.FromSlash(path)) || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != filepath.ToSlash(path) {
		return "", fmt.Errorf("payload artifact path escapes snapshot: %s", path)
	}
	return clean, nil
}

func snapshotSourcePayloadPath(path string) (string, error) {
	return snapshotArtifactPath(path)
}
