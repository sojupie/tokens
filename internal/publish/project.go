package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
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

var (
	reParen = regexp.MustCompile(`\([^)]*\)`)
	reWord  = regexp.MustCompile(`[a-zA-Z0-9]+`)
)

func cleanWords(s string) []string {
	s = reParen.ReplaceAllString(s, "")
	return reWord.FindAllString(strings.ToLower(s), -1)
}

func extractComponentSlug(pageURL string) string {
	parts := strings.Split(strings.Trim(pageURL, "/"), "/")
	for i, part := range parts {
		if part == "components" && i+1 < len(parts) {
			slug := parts[i+1]
			if idx := strings.IndexAny(slug, "?#"); idx >= 0 {
				slug = slug[:idx]
			}
			return slug
		}
	}
	return ""
}

func knownComponentSlugs(details []tokensnapshot.SourcePayloadDetail) []string {
	seen := make(map[string]bool)
	var slugs []string
	for _, detail := range details {
		for _, u := range detail.PageURLs {
			if slug := extractComponentSlug(u); slug != "" && !seen[slug] {
				seen[slug] = true
				slugs = append(slugs, slug)
			}
		}
	}
	sort.Strings(slugs)
	return slugs
}

func stemWord(w string) string {
	w = strings.ToLower(w)
	return strings.TrimSuffix(w, "s")
}

func wordsMatch(words, slugWords []string) bool {
	stemmedSlug := make(map[string]bool, len(slugWords))
	for _, sw := range slugWords {
		stemmedSlug[stemWord(sw)] = true
	}
	for _, w := range words {
		if stemmedSlug[stemWord(w)] {
			return true
		}
	}
	return false
}

func wordMatchScore(words, slugWords []string) int {
	stemmedSlug := make(map[string]bool, len(slugWords))
	for _, sw := range slugWords {
		stemmedSlug[stemWord(sw)] = true
	}
	score := 0
	for _, w := range words {
		if stemmedSlug[stemWord(w)] {
			score++
		}
	}
	return score
}

func componentSubjectKey(record tokensnapshot.ExtractedRecord, payload SourcePayloadView, fallback SubjectKey, knownSlugs []string) SubjectKey {
	words := cleanWords(record.DisplayName)
	if len(words) == 0 {
		if fallback.Kind == SubjectComponent && fallback.Slug != "" {
			return fallback
		}
		return SubjectKey{Kind: SubjectComponent, Slug: record.DisplayName}
	}

	for _, pageURL := range payload.PageURLs {
		if slug := extractComponentSlug(pageURL); slug != "" {
			if wordsMatch(words, cleanWords(slug)) {
				return SubjectKey{Kind: SubjectComponent, Slug: slug}
			}
		}
	}

	bestSlug := ""
	bestScore := 0
	for _, slug := range knownSlugs {
		score := wordMatchScore(words, cleanWords(slug))
		if score > bestScore {
			bestScore = score
			bestSlug = slug
		}
	}
	if bestScore > 0 && bestSlug != "" {
		return SubjectKey{Kind: SubjectComponent, Slug: bestSlug}
	}

	if fallback.Kind == SubjectComponent && fallback.Slug != "" {
		return fallback
	}
	return SubjectKey{Kind: SubjectComponent, Slug: strings.Join(words, "-")}
}

func styleTokenSetSubject(tokenSetName, displayName string) (SubjectKey, bool) {
	lowerName := strings.ToLower(tokenSetName)
	lowerDisplay := strings.ToLower(displayName)
	switch {
	case strings.Contains(lowerName, "md.sys.shape") || strings.Contains(lowerDisplay, "shape"):
		return SubjectKey{Kind: SubjectStyle, Slug: "shape/corner-radius"}, true
	case strings.Contains(lowerName, "md.sys.measurement") || strings.Contains(lowerName, "focus-indicator") || strings.Contains(lowerDisplay, "spacing") || strings.Contains(lowerDisplay, "focus indicator"):
		return SubjectKey{Kind: SubjectStyle, Slug: "spacing"}, true
	case strings.Contains(lowerName, "md.sys.motion") || strings.Contains(lowerName, "md.sys.state") || strings.Contains(lowerDisplay, "motion") || strings.Contains(lowerDisplay, "state opacities"):
		return SubjectKey{Kind: SubjectStyle, Slug: "motion/overview"}, true
	case strings.Contains(lowerName, "md.sys.elevation") || strings.Contains(lowerDisplay, "elevation"):
		return SubjectKey{Kind: SubjectStyle, Slug: "elevation"}, true
	case strings.Contains(lowerName, "typescale") || strings.Contains(lowerName, "typeface") || strings.Contains(lowerDisplay, "type scale") || strings.Contains(lowerDisplay, "fonts & weights"):
		return SubjectKey{Kind: SubjectStyle, Slug: "typography/type-scale"}, true
	}
	return SubjectKey{}, false
}

func tokenNameStyleSubject(tokenName string) (SubjectKey, bool) {
	switch {
	case strings.HasPrefix(tokenName, "md.sys.shape."):
		return SubjectKey{Kind: SubjectStyle, Slug: "shape/corner-radius"}, true
	case strings.HasPrefix(tokenName, "md.sys.measurement.") || strings.HasPrefix(tokenName, "md.sys.state.focus-indicator"):
		return SubjectKey{Kind: SubjectStyle, Slug: "spacing"}, true
	case strings.HasPrefix(tokenName, "md.sys.motion.") || strings.HasPrefix(tokenName, "md.sys.state."):
		return SubjectKey{Kind: SubjectStyle, Slug: "motion/overview"}, true
	case strings.HasPrefix(tokenName, "md.sys.elevation."):
		return SubjectKey{Kind: SubjectStyle, Slug: "elevation"}, true
	case strings.HasPrefix(tokenName, "md.sys.typescale.") || strings.HasPrefix(tokenName, "md.ref.typeface."):
		return SubjectKey{Kind: SubjectStyle, Slug: "typography/type-scale"}, true
	}
	return SubjectKey{}, false
}

func styleSubjectName(slug string) string {
	switch slug {
	case "shape/corner-radius":
		return "Shape"
	case "motion/overview":
		return "Motion"
	case "spacing":
		return "Spacing"
	case "elevation":
		return "Elevation"
	case "typography/type-scale":
		return "Type scale"
	default:
		return formatSubjectSlug(slug)
	}
}

func formatSubjectSlug(slug string) string {
	parts := strings.Split(slug, "/")
	last := parts[len(parts)-1]
	words := strings.Split(last, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func pageMatchesSubject(pageURL string, key SubjectKey) bool {
	if key.Kind == SubjectComponent {
		if strings.Contains(pageURL, "/components/") {
			return strings.Contains(pageURL, "/components/"+key.Slug+"/") || strings.HasSuffix(pageURL, "/components/"+key.Slug)
		}
		return !strings.Contains(pageURL, "/styles/") && strings.Contains(pageURL, key.Slug)
	}
	if key.Kind == SubjectStyle {
		if strings.Contains(pageURL, "/styles/") || strings.Contains(pageURL, "/foundations/") {
			prefix := key.Slug
			if idx := strings.IndexByte(prefix, '/'); idx >= 0 {
				prefix = prefix[:idx]
			}
			return strings.Contains(pageURL, "/styles/"+prefix) || strings.Contains(pageURL, "/foundations/"+prefix)
		}
		return !strings.Contains(pageURL, "/components/") && strings.Contains(pageURL, key.Slug)
	}
	return strings.Contains(pageURL, key.Slug)
}

func project(input snapshotInput) (Publication, error) {
	payloads, payloadSubjects := projectPayloads(input.PayloadDetails)
	componentSlugs := knownComponentSlugs(input.PayloadDetails.Records)
	builders := make(map[SubjectKey]*subjectBuilder)
	owners := tokenSetOwners(input.Graph.Records, payloads, payloadSubjects, componentSlugs)
	styleTokenSets := indexStyleTokenSets(input.Graph.Records)
	contexts := contextDescriptions(input.Graph.Records)

	for _, record := range sortedRecords(input.Graph.Records) {
		if record.Kind == "component" && record.SourcePayloadKind == "component" {
			key := componentSubjectKey(record, payloads[record.SourcePayloadID], payloadSubjects[record.SourcePayloadID], componentSlugs)
			if key.Slug == "" {
				continue
			}
			builder := getSubjectBuilder(builders, key)
			builder.view.Name = valueOr(record.DisplayName, string(key.Kind)+" "+key.Slug)
			addPayloadPages(builder, payloads[record.SourcePayloadID])
			continue
		}

		if record.Kind == "tokenSet" {
			for _, key := range recordSubjects(record, owners, styleTokenSets, payloadSubjects) {
				builder := getSubjectBuilder(builders, key)
				if key.Kind != SubjectComponent && builder.view.Name == "" {
					if key.Kind == SubjectStyle {
						builder.view.Name = styleSubjectName(key.Slug)
					} else if record.DisplayName != "" {
						builder.view.Name = record.DisplayName
					}
				}
				addPayloadPages(builder, payloads[record.SourcePayloadID])
			}
			continue
		}

		if record.Kind == "token" {
			if record.TokenName == "" {
				continue
			}
			for _, key := range recordSubjects(record, owners, styleTokenSets, payloadSubjects) {
				builder := getSubjectBuilder(builders, key)
				if key.Kind == SubjectStyle && builder.view.Name == "" {
					builder.view.Name = styleSubjectName(key.Slug)
				}
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
				for _, key := range recordSubjects(record, owners, styleTokenSets, payloadSubjects) {
					builder := getSubjectBuilder(builders, key)
					if key.Kind == SubjectStyle && builder.view.Name == "" {
						builder.view.Name = styleSubjectName(key.Slug)
					}
					addPayloadPages(builder, payloads[record.SourcePayloadID])
					if err := builder.addToken(entry.TokenName, "", entry.ResolvedValue, entry.ContextTags, entry.References, entry.Raw, record, contexts); err != nil {
						return Publication{}, err
					}
				}
			}
		}
	}

	addLifecycleNotices(builders, input.LifecycleNotices.Notices, owners, styleTokenSets, payloadSubjects)

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
		if builder.view.Name == "" {
			if key.Kind == SubjectStyle {
				builder.view.Name = styleSubjectName(key.Slug)
			} else {
				builder.view.Name = formatSubjectSlug(key.Slug)
			}
		}
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

func tokenSetOwners(records []tokensnapshot.ExtractedRecord, payloads map[string]SourcePayloadView, payloadSubjects map[string]SubjectKey, componentSlugs []string) map[string][]SubjectKey {
	owners := make(map[string][]SubjectKey)
	for _, record := range records {
		if record.Kind != "component" || record.SourcePayloadKind != "component" {
			continue
		}
		key := componentSubjectKey(record, payloads[record.SourcePayloadID], payloadSubjects[record.SourcePayloadID], componentSlugs)
		if key.Slug == "" {
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

func indexStyleTokenSets(records []tokensnapshot.ExtractedRecord) map[string]SubjectKey {
	styles := make(map[string]SubjectKey)
	for _, record := range records {
		if record.Kind != "tokenSet" {
			continue
		}
		if key, ok := styleTokenSetSubject(record.TokenSetName, record.DisplayName); ok {
			styles[record.ID] = key
		}
	}
	return styles
}

func recordSubjects(record tokensnapshot.ExtractedRecord, owners map[string][]SubjectKey, styleTokenSets map[string]SubjectKey, payloadSubjects map[string]SubjectKey) []SubjectKey {
	tokenSetID := tokenSetID(record.ID)
	if owned := owners[tokenSetID]; len(owned) > 0 {
		return owned
	}
	if styleKey, ok := styleTokenSets[tokenSetID]; ok {
		return []SubjectKey{styleKey}
	}
	if styleKey, ok := tokenNameStyleSubject(record.TokenName); ok {
		return []SubjectKey{styleKey}
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
		if pageMatchesSubject(pageURL, builder.view.Key) {
			builder.pageURLs[pageURL] = true
		}
	}
}

func noticeTargetSubject(notice tokensnapshot.LifecycleNotice, owners map[string][]SubjectKey, styleTokenSets map[string]SubjectKey, payloadSubjects map[string]SubjectKey) (SubjectKey, bool) {
	tokenSetID := tokenSetID(notice.SourceRecordID)
	if owned := owners[tokenSetID]; len(owned) > 0 {
		if payloadKey, ok := payloadSubjects[notice.SourcePayloadID]; ok {
			for _, k := range owned {
				if k == payloadKey {
					return k, true
				}
			}
		}
		for _, k := range owned {
			if strings.Contains(strings.ToLower(notice.Subject), strings.ToLower(k.Slug)) {
				return k, true
			}
		}
		return owned[0], true
	}
	if styleKey, ok := styleTokenSets[tokenSetID]; ok {
		return styleKey, true
	}
	if styleKey, ok := tokenNameStyleSubject(notice.Subject); ok {
		return styleKey, true
	}
	if key, ok := payloadSubjects[notice.SourcePayloadID]; ok {
		return key, true
	}
	return SubjectKey{}, false
}

func addLifecycleNotices(builders map[SubjectKey]*subjectBuilder, notices []tokensnapshot.LifecycleNotice, owners map[string][]SubjectKey, styleTokenSets map[string]SubjectKey, payloadSubjects map[string]SubjectKey) {
	for _, notice := range notices {
		key, ok := noticeTargetSubject(notice, owners, styleTokenSets, payloadSubjects)
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
