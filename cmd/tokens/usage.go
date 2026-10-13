package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sojupie/tokens/internal/update"
	"github.com/sojupie/tokens/internal/usage"
)

func runUsage(args []string) error {
	flags := flag.NewFlagSet("tokens usage", flag.ContinueOnError)
	directory := os.Getenv("TOKENS_USAGE_DIR")
	if directory == "" {
		directory = "/var/lib/tokens-usage"
	}
	since := "7d"
	send, weekly, svg := false, false, false
	flags.StringVar(&directory, "usage-dir", directory, "retained usage database directory")
	flags.StringVar(&since, "since", since, "report period, from 1d to 90d")
	flags.BoolVar(&svg, "svg", false, "render aggregate usage as an SVG for the README")
	flags.BoolVar(&send, "send", false, "email this report using existing SMTP settings")
	flags.BoolVar(&weekly, "weekly", false, "email once per week, resuming from the last successful digest")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage does not accept positional arguments")
	}
	if svg && (send || weekly) {
		return errors.New("--svg cannot be combined with --send or --weekly")
	}
	days, err := strconv.Atoi(strings.TrimSuffix(since, "d"))
	if err != nil || !strings.HasSuffix(since, "d") || days < 1 || days > 90 {
		return errors.New("--since must be between 1d and 90d")
	}
	// Reporting must not silently create an empty database when a mount is missing.
	if _, err := os.Stat(directory + "/usage.sqlite"); err != nil {
		return fmt.Errorf("usage database unavailable: %w", err)
	}
	store, err := usage.Open(directory, nil)
	if err != nil {
		return err
	}
	defer store.Close()
	until := time.Now().UTC()
	warning := ""
	from := until.AddDate(0, 0, -days).Truncate(time.Hour)
	if weekly {
		release, err := store.ClaimDigest()
		if err != nil {
			return err
		}
		defer release()
		until = until.Truncate(time.Hour)
		last, err := store.LastDigest()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil {
			if until.Sub(last) < 6*24*time.Hour {
				fmt.Println("Weekly digest already sent.")
				return nil
			}
			from = last
		}
		oldest := until.AddDate(0, 0, -usage.RetentionDays)
		if from.Before(oldest) {
			warning = "WARNING: some unsent history is older than the 90-day retention period and is unavailable.\n"
			from = oldest
		}
	}
	summary, err := store.Report(from, until)
	if err != nil {
		return err
	}
	if svg {
		fmt.Print(summary.SVG())
		return nil
	}
	body := warning + summary.Text()
	fmt.Print(body)
	if send || weekly {
		subject := "Token Catalog: usage report"
		if weekly {
			subject = "Token Catalog: weekly usage digest"
		}
		if err := update.SendEmailMessage(context.Background(), subject, body); err != nil {
			return err
		}
		if weekly {
			if err := store.SaveDigest(until); err != nil {
				return fmt.Errorf("digest accepted by SMTP but checkpoint failed: %w", err)
			}
		}
		fmt.Println("Usage digest accepted by SMTP.")
	}
	return nil
}
