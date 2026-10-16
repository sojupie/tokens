package publish

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sojupie/tokens/source/tokensnapshot"
)

type historySnapshot struct {
	directory string
	view      SnapshotView
	created   time.Time
}

// Rebuild history from retained snapshots, comparing each consecutive pair once.
// The first snapshot is a baseline, not a batch of invented additions.
func buildHistory(config BuildConfig, publication *Publication, currentIndex map[string]tokensnapshot.SemanticRecord) error {
	snapshots, err := historySnapshots(config, publication.CurrentSnapshot)
	if err != nil {
		return err
	}
	publication.Snapshots = nil
	publication.Changes = nil
	owners := historyTokenOwners(publication.Subjects)
	var previousIndex map[string]tokensnapshot.SemanticRecord
	for index, snapshot := range snapshots {
		nextIndex := currentIndex
		if snapshot.view.ID != publication.CurrentSnapshot.ID {
			nextIndex = nil
			if config.readHistoryCache("index", snapshot.view.ManifestHash, &nextIndex) {
				if _, err := tokensnapshot.VerifyIntegrity(snapshot.directory); err != nil {
					return err
				}
				body, err := os.ReadFile(filepath.Join(snapshot.directory, "manifest.json"))
				if err != nil || sha256Hex(body) != snapshot.view.ManifestHash {
					return fmt.Errorf("history manifest changed during publication: %s", snapshot.view.ID)
				}
				config.report("Reusing history index: " + snapshot.view.ID)
			} else {
				config.report("Verifying and indexing history snapshot: " + snapshot.view.ID)
				input, err := loadSnapshot(snapshot.directory)
				if err != nil {
					return fmt.Errorf("load history snapshot %s: %w", snapshot.view.ID, err)
				}
				if input.ManifestHash != snapshot.view.ManifestHash {
					return fmt.Errorf("history manifest changed during publication: %s", snapshot.view.ID)
				}
				nextIndex = tokensnapshot.IndexSemanticRecords(input.Graph.Records)
				config.writeHistoryCache("index", snapshot.view.ManifestHash, nextIndex)
			}
		} else {
			config.writeHistoryCache("index", snapshot.view.ManifestHash, nextIndex)
		}
		publication.Snapshots = append(publication.Snapshots, snapshot.view)
		if index == 0 {
			previousIndex = nextIndex
			continue
		}
		previous := snapshots[index-1]
		config.report(fmt.Sprintf("Comparing history %d/%d: %s -> %s", index, len(snapshots)-1, previous.view.ID, snapshot.view.ID))
		var semantic tokensnapshot.SemanticDiff
		identity := previous.view.ManifestHash + ":" + snapshot.view.ManifestHash
		if !config.readHistoryCache("diff", identity, &semantic) {
			semantic = tokensnapshot.CompareSemanticIndexes(previousIndex, nextIndex)
			config.writeHistoryCache("diff", identity, semantic)
		} else {
			config.report("Reusing history comparison: " + previous.view.ID + " -> " + snapshot.view.ID)
		}
		diff := tokensnapshot.SnapshotDiff{
			FromSnapshotID: previous.view.ID,
			ToSnapshotID:   snapshot.view.ID,
			Semantic:       semantic,
		}
		previousIndex = nextIndex
		changes := changeViews(diff)
		for i := range changes {
			changes[i].DetectedAt = snapshot.view.CreatedAt
			if owner := owners[changes[i].Name]; owner != "" {
				changes[i].Subject = owner
			} else if owner := owners["component:"+changes[i].Subject]; owner != "" {
				changes[i].Subject = owner
			}
		}
		publication.Changes = append(publication.Changes, changes...)
	}
	sortChanges(publication.Changes, publication.Snapshots)
	return nil
}

// Use actual token ownership, not guessed singular/plural component names.
// Ambiguous shared tokens keep their original subject in the global timeline.
func historyTokenOwners(subjects []SubjectView) map[string]string {
	owners := make(map[string]string)
	add := func(name, slug string) {
		if existing, found := owners[name]; found && existing != slug {
			owners[name] = ""
		} else if !found {
			owners[name] = slug
		}
	}
	for _, subject := range subjects {
		for _, token := range subject.Tokens {
			add(token.SourceName, subject.Key.Slug)
			parts := strings.Split(token.SourceName, ".")
			if subject.Key.Kind == SubjectComponent && len(parts) > 2 && parts[0] == "md" && parts[1] == "comp" {
				add("component:"+parts[2], subject.Key.Slug)
			}
		}
	}
	return owners
}

func historySnapshots(config BuildConfig, current SnapshotView) ([]historySnapshot, error) {
	directory := config.HistoryDir
	if directory == "" {
		directory = filepath.Dir(filepath.Clean(config.SnapshotDir))
	}
	paths := config.HistoryPaths
	if paths == nil {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, fmt.Errorf("read snapshot history: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				paths = append(paths, filepath.Join(directory, entry.Name()))
			}
		}
	}
	cutoff, err := time.Parse(time.RFC3339Nano, current.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("current snapshot date: %w", err)
	}
	currentPath, err := filepath.Abs(config.SnapshotDir)
	if err != nil {
		return nil, err
	}
	snapshots := []historySnapshot{{currentPath, current, cutoff}}
	ids := map[string]bool{current.ID: true}
	for _, historyPath := range paths {
		path, err := filepath.Abs(historyPath)
		if err != nil {
			return nil, err
		}
		if path == currentPath {
			continue
		}
		body, err := os.ReadFile(filepath.Join(path, "manifest.json"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read history manifest %s: %w", path, err)
		}
		var manifest tokensnapshot.Manifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			return nil, fmt.Errorf("decode history manifest %s: %w", path, err)
		}
		if manifest.SchemaVersion != tokensnapshot.SchemaVersion {
			config.report(fmt.Sprintf("Excluded history snapshot %s: unsupported schema %d", filepath.Base(path), manifest.SchemaVersion))
			continue
		}
		created, err := time.Parse(time.RFC3339Nano, manifest.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("history snapshot %s date: %w", filepath.Base(path), err)
		}
		if created.After(cutoff) {
			continue
		}
		if ids[manifest.SnapshotID] {
			return nil, fmt.Errorf("duplicate history snapshot ID %q", manifest.SnapshotID)
		}
		ids[manifest.SnapshotID] = true
		snapshots = append(snapshots, historySnapshot{path, snapshotView(manifest, sha256Hex(body)), created})
	}
	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].created.Equal(snapshots[j].created) {
			// The selected snapshot ends the timeline even if another capture shares its date.
			if snapshots[i].view.ID == current.ID {
				return false
			}
			if snapshots[j].view.ID == current.ID {
				return true
			}
			return snapshots[i].view.ID < snapshots[j].view.ID
		}
		return snapshots[i].created.Before(snapshots[j].created)
	})
	return snapshots, nil
}

func (config BuildConfig) report(message string) {
	if config.Progress != nil {
		config.Progress(message)
	}
}

func sortChanges(changes []ChangeView, snapshots []SnapshotView) {
	rank := make(map[string]int, len(snapshots))
	for i, snapshot := range snapshots {
		rank[snapshot.ID] = i
	}
	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].SnapshotID != changes[j].SnapshotID {
			return rank[changes[i].SnapshotID] > rank[changes[j].SnapshotID]
		}
		if changes[i].Subject != changes[j].Subject {
			return changes[i].Subject < changes[j].Subject
		}
		if changes[i].Name != changes[j].Name {
			return changes[i].Name < changes[j].Name
		}
		if changes[i].Type != changes[j].Type {
			return changes[i].Type < changes[j].Type
		}
		return changes[i].ID < changes[j].ID
	})
}
