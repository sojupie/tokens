package publish

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sojupie/m3-source/internal/tokensnapshot"
)

type BuildConfig struct {
	SnapshotDir   string
	OutputDir     string
	PublicationID string
	GeneratedAt   string
}

func Build(config BuildConfig) (Publication, error) {
	if err := config.validate(); err != nil {
		return Publication{}, err
	}
	input, err := loadSnapshot(config.SnapshotDir)
	if err != nil {
		return Publication{}, err
	}
	publication, err := project(input)
	if err != nil {
		return Publication{}, err
	}
	publication.PublicationID = config.PublicationID
	if publication.PublicationID == "" {
		publication.PublicationID = input.Manifest.SnapshotID
	}
	publication.GeneratedAt = config.GeneratedAt
	if publication.GeneratedAt == "" {
		publication.GeneratedAt = input.Manifest.CreatedAt
	}
	if _, err := time.Parse(time.RFC3339Nano, publication.GeneratedAt); err != nil {
		return Publication{}, fmt.Errorf("publication generatedAt is not RFC3339: %w", err)
	}
	if err := ensureEmptyDir(config.OutputDir); err != nil {
		return Publication{}, err
	}
	if err := writePublicationJSON(config.OutputDir, publication); err != nil {
		return Publication{}, err
	}
	if err := writeDatabase(filepath.Join(config.OutputDir, "tokens.sqlite"), publication); err != nil {
		return Publication{}, err
	}
	return publication, nil
}

func (config BuildConfig) validate() error {
	if config.SnapshotDir == "" || config.OutputDir == "" {
		return errors.New("snapshot directory and output directory are required")
	}
	if filepath.Clean(config.SnapshotDir) == filepath.Clean(config.OutputDir) {
		return errors.New("snapshot directory and output directory must differ")
	}
	return nil
}

func writePublicationJSON(directory string, publication Publication) error {
	body, err := json.MarshalIndent(publication, "", "  ")
	if err != nil {
		return fmt.Errorf("encode publication metadata: %w", err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(filepath.Join(directory, "publication.json"), body, 0o644); err != nil {
		return fmt.Errorf("write publication metadata: %w", err)
	}
	return nil
}

func ensureEmptyDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create publication output directory: %w", err)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read publication output directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("publication output directory is not empty: %s", path)
	}
	return nil
}

type snapshotInput struct {
	Manifest         tokensnapshot.Manifest
	ManifestHash     string
	Graph            tokensnapshot.Graph
	PayloadDetails   tokensnapshot.SourcePayloadDetails
	LifecycleNotices tokensnapshot.LifecycleNotices
}
