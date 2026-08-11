package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sojupie/m3-source/internal/m3crawl"
	"github.com/sojupie/m3-source/internal/m3snapshot"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "snapshot" {
		if err := runSnapshot(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := runCrawl(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCrawl(args []string) error {
	var config m3crawl.Config
	flags := flag.NewFlagSet("m3source", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&config.Seed, "seed", m3crawl.DefaultSeed, "starting Material page")
	flags.StringVar(&config.Sitemap, "sitemap", "", "sitemap or sitemap-index URL override")
	flags.StringVar(&config.PathPrefix, "path-prefix", "/components/", "same-origin page path prefix")
	flags.BoolVar(&config.SpecsOnly, "specs-only", true, "only open component /specs pages (use --specs-only=false to inspect other pages)")
	flags.StringVar(&config.CatalogPath, "catalog", m3crawl.DefaultCatalogPath, "curated page catalog used by default")
	flags.BoolVar(&config.Discover, "discover", false, "bypass the catalog, crawl live discovery, and report catalog drift")
	flags.BoolVar(&config.UpdateCatalog, "update-catalog", false, "crawl live discovery and replace the curated catalog")
	flags.BoolVar(&config.NoCatalog, "no-catalog", false, "bypass the curated catalog without comparing or updating it")
	flags.IntVar(&config.MaxPages, "max-pages", 500, "maximum candidate pages")
	flags.IntVar(&config.Concurrency, "concurrency", 2, "parallel browser pages")
	flags.DurationVar(&config.PageTimeout, "page-timeout", 10*time.Second, "time to observe each page")
	flags.StringVar(&config.CacheDir, "cache-dir", ".cache/m3", "content-addressed cache directory")
	flags.StringVar(&config.RunDir, "run-dir", "sources/runs/m3", "run manifest directory")
	flags.BoolVar(&config.Refresh, "refresh", false, "refetch payloads already in the cache")
	flags.BoolVar(&config.DryRun, "dry-run", false, "discover pages and payload URLs without fetching payloads")
	flags.BoolVar(&config.Headed, "headed", false, "show Chromium while observing pages")
	flags.StringVar(&config.ChromePath, "chrome-path", "", "optional Chrome/Chromium executable path")
	var all bool
	flags.BoolVar(&all, "all", false, "legacy alias for --no-catalog")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if all {
		config.NoCatalog = true
	}
	return m3crawl.Run(context.Background(), config)
}

func runSnapshot(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("snapshot requires build or verify")
	}
	switch args[0] {
	case "build":
		flags := flag.NewFlagSet("m3source snapshot build", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		config := m3snapshot.BuildConfig{}
		flags.StringVar(&config.RunManifestPath, "run-manifest", "", "completed crawler run manifest")
		flags.StringVar(&config.CatalogPath, "catalog", "", "curated catalog (defaults to the run option)")
		flags.StringVar(&config.OutputDir, "output", "", "empty directory for the immutable snapshot")
		flags.StringVar(&config.SnapshotID, "id", "", "immutable snapshot identifier")
		flags.StringVar(&config.CreatedAt, "created-at", "", "snapshot creation time (defaults to the run start time)")
		flags.StringVar(&config.ExtractorName, "extractor", "m3source", "extractor name")
		flags.StringVar(&config.ExtractorVersion, "extractor-version", "0.1.0", "extractor version")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		manifest, err := m3snapshot.Build(config)
		if err != nil {
			return err
		}
		fmt.Printf("Built snapshot %s in %s\n", manifest.SnapshotID, config.OutputDir)
		fmt.Printf("Manifest SHA-256: %s\n", manifestDigest(config.OutputDir))
		return nil
	case "verify":
		flags := flag.NewFlagSet("m3source snapshot verify", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		directory := ""
		flags.StringVar(&directory, "dir", "", "snapshot directory")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		manifest, err := m3snapshot.Verify(directory)
		if err != nil {
			return err
		}
		fmt.Printf("Verified snapshot %s\n", manifest.SnapshotID)
		fmt.Printf("Manifest SHA-256: %s\n", manifestDigest(directory))
		return nil
	default:
		return fmt.Errorf("unknown snapshot command %q", args[0])
	}
}

func manifestDigest(directory string) string {
	body, err := os.ReadFile(directory + string(os.PathSeparator) + "manifest.json")
	if err != nil {
		return "unavailable: " + err.Error()
	}
	return fmt.Sprintf("%x", sha256.Sum256(body))
}
