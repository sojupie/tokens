package guidance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMarkdownAndSavedEvidence(t *testing.T) {
	doc := document{URL: "https://example.com/components/tooltips/guidelines", HTML: `<main><h1>Tooltip guidance</h1><p>Use <strong>rich tooltips</strong> for explanations.</p><ul><li>Keep labels aligned.</li></ul><table><thead><tr><th>Variant</th><th>M3 Expressive</th></tr></thead><tbody><tr><td>Rich</td><td>Available</td></tr></tbody></table><p><a href="/components/tooltips/specs">Measurements</a></p><img src="/diagram.png" alt="16dp label inset"></main>`, Text: "Tooltip guidance\nUse rich tooltips for explanations.\nKeep labels aligned.\nVariant\tM3 Expressive\nRich\tAvailable\nMeasurements\n"}
	dir := t.TempDir()
	p := Page{URL: doc.URL}
	if err := savePage(dir, &p, doc); err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(filepath.Join(dir, p.Markdown))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# Tooltip guidance", "**rich tooltips**", "- Keep labels aligned.", "| Variant", "M3 Expressive", "Available", "https://example.com/components/tooltips/specs", "![16dp label inset](https://example.com/diagram.png)"} {
		if !strings.Contains(string(md), want) {
			t.Errorf("missing %q in:\n%s", want, md)
		}
	}
	for _, item := range []struct{ path, hash, expected string }{{p.HTML, p.HTMLSHA256, doc.HTML}, {p.Text, p.TextSHA256, doc.Text}, {p.Markdown, p.MarkdownSHA256, string(md)}} {
		data, err := os.ReadFile(filepath.Join(dir, item.path))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != item.expected || digest(string(data)) != item.hash {
			t.Errorf("evidence mismatch: %s", item.path)
		}
	}
	changed := doc
	changed.HTML = strings.Replace(doc.HTML, "Available", "Not recommended", 1)
	next, err := markdown(changed)
	if err != nil {
		t.Fatal(err)
	}
	if digest(next) == p.MarkdownSHA256 {
		t.Fatal("guidance change did not change the content hash")
	}
	if err := savePage(dir, &Page{URL: "https://example.com/empty"}, document{HTML: "<main></main>"}); err == nil {
		t.Fatal("accepted empty documentation")
	}
}

func TestMaterialContentExtraction(t *testing.T) {
	doc := document{URL: "https://m3.material.io/components/tooltips/guidelines", BaseURL: "https://m3.material.io/", HTML: `<main><nav>On this page</nav><a role="tab" href="components/tooltips/specs">Specs</a><div class="copy-button-container">Copy linkLink copied</div><h2>Placement</h2><p>Align <span class="tooltip-label">labels<span class="tooltip-parent">An unrelated definition</span></span> with supporting text.</p><a href="components/tooltips/specs">Measurements</a><img src="/diagram.png" alt="Inset"><span class="google-symbols">link</span></main>`}
	md, err := markdown(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{"On this page", "Copy link", "unrelated definition", "components/tooltips/components", "google-symbols"} {
		if strings.Contains(md, unwanted) {
			t.Errorf("unexpected %q in %s", unwanted, md)
		}
	}
	for _, want := range []string{"## Placement", "Align labels with supporting text.", "https://m3.material.io/components/tooltips/specs", "https://m3.material.io/diagram.png"} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q in %s", want, md)
		}
	}
	if !strings.Contains(doc.HTML, "Copy link") {
		t.Fatal("raw evidence was modified")
	}
}

func TestInvalidURLs(t *testing.T) {
	for _, urls := range [][]string{nil, {"file:///tmp/test"}, {"https://user:password@example.com"}, {"https://"}} {
		if _, err := Capture(context.Background(), Config{URLs: urls, Output: t.TempDir()}); err == nil {
			t.Fatalf("accepted URLs: %v", urls)
		}
	}
}

// Run explicitly with GUIDANCE_TEST_CHROME set; normal tests stay offline and
// need no browser. Includes delayed content, version retention and HTTP failures.
func TestRenderedCapture(t *testing.T) {
	chrome := os.Getenv("GUIDANCE_TEST_CHROME")
	if chrome == "" {
		t.Skip("set GUIDANCE_TEST_CHROME to run browser capture checks")
	}
	prose := strings.Repeat("Keep tooltip action labels aligned with supporting text. ", 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bad" {
			w.WriteHeader(http.StatusNotFound)
		}
		fmt.Fprintf(w, `<!doctype html><title>Guidance</title><nav>Unrelated navigation</nav><main><h1>Tooltip</h1><p>%s</p><div id="late"></div></main><script>setTimeout(()=>document.getElementById('late').innerHTML='<table><tr><th>Variant</th><th>Availability</th></tr><tr><td>Rich</td><td>Available</td></tr></table>',500)</script>`, prose)
	}))
	defer server.Close()
	cfg := Config{URLs: []string{server.URL + "/ok#section", server.URL + "/ok"}, Output: t.TempDir(), ChromePath: chrome, Timeout: 15 * time.Second}
	first, err := Capture(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	read := func(dir string) Receipt {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dir, "capture.json"))
		if err != nil {
			t.Fatal(err)
		}
		var r Receipt
		if err = json.Unmarshal(data, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	before := read(first)
	if !before.Complete || len(before.Pages) != 1 || before.Pages[0].Status != 200 {
		t.Fatalf("unexpected capture: %+v", before)
	}
	md, _ := os.ReadFile(filepath.Join(first, before.Pages[0].Markdown))
	if !strings.Contains(string(md), "Available") || strings.Contains(string(md), "Unrelated navigation") {
		t.Fatalf("wrong content: %s", md)
	}
	cfg.URLs = []string{server.URL + "/ok", server.URL + "/bad"}
	second, err := Capture(context.Background(), cfg)
	if err == nil {
		t.Fatal("accepted HTTP 404")
	}
	after := read(second)
	if first == second || after.Complete || len(after.Pages) != 2 || after.Pages[1].Status != 404 || after.Pages[1].Error == "" {
		t.Fatalf("unexpected second capture: %+v", after)
	}
	if before.Pages[0].MarkdownSHA256 != after.Pages[0].MarkdownSHA256 || !read(first).Complete {
		t.Fatal("unchanged content changed hash or previous run was altered")
	}
}

func TestMultilineTableCells(t *testing.T) {
	md, err := markdown(document{URL: "https://example.com/specs", HTML: `<main><table><tr><th>Variant</th><th>M3</th><th>M3 Expressive</th></tr><tr><td><p>Baseline</p><p>rail</p></td><td><p>Available</p></td><td><p>Not recommended</p><p>Use collapsed</p></td></tr></table></main>`})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"| Variant", "| M3", "Baseline", "Not recommended", "Use collapsed"} {
		if !strings.Contains(md, want) {
			t.Fatalf("missing %q: %s", want, md)
		}
	}
	if strings.Count(md, "|") < 12 {
		t.Fatalf("lost table structure: %s", md)
	}
}
