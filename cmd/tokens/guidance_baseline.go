package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type guidanceBaseline struct {
	BaseCommit    string `json:"baseCommit"`
	Archive       string `json:"archive"`
	ArchiveSHA256 string `json:"archiveSha256"`
	CreatedAt     string `json:"createdAt"`
}
type guidanceBaselines struct {
	SchemaVersion int                         `json:"schemaVersion"`
	Baselines     map[string]guidanceBaseline `json:"baselines"`
}

// A private Git index snapshots the actual worktree, including nonignored new
// files, without staging or committing the user's pending changes.
func captureGuidanceBaseline(args []string) error {
	flags := flag.NewFlagSet("tokens guidance baseline", flag.ContinueOnError)
	system := flags.String("design-system", "material", "guidance design system")
	output := flags.String("output", "snapshots/implementation-baselines", "local snapshot directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *system == "" || strings.ContainsAny(*system, "/\\.") {
		return fmt.Errorf("invalid baseline arguments")
	}
	index, err := os.CreateTemp("", "tokens-baseline-index-")
	if err != nil {
		return err
	}
	index.Close()
	os.Remove(index.Name())
	defer os.Remove(index.Name())
	run := func(args ...string) ([]byte, error) {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_INDEX_FILE="+index.Name())
		return cmd.Output()
	}
	head, err := run("rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if _, err = run("read-tree", "HEAD"); err != nil {
		return err
	}
	if _, err = run("add", "-A", "--", "."); err != nil {
		return err
	}
	tree, err := run("write-tree")
	if err != nil {
		return err
	}
	id := strings.TrimSpace(string(tree))
	if err = os.MkdirAll(*output, 0755); err != nil {
		return err
	}
	archive := filepath.Join(*output, id+".tar")
	if _, err = os.Stat(archive); os.IsNotExist(err) {
		if _, err = run("archive", "--format=tar", "--output="+archive, id); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	raw, err := os.ReadFile(archive)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	registryPath := filepath.Join("forge", *system, "guidance", "baselines.json")
	registry := guidanceBaselines{SchemaVersion: 1, Baselines: map[string]guidanceBaseline{}}
	if raw, err = os.ReadFile(registryPath); err == nil {
		if err = json.Unmarshal(raw, &registry); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if existing, ok := registry.Baselines[id]; ok && existing.ArchiveSHA256 != hex.EncodeToString(sum[:]) {
		return fmt.Errorf("baseline archive hash changed: %s", archive)
	}
	if _, ok := registry.Baselines[id]; !ok {
		registry.Baselines[id] = guidanceBaseline{BaseCommit: strings.TrimSpace(string(head)), Archive: filepath.ToSlash(archive), ArchiveSHA256: hex.EncodeToString(sum[:]), CreatedAt: time.Now().UTC().Format(time.RFC3339)}
		if err = os.MkdirAll(filepath.Dir(registryPath), 0755); err != nil {
			return err
		}
		raw, err = json.MarshalIndent(registry, "", "  ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(registryPath, append(raw, '\n'), 0644); err != nil {
			return err
		}
	}
	fmt.Printf("Assessment baseline: %s\nArchive: %s\n", id, archive)
	return nil
}
