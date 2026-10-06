package publish

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sojupie/tokens/internal/tokensnapshot"
)

func loadSnapshot(directory string) (snapshotInput, error) {
	manifestBody, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return snapshotInput{}, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := tokensnapshot.Verify(directory)
	if err != nil {
		return snapshotInput{}, fmt.Errorf("verify snapshot: %w", err)
	}
	if manifest.SchemaVersion != tokensnapshot.SchemaVersion {
		return snapshotInput{}, fmt.Errorf("unsupported snapshot schema version %d", manifest.SchemaVersion)
	}
	graphBody, err := readGzipFile(filepath.Join(directory, "generated-token-graph.json.gz"))
	if err != nil {
		return snapshotInput{}, fmt.Errorf("read generated token graph: %w", err)
	}
	var input snapshotInput
	input.Manifest = manifest
	input.ManifestHash = sha256Hex(manifestBody)
	if err := json.Unmarshal(graphBody, &input.Graph); err != nil {
		return snapshotInput{}, fmt.Errorf("decode generated token graph: %w", err)
	}
	if err := readJSONFile(filepath.Join(directory, "source-payload-details.json"), &input.PayloadDetails); err != nil {
		return snapshotInput{}, fmt.Errorf("read source-payload details: %w", err)
	}
	if err := readJSONFile(filepath.Join(directory, "lifecycle-notices.json"), &input.LifecycleNotices); err != nil {
		return snapshotInput{}, fmt.Errorf("read lifecycle notices: %w", err)
	}
	return input, nil
}

func sha256Hex(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

func readJSONFile(path string, destination any) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, destination)
}

func readGzipFile(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	decoded, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	return decoded, closeErr
}
