package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sojupie/tokens/source/crawl"
	"github.com/sojupie/tokens/source/update"
)

func runUpdate(args []string) error {
	command := "run"
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		command = args[0]
		args = args[1:]
	}
	id := ""
	if command == "inspect" || command == "approve" || command == "retry" || command == "notify" {
		if len(args) == 0 || args[0] == "" || args[0][0] == '-' {
			return fmt.Errorf("usage: tokens update %s <run-id> [flags]", command)
		}
		id = args[0]
		args = args[1:]
	}
	flags := flag.NewFlagSet("tokens update "+command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	config := update.Config{MaxChangeFraction: 0.25, Progress: func(message string) { fmt.Println(message) }}
	timeout := time.Hour
	flags.StringVar(&config.StateDir, "state-dir", ".tokens", "saved runs, history cache and current generation")
	if command == "run" || command == "retry" {
		flags.DurationVar(&timeout, "timeout", time.Hour, "total update time limit")
	}
	if command == "run" {
		flags.StringVar(&config.ReceiptPath, "receipt", "", "build from an existing crawl receipt instead of crawling")
		flags.StringVar(&config.HistoryDir, "history-dir", "", "retained snapshots to include (inherited from the active run when unset)")
		flags.StringVar(&config.Crawl.Seed, "seed", crawl.DefaultSeed, "starting Material page")
		flags.StringVar(&config.Crawl.Scope, "scope", crawl.ScopeTokenSourcePages, "page selection scope")
		flags.StringVar(&config.Crawl.PageCatalogPath, "page-catalog", crawl.DefaultPageCatalogPath, "reviewed page catalog")
		flags.StringVar(&config.Crawl.ChromePath, "chrome-path", "", "Chrome or Chromium executable")
		flags.StringVar(&config.Crawl.CacheDir, "cache-dir", "", "payload cache (default: inside state directory)")
		flags.IntVar(&config.Crawl.MaxPages, "max-pages", 500, "maximum pages")
		flags.IntVar(&config.Crawl.Concurrency, "concurrency", 1, "parallel browser pages")
		flags.DurationVar(&config.Crawl.PageTimeout, "page-timeout", 30*time.Second, "maximum observation time per page; finishes early after network activity settles")
		flags.Float64Var(&config.MaxChangeFraction, "max-change-fraction", 0.25, "hold changes above this fraction of previous contextual values")
		config.Crawl.PathPrefix = "/"
		config.Crawl.Refresh = true
	}
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: tokens update %s [flags]\n\nCommands: run, status, inspect <id>, approve <id>, retry <id>, notify <id>, test-notification, alert\n\n", command)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected update arguments")
	}
	config.Webhook = os.Getenv("TOKENS_NOTIFY_WEBHOOK")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	switch command {
	case "test-notification":
		if !update.NotificationsConfigured(config.Webhook) {
			return errors.New("configure SMTP or TOKENS_NOTIFY_WEBHOOK first")
		}
		if err := update.TestNotification(ctx, config.StateDir, config.Webhook); err != nil {
			return err
		}
		fmt.Println("Test notification accepted by the configured delivery service(s). Check your inbox or receiver.")
		return nil
	case "alert":
		if !update.NotificationsConfigured(config.Webhook) {
			return errors.New("configure SMTP or TOKENS_NOTIFY_WEBHOOK to receive service failure alerts")
		}
		return update.Alert(ctx, config.StateDir, config.Webhook)
	case "status":
		reports, err := update.List(config.StateDir)
		if err != nil {
			return err
		}
		if len(reports) == 0 {
			fmt.Println("No saved update runs.")
		}
		for _, report := range reports {
			fmt.Printf("%s  %-12s  %s  +%d ~%d -%d\n", report.ID, report.Status, report.Phase, report.Summary.Added, report.Summary.Updated, report.Summary.Removed)
		}
		return nil
	case "inspect":
		report, err := update.Read(config.StateDir, id)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case "approve":
		report, err := update.Approve(config.StateDir, id)
		if err != nil {
			return err
		}
		fmt.Printf("Activated %s. Restart the server to load this generation.\n", report.ID)
		return nil
	case "notify":
		if !update.NotificationsConfigured(config.Webhook) {
			return errors.New("configure SMTP or TOKENS_NOTIFY_WEBHOOK before resending notifications")
		}
		return update.NotifyRun(ctx, config.StateDir, id, config.Webhook)
	case "run", "retry":
		if timeout <= 0 {
			return errors.New("timeout must be positive")
		}
		ctx, stop := context.WithTimeout(ctx, timeout)
		defer stop()
		config.RetryRunID = id
		report, err := update.Run(ctx, config)
		if report.ID != "" {
			fmt.Printf("Run %s: %s\n", report.ID, report.Status)
		}
		if err != nil {
			return err
		}
		for _, issue := range report.Issues {
			fmt.Printf("%s: %s\n", issue.Code, issue.Message)
		}
		if report.Status == "needs_review" {
			return fmt.Errorf("run %s needs review; inspect it before approving or retrying", report.ID)
		}
		return nil
	default:
		return fmt.Errorf("unknown update command %q", command)
	}
}
