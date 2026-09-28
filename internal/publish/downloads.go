package publish

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type tokenDownload struct {
	SchemaVersion int           `json:"schemaVersion"`
	PublicationID string        `json:"publicationId"`
	SnapshotID    string        `json:"snapshotId"`
	Subjects      []SubjectView `json:"subjects"`
}

type subjectDownload struct {
	SchemaVersion int         `json:"schemaVersion"`
	PublicationID string      `json:"publicationId"`
	SnapshotID    string      `json:"snapshotId"`
	Subject       SubjectView `json:"subject"`
}

func writeDownloads(directory string, publication Publication) error {
	downloadsDirectory := filepath.Join(directory, "downloads")
	if err := os.MkdirAll(downloadsDirectory, 0o755); err != nil {
		return fmt.Errorf("create downloads directory: %w", err)
	}

	subjects := append([]SubjectView(nil), publication.Subjects...)
	sort.Slice(subjects, func(i, j int) bool {
		return subjects[i].Key.String() < subjects[j].Key.String()
	})
	subjectPaths := make(map[SubjectKey]string, len(subjects))
	for _, subject := range subjects {
		path, err := subjectDownloadPath(downloadsDirectory, subject.Key)
		if err != nil {
			return err
		}
		subjectPaths[subject.Key] = path
	}
	allTokens := tokenDownload{
		SchemaVersion: SchemaVersion,
		PublicationID: publication.PublicationID,
		SnapshotID:    publication.CurrentSnapshot.ID,
		Subjects:      subjects,
	}
	if err := writeDownloadJSON(filepath.Join(downloadsDirectory, "tokens.json"), allTokens); err != nil {
		return err
	}

	for _, subject := range subjects {
		path := subjectPaths[subject.Key]
		if err := writeDownloadJSON(path, subjectDownload{
			SchemaVersion: SchemaVersion,
			PublicationID: publication.PublicationID,
			SnapshotID:    publication.CurrentSnapshot.ID,
			Subject:       subject,
		}); err != nil {
			return err
		}
	}
	return nil
}

func subjectDownloadPath(root string, key SubjectKey) (string, error) {
	if err := validateDownloadPart(string(key.Kind), "subject kind"); err != nil {
		return "", err
	}
	if key.Slug == "" || strings.HasPrefix(key.Slug, "/") || strings.Contains(key.Slug, `\`) {
		return "", fmt.Errorf("invalid subject slug %q for normalized download", key.Slug)
	}
	parts := strings.Split(key.Slug, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("invalid subject slug %q for normalized download", key.Slug)
		}
	}
	fileName := parts[len(parts)-1] + ".json"
	pathParts := append([]string{root, string(key.Kind)}, parts[:len(parts)-1]...)
	return filepath.Join(append(pathParts, fileName)...), nil
}

func writeDownloadJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode download %s: %w", path, err)
	}
	body = append(body, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create download directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write download %s: %w", path, err)
	}
	return nil
}

func validateDownloadPart(value, label string) error {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, `/\\`) {
		return fmt.Errorf("invalid %s %q for normalized download", label, value)
	}
	return nil
}
