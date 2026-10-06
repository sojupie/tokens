package publish

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sojupie/tokens/internal/tokensnapshot"
)

type BuildConfig struct {
	SnapshotDir     string
	HistoryDir      string
	HistoryCacheDir string
	// HistoryPaths restricts history to explicitly accepted snapshots when non-nil.
	HistoryPaths  []string
	Progress      func(string)
	OutputDir     string
	PublicationID string
	GeneratedAt   string
}

func Build(config BuildConfig) (Publication, error) {
	if err := config.validate(); err != nil {
		return Publication{}, err
	}
	config.report("Verifying and loading current snapshot")
	input, err := loadSnapshot(config.SnapshotDir)
	if err != nil {
		return Publication{}, err
	}
	config.report("Projecting token data")
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
	currentIndex := tokensnapshot.IndexSemanticRecords(input.Graph.Records)
	input = snapshotInput{} // History needs the compact index, not another full graph.
	if err := buildHistory(config, &publication, currentIndex); err != nil {
		return Publication{}, err
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
	config.report("Writing publication JSON")
	if err := writePublicationJSON(stagingDir, publication); err != nil {
		return Publication{}, err
	}
	config.report("Writing SQLite database")
	if err := writeDatabase(filepath.Join(stagingDir, "tokens.sqlite"), publication); err != nil {
		return Publication{}, err
	}
	config.report("Writing token downloads")
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
	return os.MkdirTemp(parent, ".tokens-publication-")
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
	backupPath, err := os.MkdirTemp(parent, ".tokens-publication-backup-")
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
