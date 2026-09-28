package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sojupie/m3-source/internal/publish"
)

func runPublish(args []string) error {
	flags := flag.NewFlagSet("sourceloom publish", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: sourceloom publish [flags]")
		_, _ = fmt.Fprintln(flags.Output(), "\nBuilds an offline publication from one verified snapshot.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	snapshotDir := ""
	previousSnapshotDir := ""
	outputDir := "published"
	publicationID := ""
	flags.StringVar(&snapshotDir, "snapshot-dir", "", "verified snapshot directory (default: newest snapshot)")
	flags.StringVar(&previousSnapshotDir, "previous-snapshot", previousSnapshotDir, "verified snapshot to compare for change history")
	flags.StringVar(&outputDir, "output", outputDir, "empty publication output directory")
	flags.StringVar(&publicationID, "id", "", "publication identifier (default: snapshot identifier)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("publish does not accept positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if snapshotDir == "" {
		var err error
		snapshotDir, err = newestSnapshotDirectory()
		if err != nil {
			return err
		}
	}
	fmt.Printf("Using verified snapshot: %s\n", snapshotDir)
	fmt.Printf("Publication output: %s\n", outputDir)
	publicationResult, err := publish.Build(publish.BuildConfig{
		SnapshotDir:         snapshotDir,
		PreviousSnapshotDir: previousSnapshotDir,
		OutputDir:           outputDir,
		PublicationID:       publicationID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Published %s\n", publicationResult.PublicationID)
	fmt.Printf("Subjects: %d\n", len(publicationResult.Subjects))
	fmt.Printf("Source payloads: %d\n", len(publicationResult.SourcePayloads))
	return nil
}
