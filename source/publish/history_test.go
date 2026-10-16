package publish

import (
	"testing"

	"github.com/sojupie/tokens/source/tokensnapshot"
)

func TestChangeViewsAreStableAndSorted(t *testing.T) {
	diff := tokensnapshot.SnapshotDiff{
		FromSnapshotID: "snapshot-1",
		ToSnapshotID:   "snapshot-2",
		Semantic: tokensnapshot.SemanticDiff{
			Changed: []tokensnapshot.SemanticChange{{
				Key:               "token-2",
				SourcePayloadKind: "component",
				SourceSubject:     "snackbar",
				Kind:              "contextualReferenceTree",
				Name:              "md.comp.snackbar.container.width",
				Before:            `{"dimension":{"value":480,"unit":"DIPS"}}`,
				After:             `{"dimension":{"value":560,"unit":"DIPS"}}`,
			}},
			Added: []tokensnapshot.SemanticChange{{
				Key:           "token-1",
				SourceSubject: "button",
				Kind:          "contextualReferenceTree",
				Name:          "md.comp.button.container.height",
				After:         `{"dimension":{"value":40,"unit":"DIPS"}}`,
			}},
		},
	}

	changes := changeViews(diff)
	if len(changes) != 2 {
		t.Fatalf("change count = %d, want 2", len(changes))
	}
	if changes[0].Subject != "button" || changes[1].Subject != "snackbar" {
		t.Fatalf("changes are not sorted: %#v", changes)
	}
	if changes[1].Type != "changed" || changes[1].PreviousSnapshotID != "snapshot-1" || changes[1].SnapshotID != "snapshot-2" {
		t.Fatalf("change metadata = %#v", changes[1])
	}
	if changes[1].ID == "" || changes[1].ID != changeViews(diff)[1].ID {
		t.Fatal("change ID is not stable")
	}
}

func TestHistoryTokenOwnersUsePublishedComponentNames(t *testing.T) {
	owners := historyTokenOwners([]SubjectView{
		{Key: SubjectKey{Kind: SubjectComponent, Slug: "buttons"}, Tokens: []TokenView{{SourceName: "md.comp.button.height"}, {SourceName: "md.comp.elevated-button.height"}, {SourceName: "shared"}}},
		{Key: SubjectKey{Kind: SubjectStyle, Slug: "spacing"}, Tokens: []TokenView{{SourceName: "md.sys.spacing.small"}, {SourceName: "shared"}}},
	})
	if owners["component:button"] != "buttons" || owners["component:elevated-button"] != "buttons" || owners["md.sys.spacing.small"] != "spacing" || owners["shared"] != "" {
		t.Fatalf("owners = %#v", owners)
	}
}
