package tokensnapshot

import (
	"encoding/json"
	"sort"
	"strings"
)

type sourceContextGroup struct {
	Name        string `json:"displayName"`
	DefaultTag  string `json:"defaultTag"`
	Specificity int    `json:"specificity"`
}

type sourceContextTag struct {
	Name  string `json:"displayName"`
	Order int    `json:"tagOrder"`
}

type ResolvedContext struct {
	Label      string
	IsDefault  bool
	GroupOrder int
	TagOrder   int
}

func ContextDescriptions(records []ExtractedRecord) map[string]ResolvedContext {
	groups := make(map[string]sourceContextGroup)
	tags := make(map[string]sourceContextTag)
	for _, record := range records {
		switch record.Kind {
		case "contextTagGroup":
			var group sourceContextGroup
			if json.Unmarshal(record.Raw, &group) == nil {
				groups[record.ID] = group
			}
		case "tag":
			var tag sourceContextTag
			if json.Unmarshal(record.Raw, &tag) == nil {
				tags[record.ID] = tag
			}
		}
	}

	contexts := make(map[string]ResolvedContext, len(tags))
	for tagID, tag := range tags {
		groupID, _, ok := strings.Cut(tagID, "/tags/")
		if !ok {
			continue
		}
		group, ok := groups[groupID]
		if !ok || tag.Name == "" {
			continue
		}
		label := tag.Name
		if group.Name != "" {
			label = group.Name + ": " + label
		}
		contexts[tagID] = ResolvedContext{
			Label:      label,
			IsDefault:  group.DefaultTag == tagID,
			GroupOrder: group.Specificity,
			TagOrder:   tag.Order,
		}
	}
	return contexts
}

func ResolveContext(tags []string, descriptions map[string]ResolvedContext) ([]string, bool) {
	if len(tags) == 0 {
		return nil, true
	}
	type orderedLabel struct {
		label      string
		groupOrder int
		tagOrder   int
	}
	ordered := make([]orderedLabel, 0, len(tags))
	allDefault := true
	allKnown := true
	for _, tag := range tags {
		context, ok := descriptions[tag]
		if !ok {
			allKnown = false
			allDefault = false
			continue
		}
		ordered = append(ordered, orderedLabel{label: context.Label, groupOrder: context.GroupOrder, tagOrder: context.TagOrder})
		allDefault = allDefault && context.IsDefault
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].groupOrder == ordered[j].groupOrder {
			if ordered[i].tagOrder == ordered[j].tagOrder {
				return ordered[i].label < ordered[j].label
			}
			return ordered[i].tagOrder < ordered[j].tagOrder
		}
		return ordered[i].groupOrder < ordered[j].groupOrder
	})
	labels := make([]string, 0, len(ordered))
	for _, item := range ordered {
		labels = append(labels, item.label)
	}
	return labels, allKnown && allDefault
}
