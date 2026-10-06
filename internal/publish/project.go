package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/sojupie/tokens/internal/tokensnapshot"
)

type subjectBuilder struct {
	view     SubjectView
	tokens   map[string]*TokenView
	notices  map[string]bool
	pageURLs map[string]bool
}

func project(input snapshotInput) (Publication, error) {
	payloads, payloadSubjects := projectPayloads(input.PayloadDetails)
	builders := make(map[SubjectKey]*subjectBuilder)
	owners := tokenSetOwners(input.Graph.Records, payloadSubjects)
	contexts := contextDescriptions(input.Graph.Records)

	for _, record := range sortedRecords(input.Graph.Records) {
		if record.Kind == "component" && record.SourcePayloadKind == "component" {
			key, ok := payloadSubjects[record.SourcePayloadID]
			if !ok {
				continue
			}
			builder := getSubjectBuilder(builders, key)
			if builder.view.Name == "" {
				builder.view.Name = valueOr(record.DisplayName, string(key.Kind)+" "+key.Slug)
			}
			addPayloadPages(builder, payloads[record.SourcePayloadID])
			continue
		}

		if record.Kind == "tokenSet" {
			for _, key := range recordSubjects(record, owners, payloadSubjects) {
				builder := getSubjectBuilder(builders, key)
				if builder.view.Name == "" && record.DisplayName != "" {
					builder.view.Name = record.DisplayName
				}
				addPayloadPages(builder, payloads[record.SourcePayloadID])
			}
			continue
		}

		if record.Kind == "token" {
			if record.TokenName == "" {
				continue
			}
			for _, key := range recordSubjects(record, owners, payloadSubjects) {
				builder := getSubjectBuilder(builders, key)
				addPayloadPages(builder, payloads[record.SourcePayloadID])
				if err := builder.addToken(record.TokenName, record.ValueType, nil, nil, nil, record.Raw, record, contexts); err != nil {
					return Publication{}, err
				}
			}
			continue
		}

		if record.Kind == "contextualReferenceTree" {
			entries, err := contextualEntries(record.Raw)
			if err != nil {
				return Publication{}, fmt.Errorf("decode contextual reference tree %s: %w", record.ID, err)
			}
			for _, entry := range entries {
				if entry.TokenName == "" {
					continue
				}
				for _, key := range recordSubjects(record, owners, payloadSubjects) {
					builder := getSubjectBuilder(builders, key)
					addPayloadPages(builder, payloads[record.SourcePayloadID])
					if err := builder.addToken(entry.TokenName, "", entry.ResolvedValue, entry.ContextTags, entry.References, entry.Raw, record, contexts); err != nil {
						return Publication{}, err
					}
				}
			}
		}
	}

	addLifecycleNotices(builders, input.LifecycleNotices.Notices, payloadSubjects)

	publication := Publication{
		SchemaVersion:   SchemaVersion,
		CurrentSnapshot: snapshotView(input.Manifest, input.ManifestHash),
		Snapshots:       []SnapshotView{snapshotView(input.Manifest, input.ManifestHash)},
		Subjects:        make([]SubjectView, 0, len(builders)),
		SourcePayloads:  make([]SourcePayloadView, 0, len(payloads)),
	}
	keys := make([]SubjectKey, 0, len(builders))
	for key := range builders {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, key := range keys {
		builder := builders[key]
		builder.finish()
		publication.Subjects = append(publication.Subjects, builder.view)
	}
	payloadIDs := make([]string, 0, len(payloads))
	for id := range payloads {
		payloadIDs = append(payloadIDs, id)
	}
	sort.Strings(payloadIDs)
	for _, id := range payloadIDs {
		publication.SourcePayloads = append(publication.SourcePayloads, payloads[id])
	}
	return publication, nil
}

func projectPayloads(details tokensnapshot.SourcePayloadDetails) (map[string]SourcePayloadView, map[string]SubjectKey) {
	payloads := make(map[string]SourcePayloadView, len(details.Records))
	subjects := make(map[string]SubjectKey)
	for _, detail := range details.Records {
		payloads[detail.ID] = SourcePayloadView{
			ID:            detail.ID,
			URL:           detail.SourcePayloadURL,
			Hash:          detail.PayloadSha256,
			Kind:          detail.SourcePayloadKind,
			Area:          detail.SourceArea,
			SourceSubject: detail.SourceSubject,
			PageType:      detail.SourcePageType,
			PageURLs:      append([]string(nil), detail.PageURLs...),
		}
		if key, ok := payloadSubjectKey(detail.SourcePayloadKind, detail.SourceArea, detail.SourceSubject); ok {
			subjects[detail.ID] = key
		}
	}
	return payloads, subjects
}

func tokenSetOwners(records []tokensnapshot.ExtractedRecord, payloadSubjects map[string]SubjectKey) map[string][]SubjectKey {
	owners := make(map[string][]SubjectKey)
	for _, record := range records {
		if record.Kind != "component" || record.SourcePayloadKind != "component" {
			continue
		}
		key, ok := payloadSubjects[record.SourcePayloadID]
		if !ok {
			continue
		}
		var fields struct {
			TokenSets []string `json:"tokenSets"`
		}
		if json.Unmarshal(record.Raw, &fields) != nil {
			continue
		}
		for _, tokenSetID := range fields.TokenSets {
			if !containsSubject(owners[tokenSetID], key) {
				owners[tokenSetID] = append(owners[tokenSetID], key)
			}
		}
	}
	for tokenSetID := range owners {
		sort.Slice(owners[tokenSetID], func(i, j int) bool { return owners[tokenSetID][i].String() < owners[tokenSetID][j].String() })
	}
	return owners
}

func recordSubjects(record tokensnapshot.ExtractedRecord, owners map[string][]SubjectKey, payloadSubjects map[string]SubjectKey) []SubjectKey {
	tokenSetID := tokenSetID(record.ID)
	if owned := owners[tokenSetID]; len(owned) > 0 {
		return owned
	}
	if record.SourcePayloadKind == "component" {
		return nil
	}
	if key, ok := payloadSubjects[record.SourcePayloadID]; ok {
		return []SubjectKey{key}
	}
	return nil
}

func payloadSubjectKey(kind, area, subject string) (SubjectKey, bool) {
	if subject == "" {
		return SubjectKey{}, false
	}
	if kind == "component" || area == "components" {
		return SubjectKey{Kind: SubjectComponent, Slug: subject}, true
	}
	if area == "foundation" || area == "foundations" {
		return SubjectKey{Kind: SubjectFoundation, Slug: subject}, true
	}
	return SubjectKey{Kind: SubjectStyle, Slug: subject}, true
}

func getSubjectBuilder(builders map[SubjectKey]*subjectBuilder, key SubjectKey) *subjectBuilder {
	if builder, ok := builders[key]; ok {
		return builder
	}
	builder := &subjectBuilder{
		view:     SubjectView{Key: key, Tokens: []TokenView{}, LifecycleNotices: []LifecycleNoticeView{}},
		tokens:   make(map[string]*TokenView),
		notices:  make(map[string]bool),
		pageURLs: make(map[string]bool),
	}
	builders[key] = builder
	return builder
}

func (builder *subjectBuilder) addToken(name, category string, value json.RawMessage, contextTags []string, references []TokenReference, raw json.RawMessage, record tokensnapshot.ExtractedRecord, contexts map[string]resolvedContext) error {
	token, ok := builder.tokens[name]
	if !ok {
		token = &TokenView{SourceName: name, Values: []TokenValueView{}, Sources: []TokenSourceView{}}
		builder.tokens[name] = token
	}
	if category != "" && token.Category == "" {
		token.Category = category
	}
	if record.Kind == "token" && record.DisplayName != "" {
		token.DisplayName = record.DisplayName
	}
	if len(value) > 0 {
		tags := append([]string(nil), contextTags...)
		sort.Strings(tags)
		labels, isDefault := resolveContext(tags, contexts)
		valueView := findTokenValue(token.Values, tags)
		if valueView == nil {
			token.Values = append(token.Values, TokenValueView{
				ContextTags:      tags,
				ContextLabels:    labels,
				IsDefaultContext: isDefault,
				Value:            append(json.RawMessage(nil), value...),
				References:       references,
				Raw:              append(json.RawMessage(nil), raw...),
				Sources:          []TokenSourceView{},
			})
		} else if !jsonEqual(valueView.Value, value) {
			return fmt.Errorf("conflicting values for published token %s in context %s", name, strings.Join(tags, ","))
		} else {
			existing, err := json.Marshal(valueView.References)
			if err != nil {
				return err
			}
			incoming, err := json.Marshal(references)
			if err != nil {
				return err
			}
			if !jsonEqual(existing, incoming) {
				return fmt.Errorf("conflicting references for published token %s in context %s", name, strings.Join(tags, ","))
			}
		}
		source := TokenSourceView{PayloadID: record.SourcePayloadID, SourceRecordID: record.ID}
		valueView = findTokenValue(token.Values, tags)
		if !containsSource(valueView.Sources, source) {
			valueView.Sources = append(valueView.Sources, source)
		}
	}
	if len(token.Raw) == 0 && record.Kind == "token" {
		token.Raw = append(json.RawMessage(nil), raw...)
	}
	if record.State != "" && token.Lifecycle == "" {
		token.Lifecycle = record.State
	}
	source := TokenSourceView{PayloadID: record.SourcePayloadID, SourceRecordID: record.ID}
	for _, existing := range token.Sources {
		if existing == source {
			return nil
		}
	}
	token.Sources = append(token.Sources, source)
	return nil
}

func findTokenValue(values []TokenValueView, contextTags []string) *TokenValueView {
	for index := range values {
		if strings.Join(values[index].ContextTags, "\x00") == strings.Join(contextTags, "\x00") {
			return &values[index]
		}
	}
	return nil
}

func containsSource(sources []TokenSourceView, wanted TokenSourceView) bool {
	for _, source := range sources {
		if source == wanted {
			return true
		}
	}
	return false
}

func (builder *subjectBuilder) finish() {
	for _, token := range builder.tokens {
		sort.Slice(token.Sources, func(i, j int) bool {
			if token.Sources[i].PayloadID != token.Sources[j].PayloadID {
				return token.Sources[i].PayloadID < token.Sources[j].PayloadID
			}
			return token.Sources[i].SourceRecordID < token.Sources[j].SourceRecordID
		})
		sort.Slice(token.Values, func(i, j int) bool {
			if token.Values[i].IsDefaultContext != token.Values[j].IsDefaultContext {
				return token.Values[i].IsDefaultContext
			}
			left := strings.Join(token.Values[i].ContextLabels, "\x00")
			right := strings.Join(token.Values[j].ContextLabels, "\x00")
			if left != right {
				return left < right
			}
			return strings.Join(token.Values[i].ContextTags, "\x00") < strings.Join(token.Values[j].ContextTags, "\x00")
		})
		builder.view.Tokens = append(builder.view.Tokens, *token)
	}
	sort.Slice(builder.view.Tokens, func(i, j int) bool { return builder.view.Tokens[i].SourceName < builder.view.Tokens[j].SourceName })
	for pageURL := range builder.pageURLs {
		builder.view.SourcePageURLs = append(builder.view.SourcePageURLs, pageURL)
	}
	sort.Strings(builder.view.SourcePageURLs)
	sort.Slice(builder.view.LifecycleNotices, func(i, j int) bool { return builder.view.LifecycleNotices[i].ID < builder.view.LifecycleNotices[j].ID })
}

func addPayloadPages(builder *subjectBuilder, payload SourcePayloadView) {
	for _, pageURL := range payload.PageURLs {
		builder.pageURLs[pageURL] = true
	}
}

func addLifecycleNotices(builders map[SubjectKey]*subjectBuilder, notices []tokensnapshot.LifecycleNotice, payloadSubjects map[string]SubjectKey) {
	for _, notice := range notices {
		key, ok := payloadSubjects[notice.SourcePayloadID]
		if !ok {
			continue
		}
		builder := getSubjectBuilder(builders, key)
		if builder.notices[notice.ID] {
			continue
		}
		builder.notices[notice.ID] = true
		builder.view.LifecycleNotices = append(builder.view.LifecycleNotices, LifecycleNoticeView{
			ID:              notice.ID,
			State:           notice.State,
			Subject:         notice.Subject,
			Message:         notice.Message,
			SourcePayloadID: notice.SourcePayloadID,
			SourceRecordID:  notice.SourceRecordID,
		})
	}
}

type contextualEntry struct {
	TokenName     string
	ContextTags   []string
	ResolvedValue json.RawMessage
	References    []TokenReference
	Raw           json.RawMessage
}

type sourceReference struct {
	TokenName  string            `json:"tokenName"`
	ChildNodes []sourceReference `json:"childNodes"`
}

func namedReferences(nodes []sourceReference) []TokenReference {
	var references []TokenReference
	for _, node := range nodes {
		children := namedReferences(node.ChildNodes)
		if node.TokenName == "" {
			references = append(references, children...)
		} else {
			references = append(references, TokenReference{Name: node.TokenName, References: children})
		}
	}
	return references
}

func contextualEntries(raw json.RawMessage) ([]contextualEntry, error) {
	var document struct {
		Entries []struct {
			ContextTags   []string        `json:"contextTags"`
			ReferenceTree sourceReference `json:"referenceTree"`
			ResolvedValue json.RawMessage `json:"resolvedValue"`
		} `json:"contextualReferenceTree"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	entries := make([]contextualEntry, 0, len(document.Entries))
	for _, entry := range document.Entries {
		entries = append(entries, contextualEntry{
			TokenName:     entry.ReferenceTree.TokenName,
			ContextTags:   append([]string(nil), entry.ContextTags...),
			ResolvedValue: append(json.RawMessage(nil), entry.ResolvedValue...),
			References:    namedReferences(entry.ReferenceTree.ChildNodes),
			Raw:           append(json.RawMessage(nil), entry.ResolvedValue...),
		})
	}
	return entries, nil
}

func snapshotView(manifest tokensnapshot.Manifest, manifestHash string) SnapshotView {
	captureStatus := "partial"
	if manifest.Capture.Complete {
		captureStatus = "complete"
	}
	return SnapshotView{
		ID:            manifest.SnapshotID,
		CreatedAt:     manifest.CreatedAt,
		ManifestHash:  manifestHash,
		CaptureStatus: captureStatus,
		ReviewStatus:  manifest.ReviewStatus,
	}
}

func sortedRecords(records []tokensnapshot.ExtractedRecord) []tokensnapshot.ExtractedRecord {
	result := append([]tokensnapshot.ExtractedRecord(nil), records...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].SourcePayloadID != result[j].SourcePayloadID {
			return result[i].SourcePayloadID < result[j].SourcePayloadID
		}
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func tokenSetID(recordID string) string {
	parts := strings.Split(recordID, "/")
	for index, part := range parts {
		if part == "tokenSets" && index+1 < len(parts) {
			return strings.Join(parts[:index+2], "/")
		}
	}
	return ""
}

func containsSubject(subjects []SubjectKey, wanted SubjectKey) bool {
	for _, subject := range subjects {
		if subject == wanted {
			return true
		}
	}
	return false
}

func jsonEqual(left, right json.RawMessage) bool {
	var leftValue, rightValue any
	leftDecoder := json.NewDecoder(bytes.NewReader(left))
	rightDecoder := json.NewDecoder(bytes.NewReader(right))
	if leftDecoder.Decode(&leftValue) != nil || rightDecoder.Decode(&rightValue) != nil {
		return bytes.Equal(left, right)
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
