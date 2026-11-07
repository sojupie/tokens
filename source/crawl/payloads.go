// Payload storage deduplicates captured bytes by digest while retaining the URLs and page provenance that produced them.
package crawl

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"sort"
	"strings"

	"time"
)

type pageSource struct {
	PageURLs          []string
	SourcePayloadURL  string
	SourcePayloadKind string
	SourceArea        string
	SourceSubject     string
	SourcePageType    string
}

func uniqueSources(observations []Observation) []pageSource {
	byURL := map[string]*pageSource{}
	for _, observation := range observations {
		for _, payloadURL := range observation.SourcePayloadURLs {
			if payloadURL == "" {
				continue
			}
			source := byURL[payloadURL]
			if source == nil {
				area, subject, pageType, kind := sourcePayloadContext(observation.PageURL, payloadURL)
				source = &pageSource{
					SourcePayloadURL:  payloadURL,
					SourcePayloadKind: kind,
					SourceArea:        area,
					SourceSubject:     subject,
					SourcePageType:    pageType,
				}
				byURL[payloadURL] = source
			}
			source.PageURLs = mergeStrings(source.PageURLs, []string{observation.PageURL})
		}
	}
	urls := make([]string, 0, len(byURL))
	for payloadURL := range byURL {
		urls = append(urls, payloadURL)
	}
	sort.Strings(urls)
	result := make([]pageSource, 0, len(urls))
	for _, payloadURL := range urls {
		result = append(result, *byURL[payloadURL])
	}
	return result
}

func sourcePayloadContext(pageURL, payloadURL string) (area, subject, pageType, kind string) {
	area, subject, pageType = pageMetadata(pageURL)
	if area == "components" {
		return area, subject, pageType, "component"
	}
	if area == "styles" {
		parts := strings.Split(subject, "/")
		if len(parts) > 0 && parts[0] != "" {
			return area, subject, pageType, strings.ToLower(parts[0])
		}
	}
	parsedPayload, err := url.Parse(payloadURL)
	if err != nil {
		return area, subject, pageType, "unknown"
	}
	stem := strings.ToLower(strings.TrimSuffix(filepath.Base(parsedPayload.Path), filepath.Ext(parsedPayload.Path)))
	if index := strings.IndexByte(stem, '.'); index >= 0 {
		stem = stem[:index]
	}
	if strings.HasPrefix(stem, "token_type_") {
		stem = strings.TrimPrefix(stem, "token_type_")
	}
	if stem == "token_table" || stem == "" {
		stem = "unknown"
	}
	return area, subject, pageType, stem
}

func loadIndex(path string) (CacheIndex, error) {
	index := CacheIndex{SchemaVersion: 2, SourcePayloads: map[string]SourcePayloadRecord{}}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return index, nil
	}
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return index, fmt.Errorf("parse cache index: %w", err)
	}
	if index.SourcePayloads == nil {
		index.SourcePayloads = map[string]SourcePayloadRecord{}
	}
	return index, nil
}

func persistPayload(ctx context.Context, config Config, index CacheIndex, source pageSource) (SourcePayloadRecord, error) {
	if cached, ok := index.SourcePayloads[source.SourcePayloadURL]; ok && !config.Refresh {
		path := filepath.Join(config.CacheDir, cached.Artifact)
		if raw, err := readGzip(path); err == nil && sha256Hex(raw) == cached.Hash && len(raw) == cached.Bytes {
			cached.PageURLs = append([]string(nil), source.PageURLs...)
			cached.SourcePayloadURL = source.SourcePayloadURL
			cached.SourcePayloadKind = source.SourcePayloadKind
			cached.SourceArea = source.SourceArea
			cached.SourceSubject = source.SourceSubject
			cached.SourcePageType = source.SourcePageType
			cached.Reused = true
			return cached, nil
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.SourcePayloadURL, nil)
	if err != nil {
		return SourcePayloadRecord{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(request)
	if err != nil {
		return SourcePayloadRecord{}, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode >= http.StatusBadRequest {
		return SourcePayloadRecord{}, fmt.Errorf("%s", response.Status)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return SourcePayloadRecord{}, err
	}
	hashBytes := sha256.Sum256(raw)
	hash := hex.EncodeToString(hashBytes[:])
	artifact := filepath.Join("source-payloads", hash[:2], hash+".json.gz")
	if err := writeGzipIfAbsent(filepath.Join(config.CacheDir, artifact), raw); err != nil {
		return SourcePayloadRecord{}, err
	}
	record := SourcePayloadRecord{
		PageURLs:          append([]string(nil), source.PageURLs...),
		SourcePayloadURL:  source.SourcePayloadURL,
		SourcePayloadKind: source.SourcePayloadKind,
		SourceArea:        source.SourceArea,
		SourceSubject:     source.SourceSubject,
		SourcePageType:    source.SourcePageType,
		Hash:              hash,
		Artifact:          artifact,
		Bytes:             len(raw),
		RetrievedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		ETag:              response.Header.Get("ETag"),
		LastModified:      response.Header.Get("Last-Modified"),
	}
	index.SourcePayloads[source.SourcePayloadURL] = record
	return record, nil
}

func readGzip(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	body, readErr := io.ReadAll(reader)
	closeReaderErr := reader.Close()
	closeFileErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeReaderErr != nil {
		return nil, closeReaderErr
	}
	if closeFileErr != nil {
		return nil, closeFileErr
	}
	return body, nil
}

func sha256Hex(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

func writeGzipIfAbsent(path string, body []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".payload-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	writer := gzip.NewWriter(temporary)
	if _, err := writer.Write(body); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}

func writeJSONAtomic(path string, value interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".json-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(append(body, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
