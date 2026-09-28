package publish

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

func writeDatabase(path string, publication Publication) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open SQLite database: %w", err)
	}
	defer database.Close()
	if _, err := database.Exec(SchemaSQL); err != nil {
		return fmt.Errorf("create SQLite schema: %w", err)
	}
	if _, err := database.Exec(fmt.Sprintf("PRAGMA user_version = %d", SchemaVersion)); err != nil {
		return fmt.Errorf("set SQLite schema version: %w", err)
	}
	transaction, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin SQLite transaction: %w", err)
	}
	if err := insertPublication(transaction, publication); err != nil {
		_ = transaction.Rollback()
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit SQLite database: %w", err)
	}
	return nil
}

func insertPublication(transaction *sql.Tx, publication Publication) error {
	if _, err := transaction.Exec(
		"INSERT INTO publication (schema_version, publication_id, generated_at, current_snapshot_id) VALUES (?, ?, ?, ?)",
		publication.SchemaVersion, publication.PublicationID, publication.GeneratedAt, publication.CurrentSnapshot.ID,
	); err != nil {
		return fmt.Errorf("insert publication metadata: %w", err)
	}
	for _, snapshot := range publication.Snapshots {
		if _, err := transaction.Exec(
			"INSERT INTO snapshots (id, created_at, manifest_hash, capture_status, review_status) VALUES (?, ?, ?, ?, ?)",
			snapshot.ID, snapshot.CreatedAt, snapshot.ManifestHash, snapshot.CaptureStatus, snapshot.ReviewStatus,
		); err != nil {
			return fmt.Errorf("insert snapshot %s: %w", snapshot.ID, err)
		}
	}
	for _, change := range publication.Changes {
		if _, err := transaction.Exec(
			"INSERT INTO token_changes (id, type, subject, source_payload_kind, kind, name, before_json, after_json, previous_snapshot_id, snapshot_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			change.ID, change.Type, change.Subject, change.SourcePayloadKind, change.Kind, change.Name, nullableText(change.Before), nullableText(change.After), change.PreviousSnapshotID, change.SnapshotID,
		); err != nil {
			return fmt.Errorf("insert token change %s: %w", change.ID, err)
		}
	}

	payloads := append([]SourcePayloadView(nil), publication.SourcePayloads...)
	sort.Slice(payloads, func(i, j int) bool { return payloads[i].ID < payloads[j].ID })
	for _, payload := range payloads {
		pageURLs, err := json.Marshal(payload.PageURLs)
		if err != nil {
			return fmt.Errorf("encode source pages for payload %s: %w", payload.ID, err)
		}
		if _, err := transaction.Exec(
			"INSERT INTO source_payloads (id, url, hash, kind, area, source_subject, page_type, page_urls_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			payload.ID, payload.URL, payload.Hash, payload.Kind, payload.Area, payload.SourceSubject, payload.PageType, pageURLs,
		); err != nil {
			return fmt.Errorf("insert source payload %s: %w", payload.ID, err)
		}
	}

	subjects := append([]SubjectView(nil), publication.Subjects...)
	sort.Slice(subjects, func(i, j int) bool { return subjects[i].Key.String() < subjects[j].Key.String() })
	for _, subject := range subjects {
		pageURLs, err := json.Marshal(subject.SourcePageURLs)
		if err != nil {
			return fmt.Errorf("encode source pages for %s: %w", subject.Key.String(), err)
		}
		if _, err := transaction.Exec(
			"INSERT INTO subjects (kind, slug, name, source_page_urls_json) VALUES (?, ?, ?, ?)",
			subject.Key.Kind, subject.Key.Slug, subject.Name, pageURLs,
		); err != nil {
			return fmt.Errorf("insert subject %s: %w", subject.Key.String(), err)
		}
		if err := insertSubjectTokens(transaction, subject); err != nil {
			return err
		}
		if err := insertSubjectNotices(transaction, subject); err != nil {
			return err
		}
	}

	return nil
}

func insertSubjectTokens(transaction *sql.Tx, subject SubjectView) error {
	tokens := append([]TokenView(nil), subject.Tokens...)
	sort.Slice(tokens, func(i, j int) bool { return tokens[i].SourceName < tokens[j].SourceName })
	for _, token := range tokens {
		raw := jsonText(token.Raw)
		if _, err := transaction.Exec(
			"INSERT INTO tokens (subject_kind, subject_slug, source_name, category, raw_json, lifecycle) VALUES (?, ?, ?, ?, ?, ?)",
			subject.Key.Kind, subject.Key.Slug, token.SourceName, token.Category, raw, token.Lifecycle,
		); err != nil {
			return fmt.Errorf("insert token %s/%s: %w", subject.Key.String(), token.SourceName, err)
		}
		for _, tokenValue := range token.Values {
			contextTags, err := json.Marshal(tokenValue.ContextTags)
			if err != nil {
				return fmt.Errorf("encode token context %s/%s: %w", subject.Key.String(), token.SourceName, err)
			}
			sources, err := json.Marshal(tokenValue.Sources)
			if err != nil {
				return fmt.Errorf("encode token sources %s/%s: %w", subject.Key.String(), token.SourceName, err)
			}
			if _, err := transaction.Exec(
				"INSERT INTO token_values (subject_kind, subject_slug, source_name, context_tags_json, value_json, raw_json, sources_json) VALUES (?, ?, ?, ?, ?, ?, ?)",
				subject.Key.Kind, subject.Key.Slug, token.SourceName, contextTags, string(tokenValue.Value), jsonText(tokenValue.Raw), sources,
			); err != nil {
				return fmt.Errorf("insert token value %s/%s: %w", subject.Key.String(), token.SourceName, err)
			}
		}
		for _, source := range token.Sources {
			if _, err := transaction.Exec(
				"INSERT INTO token_sources (subject_kind, subject_slug, source_name, payload_id, source_record_id) VALUES (?, ?, ?, ?, ?)",
				subject.Key.Kind, subject.Key.Slug, token.SourceName, source.PayloadID, source.SourceRecordID,
			); err != nil {
				return fmt.Errorf("insert token source %s/%s: %w", subject.Key.String(), token.SourceName, err)
			}
		}
	}
	return nil
}

func insertSubjectNotices(transaction *sql.Tx, subject SubjectView) error {
	notices := append([]LifecycleNoticeView(nil), subject.LifecycleNotices...)
	sort.Slice(notices, func(i, j int) bool { return notices[i].ID < notices[j].ID })
	for _, notice := range notices {
		if _, err := transaction.Exec(
			"INSERT INTO lifecycle_notices (id, subject_kind, subject_slug, state, subject, message, source_payload_id, source_record_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			notice.ID, subject.Key.Kind, subject.Key.Slug, notice.State, notice.Subject, notice.Message, notice.SourcePayloadID, notice.SourceRecordID,
		); err != nil {
			return fmt.Errorf("insert lifecycle notice %s: %w", notice.ID, err)
		}
	}
	return nil
}

func jsonText(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
