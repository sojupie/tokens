package presentation

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sojupie/tokens/source/publish"
)

type ChangeBatch struct {
	DetectedAt string
	PreviousAt string
	Changes    []publish.ChangeView
}

func ChangedItemCount(changes []LabeledChange) int {
	items := make(map[string]bool)
	for _, change := range changes {
		items[change.Kind+"\x00"+change.Name] = true
	}
	return len(items)
}

func GroupChangelogURL(base string, group ChangeGroup, snapshot string) string {
	address, err := url.Parse(base)
	if err != nil {
		return "/changes"
	}
	query := address.Query()
	query.Del("page")
	query.Del("subject")
	query.Set("snapshot", snapshot)
	path := group.URL
	if path == "" {
		kind := publish.SubjectComponent
		if group.Changes[0].SourcePayloadKind != "" && group.Changes[0].SourcePayloadKind != "component" {
			kind = publish.SubjectStyle
		}
		path = SubjectURL(publish.SubjectView{Key: publish.SubjectKey{Kind: kind, Slug: group.Changes[0].Subject}})
	}
	return path + "/changes?" + query.Encode()
}

type LabeledChange struct {
	publish.ChangeView
	Hierarchy []string
}

type ChangeGroup struct {
	Label   string
	URL     string
	Changes []LabeledChange
}

func ChangeGroups(changes []publish.ChangeView, subjects map[publish.SubjectKey]publish.SubjectView) []ChangeGroup {
	bySlug := make(map[string]publish.SubjectView, len(subjects))
	for _, subject := range subjects {
		bySlug[subject.Key.Slug] = subject
	}
	var groups []ChangeGroup
	bySubject := make(map[string]int)
	for _, change := range changes {
		index, exists := bySubject[change.Subject]
		if !exists {
			group := ChangeGroup{Label: DisplayWords(change.Subject)}
			if subject, ok := bySlug[change.Subject]; ok {
				if subject.Name != "" {
					group.Label = SubjectName(subject)
				}
				group.URL = SubjectURL(subject)
			}
			if group.Label == "" {
				group.Label = "Other changes"
			}
			index = len(groups)
			groups = append(groups, group)
			bySubject[change.Subject] = index
		}
		group := &groups[index]
		group.Changes = append(group.Changes, LabeledChange{ChangeView: change, Hierarchy: ChangeHierarchy(change.Name)})
	}
	return groups
}

// Use the historical identifier, not today's token values or context labels.
// Keeping its segments separate preserves variant, state and part nesting,
// including for tokens that no longer exist in the current publication.
func ChangeHierarchy(name string) []string {
	parts := strings.Split(name, ".")
	if len(parts) < 3 || parts[0] != "md" {
		return nil
	}
	switch parts[1] {
	case "comp", "sys", "ref":
		parts = parts[2:]
	default:
		return nil
	}
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			labels = append(labels, TokenStateLabel(part))
		}
	}
	return labels
}

func TimelineTitle(subject string) string {
	if subject == "" {
		return "Change log"
	}
	return "Material Design 3 " + TokenPhrase(DisplayWords(subject)) + " token changes"
}

func ChangeBatches(changes []publish.ChangeView, snapshots []publish.SnapshotView) []ChangeBatch {
	dates := make(map[string]string, len(snapshots))
	for _, snapshot := range snapshots {
		dates[snapshot.ID] = snapshot.CreatedAt
	}
	var batches []ChangeBatch
	lastID := ""
	for _, change := range changes {
		if len(batches) == 0 || change.SnapshotID != lastID {
			date := change.DetectedAt
			if date == "" {
				date = dates[change.SnapshotID]
			}
			batches = append(batches, ChangeBatch{DetectedAt: date, PreviousAt: dates[change.PreviousSnapshotID]})
			lastID = change.SnapshotID
		}
		batch := &batches[len(batches)-1]
		batch.Changes = append(batch.Changes, change)
	}
	return batches
}

func ObservationDate(value string) string {
	date, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "Date unavailable"
	}
	return date.UTC().Format("2 January 2006, 15:04 UTC")
}

func HistoryCoverage(snapshots []publish.SnapshotView) string {
	if len(snapshots) == 0 {
		return "No snapshot history is available."
	}
	coverage := fmt.Sprintf("History from %s to %s · %d snapshots", ObservationDate(snapshots[0].CreatedAt), ObservationDate(snapshots[len(snapshots)-1].CreatedAt), len(snapshots))
	for _, snapshot := range snapshots {
		if snapshot.CaptureStatus == "partial" {
			return coverage + ". Includes incomplete captures; missing source data can appear as additions or removals."
		}
	}
	return coverage
}
