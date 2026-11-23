package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/tokens/forge/site"
)

func TestSiteShellSelectsCarouselPlayground(t *testing.T) {
	out := t.TempDir()
	if err := runSiteShell([]string{"--out", out}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "site-header.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(raw)
	for _, tool := range site.Default().Tools {
		want := `href="` + tool.URL + `" aria-current="false"`
		if tool.URL == site.CarouselPlaygroundURL {
			want = `href="` + tool.URL + `" aria-current="page"`
		}
		if !strings.Contains(html, want) {
			t.Errorf("site-header.html missing expected current state %s for %s", want, tool.Name)
		}
	}
}
