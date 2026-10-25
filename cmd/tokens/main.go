package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sojupie/tokens/source/crawl"
	"github.com/sojupie/tokens/source/tokensnapshot"
)

const snapshotRoot = "snapshots/material-token-data"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return runCrawl(nil)
	}
	switch args[0] {
	case "help", "-h", "--help":
		printTopHelp()
		return nil
	case "guidance":
		return runGuidance(args[1:])
	case "crawl":
		return runCrawl(args[1:])
	case "export":
		return runExport(args[1:])
	case "publish":
		return runPublish(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "usage":
		return runUsage(args[1:])
	case "forge":
		return runForge(args[1:])
	case "serve":
		return runServe(args[1:])
	case "snapshot":
		return runSnapshot(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			return runCrawl(args)
		}
		return fmt.Errorf("unknown command %q\n\n%s", args[0], topHelp)
	}
}

const topHelp = `Usage:
  tokens crawl [flags]
  tokens guidance --url URL [--url URL] [flags]
  tokens guidance validate [--file PATH] [--against REF]
  tokens publish [flags]
  tokens export [--published directory] [--context Name=Value] [--output directory]
  tokens update [run|status|inspect|approve|retry|notify] [flags]
  tokens serve [flags]
  tokens usage [--since 7d] [--send | --weekly] [flags]
  tokens forge inspect [--assets directory] [--output file]
  tokens snapshot create [flags]
  tokens snapshot verify [flags]
  tokens snapshot compare [flags]

Run "tokens <command> --help" for command-specific options.

With no arguments, tokens runs crawl.`

func printTopHelp() {
	fmt.Println(topHelp)
}

func runCrawl(args []string) error {
	var config crawl.Config
	flags := flag.NewFlagSet("tokens crawl", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: tokens crawl [flags]")
		_, _ = fmt.Fprintln(flags.Output(), "\nCrawls reviewed Material token-source pages and downloads source payloads.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	flags.StringVar(&config.Seed, "seed", crawl.DefaultSeed, "starting Material page")
	flags.StringVar(&config.Sitemap, "sitemap", "", "sitemap or sitemap-index URL override")
	flags.StringVar(&config.PathPrefix, "path-prefix", "/", "same-origin page path prefix used during live discovery")
	flags.StringVar(&config.Scope, "scope", crawl.ScopeComponentSpecs, "live discovery scope: component-specs or token-source-pages")
	flags.StringVar(&config.PageCatalogPath, "page-catalog", crawl.DefaultPageCatalogPath, "page URL catalog used by default")
	flags.BoolVar(&config.Discover, "discover", false, "bypass the page URL catalog and report page-catalog drift")
	flags.BoolVar(&config.UpdatePageCatalog, "update-page-catalog", false, "replace the page URL catalog with reviewed live discovery")
	flags.BoolVar(&config.NoPageCatalog, "no-page-catalog", false, "bypass the page URL catalog without comparing or updating it")
	flags.IntVar(&config.MaxPages, "max-pages", 500, "maximum candidate pages")
	flags.IntVar(&config.Concurrency, "concurrency", 1, "parallel browser pages")
	flags.DurationVar(&config.PageTimeout, "page-timeout", 30*time.Second, "maximum observation time per page; finishes early after network activity settles")
	flags.StringVar(&config.CacheDir, "cache-dir", ".cache/material-token-table-responses", "source-payload cache directory")
	flags.StringVar(&config.ResultsDir, "results-dir", "crawl-results", "crawl-receipt directory")
	flags.BoolVar(&config.Refresh, "refresh", false, "re-download source payloads already in the cache")
	flags.BoolVar(&config.DryRun, "dry-run", false, "discover pages and payload URLs without downloading payloads")
	flags.BoolVar(&config.RetryFailed, "retry-failed", false, "retry failed pages from the newest crawl receipt")
	flags.StringVar(&config.RetryFailedFrom, "retry-failed-from", "", "retry failed pages from this crawl receipt")
	flags.BoolVar(&config.Headed, "headed", false, "show Chromium while observing pages")
	flags.StringVar(&config.ChromePath, "chrome-path", "", "optional Chrome/Chromium executable path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("crawl does not accept positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	return crawl.Run(context.Background(), config)
}

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

const snapshotHelp = `Usage:
  tokens snapshot create [flags]
  tokens snapshot verify [flags]
  tokens snapshot compare [flags]

Defaults:
  create  uses the newest crawl receipt and creates a timestamped snapshot
  verify  verifies the newest snapshot
  compare compares the two latest snapshot runs`

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

func writeSnapshotCompareReport(writer io.Writer, diff tokensnapshot.SnapshotDiff, verbose bool) {
	fmt.Fprintf(writer, "Compared %s to %s\n", diff.FromSnapshotID, diff.ToSnapshotID)
	fmt.Fprintf(writer, "Material changes: +%d -%d changed %d\n", len(diff.Semantic.Added), len(diff.Semantic.Removed), len(diff.Semantic.Changed))
	fmt.Fprintf(writer, "Source payloads refreshed: %d\n", len(diff.SourcePayloads.Changed))
	fmt.Fprintf(writer, "Pages with changed results: %d\n", len(diff.PageResults.Changed))
	if !verbose {
		return
	}
	printSemanticChanges(writer, diff.Semantic)
}

func printSemanticChanges(writer io.Writer, diff tokensnapshot.SemanticDiff) {
	printSemanticChangeSet(writer, "Added", diff.Added)
	printSemanticChangeSet(writer, "Removed", diff.Removed)
	printSemanticChangeSet(writer, "Changed", diff.Changed)
}

func printSemanticChangeSet(writer io.Writer, label string, changes []tokensnapshot.SemanticChange) {
	if len(changes) == 0 {
		return
	}
	fmt.Fprintf(writer, "%s Material changes (%d):\n", label, len(changes))
	currentSubject := ""
	for _, change := range changes {
		if change.SourceSubject != currentSubject {
			currentSubject = change.SourceSubject
			fmt.Fprintf(writer, "  %s:\n", valueOrDash(currentSubject))
		}
		name := change.Name
		if name == "" {
			name = change.Kind
		}
		kind := semanticKindLabel(change.Kind)
		symbol := map[string]string{"Added": "+", "Removed": "-", "Changed": "~"}[label]
		fmt.Fprintf(writer, "    %s %s %s:\n", symbol, kind, name)
		for _, line := range semanticChangeLines(change, label) {
			fmt.Fprintf(writer, "      %s\n", line)
		}
	}
}

func semanticChangeLines(change tokensnapshot.SemanticChange, label string) []string {
	if label == "Added" {
		return semanticValueLines(change.After)
	}
	if label == "Removed" {
		return semanticValueLines(change.Before)
	}
	return semanticDiffLines(change.Before, change.After)
}

func semanticValueLines(value string) []string {
	if value == "" {
		return []string{"-"}
	}
	if value == `{"undefined":true}` {
		return []string{"undefined"}
	}
	leaves := flattenJSON(value)
	if len(leaves) == 0 {
		return []string{value}
	}
	lines := make([]string, 0, len(leaves))
	for _, key := range sortedKeys(leaves) {
		lines = append(lines, key+": "+leaves[key])
	}
	return lines
}

func semanticDiffLines(before, after string) []string {
	oldLeaves := flattenJSON(before)
	newLeaves := flattenJSON(after)
	keys := make(map[string]string, len(oldLeaves)+len(newLeaves))
	for key := range oldLeaves {
		keys[key] = ""
	}
	for key := range newLeaves {
		keys[key] = ""
	}
	lines := make([]string, 0)
	for _, key := range sortedKeys(keys) {
		oldValue, newValue := oldLeaves[key], newLeaves[key]
		if oldValue == newValue {
			continue
		}
		if oldValue == "" {
			oldValue = "-"
		}
		if newValue == "" {
			newValue = "-"
		}
		lines = append(lines, key+": "+oldValue+" -> "+newValue)
	}
	if len(lines) == 0 {
		return []string{valueOrDash(before) + " -> " + valueOrDash(after)}
	}
	return lines
}

func semanticKindLabel(kind string) string {
	switch kind {
	case "contextualReferenceTree":
		return "resolved value"
	case "component":
		return "component"
	case "tokenSet":
		return "token set"
	default:
		return kind
	}
}

func flattenJSON(value string) map[string]string {
	result := map[string]string{}
	var raw interface{}
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()
	if decoder.Decode(&raw) != nil {
		return result
	}
	flattenJSONValue(raw, "", result)
	return result
}

func flattenJSONValue(value interface{}, prefix string, result map[string]string) {
	switch value := value.(type) {
	case map[string]interface{}:
		if len(value) == 0 {
			result[prefixOrValue(prefix)] = "{}"
			return
		}
		if measurement, ok := value["value"]; ok {
			if unit, hasUnit := value["unit"]; hasUnit {
				result[prefixOrValue(prefix)] = scalarText(measurement) + " " + scalarText(unit)
				return
			}
			if tag, hasTag := value["tag"]; hasTag {
				result[prefixOrValue(prefix)] = scalarText(measurement) + " (" + scalarText(tag) + ")"
				return
			}
		}
		for key, child := range value {
			childPrefix := key
			if prefix != "" {
				childPrefix = prefix + "." + key
			}
			flattenJSONValue(child, childPrefix, result)
		}
	case []interface{}:
		body, _ := json.Marshal(value)
		result[prefixOrValue(prefix)] = string(body)
	default:
		result[prefixOrValue(prefix)] = fmt.Sprint(value)
	}
}

func scalarText(value interface{}) string {
	return fmt.Sprint(value)
}

func prefixOrValue(prefix string) string {
	if prefix == "" {
		return "value"
	}
	return prefix
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func writeComparisonReport(path, report string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create comparison report directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
		return fmt.Errorf("write comparison report: %w", err)
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
