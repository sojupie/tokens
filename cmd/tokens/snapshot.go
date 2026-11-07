// Snapshot commands select inputs and invoke build, verification or comparison without duplicating their source rules.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sojupie/tokens/source/crawl"
	"github.com/sojupie/tokens/source/tokensnapshot"
)

func runSnapshot(args []string) error {
	if len(args) == 0 {
		printSnapshotHelp()
		return errors.New("snapshot requires create, verify, or compare")
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printSnapshotHelp()
		return nil
	}
	switch args[0] {
	case "create":
		return runSnapshotCreate(args[1:])
	case "verify":
		return runSnapshotVerify(args[1:])
	case "compare":
		return runSnapshotCompare(args[1:])
	default:
		return fmt.Errorf("unknown snapshot command %q\n\n%s", args[0], snapshotHelp)
	}
}

func printSnapshotHelp() {
	fmt.Println(snapshotHelp)
}

func runSnapshotCreate(args []string) error {
	flags := flag.NewFlagSet("tokens snapshot create", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: tokens snapshot create [flags]")
		_, _ = fmt.Fprintln(flags.Output(), "\nCreates an immutable snapshot from a crawl receipt.")
		_, _ = fmt.Fprintln(flags.Output(), "Without flags, the newest crawl receipt is used and the ID/output are generated.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	config := tokensnapshot.BuildConfig{}
	flags.StringVar(&config.CrawlReceiptPath, "crawl-receipt", "", "crawl receipt (default: newest receipt in crawl-results)")
	flags.StringVar(&config.PageCatalogPath, "page-catalog", "", "page URL catalog (default: the crawl receipt's catalog)")
	flags.StringVar(&config.OutputDir, "output", "", "empty output directory (default: snapshots/material-token-data/<id>)")
	flags.StringVar(&config.SnapshotID, "id", "", "snapshot identifier (default: UTC crawl date and HHMM)")
	flags.StringVar(&config.CreatedAt, "created-at", "", "snapshot creation time (default: crawl start time)")
	flags.StringVar(&config.ExtractorName, "extractor", "tokens", "extractor name")
	flags.StringVar(&config.ExtractorVersion, "extractor-version", "0.1.0", "extractor version")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("snapshot create does not accept positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if err := fillSnapshotDefaults(&config); err != nil {
		return err
	}
	fmt.Printf("Using crawl receipt: %s\n", config.CrawlReceiptPath)
	fmt.Printf("Snapshot output: %s\n", config.OutputDir)
	manifest, err := tokensnapshot.Build(config)
	if err != nil {
		return err
	}
	fmt.Printf("Created snapshot %s\n", manifest.SnapshotID)
	fmt.Printf("Manifest SHA-256: %s\n", manifestDigest(config.OutputDir))
	return nil
}

func runSnapshotVerify(args []string) error {
	flags := flag.NewFlagSet("tokens snapshot verify", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: tokens snapshot verify [snapshot-directory]")
		_, _ = fmt.Fprintln(flags.Output(), "\nVerifies the newest snapshot unless a directory is supplied.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	directory := ""
	flags.StringVar(&directory, "dir", "", "snapshot directory (default: newest snapshot)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 1 {
		return errors.New("snapshot verify accepts at most one directory")
	}
	if flags.NArg() == 1 {
		if directory != "" {
			return errors.New("snapshot verify accepts the directory either positionally or with --dir, not both")
		}
		directory = flags.Arg(0)
	}
	if directory == "" {
		var err error
		directory, err = newestSnapshotDirectory()
		if err != nil {
			return err
		}
	}
	manifest, err := tokensnapshot.Verify(directory)
	if err != nil {
		return err
	}
	fmt.Printf("Verified snapshot %s\n", manifest.SnapshotID)
	fmt.Printf("Manifest SHA-256: %s\n", manifestDigest(directory))
	return nil
}

func runSnapshotCompare(args []string) error {
	flags := flag.NewFlagSet("tokens snapshot compare", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: tokens snapshot compare [older newer]")
		_, _ = fmt.Fprintln(flags.Output(), "\nCompares the two latest snapshot runs unless directories are supplied.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	fromDir := ""
	toDir := ""
	verbose := false
	output := ""
	flags.StringVar(&fromDir, "from", "", "older snapshot directory")
	flags.StringVar(&toDir, "to", "", "newer snapshot directory")
	flags.BoolVar(&verbose, "verbose", false, "print payload details and record change summaries")
	flags.StringVar(&output, "output", "", "write the same comparison report to this path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 && (fromDir != "" || toDir != "") {
		return errors.New("snapshot compare accepts directories either positionally or with --from/--to, not both")
	}
	if flags.NArg() > 2 {
		return errors.New("snapshot compare accepts exactly two directories")
	}
	if flags.NArg() == 2 {
		fromDir, toDir = flags.Arg(0), flags.Arg(1)
	}
	if fromDir == "" && toDir == "" {
		var err error
		fromDir, toDir, err = newestSnapshotPair()
		if err != nil {
			return err
		}
	} else if fromDir == "" || toDir == "" {
		return errors.New("snapshot compare requires both --from and --to")
	}
	diff, err := tokensnapshot.Diff(fromDir, toDir)
	if err != nil {
		return err
	}
	var report strings.Builder
	writeSnapshotCompareReport(&report, diff, verbose)
	fmt.Print(report.String())
	if output != "" {
		if err := writeComparisonReport(output, report.String()); err != nil {
			return err
		}
		fmt.Printf("Wrote comparison report: %s\n", output)
	}
	return nil
}

func fillSnapshotDefaults(config *tokensnapshot.BuildConfig) error {
	if config.CrawlReceiptPath == "" {
		var err error
		config.CrawlReceiptPath, err = newestJSONFile("crawl-results")
		if err != nil {
			return fmt.Errorf("find crawl receipt: %w", err)
		}
	}
	receipt, err := readCrawlReceipt(config.CrawlReceiptPath)
	if err != nil {
		return err
	}
	if config.SnapshotID == "" {
		config.SnapshotID, err = defaultSnapshotID(receipt)
		if err != nil {
			return err
		}
	}
	if config.OutputDir == "" {
		config.OutputDir = filepath.Join(snapshotRoot, config.SnapshotID)
	}
	return nil
}

func readCrawlReceipt(path string) (crawl.CrawlReceipt, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return crawl.CrawlReceipt{}, fmt.Errorf("read crawl receipt: %w", err)
	}
	var receipt crawl.CrawlReceipt
	if err := json.Unmarshal(body, &receipt); err != nil {
		return crawl.CrawlReceipt{}, fmt.Errorf("parse crawl receipt: %w", err)
	}
	if receipt.SchemaVersion != 2 {
		return crawl.CrawlReceipt{}, fmt.Errorf("unsupported crawl receipt schema version %d", receipt.SchemaVersion)
	}
	return receipt, nil
}

func defaultSnapshotID(receipt crawl.CrawlReceipt) (string, error) {
	started, err := time.Parse(time.RFC3339Nano, receipt.StartedAt)
	if err != nil {
		return "", fmt.Errorf("crawl receipt has invalid startedAt: %w", err)
	}
	return "material-" + started.UTC().Format("2006-01-02-1504"), nil
}

func newestJSONFile(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return "", fmt.Errorf("no crawl receipts found in %s", directory)
	}
	return files[len(files)-1], nil
}

func newestSnapshotDirectory() (string, error) {
	entries, err := snapshotDirectories()
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no snapshots found in %s", snapshotRoot)
	}
	return entries[len(entries)-1], nil
}

func newestSnapshotPair() (string, string, error) {
	entries, err := snapshotDirectories()
	if err != nil {
		return "", "", err
	}
	if len(entries) < 2 {
		return "", "", fmt.Errorf("need at least two snapshots in %s", snapshotRoot)
	}
	return entries[len(entries)-2], entries[len(entries)-1], nil
}

func snapshotDirectories() ([]string, error) {
	entries, err := os.ReadDir(snapshotRoot)
	if err != nil {
		return nil, err
	}
	type snapshotRun struct {
		directory string
		createdAt time.Time
	}
	runs := make([]snapshotRun, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			directory := filepath.Join(snapshotRoot, entry.Name())
			body, readErr := os.ReadFile(filepath.Join(directory, "manifest.json"))
			if readErr != nil {
				continue
			}
			var metadata struct {
				CreatedAt string `json:"createdAt"`
			}
			if json.Unmarshal(body, &metadata) != nil {
				continue
			}
			createdAt, parseErr := time.Parse(time.RFC3339Nano, metadata.CreatedAt)
			if parseErr != nil {
				continue
			}
			runs = append(runs, snapshotRun{directory: directory, createdAt: createdAt})
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].createdAt.Equal(runs[j].createdAt) {
			return runs[i].directory < runs[j].directory
		}
		return runs[i].createdAt.Before(runs[j].createdAt)
	})
	directories := make([]string, 0, len(runs))
	for _, run := range runs {
		directories = append(directories, run.directory)
	}
	return directories, nil
}

func manifestDigest(directory string) string {
	body, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return "unavailable: " + err.Error()
	}
	return fmt.Sprintf("%x", sha256.Sum256(body))
}
