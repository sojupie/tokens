package m3snapshot

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sojupie/m3-source/internal/m3crawl"
)

const SchemaVersion = 1

type BuildConfig struct {
	RunManifestPath  string
	CatalogPath      string
	OutputDir        string
	SnapshotID       string
	CreatedAt        string
	ExtractorName    string
	ExtractorVersion string
}

type Manifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	SnapshotID    string        `json:"snapshotId"`
	CreatedAt     string        `json:"createdAt"`
	Source        SourceInfo    `json:"source"`
	Capture       CaptureInfo   `json:"capture"`
	Extractor     ExtractorInfo `json:"extractor"`
	Artifacts     []Artifact    `json:"artifacts"`
}

type SourceInfo struct {
	System        string `json:"system"`
	CatalogSha256 string `json:"catalogSha256"`
	PageCount     int    `json:"pageCount"`
}

type CaptureInfo struct {
	RunID            string `json:"runId"`
	RunManifest      string `json:"runManifest"`
	PagesAttempted   int    `json:"pagesAttempted"`
	PagesSucceeded   int    `json:"pagesSucceeded"`
	PayloadsObserved int    `json:"payloadsObserved"`
	PayloadsFetched  int    `json:"payloadsFetched"`
	Complete         bool   `json:"complete"`
}

type ExtractorInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	SchemaVersion int    `json:"schemaVersion"`
}

type Artifact struct {
	Path            string `json:"path"`
	MediaType       string `json:"mediaType"`
	ContentEncoding string `json:"contentEncoding,omitempty"`
	Sha256          string `json:"sha256"`
	Bytes           int    `json:"bytes"`
}

type Graph struct {
	SchemaVersion int                   `json:"schemaVersion"`
	SnapshotID    string                `json:"snapshotId"`
	RunID         string                `json:"runId"`
	Pages         []string              `json:"pages"`
	Observations  []m3crawl.Observation `json:"observations"`
	Payloads      []GraphPayload        `json:"payloads"`
}

type GraphPayload struct {
	PageURL       string `json:"pageUrl"`
	TokenTableURL string `json:"tokenTableUrl"`
	Hash          string `json:"hash,omitempty"`
	Artifact      string `json:"artifact,omitempty"`
	Bytes         int    `json:"bytes,omitempty"`
	Error         string `json:"error,omitempty"`
}

type Provenance struct {
	SchemaVersion int                `json:"schemaVersion"`
	Records       []ProvenanceRecord `json:"records"`
}

type ProvenanceRecord struct {
	ID            string `json:"id"`
	PageURL       string `json:"pageUrl"`
	PayloadURL    string `json:"payloadUrl"`
	PayloadSha256 string `json:"payloadSha256,omitempty"`
	Artifact      string `json:"artifact,omitempty"`
	RetrievedAt   string `json:"retrievedAt,omitempty"`
	ETag          string `json:"etag,omitempty"`
	LastModified  string `json:"lastModified,omitempty"`
	Error         string `json:"error,omitempty"`
}

type Coverage struct {
	SchemaVersion int            `json:"schemaVersion"`
	Pages         []CoveragePage `json:"pages"`
}

type CoveragePage struct {
	URL                string `json:"url"`
	Status             int    `json:"status"`
	TokenTableObserved bool   `json:"tokenTableObserved"`
	PayloadsObserved   int    `json:"payloadsObserved"`
	PayloadsFetched    int    `json:"payloadsFetched"`
	PayloadError       string `json:"payloadError,omitempty"`
	Error              string `json:"error,omitempty"`
}

func Build(config BuildConfig) (Manifest, error) {
	if err := config.validate(); err != nil {
		return Manifest{}, err
	}
	run, err := readRunManifest(config.RunManifestPath)
	if err != nil {
		return Manifest{}, err
	}
	if err := validateCapture(run); err != nil {
		return Manifest{}, err
	}
	createdAt := config.CreatedAt
	if createdAt == "" {
		createdAt = run.StartedAt
	}
	if createdAt == "" {
		return Manifest{}, errors.New("createdAt is required when the run manifest has no startedAt")
	}
	catalogPath := config.CatalogPath
	if catalogPath == "" {
		catalogPath = run.Options.CatalogPath
	}
	catalog, err := readCanonicalJSON(catalogPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("read catalog: %w", err)
	}
	graph, provenance, coverage := buildDocuments(config.SnapshotID, run)
	graphJSON, err := marshalDocument(graph)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode graph: %w", err)
	}
	provenanceJSON, err := marshalDocument(provenance)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode provenance: %w", err)
	}
	coverageJSON, err := marshalDocument(coverage)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode coverage: %w", err)
	}
	graphGzip, err := gzipJSON(graphJSON)
	if err != nil {
		return Manifest{}, fmt.Errorf("compress graph: %w", err)
	}
	if err := ensureEmptyDir(config.OutputDir); err != nil {
		return Manifest{}, err
	}
	files := map[string][]byte{
		"catalog.json":           catalog,
		"material.graph.json.gz": graphGzip,
		"provenance.json":        provenanceJSON,
		"coverage.json":          coverageJSON,
	}
	artifacts := make([]Artifact, 0, len(files))
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		body := files[path]
		if err := writeFile(filepath.Join(config.OutputDir, path), body); err != nil {
			return Manifest{}, fmt.Errorf("write %s: %w", path, err)
		}
		artifacts = append(artifacts, Artifact{
			Path:            path,
			MediaType:       mediaType(path),
			ContentEncoding: encoding(path),
			Sha256:          sha256Hex(body),
			Bytes:           len(body),
		})
	}
	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		SnapshotID:    config.SnapshotID,
		CreatedAt:     createdAt,
		Source: SourceInfo{
			System:        "material-3",
			CatalogSha256: sha256Hex(catalog),
			PageCount:     len(run.Pages),
		},
		Capture: CaptureInfo{
			RunID:            run.RunID,
			RunManifest:      filepath.Base(config.RunManifestPath),
			PagesAttempted:   len(run.Pages),
			PagesSucceeded:   successfulPageCount(run),
			PayloadsObserved: len(run.Payloads),
			PayloadsFetched:  successfulPayloadCount(run),
			Complete:         successfulPageCount(run) == len(run.Pages) && len(run.Pages) > 0 && successfulPayloadCount(run) == len(run.Payloads) && len(run.Payloads) > 0,
		},
		Extractor: ExtractorInfo{
			Name:          valueOr(config.ExtractorName, "m3source"),
			Version:       valueOr(config.ExtractorVersion, "0.1.0"),
			SchemaVersion: SchemaVersion,
		},
		Artifacts: artifacts,
	}
	manifestJSON, err := marshalDocument(manifest)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode manifest: %w", err)
	}
	if err := writeFile(filepath.Join(config.OutputDir, "manifest.json"), manifestJSON); err != nil {
		return Manifest{}, fmt.Errorf("write manifest: %w", err)
	}
	checksums := checksumFile(manifestJSON, artifacts)
	if err := writeFile(filepath.Join(config.OutputDir, "checksums.txt"), []byte(checksums)); err != nil {
		return Manifest{}, fmt.Errorf("write checksums: %w", err)
	}
	return manifest, nil
}

func successfulPageCount(run m3crawl.RunManifest) int {
	count := 0
	for _, observation := range run.Observations {
		if observation.Status >= 200 && observation.Status < 400 && observation.Error == "" {
			count++
		}
	}
	return count
}

func successfulPayloadCount(run m3crawl.RunManifest) int {
	count := 0
	for _, payload := range run.Payloads {
		if payload.Hash != "" && payload.Error == "" {
			count++
		}
	}
	return count
}

func Verify(directory string) (Manifest, error) {
	manifestBody, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBody, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	if manifest.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("unsupported snapshot schema version %d", manifest.SchemaVersion)
	}
	for _, artifact := range manifest.Artifacts {
		body, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(artifact.Path)))
		if err != nil {
			return Manifest{}, fmt.Errorf("read artifact %s: %w", artifact.Path, err)
		}
		if got := sha256Hex(body); got != artifact.Sha256 {
			return Manifest{}, fmt.Errorf("artifact %s hash mismatch: got %s, want %s", artifact.Path, got, artifact.Sha256)
		}
		if len(body) != artifact.Bytes {
			return Manifest{}, fmt.Errorf("artifact %s byte count mismatch", artifact.Path)
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
	return manifest, nil
}

func (config BuildConfig) validate() error {
	if config.RunManifestPath == "" || config.OutputDir == "" || config.SnapshotID == "" {
		return errors.New("run manifest, output directory, and snapshot ID are required")
	}
	if strings.ContainsAny(config.SnapshotID, `/\\`) {
		return errors.New("snapshot ID must not contain path separators")
	}
	return nil
}

func readRunManifest(path string) (m3crawl.RunManifest, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return m3crawl.RunManifest{}, fmt.Errorf("read run manifest: %w", err)
	}
	var run m3crawl.RunManifest
	if err := json.Unmarshal(body, &run); err != nil {
		return m3crawl.RunManifest{}, fmt.Errorf("parse run manifest: %w", err)
	}
	if run.SchemaVersion != 1 {
		return m3crawl.RunManifest{}, fmt.Errorf("unsupported run manifest schema version %d", run.SchemaVersion)
	}
	return run, nil
}

func validateCapture(run m3crawl.RunManifest) error {
	if successfulPageCount(run) == 0 {
		return errors.New("cannot build snapshot: capture has no successful pages")
	}
	if successfulPayloadCount(run) == 0 {
		return errors.New("cannot build snapshot: capture has no successful payloads")
	}
	return nil
}

func buildDocuments(snapshotID string, run m3crawl.RunManifest) (Graph, Provenance, Coverage) {
	graphPayloads := make([]GraphPayload, 0, len(run.Payloads))
	provenanceRecords := make([]ProvenanceRecord, 0, len(run.Payloads))
	payloadsByPage := map[string][]m3crawl.PayloadRecord{}
	for index, payload := range run.Payloads {
		graphPayloads = append(graphPayloads, GraphPayload{
			PageURL:       payload.PageURL,
			TokenTableURL: payload.TokenTableURL,
			Hash:          payload.Hash,
			Artifact:      payload.Artifact,
			Bytes:         payload.Bytes,
			Error:         payload.Error,
		})
		provenanceRecords = append(provenanceRecords, ProvenanceRecord{
			ID:            fmt.Sprintf("payload-%04d", index+1),
			PageURL:       payload.PageURL,
			PayloadURL:    payload.TokenTableURL,
			PayloadSha256: payload.Hash,
			Artifact:      payload.Artifact,
			RetrievedAt:   payload.RetrievedAt,
			ETag:          payload.ETag,
			LastModified:  payload.LastModified,
			Error:         payload.Error,
		})
		payloadsByPage[payload.PageURL] = append(payloadsByPage[payload.PageURL], payload)
	}
	coveragePages := make([]CoveragePage, 0, len(run.Observations))
	for _, observation := range run.Observations {
		pagePayloads := payloadsByPage[observation.PageURL]
		payloadsFetched := 0
		payloadErrors := make([]string, 0)
		for _, payload := range pagePayloads {
			if payload.Hash != "" && payload.Error == "" {
				payloadsFetched++
			}
			if payload.Error != "" {
				payloadErrors = append(payloadErrors, payload.Error)
			}
		}
		coveragePages = append(coveragePages, CoveragePage{
			URL:                observation.PageURL,
			Status:             observation.Status,
			TokenTableObserved: observation.TokenTableURL != "",
			PayloadsObserved:   len(pagePayloads),
			PayloadsFetched:    payloadsFetched,
			PayloadError:       strings.Join(payloadErrors, "; "),
			Error:              observation.Error,
		})
	}
	return Graph{
		SchemaVersion: SchemaVersion,
		SnapshotID:    snapshotID,
		RunID:         run.RunID,
		Pages:         append([]string(nil), run.Pages...),
		Observations:  append([]m3crawl.Observation(nil), run.Observations...),
		Payloads:      graphPayloads,
	}, Provenance{SchemaVersion: SchemaVersion, Records: provenanceRecords}, Coverage{SchemaVersion: SchemaVersion, Pages: coveragePages}
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
	writer.Header.ModTime = time.Time{}
	writer.Header.Name = ""
	writer.Header.Comment = ""
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
