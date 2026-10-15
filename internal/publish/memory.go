package publish

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type PublishedState struct {
	schemaVersion   int
	PublicationID   string
	GeneratedAt     string
	CurrentSnapshot SnapshotView
	Snapshots       []SnapshotView
	Changes         []ChangeView
	Subjects        map[SubjectKey]SubjectView
	SourcePayloads  map[string]SourcePayloadView
}

func Load(databasePath string) (PublishedState, error) {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return PublishedState{}, fmt.Errorf("open publication database: %w", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec("PRAGMA query_only = ON"); err != nil {
		return PublishedState{}, fmt.Errorf("set publication database read-only: %w", err)
	}

	state, err := loadPublication(database)
	if err != nil {
		return PublishedState{}, err
	}
	if err := loadSnapshots(database, &state); err != nil {
		return PublishedState{}, err
	}
	if err := loadSourcePayloads(database, &state); err != nil {
		return PublishedState{}, err
	}
	if err := loadChanges(database, &state); err != nil {
		return PublishedState{}, err
	}
	if err := loadSubjects(database, &state); err != nil {
		return PublishedState{}, err
	}
	return state, nil
}

func loadPublication(database *sql.DB) (PublishedState, error) {
	var state PublishedState
	var schemaVersion int
	var currentSnapshotID string
	if err := database.QueryRow(
		"SELECT schema_version, publication_id, generated_at, current_snapshot_id FROM publication LIMIT 1",
	).Scan(&schemaVersion, &state.PublicationID, &state.GeneratedAt, &currentSnapshotID); err != nil {
		return PublishedState{}, fmt.Errorf("read publication metadata: %w", err)
	}
	if schemaVersion != SchemaVersion && schemaVersion != 3 {
		return PublishedState{}, fmt.Errorf("unsupported publication schema version %d; rebuild with tokens publish", schemaVersion)
	}
	state.schemaVersion = schemaVersion
	state.Subjects = make(map[SubjectKey]SubjectView)
	state.SourcePayloads = make(map[string]SourcePayloadView)
	state.CurrentSnapshot.ID = currentSnapshotID
	return state, nil
}

func loadSnapshots(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT id, created_at, manifest_hash, capture_status, review_status
		FROM snapshots
		ORDER BY created_at, id`)
	if err != nil {
		return fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var snapshot SnapshotView
		if err := rows.Scan(&snapshot.ID, &snapshot.CreatedAt, &snapshot.ManifestHash, &snapshot.CaptureStatus, &snapshot.ReviewStatus); err != nil {
			return fmt.Errorf("read snapshot: %w", err)
		}
		state.Snapshots = append(state.Snapshots, snapshot)
		if snapshot.ID == state.CurrentSnapshot.ID {
			state.CurrentSnapshot = snapshot
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read snapshots: %w", err)
	}
	if state.CurrentSnapshot.CreatedAt == "" {
		return fmt.Errorf("current snapshot %q is not present", state.CurrentSnapshot.ID)
	}
	dates := make(map[string]time.Time, len(state.Snapshots))
	for _, snapshot := range state.Snapshots {
		date, err := time.Parse(time.RFC3339Nano, snapshot.CreatedAt)
		if err != nil {
			return fmt.Errorf("snapshot %s date: %w", snapshot.ID, err)
		}
		dates[snapshot.ID] = date
	}
	sort.Slice(state.Snapshots, func(i, j int) bool {
		first, second := state.Snapshots[i], state.Snapshots[j]
		if dates[first.ID].Equal(dates[second.ID]) {
			if first.ID == state.CurrentSnapshot.ID {
				return false
			}
			if second.ID == state.CurrentSnapshot.ID {
				return true
			}
			return first.ID < second.ID
		}
		return dates[first.ID].Before(dates[second.ID])
	})
	return nil
}

func loadSourcePayloads(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT id, url, hash, kind, area, source_subject, page_type, page_urls_json
		FROM source_payloads
		ORDER BY id`)
	if err != nil {
		return fmt.Errorf("query source payloads: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var payload SourcePayloadView
		var pageURLs string
		if err := rows.Scan(&payload.ID, &payload.URL, &payload.Hash, &payload.Kind, &payload.Area, &payload.SourceSubject, &payload.PageType, &pageURLs); err != nil {
			return fmt.Errorf("read source payload: %w", err)
		}
		if err := json.Unmarshal([]byte(pageURLs), &payload.PageURLs); err != nil {
			return fmt.Errorf("decode source pages for payload %s: %w", payload.ID, err)
		}
		state.SourcePayloads[payload.ID] = payload
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read source payloads: %w", err)
	}
	return nil
}
