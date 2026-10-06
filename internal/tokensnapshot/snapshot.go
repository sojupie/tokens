package tokensnapshot

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

	"github.com/sojupie/tokens/internal/crawl"
)

const SchemaVersion = 3

type BuildConfig struct {
	CrawlReceiptPath string
	PageCatalogPath  string
	OutputDir        string
	SnapshotID       string
	CreatedAt        string
	ExtractorName    string
	ExtractorVersion string
}

type Manifest struct {
	Type          string        `json:"type"`
	SchemaVersion int           `json:"schemaVersion"`
	SnapshotID    string        `json:"snapshotId"`
	CreatedAt     string        `json:"createdAt"`
	ReviewStatus  string        `json:"reviewStatus"`
	Source        SourceInfo    `json:"source"`
	Capture       CaptureInfo   `json:"capture"`
	Extractor     ExtractorInfo `json:"extractor"`
	Artifacts     []Artifact    `json:"artifacts"`
}

type SourceInfo struct {
	System            string `json:"system"`
	PageCatalogSha256 string `json:"pageCatalogSha256"`
	PageCount         int    `json:"pageCount"`
}

type CaptureInfo struct {
	CrawlID          string `json:"crawlId"`
	CrawlReceipt     string `json:"crawlReceipt"`
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
	SchemaVersion  int                 `json:"schemaVersion"`
	SnapshotID     string              `json:"snapshotId"`
	CrawlID        string              `json:"crawlId"`
	Pages          []string            `json:"pages"`
	Observations   []crawl.Observation `json:"observations"`
	SourcePayloads []GraphPayload      `json:"sourcePayloads"`
	Records        []ExtractedRecord   `json:"records"`
}

type GraphPayload struct {
	PageURLs          []string `json:"pageUrls"`
	SourcePayloadURL  string   `json:"sourcePayloadUrl"`
	SourcePayloadKind string   `json:"sourcePayloadKind,omitempty"`
	SourceArea        string   `json:"sourceArea,omitempty"`
	SourceSubject     string   `json:"sourceSubject,omitempty"`
	SourcePageType    string   `json:"sourcePageType,omitempty"`
	Hash              string   `json:"hash,omitempty"`
	Artifact          string   `json:"artifact,omitempty"`
	Bytes             int      `json:"bytes,omitempty"`
	Records           int      `json:"records"`
	ExtractionError   string   `json:"extractionError,omitempty"`
	Error             string   `json:"error,omitempty"`
}

type SourcePayloadDetails struct {
	SchemaVersion int                   `json:"schemaVersion"`
	Records       []SourcePayloadDetail `json:"records"`
}

type SourcePayloadDetail struct {
	ID                string   `json:"id"`
	PageURLs          []string `json:"pageUrls"`
	SourcePayloadURL  string   `json:"sourcePayloadUrl"`
	SourcePayloadKind string   `json:"sourcePayloadKind,omitempty"`
	SourceArea        string   `json:"sourceArea,omitempty"`
	SourceSubject     string   `json:"sourceSubject,omitempty"`
	SourcePageType    string   `json:"sourcePageType,omitempty"`
	PayloadSha256     string   `json:"payloadSha256,omitempty"`
	Artifact          string   `json:"artifact,omitempty"`
	RetrievedAt       string   `json:"retrievedAt,omitempty"`
	ETag              string   `json:"etag,omitempty"`
	LastModified      string   `json:"lastModified,omitempty"`
	Error             string   `json:"error,omitempty"`
}

type PageResults struct {
	SchemaVersion int          `json:"schemaVersion"`
	Pages         []PageResult `json:"pages"`
}

type LifecycleNotices struct {
	SchemaVersion int               `json:"schemaVersion"`
	Notices       []LifecycleNotice `json:"notices"`
}

type LifecycleNotice struct {
	ID              string `json:"id"`
	Scope           string `json:"scope"`
	State           string `json:"state"`
	Subject         string `json:"subject"`
	Message         string `json:"message"`
	SourcePayloadID string `json:"sourcePayloadId"`
	SourceRecordID  string `json:"sourceRecordId"`
	RecordKind      string `json:"recordKind"`
	SourceField     string `json:"sourceField"`
}

type PageResult struct {
	URL                    string `json:"url"`
	Status                 int    `json:"status"`
	SourcePayloadsObserved bool   `json:"sourcePayloadsObserved"`
	PayloadsObserved       int    `json:"payloadsObserved"`
	PayloadsFetched        int    `json:"payloadsFetched"`
	RecordsExtracted       int    `json:"recordsExtracted"`
	PayloadError           string `json:"payloadError,omitempty"`
	ExtractionError        string `json:"extractionError,omitempty"`
	Error                  string `json:"error,omitempty"`
}

func Build(config BuildConfig) (Manifest, error) {
	if err := config.validate(); err != nil {
		return Manifest{}, err
	}
	run, err := readCrawlReceipt(config.CrawlReceiptPath)
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
		return Manifest{}, errors.New("createdAt is required when the crawl receipt has no startedAt")
	}
	pageCatalogPath := config.PageCatalogPath
	if pageCatalogPath == "" {
		pageCatalogPath = run.Options.PageCatalogPath
	}
	pageCatalog, err := readCanonicalJSON(pageCatalogPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("read page catalog: %w", err)
	}
	if err := crawl.ValidatePageCatalog(pageCatalog); err != nil {
		return Manifest{}, fmt.Errorf("validate page catalog: %w", err)
	}
	graph, payloadDetails, pageResults, lifecycleNotices := buildDocuments(config.SnapshotID, run)
	graphJSON, err := marshalDocument(graph)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode graph: %w", err)
	}
	payloadDetailsJSON, err := marshalDocument(payloadDetails)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode source-payload details: %w", err)
	}
	pageResultsJSON, err := marshalDocument(pageResults)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode page results: %w", err)
	}
	lifecycleNoticesJSON, err := marshalDocument(lifecycleNotices)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode lifecycle notices: %w", err)
	}
	graphGzip, err := gzipJSON(graphJSON)
	if err != nil {
		return Manifest{}, fmt.Errorf("compress graph: %w", err)
	}
	payloadFiles, err := readPayloadArtifacts(run)
	if err != nil {
		return Manifest{}, fmt.Errorf("read payload artifacts: %w", err)
	}
	if err := ensureEmptyDir(config.OutputDir); err != nil {
		return Manifest{}, err
	}
	files := map[string][]byte{
		"page-catalog.json":             pageCatalog,
		"generated-token-graph.json.gz": graphGzip,
		"source-payload-details.json":   payloadDetailsJSON,
		"page-results.json":             pageResultsJSON,
		"lifecycle-notices.json":        lifecycleNoticesJSON,
	}
	for path, body := range payloadFiles {
		files[path] = body
	}
	if err := validateSnapshotDocuments(files); err != nil {
		return Manifest{}, err
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
		Type:          "material-token-data-snapshot",
		SchemaVersion: SchemaVersion,
		SnapshotID:    config.SnapshotID,
		CreatedAt:     createdAt,
		ReviewStatus:  "provisional",
		Source: SourceInfo{
			System:            "material-3",
			PageCatalogSha256: sha256Hex(pageCatalog),
			PageCount:         len(run.Pages),
		},
		Capture: CaptureInfo{
			CrawlID:          run.CrawlID,
			CrawlReceipt:     filepath.Base(config.CrawlReceiptPath),
			PagesAttempted:   len(run.Pages),
			PagesSucceeded:   successfulPageCount(run),
			PayloadsObserved: uniqueSuccessfulPayloadURLs(run),
			PayloadsFetched:  successfulPayloadCount(run),
			Complete:         successfulPageCount(run) == len(run.Pages) && len(run.Pages) > 0 && successfulPayloadCount(run) == uniqueSuccessfulPayloadURLs(run) && uniqueSuccessfulPayloadURLs(run) > 0,
		},
		Extractor: ExtractorInfo{
			Name:          valueOr(config.ExtractorName, "tokens"),
			Version:       valueOr(config.ExtractorVersion, "0.1.0"),
			SchemaVersion: SchemaVersion,
		},
		Artifacts: artifacts,
	}
	manifestJSON, err := marshalDocument(manifest)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode manifest: %w", err)
	}
	if err := validateJSONSchema("manifest.schema.json", manifestJSON); err != nil {
		return Manifest{}, err
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

func successfulPageCount(run crawl.CrawlReceipt) int {
	count := 0
	for _, observation := range run.Observations {
		if observation.Status >= 200 && observation.Status < 400 && observation.Error == "" {
			count++
		}
	}
	return count
}

func successfulPayloadCount(run crawl.CrawlReceipt) int {
	seen := map[string]bool{}
	count := 0
	for _, payload := range run.SourcePayloads {
		if payload.SourcePayloadURL != "" && payload.Hash != "" && payload.Error == "" && !seen[payload.SourcePayloadURL] {
			seen[payload.SourcePayloadURL] = true
			count++
		}
	}
	return count
}

func uniqueSuccessfulPayloadURLs(run crawl.CrawlReceipt) int {
	seen := map[string]bool{}
	for _, payload := range run.SourcePayloads {
		if payload.SourcePayloadURL != "" {
			seen[payload.SourcePayloadURL] = true
		}
	}
	return len(seen)
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
		body, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(path)))
		if err != nil {
			return Manifest{}, fmt.Errorf("read artifact %s: %w", artifact.Path, err)
		}
		if got := sha256Hex(body); got != artifact.Sha256 {
			return Manifest{}, fmt.Errorf("artifact %s hash mismatch: got %s, want %s", artifact.Path, got, artifact.Sha256)
		}
		if len(body) != artifact.Bytes {
			return Manifest{}, fmt.Errorf("artifact %s byte count mismatch", artifact.Path)
		}
		if artifact.ContentEncoding == "gzip" {
			if _, err := readGzip(filepath.Join(directory, filepath.FromSlash(path))); err != nil {
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
	if manifest.SchemaVersion == SchemaVersion {
		if err := validateJSONSchema("manifest.schema.json", manifestBody); err != nil {
			return Manifest{}, err
		}
		documents := make(map[string][]byte)
		for _, name := range []string{"generated-token-graph.json.gz", "source-payload-details.json", "page-results.json", "lifecycle-notices.json"} {
			body, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				return Manifest{}, fmt.Errorf("read required document %s: %w", name, err)
			}
			documents[name] = body
		}
		if err := validateSnapshotDocuments(documents); err != nil {
			return Manifest{}, err
		}
	}
	pageCatalogBody, err := os.ReadFile(filepath.Join(directory, "page-catalog.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read page catalog: %w", err)
	}
	if err := crawl.ValidatePageCatalog(pageCatalogBody); err != nil {
		return Manifest{}, fmt.Errorf("validate page catalog: %w", err)
	}
	graph, err := readGraph(directory)
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
		raw, err := readGzip(filepath.Join(directory, filepath.FromSlash(path)))
		if err != nil {
			return Manifest{}, fmt.Errorf("read source payload %s: %w", payload.SourcePayloadURL, err)
		}
		if sha256Hex(raw) != payload.Hash || len(raw) != payload.Bytes {
			return Manifest{}, fmt.Errorf("source payload %s hash or byte count mismatch", payload.SourcePayloadURL)
		}
	}
	return manifest, nil
}

func (config BuildConfig) validate() error {
	if config.CrawlReceiptPath == "" || config.OutputDir == "" || config.SnapshotID == "" {
		return errors.New("crawl receipt, output directory, and snapshot ID are required")
	}
	if strings.ContainsAny(config.SnapshotID, `/\\`) {
		return errors.New("snapshot ID must not contain path separators")
	}
	return nil
}

func readCrawlReceipt(path string) (crawl.CrawlReceipt, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return crawl.CrawlReceipt{}, fmt.Errorf("read crawl receipt: %w", err)
	}
	var run crawl.CrawlReceipt
	if err := json.Unmarshal(body, &run); err != nil {
		return crawl.CrawlReceipt{}, fmt.Errorf("parse crawl receipt: %w", err)
	}
	if run.SchemaVersion != 2 {
		return crawl.CrawlReceipt{}, fmt.Errorf("unsupported crawl receipt schema version %d", run.SchemaVersion)
	}
	return run, nil
}

func validateCapture(run crawl.CrawlReceipt) error {
	if successfulPageCount(run) == 0 {
		return errors.New("cannot build snapshot: capture has no successful pages")
	}
	if successfulPayloadCount(run) == 0 {
		return errors.New("cannot build snapshot: capture has no successful payloads")
	}
	return nil
}

func buildDocuments(snapshotID string, run crawl.CrawlReceipt) (Graph, SourcePayloadDetails, PageResults, LifecycleNotices) {
	graphPayloads := make([]GraphPayload, 0, len(run.SourcePayloads))
	graphRecords := make([]ExtractedRecord, 0)
	payloadDetails := make([]SourcePayloadDetail, 0, len(run.SourcePayloads))
	payloadsByPage := map[string][]crawl.SourcePayloadRecord{}
	graphPayloadByURL := map[string]GraphPayload{}
	for _, payload := range run.SourcePayloads {
		sourcePayloadID := sourcePayloadID(payload)
		artifactPath, artifactErr := snapshotSourcePayloadPath(payload.Artifact)
		var records []ExtractedRecord
		extractionErr := artifactErr
		if extractionErr == nil {
			records, extractionErr = extractPayload(run.Options.CacheDir, sourcePayload{
				Artifact: payload.Artifact,
				Hash:     payload.Hash,
				Error:    payload.Error,
			}, sourcePayloadID)
		}
		for index := range records {
			records[index].SourcePayloadKind = payload.SourcePayloadKind
			records[index].SourceArea = payload.SourceArea
			records[index].SourceSubject = payload.SourceSubject
			records[index].SourcePageType = payload.SourcePageType
		}
		graphPayloads = append(graphPayloads, GraphPayload{
			PageURLs:          append([]string(nil), payload.PageURLs...),
			SourcePayloadURL:  payload.SourcePayloadURL,
			SourcePayloadKind: payload.SourcePayloadKind,
			SourceArea:        payload.SourceArea,
			SourceSubject:     payload.SourceSubject,
			SourcePageType:    payload.SourcePageType,
			Hash:              payload.Hash,
			Artifact:          artifactPath,
			Bytes:             payload.Bytes,
			Records:           len(records),
			ExtractionError:   errorString(extractionErr),
			Error:             payload.Error,
		})
		graphPayloadByURL[payload.SourcePayloadURL] = graphPayloads[len(graphPayloads)-1]
		graphRecords = append(graphRecords, records...)
		payloadDetails = append(payloadDetails, SourcePayloadDetail{
			ID:                sourcePayloadID,
			PageURLs:          append([]string(nil), payload.PageURLs...),
			SourcePayloadURL:  payload.SourcePayloadURL,
			SourcePayloadKind: payload.SourcePayloadKind,
			SourceArea:        payload.SourceArea,
			SourceSubject:     payload.SourceSubject,
			SourcePageType:    payload.SourcePageType,
			PayloadSha256:     payload.Hash,
			Artifact:          artifactPath,
			RetrievedAt:       payload.RetrievedAt,
			ETag:              payload.ETag,
			LastModified:      payload.LastModified,
			Error:             payload.Error,
		})
		for _, pageURL := range payload.PageURLs {
			payloadsByPage[pageURL] = append(payloadsByPage[pageURL], payload)
		}
	}
	pageResults := make([]PageResult, 0, len(run.Observations))
	for _, observation := range run.Observations {
		pagePayloads := payloadsByPage[observation.PageURL]
		payloadsFetched := 0
		recordsExtracted := 0
		payloadErrors := make([]string, 0)
		extractionErrors := make([]string, 0)
		for _, payload := range pagePayloads {
			if payload.Hash != "" && payload.Error == "" {
				payloadsFetched++
			}
			if payload.Error != "" {
				payloadErrors = append(payloadErrors, payload.Error)
			}
			if graphPayload, ok := graphPayloadByURL[payload.SourcePayloadURL]; ok {
				recordsExtracted += graphPayload.Records
				if graphPayload.ExtractionError != "" {
					extractionErrors = append(extractionErrors, graphPayload.ExtractionError)
				}
			}
		}
		pageResults = append(pageResults, PageResult{
			URL:                    observation.PageURL,
			Status:                 observation.Status,
			SourcePayloadsObserved: len(observation.SourcePayloadURLs) > 0,
			PayloadsObserved:       len(pagePayloads),
			PayloadsFetched:        payloadsFetched,
			RecordsExtracted:       recordsExtracted,
			PayloadError:           strings.Join(payloadErrors, "; "),
			ExtractionError:        strings.Join(extractionErrors, "; "),
			Error:                  observation.Error,
		})
	}
	return Graph{
		SchemaVersion:  SchemaVersion,
		SnapshotID:     snapshotID,
		CrawlID:        run.CrawlID,
		Pages:          append([]string{}, run.Pages...),
		Observations:   append([]crawl.Observation{}, run.Observations...),
		SourcePayloads: graphPayloads,
		Records:        graphRecords,
	}, SourcePayloadDetails{SchemaVersion: SchemaVersion, Records: payloadDetails}, PageResults{SchemaVersion: SchemaVersion, Pages: pageResults}, LifecycleNotices{SchemaVersion: 1, Notices: lifecycleNotices(graphRecords)}
}

func sourcePayloadID(payload crawl.SourcePayloadRecord) string {
	return "payload-" + sha256Hex([]byte(payload.SourcePayloadURL+"\x00"+payload.Hash))
}

func readPayloadArtifacts(run crawl.CrawlReceipt) (map[string][]byte, error) {
	files := make(map[string][]byte)
	for _, payload := range run.SourcePayloads {
		if payload.Hash == "" || payload.Error != "" {
			continue
		}
		if run.Options.CacheDir == "" {
			return nil, errors.New("cache directory is required to preserve successful payload artifacts")
		}
		cachePath, err := snapshotArtifactPath(payload.Artifact)
		if err != nil {
			return nil, err
		}
		path, err := snapshotSourcePayloadPath(cachePath)
		if err != nil {
			return nil, err
		}
		compressedPath := filepath.Join(run.Options.CacheDir, filepath.FromSlash(cachePath))
		body, err := os.ReadFile(compressedPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", payload.Artifact, err)
		}
		raw, err := readGzip(compressedPath)
		if err != nil {
			return nil, fmt.Errorf("read gzip %s: %w", payload.Artifact, err)
		}
		if sha256Hex(raw) != payload.Hash || len(raw) != payload.Bytes {
			return nil, fmt.Errorf("source payload %s does not match recorded hash or byte count", payload.SourcePayloadURL)
		}
		if existing, ok := files[path]; ok && sha256Hex(existing) != sha256Hex(body) {
			return nil, fmt.Errorf("payload artifact path %s has conflicting contents", path)
		}
		files[path] = body
	}
	return files, nil
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
