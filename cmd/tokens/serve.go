package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sojupie/tokens/internal/publish"
	"github.com/sojupie/tokens/internal/web"
)

func runServe(args []string) error {
	flags := flag.NewFlagSet("tokens serve", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: tokens serve [flags]")
		_, _ = fmt.Fprintln(flags.Output(), "\nServes a read-only publication over HTTP.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	publishedDir := "published"
	listenAddress := "127.0.0.1:8080"
	publicOrigin := os.Getenv("TOKENS_PUBLIC_ORIGIN")
	flags.StringVar(&publishedDir, "published", publishedDir, "publication directory")
	flags.StringVar(&listenAddress, "listen", listenAddress, "HTTP listen address")
	flags.StringVar(&publicOrigin, "public-origin", publicOrigin, "public http(s) base URL, optionally including a path, for canonical URLs and sitemap; unset disables indexing")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("serve does not accept positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	// Resolve once: a later activation must not mix old SQL with new downloads.
	publishedDir, err := publish.ResolveGeneration(publishedDir)
	if err != nil {
		return err
	}
	state, err := publish.Load(filepath.Join(publishedDir, "tokens.sqlite"))
	if err != nil {
		return err
	}
	publicationServer := web.New(state, filepath.Join(publishedDir, "downloads"))
	if err := publicationServer.SetPublicOrigin(publicOrigin); err != nil {
		return err
	}
	server := &http.Server{
		Addr:              listenAddress,
		Handler:           publicationServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Printf("Serving publication %s on http://%s\n", state.PublicationID, listenAddress)
	return server.ListenAndServe()
}
