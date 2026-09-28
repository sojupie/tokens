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
	SnapshotDir         string
	PreviousSnapshotDir string
	OutputDir           string
	PublicationID       string
	GeneratedAt         string
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
	if config.PreviousSnapshotDir != "" {
		previous, err := loadSnapshot(config.PreviousSnapshotDir)
		if err != nil {
			return Publication{}, fmt.Errorf("load previous snapshot: %w", err)
		}
		if previous.Manifest.SnapshotID == input.Manifest.SnapshotID {
			return Publication{}, errors.New("previous snapshot must have a different snapshot ID")
		}
		diff, err := tokensnapshot.Diff(config.PreviousSnapshotDir, config.SnapshotDir)
		if err != nil {
			return Publication{}, fmt.Errorf("compare snapshots for publication history: %w", err)
		}
		publication.Changes = changeViews(diff)
		publication.Snapshots = []SnapshotView{
			snapshotView(previous.Manifest, previous.ManifestHash),
			snapshotView(input.Manifest, input.ManifestHash),
		}
	}
	stagingDir, err := createStagingDir(config.OutputDir)
	if err != nil {
		return Publication{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(stagingDir)
		}
	}()
	if err := writePublicationJSON(stagingDir, publication); err != nil {
		return Publication{}, err
	}
	if err := writeDatabase(filepath.Join(stagingDir, "tokens.sqlite"), publication); err != nil {
		return Publication{}, err
	}
	if err := writeDownloads(stagingDir, publication); err != nil {
		return Publication{}, err
	}
	if err := replaceOutput(stagingDir, config.OutputDir); err != nil {
		return Publication{}, err
	}
	committed = true
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

func createStagingDir(outputPath string) (string, error) {
	parent := filepath.Dir(filepath.Clean(outputPath))
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("create publication parent directory: %w", err)
	}
	return os.MkdirTemp(parent, ".sourceloom-publication-")
}

func replaceOutput(stagingPath, outputPath string) error {
	outputInfo, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		if err := os.Rename(stagingPath, outputPath); err != nil {
			return fmt.Errorf("install publication: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect publication output: %w", err)
	}
	if !outputInfo.IsDir() {
		return fmt.Errorf("publication output is not a directory: %s", outputPath)
	}

	parent := filepath.Dir(filepath.Clean(outputPath))
	backupPath, err := os.MkdirTemp(parent, ".sourceloom-publication-backup-")
	if err != nil {
		return fmt.Errorf("create publication backup path: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("prepare publication backup path: %w", err)
	}
	if err := os.Rename(outputPath, backupPath); err != nil {
		return fmt.Errorf("stage existing publication: %w", err)
	}
	if err := os.Rename(stagingPath, outputPath); err != nil {
		_ = os.Rename(backupPath, outputPath)
		return fmt.Errorf("install publication: %w", err)
	}
	if err := os.RemoveAll(backupPath); err != nil {
		return fmt.Errorf("remove previous publication: %w", err)
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
