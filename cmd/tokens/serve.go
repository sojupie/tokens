package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sojupie/tokens/catalog/usage"
	"github.com/sojupie/tokens/catalog/web"
	"github.com/sojupie/tokens/source/publish"
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
	usageDir := os.Getenv("TOKENS_USAGE_DIR")
	flags.StringVar(&usageDir, "usage-dir", usageDir, "directory for retained usage totals; empty disables usage logging")
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
	if usageDir != "" {
		store, err := usage.Open(usageDir, os.Stdout)
		if err != nil {
			return fmt.Errorf("open usage store: %w", err)
		}
		defer store.Close()
		publicationServer.SetUsage(store)
	}
	server := &http.Server{
		Addr:              listenAddress,
		Handler:           publicationServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Printf("Serving publication %s on http://%s\n", state.PublicationID, listenAddress)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	shutdownDone := make(chan struct{})
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		close(shutdownDone)
	}()
	err = server.ListenAndServe()
	stop()
	<-shutdownDone
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
