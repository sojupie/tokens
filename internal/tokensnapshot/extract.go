package tokensnapshot

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExtractedRecord keeps the upstream object intact while exposing the fields
// needed to index the graph without knowing every Material object shape.
type ExtractedRecord struct {
	ID              string          `json:"id"`
	Kind            string          `json:"kind"`
	SourcePayloadID string          `json:"sourcePayloadId"`
	Ordinal         int             `json:"ordinal"`
	Name            string          `json:"name,omitempty"`
	TokenName       string          `json:"tokenName,omitempty"`
	DisplayName     string          `json:"displayName,omitempty"`
	TokenSetName    string          `json:"tokenSetName,omitempty"`
	State           string          `json:"state,omitempty"`
	ValueType       string          `json:"valueType,omitempty"`
	Raw             json.RawMessage `json:"raw"`
}

type m3Payload struct {
	Artifact string
	Hash     string
	Error    string
}

func extractPayload(cacheDir string, payload m3Payload, sourcePayloadID string) ([]ExtractedRecord, error) {
	if cacheDir == "" || payload.Artifact == "" || payload.Hash == "" || payload.Error != "" {
		return nil, nil
	}
	body, err := readGzip(filepath.Join(cacheDir, filepath.FromSlash(payload.Artifact)))
	if err != nil {
		return nil, err
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(body, &document); err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	systemRaw, ok := document["system"]
	if !ok {
		return nil, fmt.Errorf("payload has no system object")
	}
	var system map[string]json.RawMessage
	if err := json.Unmarshal(systemRaw, &system); err != nil {
		return nil, fmt.Errorf("decode system object: %w", err)
	}
	records := []ExtractedRecord{newExtractedRecord("system", sourcePayloadID, "system", 0, systemRaw)}
	fields := []struct {
		name string
		kind string
	}{
		{name: "parentVersions", kind: "parentVersion"},
		{name: "tokenSets", kind: "tokenSet"},
		{name: "versions", kind: "version"},
		{name: "components", kind: "component"},
		{name: "componentDisplayGroups", kind: "componentDisplayGroup"},
		{name: "connections", kind: "connection"},
		{name: "contextTagGroups", kind: "contextTagGroup"},
		{name: "tags", kind: "tag"},
		{name: "tokens", kind: "token"},
		{name: "displayGroups", kind: "displayGroup"},
		{name: "values", kind: "value"},
	}
	for _, field := range fields {
		items, ok := system[field.name]
		if !ok {
			continue
		}
		var values []json.RawMessage
		if err := json.Unmarshal(items, &values); err != nil {
			return nil, fmt.Errorf("decode system.%s: %w", field.name, err)
		}
		for index, value := range values {
			records = append(records, newExtractedRecord(field.kind, sourcePayloadID, recordID(value, index), index, value))
		}
	}
	if trees, ok := system["contextualReferenceTrees"]; ok {
		var values map[string]json.RawMessage
		if err := json.Unmarshal(trees, &values); err != nil {
			return nil, fmt.Errorf("decode system.contextualReferenceTrees: %w", err)
		}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for index, key := range keys {
			records = append(records, newExtractedRecord("contextualReferenceTree", sourcePayloadID, key, index, values[key]))
		}
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Kind != records[j].Kind {
			return records[i].Kind < records[j].Kind
		}
		if records[i].TokenName != records[j].TokenName {
			return records[i].TokenName < records[j].TokenName
		}
		if records[i].ID != records[j].ID {
			return records[i].ID < records[j].ID
		}
		if records[i].Ordinal != records[j].Ordinal {
			return records[i].Ordinal < records[j].Ordinal
		}
		return string(records[i].Raw) < string(records[j].Raw)
	})
	return records, nil
}

func newExtractedRecord(kind, sourcePayloadID, id string, ordinal int, raw json.RawMessage) ExtractedRecord {
	record := ExtractedRecord{ID: id, Kind: kind, SourcePayloadID: sourcePayloadID, Ordinal: ordinal, Raw: append(json.RawMessage(nil), raw...)}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return record
	}
	record.Name = jsonString(fields["name"])
	record.TokenName = jsonString(fields["tokenName"])
	record.DisplayName = jsonString(fields["displayName"])
	record.TokenSetName = jsonString(fields["tokenSetName"])
	record.State = jsonString(fields["state"])
	record.ValueType = jsonString(fields["tokenValueType"])
	if record.ID == "" {
		record.ID = record.Name
	}
	return record
}

func lifecycleNotices(records []ExtractedRecord) []LifecycleNotice {
	notices := make([]LifecycleNotice, 0)
	for _, record := range records {
		var fields map[string]json.RawMessage
		if json.Unmarshal(record.Raw, &fields) != nil {
			continue
		}
		if message := deprecationMessage(fields["deprecationMessage"]); message != "" {
			notices = append(notices, newLifecycleNotice(record, "deprecationMessage", "deprecated", message))
		}
		for _, field := range []string{"description", "displayName", "name", "tokenName", "tokenSetName"} {
			message := jsonString(fields[field])
			state := lifecycleState(message)
			if state == "" {
				continue
			}
			notices = append(notices, newLifecycleNotice(record, field, state, message))
		}
	}
	sort.Slice(notices, func(i, j int) bool { return notices[i].ID < notices[j].ID })
	return notices
}

func deprecationMessage(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var object struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &object) == nil && object.Message != "" {
		return object.Message
	}
	return jsonString(raw)
}

func lifecycleState(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "not recommended") || strings.Contains(lower, "no longer recommended") {
		return "not-recommended"
	}
	if strings.Contains(lower, "deprecated") {
		return "deprecated"
	}
	return ""
}

func newLifecycleNotice(record ExtractedRecord, sourceField, state, message string) LifecycleNotice {
	subject := record.TokenName
	if subject == "" {
		subject = record.TokenSetName
	}
	if subject == "" {
		subject = record.Name
	}
	if subject == "" {
		subject = record.ID
	}
	return LifecycleNotice{
		ID:              "notice-" + sha256Hex([]byte(record.SourcePayloadID+"\x00"+record.Kind+"\x00"+record.ID+"\x00"+sourceField+"\x00"+message)),
		Scope:           lifecycleScope(record.Kind),
		State:           state,
		Subject:         subject,
		Message:         message,
		SourcePayloadID: record.SourcePayloadID,
		SourceRecordID:  record.ID,
		RecordKind:      record.Kind,
		SourceField:     sourceField,
	}
}

func lifecycleScope(kind string) string {
	switch kind {
	case "component":
		return "component"
	case "tokenSet":
		return "style"
	case "token":
		return "token"
	default:
		return "record"
	}
}

func recordID(raw json.RawMessage, index int) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) == nil {
		for _, key := range []string{"name", "tokenName", "id"} {
			if value := jsonString(fields[key]); value != "" {
				return value
			}
		}
	}
	return fmt.Sprintf("item-%04d", index+1)
}

func jsonString(raw json.RawMessage) string {
	var value string
	if len(raw) > 0 && json.Unmarshal(raw, &value) == nil {
		return value
	}
	return ""
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
