// Document projection preserves raw payload references and failed page results alongside the normalized graph.
package tokensnapshot

import (
	"errors"
	"fmt"

	"os"
	"path/filepath"

	"strings"

	"github.com/sojupie/tokens/source/crawl"
)

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
	return "payload-" + sha256Hex([]byte(sourcePayloadKey(payload.SourcePayloadURL)))
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
