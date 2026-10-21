package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/sojupie/tokens/forge/cssusage"
	"os"
)

func runForge(args []string) error {
	if len(args) == 0 || args[0] != "inspect" {
		return errors.New("usage: tokens forge inspect [--assets directory] [--output file]")
	}
	flags := flag.NewFlagSet("tokens forge inspect", flag.ContinueOnError)
	assets := flags.String("assets", "", "generated CSS directory; omit to inspect the checked-in CSS")
	output := flags.String("output", "", "write the CSS usage report as JSON")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	report, err := cssusage.Run(*assets)
	if err != nil {
		return err
	}
	if *output != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(*output, append(data, '\n'), 0644); err != nil {
			return err
		}
	}
	fmt.Printf("CSS declarations: %d token references, %d mixed, %d manual, %d unresolved. Active overrides: %d.\n", report.Count("Token"), report.Count("Mixed"), report.Count("Manual"), report.Count("Unresolved"), len(report.Overrides))
	return nil
}
