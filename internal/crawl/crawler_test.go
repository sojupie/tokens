package crawl

import (
	"net/http"
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
	if err := writePageCatalog(path, "https://m3.material.io/sitemap.xml", pages); err != nil {
		t.Fatal(err)
	}
	got, err := loadPageCatalog(path, 0)
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

func TestFailedPagesIncludesPageAndPayloadFailures(t *testing.T) {
	pages := []string{"https://example.test/a/specs", "https://example.test/b/specs", "https://example.test/c/specs", "https://example.test/d/specs"}
	receipt := CrawlReceipt{
		Pages: pages,
		Observations: []Observation{
			{PageURL: pages[0], Status: http.StatusOK, TokenTableURL: "https://example.test/a.json"},
			{PageURL: pages[1], Status: http.StatusInternalServerError},
			{PageURL: pages[2], Status: http.StatusOK, TokenTableURL: "https://example.test/c.json"},
			{PageURL: pages[3], Status: http.StatusOK},
		},
		SourcePayloads: []SourcePayloadRecord{
			{PageURL: pages[0], TokenTableURL: "https://example.test/a.json", Hash: "hash-a"},
			{PageURL: pages[2], TokenTableURL: "https://example.test/c.json", Error: "404 Not Found"},
		},
	}
	got, err := failedPages(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != pages[1] || got[1] != pages[2] || got[2] != pages[3] {
		t.Fatalf("unexpected failed pages: %#v", got)
	}
}

func TestFailedPagesAcceptsPayloadAttributedToAnotherPage(t *testing.T) {
	pages := []string{"https://example.test/a/specs", "https://example.test/b/specs"}
	sharedPayload := "https://example.test/shared.json"
	receipt := CrawlReceipt{
		Pages: pages,
		Observations: []Observation{
			{PageURL: pages[0], Status: http.StatusOK, TokenTableURL: sharedPayload},
			{PageURL: pages[1], Status: http.StatusOK, TokenTableURL: sharedPayload},
		},
		SourcePayloads: []SourcePayloadRecord{
			{PageURL: pages[0], TokenTableURL: sharedPayload, Hash: "shared-hash"},
		},
	}
	got, err := failedPages(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("shared successful payload should satisfy both pages: %#v", got)
	}
}

func TestMergeRetryResultsKeepsSuccessfulPages(t *testing.T) {
	pages := []string{"https://example.test/a/specs", "https://example.test/b/specs"}
	previousObservations := []Observation{
		{PageURL: pages[0], Status: http.StatusOK, TokenTableURL: "https://example.test/a.json"},
		{PageURL: pages[1], Status: http.StatusInternalServerError},
	}
	retriedObservations := []Observation{
		{PageURL: pages[1], Status: http.StatusOK, TokenTableURL: "https://example.test/b.json"},
	}
	mergedObservations := mergeObservations(previousObservations, retriedObservations, []string{pages[1]})
	if len(mergedObservations) != 2 || mergedObservations[0].Status != http.StatusOK || mergedObservations[1].Status != http.StatusOK {
		t.Fatalf("unexpected merged observations: %#v", mergedObservations)
	}
	mergedPayloads := mergeSourcePayloads(
		[]SourcePayloadRecord{{PageURL: pages[0], Hash: "hash-a"}, {PageURL: pages[1], Hash: "old-hash"}},
		[]SourcePayloadRecord{{PageURL: pages[1], Hash: "new-hash"}},
		[]string{pages[1]},
	)
	if len(mergedPayloads) != 2 || mergedPayloads[0].Hash != "hash-a" || mergedPayloads[1].Hash != "new-hash" {
		t.Fatalf("unexpected merged payloads: %#v", mergedPayloads)
	}
}
