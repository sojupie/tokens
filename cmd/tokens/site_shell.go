package main

import (
	"context"
	"flag"
	"github.com/sojupie/tokens/catalog/web"
	"github.com/sojupie/tokens/source/publish"
	"os"
	"path/filepath"
	"strings"

	"github.com/sojupie/tokens/forge/site"
)

func runSiteShell(args []string) error {
	flags := flag.NewFlagSet("tokens site-shell", flag.ContinueOnError)
	out := flags.String("out", ".", "Output directory for shared site assets and HTML fragments")
	published := flags.String("published", "", "Optional publication directory supplying component and style navigation")
	origin := flags.String("catalog", "https://tokencatalog.dev/md3", "Catalog URL including collection path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0755); err != nil {
		return err
	}
	config := site.Default()
	if *published != "" {
		state, err := publish.Load(filepath.Join(*published, "tokens.sqlite"))
		if err != nil {
			return err
		}
		config.Navigation = web.SiteNavigation(state)
	}
	config.Catalog = strings.TrimRight(*origin, "/")
	config.ActiveTool = config.Tools[1].URL
	for _, name := range []string{"site.css", "site.js", "logo.svg", "logo-250x250.png", "logo-250x250.webp"} {
		data, err := site.Files.ReadFile(name)
		if err != nil {
			return err
		}
		target := name
		if name == "logo.svg" {
			target = "site-logo.svg"
		}
		if name == "logo-250x250.png" {
			target = "site-logo.png"
		}
		if name == "logo-250x250.webp" {
			target = "site-logo.webp"
		}
		if err = os.WriteFile(filepath.Join(*out, target), data, 0644); err != nil {
			return err
		}
	}
	header, err := os.Create(filepath.Join(*out, "site-header.html"))
	if err != nil {
		return err
	}
	err = site.Header(config).Render(context.Background(), header)
	closeErr := header.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	footer, err := os.Create(filepath.Join(*out, "site-footer.html"))
	if err != nil {
		return err
	}
	err = site.Footer(config).Render(context.Background(), footer)
	closeErr = footer.Close()
	if err != nil {
		return err
	}
	return closeErr
}
