package presentation

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sojupie/tokens/source/publish"
)

type ChangeBatch struct {
	SnapshotID string
	DetectedAt string
	PreviousAt string
	Changes    []publish.ChangeView
}

type ChangeTimelineBatch struct {
	SnapshotID string
	DetectedAt string
	PreviousAt string
	Groups     []ChangeGroup
	Hidden     bool
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
	for key, values := range query {
		if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
			query.Del(key)
		}
	}
	query.Set("snapshot", snapshot)
	path := group.URL
	if path == "" {
		kind := publish.SubjectComponent
		if len(group.Changes) > 0 && group.Changes[0].SourcePayloadKind != "" && group.Changes[0].SourcePayloadKind != "component" {
			kind = publish.SubjectStyle
		}
		slug := group.Subject
		if slug == "" && len(group.Changes) > 0 {
			slug = group.Changes[0].Subject
		}
		path = SubjectURL(publish.SubjectView{Key: publish.SubjectKey{Kind: kind, Slug: slug}})
	}
	return path + "/changes?" + query.Encode()
}

type LabeledChange struct {
	publish.ChangeView
	Hierarchy []string
	Search    string
	Hidden    bool
}

type ChangeGroup struct {
	Subject    string
	Label      string
	URL        string
	Changes    []LabeledChange
	Count      int
	CountsJSON string
	Hidden     bool
}

func ChangeSearchText(name string) string {
	hierarchy := ChangeHierarchy(name)
	if len(hierarchy) == 0 {
		return strings.ToLower(name)
	}
	return strings.ToLower(name + " " + strings.Join(hierarchy, " "))
}

func ChangeMatchesFilters(change publish.ChangeView, filters ChangeFilters) bool {
	if filters.Subject != "" && change.Subject != filters.Subject {
		return false
	}
	if filters.Snapshot != "" && change.SnapshotID != filters.Snapshot {
		return false
	}
	if filters.Type != "" && change.Type != filters.Type {
		return false
	}
	if filters.Query != "" && !strings.Contains(ChangeSearchText(change.Name), strings.ToLower(filters.Query)) {
		return false
	}
	return true
}

func ChangeTypeCountsJSON(changes []LabeledChange) string {
	all := make(map[string]bool)
	byType := make(map[string]map[string]bool)
	for _, change := range changes {
		key := change.Kind + "\x00" + change.Name
		all[key] = true
		if byType[change.Type] == nil {
			byType[change.Type] = make(map[string]bool)
		}
		byType[change.Type][key] = true
	}
	counts := make(map[string]int, len(byType)+1)
	counts[""] = len(all)
	for changeType, items := range byType {
		counts[changeType] = len(items)
	}
	encoded, _ := json.Marshal(counts)
	return string(encoded)
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
			group := ChangeGroup{Subject: change.Subject, Label: DisplayWords(change.Subject)}
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
		hierarchy := ChangeHierarchy(change.Name)
		group.Changes = append(group.Changes, LabeledChange{
			ChangeView: change,
			Hierarchy:  hierarchy,
			Search:     ChangeSearchText(change.Name),
		})
	}
	return groups
}

func BuildChangeTimeline(mounted []publish.ChangeView, snapshots []publish.SnapshotView, subjects map[publish.SubjectKey]publish.SubjectView, filters ChangeFilters, global bool, local bool) []ChangeTimelineBatch {
	batches := ChangeBatches(mounted, snapshots)
	result := make([]ChangeTimelineBatch, 0, len(batches))
	for _, batch := range batches {
		groups := ChangeGroups(batch.Changes, subjects)
		batchHidden := true
		for i := range groups {
			group := &groups[i]
			if global {
				if local {
					group.CountsJSON = ChangeTypeCountsJSON(group.Changes)
				}
				matched := make([]LabeledChange, 0, len(group.Changes))
				for _, change := range group.Changes {
					if !local || ChangeMatchesFilters(change.ChangeView, filters) {
						matched = append(matched, change)
					}
				}
				group.Count = ChangedItemCount(matched)
				group.Hidden = group.Count == 0
				if !group.Hidden {
					batchHidden = false
				}
			} else {
				groupHidden := true
				for j := range group.Changes {
					if local && !ChangeMatchesFilters(group.Changes[j].ChangeView, filters) {
						group.Changes[j].Hidden = true
					} else {
						groupHidden = false
					}
				}
				group.Hidden = groupHidden
				if !groupHidden {
					batchHidden = false
				}
			}
		}
		result = append(result, ChangeTimelineBatch{
			SnapshotID: batch.SnapshotID,
			DetectedAt: batch.DetectedAt,
			PreviousAt: batch.PreviousAt,
			Groups:     groups,
			Hidden:     batchHidden,
		})
	}
	return result
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
			batches = append(batches, ChangeBatch{SnapshotID: change.SnapshotID, DetectedAt: date, PreviousAt: dates[change.PreviousSnapshotID]})
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
			return coverage + ". Some source captures were incomplete."
		}
	}
	return coverage
}
