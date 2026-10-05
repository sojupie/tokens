package tokensnapshot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SnapshotDiff struct {
	SchemaVersion      int                 `json:"schemaVersion"`
	FromSnapshotID     string              `json:"fromSnapshotId"`
	ToSnapshotID       string              `json:"toSnapshotId"`
	Changed            bool                `json:"changed"`
	PageCatalogChanged bool                `json:"pageCatalogChanged"`
	CaptureChanged     bool                `json:"captureChanged"`
	Pages              PageSetDiff         `json:"pages"`
	SourcePayloads     SourcePayloadDiff   `json:"sourcePayloads"`
	Records            RecordSetDiff       `json:"records"`
	PageResults        PageResultsDiff     `json:"pageResults"`
	LifecycleNotices   LifecycleNoticeDiff `json:"lifecycleNotices"`
	Semantic           SemanticDiff        `json:"semantic"`
}

type PageSetDiff struct {
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

type SourcePayloadDiff struct {
	Added   []SourcePayloadChange `json:"added,omitempty"`
	Removed []SourcePayloadChange `json:"removed,omitempty"`
	Changed []SourcePayloadChange `json:"changed,omitempty"`
}

type SourcePayloadChange struct {
	Key              string   `json:"key"`
	PageURLs         []string `json:"pageUrls,omitempty"`
	SourcePayloadURL string   `json:"sourcePayloadUrl"`
	BeforePageURLs   []string `json:"beforePageUrls,omitempty"`
	AfterPageURLs    []string `json:"afterPageUrls,omitempty"`
	BeforeKind       string   `json:"beforeKind,omitempty"`
	AfterKind        string   `json:"afterKind,omitempty"`
	BeforeHash       string   `json:"beforeHash,omitempty"`
	AfterHash        string   `json:"afterHash,omitempty"`
	BeforeError      string   `json:"beforeError,omitempty"`
	AfterError       string   `json:"afterError,omitempty"`
	BeforeBytes      int      `json:"beforeBytes,omitempty"`
	AfterBytes       int      `json:"afterBytes,omitempty"`
}

type PageResultsDiff struct {
	Added   []PageResultChange `json:"added,omitempty"`
	Removed []PageResultChange `json:"removed,omitempty"`
	Changed []PageResultChange `json:"changed,omitempty"`
}

type LifecycleNoticeDiff struct {
	Added   []LifecycleNotice `json:"added,omitempty"`
	Removed []LifecycleNotice `json:"removed,omitempty"`
}

type PageResultChange struct {
	URL    string     `json:"url"`
	Before PageResult `json:"before"`
	After  PageResult `json:"after"`
}

type RecordSetDiff struct {
	Added   []RecordChange `json:"added,omitempty"`
	Removed []RecordChange `json:"removed,omitempty"`
	Changed []RecordChange `json:"changed,omitempty"`
}

type RecordChange struct {
	Key          string `json:"key"`
	Kind         string `json:"kind"`
	TokenName    string `json:"tokenName,omitempty"`
	BeforeSha256 string `json:"beforeSha256,omitempty"`
	AfterSha256  string `json:"afterSha256,omitempty"`
}

type SemanticDiff struct {
	Added   []SemanticChange `json:"added,omitempty"`
	Removed []SemanticChange `json:"removed,omitempty"`
	Changed []SemanticChange `json:"changed,omitempty"`
}

type SemanticChange struct {
	Key               string `json:"key"`
	SourcePayloadKind string `json:"sourcePayloadKind,omitempty"`
	SourceSubject     string `json:"sourceSubject,omitempty"`
	Kind              string `json:"kind"`
	Name              string `json:"name"`
	Before            string `json:"before,omitempty"`
	After             string `json:"after,omitempty"`
}

func Diff(fromDir, toDir string) (SnapshotDiff, error) {
	fromManifest, err := Verify(fromDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("verify from snapshot: %w", err)
	}
	toManifest, err := Verify(toDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("verify to snapshot: %w", err)
	}
	fromGraph, err := readGraph(fromDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("read from graph: %w", err)
	}
	toGraph, err := readGraph(toDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("read to graph: %w", err)
	}
	fromPageResults, err := readPageResults(fromDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("read from page results: %w", err)
	}
	toPageResults, err := readPageResults(toDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("read to page results: %w", err)
	}
	fromLifecycle, err := readLifecycleNotices(fromDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("read from lifecycle notices: %w", err)
	}
	toLifecycle, err := readLifecycleNotices(toDir)
	if err != nil {
		return SnapshotDiff{}, fmt.Errorf("read to lifecycle notices: %w", err)
	}

	diff := SnapshotDiff{
		SchemaVersion:      SchemaVersion,
		FromSnapshotID:     fromManifest.SnapshotID,
		ToSnapshotID:       toManifest.SnapshotID,
		PageCatalogChanged: pageCatalogHash(fromManifest.Source) != pageCatalogHash(toManifest.Source),
		CaptureChanged:     captureChanged(fromManifest.Capture, toManifest.Capture),
		Pages:              diffPages(fromGraph.Pages, toGraph.Pages),
		SourcePayloads:     diffSourcePayloads(fromGraph.SourcePayloads, toGraph.SourcePayloads),
		Records:            diffRecords(fromGraph.Records, toGraph.Records),
		PageResults:        diffPageResults(fromPageResults.Pages, toPageResults.Pages),
		LifecycleNotices:   diffLifecycleNotices(fromLifecycle.Notices, toLifecycle.Notices),
		Semantic:           diffSemanticRecords(fromGraph.Records, toGraph.Records),
	}
	diff.Changed = diff.PageCatalogChanged || diff.CaptureChanged || len(diff.Pages.Added) > 0 || len(diff.Pages.Removed) > 0 || len(diff.SourcePayloads.Added) > 0 || len(diff.SourcePayloads.Removed) > 0 || len(diff.SourcePayloads.Changed) > 0 || len(diff.Records.Added) > 0 || len(diff.Records.Removed) > 0 || len(diff.Records.Changed) > 0 || len(diff.PageResults.Added) > 0 || len(diff.PageResults.Removed) > 0 || len(diff.PageResults.Changed) > 0 || len(diff.LifecycleNotices.Added) > 0 || len(diff.LifecycleNotices.Removed) > 0
	return diff, nil
}

func pageCatalogHash(source SourceInfo) string {
	return source.PageCatalogSha256
}

func captureChanged(from, to CaptureInfo) bool {
	return from.PagesAttempted != to.PagesAttempted ||
		from.PagesSucceeded != to.PagesSucceeded ||
		from.PayloadsObserved != to.PayloadsObserved ||
		from.PayloadsFetched != to.PayloadsFetched ||
		from.Complete != to.Complete
}

func readGraph(directory string) (Graph, error) {
	file, err := os.Open(filepath.Join(directory, "generated-token-graph.json.gz"))
	if err != nil {
		return Graph{}, err
	}
	decompressor, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return Graph{}, err
	}
	var graph Graph
	err = json.NewDecoder(decompressor).Decode(&graph)
	closeErr := decompressor.Close()
	fileCloseErr := file.Close()
	if err != nil {
		return Graph{}, err
	}
	if closeErr != nil {
		return Graph{}, closeErr
	}
	if fileCloseErr != nil {
		return Graph{}, fileCloseErr
	}
	if graph.SchemaVersion != 2 && graph.SchemaVersion != SchemaVersion {
		return Graph{}, fmt.Errorf("unsupported graph schema version %d", graph.SchemaVersion)
	}
	return graph, nil
}

func readPageResults(directory string) (PageResults, error) {
	body, err := os.ReadFile(filepath.Join(directory, "page-results.json"))
	if err != nil {
		return PageResults{}, err
	}
	var pageResults PageResults
	if err := json.Unmarshal(body, &pageResults); err != nil {
		return PageResults{}, err
	}
	if pageResults.SchemaVersion != 2 && pageResults.SchemaVersion != SchemaVersion {
		return PageResults{}, fmt.Errorf("unsupported page-results schema version %d", pageResults.SchemaVersion)
	}
	return pageResults, nil
}

func readLifecycleNotices(directory string) (LifecycleNotices, error) {
	body, err := os.ReadFile(filepath.Join(directory, "lifecycle-notices.json"))
	if errors.Is(err, os.ErrNotExist) {
		return LifecycleNotices{SchemaVersion: 1, Notices: []LifecycleNotice{}}, nil
	}
	if err != nil {
		return LifecycleNotices{}, err
	}
	var notices LifecycleNotices
	if err := json.Unmarshal(body, &notices); err != nil {
		return LifecycleNotices{}, err
	}
	if notices.SchemaVersion != 1 {
		return LifecycleNotices{}, fmt.Errorf("unsupported lifecycle-notices schema version %d", notices.SchemaVersion)
	}
	return notices, nil
}

func diffLifecycleNotices(from, to []LifecycleNotice) LifecycleNoticeDiff {
	fromByID := make(map[string]LifecycleNotice, len(from))
	toByID := make(map[string]LifecycleNotice, len(to))
	for _, notice := range from {
		fromByID[notice.ID] = notice
	}
	for _, notice := range to {
		toByID[notice.ID] = notice
	}
	diff := LifecycleNoticeDiff{}
	for id, notice := range toByID {
		if _, exists := fromByID[id]; !exists {
			diff.Added = append(diff.Added, notice)
		}
	}
	for id, notice := range fromByID {
		if _, exists := toByID[id]; !exists {
			diff.Removed = append(diff.Removed, notice)
		}
	}
	sort.Slice(diff.Added, func(i, j int) bool { return diff.Added[i].ID < diff.Added[j].ID })
	sort.Slice(diff.Removed, func(i, j int) bool { return diff.Removed[i].ID < diff.Removed[j].ID })
	return diff
}

func diffPages(from, to []string) PageSetDiff {
	fromSet := stringSet(from)
	toSet := stringSet(to)
	diff := PageSetDiff{}
	for page := range toSet {
		if !fromSet[page] {
			diff.Added = append(diff.Added, page)
		}
	}
	for page := range fromSet {
		if !toSet[page] {
			diff.Removed = append(diff.Removed, page)
		}
	}
	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	return diff
}

func diffSourcePayloads(from, to []GraphPayload) SourcePayloadDiff {
	fromByKey := payloadMap(from)
	toByKey := payloadMap(to)
	diff := SourcePayloadDiff{}
	for key, payload := range toByKey {
		before, exists := fromByKey[key]
		if !exists {
			diff.Added = append(diff.Added, payloadChange(payload, GraphPayload{}))
		} else if before.Hash != payload.Hash || before.Error != payload.Error || before.Bytes != payload.Bytes || before.SourcePayloadKind != payload.SourcePayloadKind || before.SourceArea != payload.SourceArea || before.SourceSubject != payload.SourceSubject || before.SourcePageType != payload.SourcePageType || !sameStrings(before.PageURLs, payload.PageURLs) {
			diff.Changed = append(diff.Changed, payloadChange(payload, before))
		}
	}
	for key, payload := range fromByKey {
		if _, exists := toByKey[key]; !exists {
			diff.Removed = append(diff.Removed, payloadChange(GraphPayload{}, payload))
		}
	}
	sortSourcePayloadChanges(diff.Added)
	sortSourcePayloadChanges(diff.Removed)
	sortSourcePayloadChanges(diff.Changed)
	return diff
}

func diffPageResults(from, to []PageResult) PageResultsDiff {
	fromByURL := pageResultsMap(from)
	toByURL := pageResultsMap(to)
	diff := PageResultsDiff{}
	for url, after := range toByURL {
		before, exists := fromByURL[url]
		if !exists {
			diff.Added = append(diff.Added, PageResultChange{URL: url, After: after})
		} else if before != after {
			diff.Changed = append(diff.Changed, PageResultChange{URL: url, Before: before, After: after})
		}
	}
	for url, before := range fromByURL {
		if _, exists := toByURL[url]; !exists {
			diff.Removed = append(diff.Removed, PageResultChange{URL: url, Before: before})
		}
	}
	sort.Slice(diff.Added, func(i, j int) bool { return diff.Added[i].URL < diff.Added[j].URL })
	sort.Slice(diff.Removed, func(i, j int) bool { return diff.Removed[i].URL < diff.Removed[j].URL })
	sort.Slice(diff.Changed, func(i, j int) bool { return diff.Changed[i].URL < diff.Changed[j].URL })
	return diff
}

func diffRecords(from, to []ExtractedRecord) RecordSetDiff {
	fromByKey := recordMap(from)
	toByKey := recordMap(to)
	diff := RecordSetDiff{}
	for key, after := range toByKey {
		before, exists := fromByKey[key]
		if !exists {
			diff.Added = append(diff.Added, recordChange(after, ExtractedRecord{}))
		} else if sha256Hex(before.Raw) != sha256Hex(after.Raw) {
			diff.Changed = append(diff.Changed, recordChange(after, before))
		}
	}
	for key, before := range fromByKey {
		if _, exists := toByKey[key]; !exists {
			diff.Removed = append(diff.Removed, recordChange(ExtractedRecord{}, before))
		}
	}
	sortRecordChanges(diff.Added)
	sortRecordChanges(diff.Removed)
	sortRecordChanges(diff.Changed)
	return diff
}

func diffSemanticRecords(from, to []ExtractedRecord) SemanticDiff {
	return CompareSemanticIndexes(IndexSemanticRecords(from), IndexSemanticRecords(to))
}

// CompareSemanticIndexes compares token meaning without building the diagnostic
// payload, page, and raw-record diffs. Callers must verify snapshots first.
func CompareSemanticIndexes(fromByKey, toByKey map[string]SemanticRecord) SemanticDiff {
	diff := SemanticDiff{}
	for key, after := range toByKey {
		before, exists := fromByKey[key]
		if !exists {
			diff.Added = append(diff.Added, semanticChange(after, SemanticRecord{}))
		} else if before.Value != after.Value {
			diff.Changed = append(diff.Changed, semanticChange(after, before))
		}
	}
	for key, before := range fromByKey {
		if _, exists := toByKey[key]; !exists {
			diff.Removed = append(diff.Removed, semanticChange(SemanticRecord{}, before))
		}
	}
	sortSemanticChanges(diff.Added)
	sortSemanticChanges(diff.Removed)
	sortSemanticChanges(diff.Changed)
	return diff
}

type SemanticRecord struct {
	Key               string
	SourcePayloadKind string
	SourceSubject     string
	Kind              string
	Name              string
	Value             string
}

// IndexSemanticRecords retains only the token meaning needed by a timeline.
func IndexSemanticRecords(records []ExtractedRecord) map[string]SemanticRecord {
	result := make(map[string]SemanticRecord)
	contexts := ContextDescriptions(records)
	for _, record := range records {
		if record.Kind == "contextualReferenceTree" {
			for _, semantic := range contextualSemanticRecords(record, contexts) {
				if current, exists := result[semantic.Key]; !exists || semanticRecordPreferred(semantic, current) {
					result[semantic.Key] = semantic
				}
			}
			continue
		}
		semantic, ok := semanticRecord(record)
		if ok {
			if current, exists := result[semantic.Key]; !exists || semanticRecordPreferred(semantic, current) {
				result[semantic.Key] = semantic
			}
		}
	}
	return result
}

func semanticRecord(record ExtractedRecord) (SemanticRecord, bool) {
	semantic := SemanticRecord{
		SourcePayloadKind: record.SourcePayloadKind,
		SourceSubject:     record.SourceSubject,
		Kind:              record.Kind,
	}
	switch record.Kind {
	case "contextualReferenceTree":
		return SemanticRecord{}, false // Contextual records are expanded by IndexSemanticRecords.
	case "component":
		if record.SourcePayloadKind != "component" {
			return SemanticRecord{}, false
		}
		if record.DisplayName == "" {
			return SemanticRecord{}, false
		}
		semantic.Name = record.DisplayName
		semantic.Value = selectedFields(record.Raw, "displayName", "state", "definition")
	case "tokenSet":
		name := record.TokenSetName
		if name == "" {
			name = record.DisplayName
		}
		if name == "" {
			return SemanticRecord{}, false
		}
		semantic.Name = name
		semantic.Value = selectedFields(record.Raw, "displayName", "state", "description")
	default:
		return SemanticRecord{}, false
	}
	semantic.SourceSubject = semanticSubject(semantic.Name, record.SourceSubject)
	semantic.Key = fmt.Sprintf("%s\x00%s", semantic.Kind, semantic.Name)
	return semantic, true
}

func semanticRecordPreferred(candidate, current SemanticRecord) bool {
	if candidate.SourcePayloadKind == current.SourcePayloadKind {
		return candidate.SourceSubject < current.SourceSubject
	}
	return candidate.SourcePayloadKind == "component"
}

func semanticSubject(name, fallback string) string {
	parts := strings.Split(name, ".")
	if len(parts) >= 3 {
		switch parts[1] {
		case "comp":
			return parts[2]
		case "ref":
			return "reference tokens"
		case "sys":
			return "system tokens"
		}
	}
	return fallback
}

func selectedFields(raw json.RawMessage, names ...string) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return ""
	}
	selected := make(map[string]json.RawMessage, len(names))
	for _, name := range names {
		if value, ok := fields[name]; ok {
			selected[name] = value
		}
	}
	body, err := json.Marshal(selected)
	if err != nil {
		return ""
	}
	return string(body)
}

func compactJSON(raw json.RawMessage) string {
	var compacted bytes.Buffer
	if json.Compact(&compacted, raw) != nil {
		return string(raw)
	}
	return compacted.String()
}

func semanticChange(after, before SemanticRecord) SemanticChange {
	value := after
	if value.Name == "" {
		value = before
	}
	return SemanticChange{
		Key:               value.Key,
		SourcePayloadKind: value.SourcePayloadKind,
		SourceSubject:     value.SourceSubject,
		Kind:              value.Kind,
		Name:              value.Name,
		Before:            before.Value,
		After:             after.Value,
	}
}

func sortSemanticChanges(values []SemanticChange) {
	sort.Slice(values, func(i, j int) bool {
		if values[i].SourceSubject != values[j].SourceSubject {
			return values[i].SourceSubject < values[j].SourceSubject
		}
		if values[i].Kind != values[j].Kind {
			return values[i].Kind < values[j].Kind
		}
		if values[i].Name != values[j].Name {
			return values[i].Name < values[j].Name
		}
		return values[i].Key < values[j].Key
	})
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func payloadMap(values []GraphPayload) map[string]GraphPayload {
	result := make(map[string]GraphPayload, len(values))
	for _, value := range values {
		result[value.SourcePayloadURL] = value
	}
	return result
}

func pageResultsMap(values []PageResult) map[string]PageResult {
	result := make(map[string]PageResult, len(values))
	for _, value := range values {
		result[value.URL] = value
	}
	return result
}

func recordMap(values []ExtractedRecord) map[string]ExtractedRecord {
	result := make(map[string]ExtractedRecord, len(values))
	for _, value := range values {
		result[recordKey(value)] = value
	}
	return result
}

func recordKey(value ExtractedRecord) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%04d", value.SourcePayloadID, value.Kind, value.ID, value.Ordinal)
}

func payloadChange(after, before GraphPayload) SourcePayloadChange {
	pageURLs, sourcePayloadURL := after.PageURLs, after.SourcePayloadURL
	if sourcePayloadURL == "" {
		pageURLs, sourcePayloadURL = before.PageURLs, before.SourcePayloadURL
	}
	return SourcePayloadChange{
		Key:              sourcePayloadURL,
		PageURLs:         pageURLs,
		SourcePayloadURL: sourcePayloadURL,
		BeforePageURLs:   before.PageURLs,
		AfterPageURLs:    after.PageURLs,
		BeforeKind:       before.SourcePayloadKind,
		AfterKind:        after.SourcePayloadKind,
		BeforeHash:       before.Hash,
		AfterHash:        after.Hash,
		BeforeError:      before.Error,
		AfterError:       after.Error,
		BeforeBytes:      before.Bytes,
		AfterBytes:       after.Bytes,
	}
}

func sortSourcePayloadChanges(values []SourcePayloadChange) {
	sort.Slice(values, func(i, j int) bool { return values[i].Key < values[j].Key })
}

func recordChange(after, before ExtractedRecord) RecordChange {
	value := after
	if value.ID == "" {
		value = before
	}
	return RecordChange{
		Key:          recordKey(value),
		Kind:         value.Kind,
		TokenName:    value.TokenName,
		BeforeSha256: sha256Hex(before.Raw),
		AfterSha256:  sha256Hex(after.Raw),
	}
}

func sortRecordChanges(values []RecordChange) {
	sort.Slice(values, func(i, j int) bool { return values[i].Key < values[j].Key })
}
