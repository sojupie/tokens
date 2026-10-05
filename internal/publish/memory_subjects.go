package publish

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func loadSubjects(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT kind, slug, name, source_page_urls_json
		FROM subjects
		ORDER BY kind, slug`)
	if err != nil {
		return fmt.Errorf("query subjects: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var subject SubjectView
		var pageURLs string
		if err := rows.Scan(&subject.Key.Kind, &subject.Key.Slug, &subject.Name, &pageURLs); err != nil {
			return fmt.Errorf("read subject: %w", err)
		}
		if err := json.Unmarshal([]byte(pageURLs), &subject.SourcePageURLs); err != nil {
			return fmt.Errorf("decode source pages for %s: %w", subject.Key.String(), err)
		}
		subject.Tokens = []TokenView{}
		subject.LifecycleNotices = []LifecycleNoticeView{}
		state.Subjects[subject.Key] = subject
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read subjects: %w", err)
	}
	if err := loadTokens(database, state); err != nil {
		return err
	}
	return loadLifecycleNotices(database, state)
}

func loadTokens(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT subject_kind, subject_slug, source_name, display_name, category, raw_json, lifecycle
		FROM tokens
		ORDER BY subject_kind, subject_slug, source_name`)
	if err != nil {
		return fmt.Errorf("query tokens: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key SubjectKey
		var token TokenView
		var raw, lifecycle sql.NullString
		if err := rows.Scan(&key.Kind, &key.Slug, &token.SourceName, &token.DisplayName, &token.Category, &raw, &lifecycle); err != nil {
			return fmt.Errorf("read token: %w", err)
		}
		if _, ok := state.Subjects[key]; !ok {
			return fmt.Errorf("token %s belongs to missing subject", token.SourceName)
		}
		if raw.Valid {
			token.Raw = json.RawMessage(raw.String)
		}
		token.Lifecycle = lifecycle.String
		token.Values = []TokenValueView{}
		token.Sources = []TokenSourceView{}
		subject := state.Subjects[key]
		subject.Tokens = append(subject.Tokens, token)
		state.Subjects[key] = subject
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read tokens: %w", err)
	}
	if err := loadTokenValues(database, state); err != nil {
		return err
	}
	return loadTokenSources(database, state)
}

func loadTokenValues(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT subject_kind, subject_slug, source_name, context_tags_json, context_labels_json, is_default_context, value_json, references_json, raw_json, sources_json
		FROM token_values
		ORDER BY subject_kind, subject_slug, source_name, context_tags_json`)
	if err != nil {
		return fmt.Errorf("query token values: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key SubjectKey
		var sourceName, contextTags, contextLabels, value, references, sources string
		var isDefaultContext bool
		var raw sql.NullString
		if err := rows.Scan(&key.Kind, &key.Slug, &sourceName, &contextTags, &contextLabels, &isDefaultContext, &value, &references, &raw, &sources); err != nil {
			return fmt.Errorf("read token value: %w", err)
		}
		subject, ok := state.Subjects[key]
		if !ok {
			return fmt.Errorf("token value %s belongs to missing subject", sourceName)
		}
		token, ok := findSubjectToken(&subject, sourceName)
		if !ok {
			return fmt.Errorf("token value %s belongs to missing token", sourceName)
		}
		var tokenValue TokenValueView
		if err := json.Unmarshal([]byte(contextTags), &tokenValue.ContextTags); err != nil {
			return fmt.Errorf("decode context tags for %s: %w", sourceName, err)
		}
		if err := json.Unmarshal([]byte(contextLabels), &tokenValue.ContextLabels); err != nil {
			return fmt.Errorf("decode context labels for %s: %w", sourceName, err)
		}
		tokenValue.IsDefaultContext = isDefaultContext
		tokenValue.Value = json.RawMessage(value)
		if err := json.Unmarshal([]byte(references), &tokenValue.References); err != nil {
			return fmt.Errorf("decode references for %s: %w", sourceName, err)
		}
		if raw.Valid {
			tokenValue.Raw = json.RawMessage(raw.String)
		}
		if err := json.Unmarshal([]byte(sources), &tokenValue.Sources); err != nil {
			return fmt.Errorf("decode value sources for %s: %w", sourceName, err)
		}
		token.Values = append(token.Values, tokenValue)
		state.Subjects[key] = subject
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read token values: %w", err)
	}
	return nil
}

func loadTokenSources(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT subject_kind, subject_slug, source_name, payload_id, source_record_id
		FROM token_sources
		ORDER BY subject_kind, subject_slug, source_name, payload_id, source_record_id`)
	if err != nil {
		return fmt.Errorf("query token sources: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key SubjectKey
		var sourceName string
		var source TokenSourceView
		if err := rows.Scan(&key.Kind, &key.Slug, &sourceName, &source.PayloadID, &source.SourceRecordID); err != nil {
			return fmt.Errorf("read token source: %w", err)
		}
		subject, ok := state.Subjects[key]
		if !ok {
			return fmt.Errorf("token source %s belongs to missing subject", sourceName)
		}
		token, ok := findSubjectToken(&subject, sourceName)
		if !ok {
			return fmt.Errorf("token source %s belongs to missing token", sourceName)
		}
		token.Sources = append(token.Sources, source)
		state.Subjects[key] = subject
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read token sources: %w", err)
	}
	return nil
}

func loadLifecycleNotices(database *sql.DB, state *PublishedState) error {
	rows, err := database.Query(`
		SELECT id, subject_kind, subject_slug, state, subject, message, source_payload_id, source_record_id
		FROM lifecycle_notices
		ORDER BY subject_kind, subject_slug, id`)
	if err != nil {
		return fmt.Errorf("query lifecycle notices: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var notice LifecycleNoticeView
		var key SubjectKey
		if err := rows.Scan(&notice.ID, &key.Kind, &key.Slug, &notice.State, &notice.Subject, &notice.Message, &notice.SourcePayloadID, &notice.SourceRecordID); err != nil {
			return fmt.Errorf("read lifecycle notice: %w", err)
		}
		subject, ok := state.Subjects[key]
		if !ok {
			return fmt.Errorf("lifecycle notice %s belongs to missing subject", notice.ID)
		}
		subject.LifecycleNotices = append(subject.LifecycleNotices, notice)
		state.Subjects[key] = subject
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read lifecycle notices: %w", err)
	}
	return nil
}

func findSubjectToken(subject *SubjectView, sourceName string) (*TokenView, bool) {
	for index := range subject.Tokens {
		if subject.Tokens[index].SourceName == sourceName {
			return &subject.Tokens[index], true
		}
	}
	return nil, false
}
