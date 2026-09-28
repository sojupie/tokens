package publish

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/sojupie/m3-source/internal/tokensnapshot"
)

func changeViews(diff tokensnapshot.SnapshotDiff) []ChangeView {
	changes := make([]ChangeView, 0, len(diff.Semantic.Added)+len(diff.Semantic.Removed)+len(diff.Semantic.Changed))
	appendChanges := func(changeType string, values []tokensnapshot.SemanticChange) {
		for _, value := range values {
			changes = append(changes, ChangeView{
				ID:                 changeID(diff.FromSnapshotID, diff.ToSnapshotID, changeType, value),
				Type:               changeType,
				Subject:            value.SourceSubject,
				SourcePayloadKind:  value.SourcePayloadKind,
				Kind:               value.Kind,
				Name:               value.Name,
				Before:             value.Before,
				After:              value.After,
				PreviousSnapshotID: diff.FromSnapshotID,
				SnapshotID:         diff.ToSnapshotID,
			})
		}
	}
	appendChanges("added", diff.Semantic.Added)
	appendChanges("removed", diff.Semantic.Removed)
	appendChanges("changed", diff.Semantic.Changed)
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Subject != changes[j].Subject {
			return changes[i].Subject < changes[j].Subject
		}
		if changes[i].Name != changes[j].Name {
			return changes[i].Name < changes[j].Name
		}
		return changes[i].Type < changes[j].Type
	})
	return changes
}

func changeID(previousSnapshotID, snapshotID, changeType string, change tokensnapshot.SemanticChange) string {
	input := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", previousSnapshotID, snapshotID, changeType, change.Kind, change.Key)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}
