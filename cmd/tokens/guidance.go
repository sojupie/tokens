package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sojupie/tokens/source/crawl"
	"github.com/sojupie/tokens/source/guidance"
)

type guidanceURLs []string

func (u *guidanceURLs) String() string         { return fmt.Sprint([]string(*u)) }
func (u *guidanceURLs) Set(value string) error { *u = append(*u, value); return nil }

func runGuidance(args []string) error {
	if len(args) > 0 && args[0] == "baseline" {
		return captureGuidanceBaseline(args[1:])
	}
	if len(args) > 0 && args[0] == "validate" {
		return validateGuidance(args[1:])
	}
	flags := flag.NewFlagSet("tokens guidance", flag.ContinueOnError)
	var cfg guidance.Config
	var urls, components guidanceURLs
	catalog := flags.String("page-catalog", crawl.DefaultPageCatalogPath, "existing token crawler page catalog")
	list := flags.Bool("list", false, "print selected URLs without capturing")
	flags.Var(&components, "component", "component slug; repeat to select several, defaults to all catalog components")
	flags.Var(&urls, "url", "documentation URL; repeat to capture several pages")
	flags.StringVar(&cfg.Output, "output", "snapshots/material-guidance", "versioned capture directory")
	flags.StringVar(&cfg.Selector, "selector", "main", "rendered documentation CSS selector")
	flags.StringVar(&cfg.ChromePath, "chrome-path", "", "Chrome/Chromium executable path")
	flags.DurationVar(&cfg.Timeout, "page-timeout", 45*time.Second, "maximum render and extraction time per page")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("guidance accepts URLs through --url")
	}
	if len(urls) == 0 || len(components) > 0 {
		tokenPages, err := crawl.DiscoverPages(context.Background(), crawl.Config{PageCatalogPath: *catalog})
		if err != nil {
			return err
		}
		pages, err := guidance.MaterialComponentPages(tokenPages, components)
		if err != nil {
			return err
		}
		urls = append(guidanceURLs(pages), urls...)
	}
	cfg.URLs = urls
	if *list {
		seen := map[string]bool{}
		for _, raw := range cfg.URLs {
			if !seen[raw] {
				fmt.Println(raw)
				seen[raw] = true
			}
		}
		return nil
	}
	path, err := guidance.Capture(context.Background(), cfg)
	if path != "" {
		fmt.Println("Documentation capture:", path)
	}
	return err
}

// Validation stays offline: Git resolves implementation commits; captured content
// and the truth of assessment evidence still require curation.
func validateGuidance(args []string) error {
	flags := flag.NewFlagSet("tokens guidance validate", flag.ContinueOnError)
	file := flags.String("file", "", "one canonical checks.json; defaults to every component")
	against := flags.String("against", "HEAD", "Git revision used to check retained IDs and append-only history")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("guidance validate does not accept positional arguments")
	}
	if _, err := exec.Command("git", "rev-parse", "--verify", *against+"^{commit}").Output(); err != nil {
		return fmt.Errorf("invalid comparison commit %q", *against)
	}
	files := []string{*file}
	if *file == "" {
		var err error
		files, err = filepath.Glob("forge/*/guidance/*/checks.json")
		if err != nil {
			return err
		}
	}
	if *file == "" {
		archives, err := filepath.Glob("forge/*/guidance/*/retired-checks.json")
		if err != nil {
			return err
		}
		for _, archive := range archives {
			active := filepath.Join(filepath.Dir(archive), "checks.json")
			if _, err := os.Stat(active); err != nil {
				return fmt.Errorf("archive requires its active checks.json file: %s", archive)
			}
		}
		tracked, err := exec.Command("git", "ls-tree", "-r", "--name-only", *against, "--", "forge").Output()
		if err != nil {
			return err
		}
		for _, path := range strings.Split(strings.TrimSpace(string(tracked)), "\n") {
			match, _ := filepath.Match("forge/*/guidance/*/checks.json", path)
			if match {
				if _, err := os.Stat(path); err != nil {
					return fmt.Errorf("canonical guidance file was deleted or is unreadable: %s; retain its checks and record removals", path)
				}
			}
		}
	}
	if len(files) == 0 {
		return errors.New("no component checks.json files found")
	}
	for _, path := range files {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return err
		}
		path = filepath.ToSlash(relative)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var previous []byte
		if exec.Command("git", "cat-file", "-e", *against+":"+path).Run() == nil {
			previous, err = exec.Command("git", "show", *against+":"+path).Output()
			if err != nil {
				return err
			}
		}
		archivePath := filepath.ToSlash(filepath.Join(filepath.Dir(path), "retired-checks.json"))
		retired, err := os.ReadFile(archivePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		var previousRetired []byte
		if exec.Command("git", "cat-file", "-e", *against+":"+archivePath).Run() == nil {
			previousRetired, err = exec.Command("git", "show", *against+":"+archivePath).Output()
			if err != nil {
				return err
			}
		}
		summary, err := guidance.ValidateCheckFiles(data, retired, previous, previousRetired)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		canonical := "forge/" + summary.DesignSystem + "/guidance/" + summary.Component + "/checks.json"
		if path != canonical {
			return fmt.Errorf("component checks belong at %s, got %s", canonical, path)
		}
		if len(summary.Baselines) > 0 {
			baselineFile := "forge/" + summary.DesignSystem + "/guidance/baselines.json"
			raw, err := os.ReadFile(baselineFile)
			if err != nil {
				return err
			}
			var registry guidanceBaselines
			if err = json.Unmarshal(raw, &registry); err != nil {
				return err
			}
			for _, id := range summary.Baselines {
				b, ok := registry.Baselines[id]
				if !ok {
					return fmt.Errorf("%s: unknown assessment baseline %s", path, id)
				}
				if len(b.ArchiveSHA256) != 64 || b.Archive == "" {
					return fmt.Errorf("%s: invalid baseline archive record", baselineFile)
				}
				if err = exec.Command("git", "cat-file", "-e", b.BaseCommit+"^{commit}").Run(); err != nil {
					return fmt.Errorf("baseline base commit is absent: %s", b.BaseCommit)
				}
			}
		}
		for _, commit := range summary.Commits {
			if err := exec.Command("git", "cat-file", "-e", commit+"^{commit}").Run(); err != nil {
				return fmt.Errorf("%s: implementation commit does not exist locally: %s", path, commit)
			}
		}
		fmt.Printf("%s: valid format and history; %d active checks, %d removed, %d withdrawn, %d pending/unclear sections\n", path, summary.Checks, summary.Removed, summary.Withdrawn, summary.PendingSections)
	}
	fmt.Println("Validation does not execute evidence or establish component conformance.")
	return nil
}
