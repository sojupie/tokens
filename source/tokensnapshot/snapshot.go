// Build turns a capture receipt into an immutable snapshot only after checking its capture requirements.
package tokensnapshot

import (
	"encoding/json"
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/crawl"
)

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
	graphJSON, err := json.Marshal(graph)
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
