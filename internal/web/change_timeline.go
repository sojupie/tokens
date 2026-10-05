package web

import (
	"fmt"
	"strings"
	"time"

	"github.com/sojupie/m3-source/internal/publish"
)

type changeBatch struct {
	DetectedAt string
	PreviousAt string
	Changes    []publish.ChangeView
}

type labeledChange struct {
	publish.ChangeView
	Label     string
	Hierarchy []string
}

type changeGroup struct {
	Label   string
	URL     string
	Changes []labeledChange
}

func changeGroups(changes []publish.ChangeView, subjects map[publish.SubjectKey]publish.SubjectView) []changeGroup {
	bySlug := make(map[string]publish.SubjectView, len(subjects))
	for _, subject := range subjects {
		bySlug[subject.Key.Slug] = subject
	}
	var groups []changeGroup
	lastSubject := ""
	var names map[string]string
	for _, change := range changes {
		if len(groups) == 0 || lastSubject != change.Subject {
			group := changeGroup{Label: displayWords(change.Subject)}
			names = make(map[string]string)
			if subject, ok := bySlug[change.Subject]; ok {
				if subject.Name != "" {
					group.Label = subject.Name
				}
				group.URL = subjectURL(subject)
				labelCounts := make(map[string]int)
				for _, token := range subject.Tokens {
					labelCounts[strings.ToLower(tokenDisplayName(token))]++
				}
				for _, token := range subject.Tokens {
					label := tokenDisplayName(token)
					// Some source labels are just "Type style", or omit the variant.
					// Use the readable full token path when a label is ambiguous.
					if labelCounts[strings.ToLower(label)] > 1 {
						label = readableChangeName(token.SourceName)
					}
					names[token.SourceName] = label
				}
			}
			if group.Label == "" {
				group.Label = "Other changes"
			}
			groups = append(groups, group)
			lastSubject = change.Subject
		}
		label := names[change.Name]
		if label == "" {
			label = readableChangeName(change.Name)
		}
		group := &groups[len(groups)-1]
		group.Changes = append(group.Changes, labeledChange{ChangeView: change, Label: label, Hierarchy: changeHierarchy(change.Name)})
	}
	return groups
}

// Use the historical identifier, not today's token values or context labels.
// Keeping its segments separate preserves variant, state and part nesting,
// including for tokens that no longer exist in the current publication.
func changeHierarchy(name string) []string {
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
			labels = append(labels, tokenStateLabel(part))
		}
	}
	return labels
}

func readableChangeName(name string) string {
	if !strings.HasPrefix(name, "md.") {
		return name
	}
	name = strings.TrimPrefix(strings.TrimPrefix(name, "md.sys."), "md.ref.")
	return tokenDisplayName(publish.TokenView{SourceName: name})
}

func timelineTitle(subject string, subjects map[publish.SubjectKey]publish.SubjectView) string {
	if subject == "" {
		return "Change log"
	}
	for _, item := range subjects {
		if item.Key.Slug == subject && item.Name != "" {
			return item.Name + " changes"
		}
	}
	return displayWords(subject) + " changes"
}

func changeBatches(changes []publish.ChangeView, snapshots []publish.SnapshotView) []changeBatch {
	dates := make(map[string]string, len(snapshots))
	for _, snapshot := range snapshots {
		dates[snapshot.ID] = snapshot.CreatedAt
	}
	var batches []changeBatch
	lastID := ""
	for _, change := range changes {
		if len(batches) == 0 || change.SnapshotID != lastID {
			date := change.DetectedAt
			if date == "" {
				date = dates[change.SnapshotID]
			}
			batches = append(batches, changeBatch{DetectedAt: date, PreviousAt: dates[change.PreviousSnapshotID]})
			lastID = change.SnapshotID
		}
		batch := &batches[len(batches)-1]
		batch.Changes = append(batch.Changes, change)
	}
	return batches
}

func observationDate(value string) string {
	date, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "Date unavailable"
	}
	return date.UTC().Format("2 January 2006, 15:04 UTC")
}

func historyCoverage(snapshots []publish.SnapshotView) string {
	if len(snapshots) == 0 {
		return "No snapshot history is available."
	}
	return fmt.Sprintf("History from %s to %s · %d snapshots", observationDate(snapshots[0].CreatedAt), observationDate(snapshots[len(snapshots)-1].CreatedAt), len(snapshots))
}
