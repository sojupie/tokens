// Snapshot storage canonicalizes JSON and records checksums so later verification can detect any change to published evidence.
package tokensnapshot

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"fmt"

	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func readCanonicalJSON(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return canonicalJSON(body)
}

func canonicalJSON(body []byte) ([]byte, error) {
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return marshalDocument(value)
}

func marshalDocument(value interface{}) ([]byte, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(body, '\n'), nil
}

func gzipJSON(body []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	writer.ModTime = time.Time{}
	writer.Name = ""
	writer.Comment = ""
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ensureEmptyDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("snapshot output directory is not empty: %s", path)
	}
	return nil
}

func writeFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func checksumFile(manifest []byte, artifacts []Artifact) string {
	lines := []string{fmt.Sprintf("%s  manifest.json", sha256Hex(manifest))}
	for _, artifact := range artifacts {
		lines = append(lines, fmt.Sprintf("%s  %s", artifact.Sha256, artifact.Path))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func mediaType(path string) string {
	if strings.HasSuffix(path, ".gz") {
		return "application/json"
	}
	return "application/json"
}

func encoding(path string) string {
	if strings.HasSuffix(path, ".gz") {
		return "gzip"
	}
	return ""
}

func sha256Hex(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
