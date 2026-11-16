package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sojupie/tokens/catalog/web"
)

func runTableShell(args []string) error {
	flags := flag.NewFlagSet("tokens table-shell", flag.ContinueOnError)
	assets := flags.String("assets", "", "Generated Catalog asset directory")
	out := flags.String("out", "", "Output directory for the static table and its assets")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *assets == "" || *out == "" || flags.NArg() != 0 {
		return fmt.Errorf("usage: tokens table-shell --assets DIR --out DIR < table.json")
	}
	return web.ExportStaticTokenTable(context.Background(), os.Stdin, *assets, *out)
}
