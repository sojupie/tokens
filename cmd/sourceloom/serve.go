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

	"github.com/sojupie/m3-source/internal/publish"
	"github.com/sojupie/m3-source/internal/web"
)

func runServe(args []string) error {
	flags := flag.NewFlagSet("sourceloom serve", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: sourceloom serve [flags]")
		_, _ = fmt.Fprintln(flags.Output(), "\nServes a read-only publication over HTTP.")
		_, _ = fmt.Fprintln(flags.Output(), "")
		flags.PrintDefaults()
	}
	publishedDir := "published"
	listenAddress := "127.0.0.1:8080"
	flags.StringVar(&publishedDir, "published", publishedDir, "publication directory")
	flags.StringVar(&listenAddress, "listen", listenAddress, "HTTP listen address")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("serve does not accept positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	state, err := publish.Load(filepath.Join(publishedDir, "tokens.sqlite"))
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              listenAddress,
		Handler:           web.New(state).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Printf("Serving publication %s on http://%s\n", state.PublicationID, listenAddress)
	return server.ListenAndServe()
}
