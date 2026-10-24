package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/sojupie/tokens/source/guidance"
)

type guidanceURLs []string

func (u *guidanceURLs) String() string         { return fmt.Sprint([]string(*u)) }
func (u *guidanceURLs) Set(value string) error { *u = append(*u, value); return nil }

func runGuidance(args []string) error {
	flags := flag.NewFlagSet("tokens guidance", flag.ContinueOnError)
	var cfg guidance.Config
	var urls guidanceURLs
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
	cfg.URLs = urls
	path, err := guidance.Capture(context.Background(), cfg)
	if path != "" {
		fmt.Println("Documentation capture:", path)
	}
	return err
}
