package publish

import (
	"database/sql"
	"fmt"
)

func loadChanges(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT id, type, subject, source_payload_kind, kind, name, before_json, after_json, previous_snapshot_id, snapshot_id
		FROM token_changes
		ORDER BY subject, name, type, id`)
	if err != nil {
		return fmt.Errorf("query token changes: %w", err)
	}
	defer rows.Close()
	dates := make(map[string]string, len(state.Snapshots))
	for _, snapshot := range state.Snapshots {
		dates[snapshot.ID] = snapshot.CreatedAt
	}
	for rows.Next() {
		var change ChangeView
		var before, after sql.NullString
		if err := rows.Scan(&change.ID, &change.Type, &change.Subject, &change.SourcePayloadKind, &change.Kind, &change.Name, &before, &after, &change.PreviousSnapshotID, &change.SnapshotID); err != nil {
			return fmt.Errorf("read token change: %w", err)
		}
		if before.Valid {
			change.Before = before.String
		}
		if after.Valid {
			change.After = after.String
		}
		change.DetectedAt = dates[change.SnapshotID]
		state.Changes = append(state.Changes, change)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read token changes: %w", err)
	}
	sortChanges(state.Changes, state.Snapshots)
	return nil
}
