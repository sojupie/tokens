package m3crawl

import (
	"net/url"
	"testing"
)

func TestParseRobotsSitemaps(t *testing.T) {
	origin, _ := url.Parse("https://m3.material.io/components/buttons/specs")
	got := parseRobotsSitemaps([]byte("User-agent: *\nSitemap: /sitemap.xml\nSITEMAP: https://example.test/nope.xml\n"), origin)
	if len(got) != 2 || got[0] != "https://m3.material.io/sitemap.xml" {
		t.Fatalf("unexpected sitemap URLs: %#v", got)
	}
}

func TestCanonicalPageURL(t *testing.T) {
	origin, _ := url.Parse("https://m3.material.io/components/buttons/specs")
	if got := canonicalPageURL("/components/buttons/overview#tokens", origin, "/components/"); got != "https://m3.material.io/components/buttons/overview" {
		t.Fatalf("unexpected canonical URL: %q", got)
	}
	if got := canonicalPageURL("https://example.test/components/other", origin, "/components/"); got != "" {
		t.Fatalf("cross-origin URL should be rejected: %q", got)
	}
	if got := canonicalPageURL("/static/app.js", origin, "/components/"); got != "" {
		t.Fatalf("asset URL should be rejected: %q", got)
	}
}

func TestIsSpecsPage(t *testing.T) {
	for _, page := range []string{
		"https://m3.material.io/components/buttons/specs",
		"https://m3.material.io/components/buttons/specs/",
	} {
		if !isSpecsPage(page) {
			t.Fatalf("expected specs page: %s", page)
		}
	}
	for _, page := range []string{
		"https://m3.material.io/components/buttons/overview",
		"https://m3.material.io/components/buttons/specs/details",
	} {
		if isSpecsPage(page) {
			t.Fatalf("did not expect specs page: %s", page)
		}
	}
}

func TestCatalogRoundTripAndDiff(t *testing.T) {
	path := t.TempDir() + "/pages.json"
	pages := []string{
		"https://m3.material.io/components/buttons/specs",
		"https://m3.material.io/components/cards/specs",
	}
	if err := writeCatalog(path, "https://m3.material.io/sitemap.xml", pages); err != nil {
		t.Fatal(err)
	}
	got, err := loadCatalogPages(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != pages[0] || got[1] != pages[1] {
		t.Fatalf("unexpected catalog pages: %#v", got)
	}
	added, removed := pageDiff(got, []string{
		"https://m3.material.io/components/buttons/specs",
		"https://m3.material.io/components/dialogs/specs",
	})
	if len(added) != 1 || added[0] != "https://m3.material.io/components/dialogs/specs" {
		t.Fatalf("unexpected additions: %#v", added)
	}
	if len(removed) != 1 || removed[0] != "https://m3.material.io/components/cards/specs" {
		t.Fatalf("unexpected removals: %#v", removed)
	}
}
