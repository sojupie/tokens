package publish

import (
	"encoding/json"
	"strings"

	"github.com/sojupie/tokens/internal/tokensnapshot"
)

// TokenPresentation preserves the source's table hierarchy without exposing
// revision IDs or changing a token's typed value.
type TokenPresentation struct {
	SetName  string         `json:"setName"`
	SetLabel string         `json:"setLabel"`
	SetOrder int            `json:"setOrder"`
	Groups   []DisplayGroup `json:"groups,omitempty"`
	Order    int            `json:"order"`
}

type DisplayGroup struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

func tokenPresentations(records []tokensnapshot.ExtractedRecord) map[string]*TokenPresentation {
	type sourceFields struct {
		TokenSetName              string `json:"tokenSetName"`
		DisplayName               string `json:"displayName"`
		DisplayGroup              string `json:"displayGroup"`
		ParentGroup               string `json:"parentGroup"`
		Order                     int    `json:"order"`
		OrderInTokenSet           *int   `json:"orderInTokenSet"`
		OrderInDisplayGroup       *int   `json:"orderInDisplayGroup"`
		OrderInParentTokenSet     int    `json:"orderInParentTokenSet"`
		OrderInParentDisplayGroup int    `json:"orderInParentDisplayGroup"`
	}
	fields := make(map[string]sourceFields)
	shared := make(map[string]sourceFields)
	for _, record := range records {
		if record.Kind != "tokenSet" && record.Kind != "displayGroup" && record.Kind != "token" {
			continue
		}
		var value sourceFields
		if json.Unmarshal(record.Raw, &value) == nil {
			fields[record.SourcePayloadID+"\x00"+record.ID] = value
			shared[record.ID] = value
		}
	}
	result := make(map[string]*TokenPresentation)
	for _, record := range records {
		if record.Kind != "token" {
			continue
		}
		prefix := record.SourcePayloadID + "\x00"
		set, found := fields[prefix+tokenSetID(record.ID)]
		if !found {
			set = shared[tokenSetID(record.ID)]
		}
		if set.TokenSetName == "" || set.DisplayName == "" {
			continue
		}
		token := fields[prefix+record.ID]
		presentation := &TokenPresentation{SetName: set.TokenSetName, SetLabel: strings.TrimSpace(set.DisplayName), SetOrder: set.Order, Order: token.Order}
		if token.OrderInTokenSet != nil {
			presentation.Order = *token.OrderInTokenSet
		}
		if token.OrderInDisplayGroup != nil {
			presentation.Order = *token.OrderInDisplayGroup
		}
		seen := make(map[string]bool)
		for id := token.DisplayGroup; id != "" && !seen[id]; {
			seen[id] = true
			group, found := fields[prefix+id]
			if !found {
				group, found = shared[id]
			}
			if !found {
				break
			}
			order := group.OrderInParentTokenSet
			if group.ParentGroup != "" {
				order = group.OrderInParentDisplayGroup
			}
			if group.DisplayName != "" {
				presentation.Groups = append(presentation.Groups, DisplayGroup{Name: strings.TrimSpace(group.DisplayName), Order: order})
			}
			id = group.ParentGroup
		}
		for i, j := 0, len(presentation.Groups)-1; i < j; i, j = i+1, j-1 {
			presentation.Groups[i], presentation.Groups[j] = presentation.Groups[j], presentation.Groups[i]
		}
		result[prefix+record.ID] = presentation
	}
	return result
}
