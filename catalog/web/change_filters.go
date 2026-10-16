package web

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

type changeOption struct {
	Value string
	Label string
}

type changeFilters struct {
	Query        string
	Subject      string
	Snapshot     string
	Type         string
	Subjects     []changeOption
	Snapshots    []changeOption
	FixedSubject bool
}

func filterChangeHistory(state publish.PublishedState, subject string, query url.Values) ([]publish.ChangeView, changeFilters) {
	filters := changeFilters{Query: strings.TrimSpace(query.Get("q")), Subject: query.Get("subject"), Snapshot: query.Get("snapshot"), Type: query.Get("type"), FixedSubject: subject != ""}
	if subject != "" {
		filters.Subject = subject
	}
	labels := make(map[string]string)
	for _, item := range state.Subjects {
		labels[item.Key.Slug] = item.Name
	}
	subjects := make(map[string]bool)
	snapshots := make(map[string]bool)
	var matches []publish.ChangeView
	search := strings.ToLower(filters.Query)
	for _, change := range state.Changes {
		if subject != "" && change.Subject != subject {
			continue
		}
		subjects[change.Subject] = true
		snapshots[change.SnapshotID] = true
		if filters.Subject != "" && change.Subject != filters.Subject || filters.Snapshot != "" && change.SnapshotID != filters.Snapshot || filters.Type != "" && change.Type != filters.Type {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(change.Name+" "+strings.Join(changeHierarchy(change.Name), " ")), search) {
			continue
		}
		matches = append(matches, change)
	}
	for slug := range subjects {
		label := labels[slug]
		if label == "" {
			label = displayWords(slug)
		}
		filters.Subjects = append(filters.Subjects, changeOption{slug, label})
	}
	sort.Slice(filters.Subjects, func(i, j int) bool { return filters.Subjects[i].Label < filters.Subjects[j].Label })
	for i := len(state.Snapshots) - 1; i >= 0; i-- {
		snapshot := state.Snapshots[i]
		if snapshots[snapshot.ID] {
			filters.Snapshots = append(filters.Snapshots, changeOption{snapshot.ID, observationDate(snapshot.CreatedAt)})
		}
	}
	return matches, filters
}

func changeCountSummary(changes []publish.ChangeView) string {
	counts := make(map[string]int)
	for _, change := range changes {
		counts[change.Type]++
	}
	return fmt.Sprintf("%d added · %d updated · %d removed", counts["added"], counts["changed"], counts["removed"])
}
