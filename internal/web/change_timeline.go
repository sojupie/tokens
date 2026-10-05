package web

import (
	"fmt"
	"time"

	"github.com/sojupie/m3-source/internal/publish"
)

type changeBatch struct {
	DetectedAt string
	PreviousAt string
	Changes    []publish.ChangeView
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
