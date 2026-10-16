package crawl

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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
	if got := canonicalPageURL("/components-extra/buttons/specs", origin, "/components/"); got != "" {
		t.Fatalf("sibling path should be rejected: %q", got)
	}
}

func TestUniqueSourcesPreservesPayloadsAndContext(t *testing.T) {
	shared := "https://m3.material.io/_dsm/data/dsdb-m3/ELEVATION.json"
	component := "https://m3.material.io/_dsm/data/dsdb-m3/TOKEN_TABLE.json"
	sources := uniqueSources([]Observation{
		{PageURL: "https://m3.material.io/components/buttons/specs", SourcePayloadURLs: []string{component, shared}},
		{PageURL: "https://m3.material.io/styles/elevation/tokens", SourcePayloadURLs: []string{shared}},
	})
	if len(sources) != 2 {
		t.Fatalf("got %d sources, want 2: %#v", len(sources), sources)
	}
	for _, source := range sources {
		if source.SourcePayloadURL == shared {
			if len(source.PageURLs) != 2 || source.SourcePayloadKind != "component" {
				t.Fatalf("shared source lost page context: %#v", source)
			}
		}
	}
}

func TestPersistPayloadRedownloadsCorruptCache(t *testing.T) {
	body := []byte(`{"system":{"tokens":[]}}`)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write(body)
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	artifact := "source-payloads/aa/payload.json.gz"
	artifactPath := filepath.Join(cacheDir, filepath.FromSlash(artifact))
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, _ = io.WriteString(writer, `{"system":{"tokens":[{"corrupt":true}]}}`)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifactPath, compressed.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	index := CacheIndex{SchemaVersion: 2, SourcePayloads: map[string]SourcePayloadRecord{
		server.URL: {SourcePayloadURL: server.URL, Artifact: artifact, Hash: sha256Hex(body), Bytes: len(body)},
	}}
	record, err := persistPayload(context.Background(), Config{CacheDir: cacheDir}, index, pageSource{
		PageURLs:          []string{"https://m3.material.io/styles/elevation/tokens"},
		SourcePayloadURL:  server.URL,
		SourcePayloadKind: "elevation",
	})
	if err != nil {
		t.Fatal(err)
	}
	if record.Reused || record.Hash != sha256Hex(body) || record.SourcePayloadKind != "elevation" {
		t.Fatalf("corrupt cache was reused or context was lost: %#v", record)
	}
}

func TestSourcePayloadPattern(t *testing.T) {
	for _, target := range []string{
		"https://m3.material.io/_dsm/data/dsdb-m3/2026-08-05_09-00-19/TOKEN_TABLE.52865380210a9b69.json",
		"https://m3.material.io/_dsm/data/dsdb-m3/2026-08-05_09-00-19/TOKEN_TYPE_UNSPECIFIED.20543ce1889f27d9.json",
		"https://m3.material.io/_dsm/data/dsdb-m3/2026-08-05_09-00-19/TOKEN_TYPE_COLOR.json?alt=media",
		"https://m3.material.io/_dsm/data/dsdb-m3/ELEVATION.20543ce18892f7d9.json",
		"https://m3.material.io/_dsm/data/dsdb-m3/SPACING.20543ce18892f7d9.json",
		"https://m3.material.io/_dsm/data/dsdb-m3/TYPE_SCALE.20543ce18892f7d9.json",
	} {
		if !sourcePayloadRE.MatchString(target) {
			t.Fatalf("expected source payload URL to match: %s", target)
		}
	}
	for _, target := range []string{
		"https://m3.material.io/_dsm/data/OTHER.json",
		"https://m3.material.io/_dsm/data/TOKEN_TYPE_UNSPECIFIED.txt",
		"https://m3.material.io/_dsm/data/dsdb-m3/ordinary-page-data.json",
		"https://m3.material.io/api/ELEVATION.20543ce18892f7d9.json",
	} {
		if sourcePayloadRE.MatchString(target) {
			t.Fatalf("did not expect source payload URL to match: %s", target)
		}
	}
}

func TestPageScopes(t *testing.T) {
	for _, page := range []string{
		"https://m3.material.io/components/buttons/specs",
		"https://m3.material.io/components/buttons/specs/",
	} {
		if !isComponentSpecsPage(page) || !isPageInScope(page, ScopeComponentSpecs) || !isPageInScope(page, ScopeTokenSourcePages) {
			t.Fatalf("expected specs page: %s", page)
		}
	}
	for _, page := range []string{
		"https://m3.material.io/components/buttons/overview",
		"https://m3.material.io/components/buttons/specs/details",
	} {
		if isComponentSpecsPage(page) || isPageInScope(page, ScopeTokenSourcePages) {
			t.Fatalf("did not expect specs page: %s", page)
		}
	}
	for _, page := range []string{
		"https://m3.material.io/styles/elevation/tokens",
		"https://m3.material.io/styles/motion/overview/specs",
		"https://m3.material.io/styles/typography/type-scale-tokens",
		"https://m3.material.io/styles/shape/corner-radius-scale",
	} {
		if !isStyleTokenSourcePage(page) || !isPageInScope(page, ScopeTokenSourcePages) || isPageInScope(page, ScopeComponentSpecs) {
			t.Fatalf("expected style token-source page: %s", page)
		}
	}
	if isStyleTokenSourcePage("https://m3.material.io/styles/motion/overview") {
		t.Fatal("did not expect style overview page")
	}
}

func TestCatalogRoundTripAndDiff(t *testing.T) {
	path := t.TempDir() + "/pages.json"
	pages := []string{
		"https://m3.material.io/components/buttons/specs",
		"https://m3.material.io/styles/elevation/tokens",
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
	if len(removed) != 1 || removed[0] != "https://m3.material.io/styles/elevation/tokens" {
		t.Fatalf("unexpected removals: %#v", removed)
	}
}

func TestPageMetadata(t *testing.T) {
	tests := []struct {
		page, area, subject, pageType string
	}{
		{"https://m3.material.io/components/navigation-bar/specs", "components", "navigation-bar", "specs"},
		{"https://m3.material.io/styles/motion/overview/specs", "styles", "motion/overview", "specs"},
		{"https://m3.material.io/styles/spacing/tokens", "styles", "spacing", "tokens"},
		{"https://m3.material.io/styles/typography/type-scale-tokens", "styles", "typography/type-scale", "tokens"},
		{"https://m3.material.io/styles/shape/corner-radius-scale", "styles", "shape/corner-radius", "tokens"},
	}
	for _, test := range tests {
		area, subject, pageType := pageMetadata(test.page)
		if area != test.area || subject != test.subject || pageType != test.pageType {
			t.Fatalf("pageMetadata(%q) = %q, %q, %q", test.page, area, subject, pageType)
		}
	}
}

func TestFailedPagesIncludesPageAndPayloadFailures(t *testing.T) {
	pages := []string{"https://example.test/a/specs", "https://example.test/b/specs", "https://example.test/c/specs", "https://example.test/d/specs"}
	receipt := CrawlReceipt{
		Pages: pages,
		Observations: []Observation{
			{PageURL: pages[0], Status: http.StatusOK, SourcePayloadURLs: []string{"https://example.test/a.json"}},
			{PageURL: pages[1], Status: http.StatusInternalServerError},
			{PageURL: pages[2], Status: http.StatusOK, SourcePayloadURLs: []string{"https://example.test/c.json"}},
			{PageURL: pages[3], Status: http.StatusOK},
		},
		SourcePayloads: []SourcePayloadRecord{
			{PageURLs: []string{pages[0]}, SourcePayloadURL: "https://example.test/a.json", Hash: "hash-a"},
			{PageURLs: []string{pages[2]}, SourcePayloadURL: "https://example.test/c.json", Error: "404 Not Found"},
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
			{PageURL: pages[0], Status: http.StatusOK, SourcePayloadURLs: []string{sharedPayload}},
			{PageURL: pages[1], Status: http.StatusOK, SourcePayloadURLs: []string{sharedPayload}},
		},
		SourcePayloads: []SourcePayloadRecord{
			{PageURLs: pages, SourcePayloadURL: sharedPayload, Hash: "shared-hash"},
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
		{PageURL: pages[0], Status: http.StatusOK, SourcePayloadURLs: []string{"https://example.test/a.json"}},
		{PageURL: pages[1], Status: http.StatusInternalServerError},
	}
	retriedObservations := []Observation{
		{PageURL: pages[1], Status: http.StatusOK, SourcePayloadURLs: []string{"https://example.test/b.json"}},
	}
	mergedObservations := mergeObservations(previousObservations, retriedObservations, []string{pages[1]})
	if len(mergedObservations) != 2 || mergedObservations[0].Status != http.StatusOK || mergedObservations[1].Status != http.StatusOK {
		t.Fatalf("unexpected merged observations: %#v", mergedObservations)
	}
	mergedPayloads := mergeSourcePayloads(
		[]SourcePayloadRecord{{PageURLs: []string{pages[0]}, SourcePayloadURL: "https://example.test/a.json", Hash: "hash-a"}, {PageURLs: []string{pages[1]}, SourcePayloadURL: "https://example.test/b.json", Hash: "old-hash"}},
		[]SourcePayloadRecord{{PageURLs: []string{pages[1]}, SourcePayloadURL: "https://example.test/b.json", Hash: "new-hash"}},
		[]string{pages[1]},
	)
	if len(mergedPayloads) != 2 || mergedPayloads[0].Hash != "hash-a" || mergedPayloads[1].Hash != "new-hash" {
		t.Fatalf("unexpected merged payloads: %#v", mergedPayloads)
	}
}
