// Command dispatch stays separate from command implementations so CLI entry points can be reviewed in one place.
package main

import (
	"fmt"

	"os"

	"strings"
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
	case "site-shell":
		return runSiteShell(args[1:])
	case "illustrations":
		return runIllustrations(args[1:])
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
  tokens guidance [--component SLUG] [--url URL] [--list] [flags]
  tokens guidance baseline [--design-system material] [flags]
  tokens guidance validate [--file PATH] [--against REF]
  tokens publish [flags]
  tokens export [--published directory] [--context Name=Value] [--output directory]
  tokens update [run|status|inspect|approve|retry|notify] [flags]
  tokens serve [flags]
  tokens illustrations [--published directory] [--assets directory] [--chrome-path path]
  tokens site-shell --out directory [--catalog URL]
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

const snapshotHelp = `Usage:
  tokens snapshot create [flags]
  tokens snapshot verify [flags]
  tokens snapshot compare [flags]

Defaults:
  create  uses the newest crawl receipt and creates a timestamped snapshot
  verify  verifies the newest snapshot
  compare compares the two latest snapshot runs`
