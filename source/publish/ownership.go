// Ownership maps captured records to catalog subjects. Material-specific token mappings stay explicit rather than becoming generic source rules.
package publish

import (
	"encoding/json"

	"sort"
	"strings"

	"github.com/sojupie/tokens/source/tokensnapshot"
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

	var payloadSlugs []string
	for _, pageURL := range payload.PageURLs {
		if slug := extractComponentSlug(pageURL); slug != "" {
			payloadSlugs = append(payloadSlugs, slug)
		}
	}
	if slug := bestComponentSlug(words, payloadSlugs); slug != "" {
		return SubjectKey{Kind: SubjectComponent, Slug: slug}
	}
	if slug := bestComponentSlug(words, knownSlugs); slug != "" {
		return SubjectKey{Kind: SubjectComponent, Slug: slug}
	}

	if fallback.Kind == SubjectComponent && fallback.Slug != "" {
		return fallback
	}
	return SubjectKey{Kind: SubjectComponent, Slug: strings.Join(words, "-")}
}

// Prefer the most specific name match; a tie is not evidence of ownership.
func bestComponentSlug(words, slugs []string) string {
	best, bestScore, bestExtra, tied := "", 0, 0, false
	for _, slug := range slugs {
		slugWords := cleanWords(slug)
		score := wordMatchScore(words, slugWords)
		extra := len(slugWords) - score
		if score > bestScore || score > 0 && score == bestScore && extra < bestExtra {
			best, bestScore, bestExtra, tied = slug, score, extra, false
		} else if score > 0 && score == bestScore && extra == bestExtra && best != slug {
			tied = true
		}
	}
	if tied {
		return ""
	}
	return best
}

func styleTokenSetSubject(tokenSetName, displayName string) (SubjectKey, bool) {
	lowerName := strings.ToLower(tokenSetName)
	lowerDisplay := ""
	if tokenSetName == "" {
		lowerDisplay = strings.ToLower(displayName)
	}
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
