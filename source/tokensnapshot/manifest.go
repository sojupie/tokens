// Snapshot documents record capture completeness and artifact digests independently from the normalized token graph.
package tokensnapshot

import (
	"github.com/sojupie/tokens/source/crawl"
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
