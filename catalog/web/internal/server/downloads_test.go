// Integration checks for download.
package server

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
	"golang.org/x/net/html"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectionDownloadsAndSharedControls(t *testing.T) {
	state := publish.PublishedState{CurrentSnapshot: publish.SnapshotView{ID: "fixture", CreatedAt: "2026-10-07T00:00:00Z"}, Subjects: map[publish.SubjectKey]publish.SubjectView{}}
	for _, key := range []publish.SubjectKey{{Kind: publish.SubjectComponent, Slug: "buttons"}, {Kind: publish.SubjectStyle, Slug: "color"}} {
		state.Subjects[key] = publish.SubjectView{Key: key, Name: key.Slug, Tokens: []publish.TokenView{{SourceName: "md.sys.color.primary"}}}
	}
	handler := New(state, "").Handler()
	for _, scope := range []string{"all", "components", "styles"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/tokens.json?scope="+scope+"&q=no-match", nil))
		if response.Code != 200 || !strings.Contains(response.Header().Get("Content-Disposition"), "attachment") {
			t.Fatalf("download %s: %d %v", scope, response.Code, response.Header())
		}
		var document struct {
			CapturedAt string                  `json:"capturedAt"`
			Subjects   []publish.PublicSubject `json:"subjects"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
			t.Fatal(err)
		}
		want := 1
		if scope == "all" {
			want = 2
		}
		if len(document.Subjects) != want || document.CapturedAt != state.CurrentSnapshot.CreatedAt {
			t.Fatalf("wrong export %s: %#v", scope, document)
		}
		for _, subject := range document.Subjects {
			if scope == "components" && subject.Kind != publish.SubjectComponent || scope == "styles" && subject.Kind == publish.SubjectComponent {
				t.Fatalf("wrong subject in %s", scope)
			}
			if len(subject.Tokens) != 1 {
				t.Fatal("missing tokens")
			}
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/tokens.json?scope=invalid", nil))
	if response.Code != 400 {
		t.Fatal("accepted invalid collection scope")
	}
	for _, path := range []string{"/", "/components", "/styles", "/components/buttons", "/styles/color", "/changes"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		body := response.Body.String()
		if response.Code != 200 {
			t.Fatalf("page %s: %d", path, response.Code)
		}
		for _, text := range []string{`class="tokens-brand-logo"`, `rel="icon"`, `id="tokens-site-download"`, `data-export-open`, `data-dialog-surface`, `CSS and DTCG exports include required token references`} {
			if !strings.Contains(body, text) {
				t.Fatalf("%s missing %s", path, text)
			}
		}
		if strings.Count(body, `id="tokens-site-download"`) != 1 || strings.Count(body, `src="/static/token-table.js`) != 1 {
			t.Fatalf("duplicate dialog or script on %s", path)
		}

		document, err := html.ParseWithOptions(strings.NewReader(body), html.ParseOptionEnableScripting(false))
		if err != nil {
			t.Fatal(err)
		}
		var checkFallback func(*html.Node, bool)
		checkFallback = func(node *html.Node, inNoscript bool) {
			inNoscript = inNoscript || node.Type == html.ElementNode && node.Data == "noscript"
			if inNoscript && node.Type == html.ElementNode && node.Data == "style" {
				t.Fatalf("%s embeds fallback CSS that activates on history restoration", path)
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				checkFallback(child, inNoscript)
			}
		}
		checkFallback(document, false)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/component/buttons.json?scope=everything", nil))
	if response.Code != 303 || response.Header().Get("Location") != "/downloads/tokens.json?scope=all" {
		t.Fatal("subject-wide export does not lead to all tokens")
	}
}

func TestNativeDownloadSignalsPreparationResult(t *testing.T) {
	server := scopedExportFixture()
	for _, test := range []struct{ query, result string }{
		{"scope=subject&kind=component&slug=buttons&context.Theme=Light&format=dtcg", "ready"},
		{"scope=subject&kind=component&slug=missing&format=dtcg", "error"},
	} {
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export?"+test.query+"&_download=test-123", nil))
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != "tokens-download-test-123" || cookies[0].Value != test.result {
			t.Fatalf("%d: %v", w.Code, cookies)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("download acknowledgement must not be cached")
		}
		if test.result == "ready" && !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment;") {
			t.Fatal("lost native download")
		}
	}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export?scope=subject&kind=component&slug=buttons&context.Theme=Light&format=dtcg", nil))
	if len(w.Result().Cookies()) != 0 {
		t.Fatal("ordinary API downloads must not set acknowledgement cookies")
	}
}

func scopedExportFixture() *Server {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	values := []publish.TokenValueView{{ContextLabels: []string{"Theme: Light"}, Value: json.RawMessage(`{"numeric":1}`), References: []publish.TokenReference{{Name: "md.sys.opacity"}}}, {ContextLabels: []string{"Theme: Dark"}, Value: json.RawMessage(`{"numeric":2}`), References: []publish.TokenReference{{Name: "md.sys.opacity"}}}}
	subject := publish.SubjectView{Key: key, Name: "Buttons", Tokens: []publish.TokenView{{SourceName: "md.comp.button.outlined.icon.opacity", Category: "COLOR", Values: values}, {SourceName: "md.comp.button.outlined.container.opacity", Category: "COLOR", Values: values}, {SourceName: "md.comp.button.filled.icon.opacity", Category: "COLOR", Values: values}}}
	server := New(publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: subject}}, "")
	server.cssValues = map[string]map[string]string{dtcg.ValueKey(values[0].Value): {"": "1"}, dtcg.ValueKey(values[1].Value): {"": "2"}}
	return server
}

func TestScopedExportsPreserveContextAndRequiredReferences(t *testing.T) {
	server := scopedExportFixture()
	if err := server.SetPublicOrigin("https://example.test/md3"); err != nil {
		t.Fatal(err)
	}
	for scope, count := range map[string]int{"view": 1, "set": 2, "subject": 3, "collection": 3} {
		query := url.Values{"kind": {"component"}, "slug": {"buttons"}, "scope": {scope}, "set": {"button|color|outlined"}, "q": {"icon"}, "context.Theme": {"Dark"}}
		for _, format := range []string{"summary", "dtcg", "css", "source", "report"} {
			query.Set("format", format)
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/downloads/export?"+query.Encode(), nil))
			if response.Code != 200 {
				t.Fatalf("%s %s: %d %s", scope, format, response.Code, response.Body.String())
			}
			if format == "summary" {
				var summary exportSummary
				if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
					t.Fatal(err)
				}
				if summary.Selected != count || summary.Dependencies != 1 || summary.Context["Theme"] != "Dark" {
					t.Fatalf("wrong selection: %#v", summary)
				}
			}
			if format == "css" {
				css := response.Body.String()
				if !strings.Contains(css, "--md-sys-opacity: 2;") || !strings.Contains(css, "var(--md-sys-opacity)") {
					t.Fatalf("broken references: %s", css)
				}
				if scope == "view" && strings.Contains(css, "container") {
					t.Fatal("included token outside search")
				}
			}
			if format == "dtcg" && scope == "view" && strings.Contains(response.Body.String(), "container.opacity") {
				t.Fatal("DTCG lost table search")
			}
		}
	}
	for _, query := range []string{"scope=bad", "scope=subject&kind=component&slug=missing", "scope=view&kind=component&slug=buttons&set=missing", "scope=collection&context.Theme=Unknown"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/md3/downloads/export?format=css&"+query, nil))
		if response.Code != 400 {
			t.Fatalf("accepted invalid request: %s (%d)", query, response.Code)
		}
	}
}

func TestSourceExportKeepsSharedTokensWithinTheRequestedSubject(t *testing.T) {
	server := scopedExportFixture()
	buttons := server.state.Subjects[publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}]
	other := publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "other"}
	server.state.Subjects[other] = publish.SubjectView{Key: other, Tokens: buttons.Tokens[:1]}
	handler := server.Handler()

	for scope, want := range map[string]int{"view": 1, "set": 1, "subject": 1, "collection": 2} {
		query := url.Values{"kind": {"component"}, "slug": {"buttons"}, "scope": {scope}, "set": {"button|color|outlined"}, "q": {"icon"}, "format": {"source"}}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
		var document struct {
			Subjects []publish.PublicSubject
		}
		if err := json.Unmarshal(response.Body.Bytes(), &document); response.Code != 200 || err != nil {
			t.Fatalf("%s: %d %s", scope, response.Code, response.Body.String())
		}
		if len(document.Subjects) != want || (want == 1 && document.Subjects[0].Slug != "buttons") {
			t.Fatalf("%s: unexpected subjects %+v", scope, document.Subjects)
		}
	}
}

func TestImplementationArchiveIncludesNestedAssetsAndRunnableExample(t *testing.T) {
	server := scopedExportFixture()
	server.themeBundles = map[string][]byte{"shared": []byte(":root{}"), "switch": []byte(":root{}")}
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "switch"}
	server.state.Subjects[key] = publish.SubjectView{Key: key, Name: "Switch"}
	dir := t.TempDir()
	server.tokenAssetsDir = dir
	if err := os.WriteFile(filepath.Join(dir, "components-coverage.json"), []byte(`{"configurations":[{"family":"switch","variant":"icon","state":"enabled"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	query := url.Values{"scope": {"subject"}, "kind": {"component"}, "slug": {"switch"}, "content": {"implementation"}, "format": {"zip"}}

	draft := cloneExportQuery(query)
	draft["format"] = []string{"dtcg", "css"}
	draft.Set("context.Platform", "Web")
	draft.Set("context.Theme", "Light")
	dialog := server.exportDialog(context.Background(), draft)
	if !dialog.Ready || dialog.Error != "" || len(dialog.Query["_token-format"]) != 2 {
		t.Fatalf("implementation dialog: %+v", dialog)
	}
	draft = cloneExportQuery(dialog.Query)
	draft.Set("content", "tokens")
	draft.Set("slug", "buttons")
	draft.Del("context.Platform")
	draft.Set("context.Theme", "Light")
	dialog = server.exportDialog(context.Background(), draft)
	if !dialog.Ready || len(dialog.Query["format"]) != 2 {
		t.Fatalf("token format round trip: %+v", dialog)
	}
	files, _, err := server.implementationFiles(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.html", "tokens.css", "components.css", "buttons.js", "fields.js", "runtime/utils.js", "runtime/switches.js", "README.txt"} {
		if len(files[name]) == 0 {
			t.Fatalf("missing %s", name)
		}
	}
	if !strings.Contains(string(files["index.html"]), `role="switch"`) || !strings.Contains(string(files["index.html"]), "<main data-component-assets>") || strings.Contains(string(files["index.html"]), "/static/") {
		t.Fatal("example is not standalone")
	}
	for _, invalid := range []url.Values{
		{"scope": {"subject"}, "kind": {"component"}, "slug": {"switch"}, "content": {"implementation"}, "format": {"zip"}, "context.Theme": {"Invalid"}},
		{"scope": {"subject"}, "kind": {"component"}, "slug": {"switch"}, "content": {"implementation"}, "format": {"zip"}, "context.Platform": {"Android"}},
	} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+invalid.Encode(), nil))
		if response.Code != 400 {
			t.Fatal("silently accepted unsupported implementation context")
		}
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestExportMultipleFormatsAndContexts(t *testing.T) {
	server := scopedExportFixture()
	query := url.Values{"scope": {"subject"}, "kind": {"component"}, "slug": {"buttons"}, "format": {"css", "dtcg"}, "context.Theme": {"Light", "Dark"}}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 200 {
		t.Fatalf("%d: %s", response.Code, response.Body.String())
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.File) != 9 {
		t.Fatalf("expected four exports, four reports and manifest, got %d", len(archive.File))
	}
	light, dark, css, dtcg, reports := 0, 0, 0, 0, 0
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(file.Name, ".css") {
			css++
			if strings.Contains(string(data), "--md-sys-opacity: 1;") {
				light++
			}
			if strings.Contains(string(data), "--md-sys-opacity: 2;") {
				dark++
			}
		}
		if strings.HasSuffix(file.Name, "-report.json") {
			reports++
			var report struct {
				Context   map[string]string `json:"context"`
				Selection exportSummary     `json:"selection"`
			}
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			if report.Selection.Selected != 3 || (report.Context["Theme"] != "Light" && report.Context["Theme"] != "Dark") {
				t.Fatalf("invalid bundled report: %s", data)
			}
		}
		if strings.HasSuffix(file.Name, ".tokens.json") {
			dtcg++
		}
	}
	if light != 1 || dark != 1 || css != 2 || dtcg != 2 || reports != 4 {
		t.Fatalf("wrong context/format exports: %d %d %d %d", light, dark, css, dtcg)
	}
	query.Set("summary", "1")
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"combinations":4`) {
		t.Fatal(response.Body.String())
	}
	query.Set("context.Theme", "Unknown")
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 400 {
		t.Fatalf("accepted invalid batch: %d", response.Code)
	}
	query.Del("context.Theme")
	query["format"] = []string{"css", "css"}
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 400 {
		t.Fatalf("accepted duplicate formats: %d", response.Code)
	}
}

func TestExportContextsFollowDefinedCombinations(t *testing.T) {
	server := scopedExportFixture()
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	subject := server.state.Subjects[key]
	subject.Tokens = append(subject.Tokens, publish.TokenView{SourceName: "md.comp.button.label", Values: []publish.TokenValueView{
		{ContextLabels: []string{"Platform: Web", "Font: Static"}, Value: json.RawMessage(`{"numeric":1}`)},
		{ContextLabels: []string{"Platform: Web", "Font: iOS bold"}, Value: json.RawMessage(`{"numeric":2}`)},
		{ContextLabels: []string{"Platform: iOS", "Font: Static", "iOS weight setting: Default"}, Value: json.RawMessage(`{"numeric":1}`)},
		{ContextLabels: []string{"Platform: iOS", "Font: Static", "iOS weight setting: Bold"}, Value: json.RawMessage(`{"numeric":2}`)},
		{ContextLabels: []string{"Platform: Web", "iOS dark theme elevation: Elevated"}, Value: json.RawMessage(`{"undefined":true}`)},
		{ContextLabels: []string{"Platform: Web", "iOS dark theme elevation: None"}, Value: json.RawMessage(`{"undefined":true}`)},
	}})
	server.state.Subjects[key] = subject
	query := url.Values{"scope": {"subject"}, "kind": {"component"}, "slug": {"buttons"}, "context.Platform": {"Web"}, "context.Font": {"Static"}, "context.iOS weight setting": {"Default", "Bold"}, "format": {"dtcg"}}
	choices, err := server.exportContextChoices(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices["iOS weight setting"]) != 0 || len(choices["iOS dark theme elevation"]) != 0 {
		t.Fatalf("exposed irrelevant or undefined values: %v", choices)
	}
	if strings.Join(choices["Font"], ",") != "Static,iOS bold" {
		t.Fatalf("guessed applicability from label instead of captured values: %v", choices)
	}
	active, err := server.activeExportQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	if active.Has("context.iOS weight setting") {
		t.Fatal("inactive setting produces duplicate permutations")
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 200 || !strings.Contains(response.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("irrelevant alternatives must yield one document: %d %s", response.Code, response.Body.String())
	}
	query.Set("context.Platform", "iOS")
	choices, err = server.exportContextChoices(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices["iOS weight setting"]) != 2 {
		t.Fatalf("missing defined iOS settings: %v", choices)
	}
	query.Set("options", "1")
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"iOS weight setting"`) {
		t.Fatal(response.Body.String())
	}
}

func TestExportChoicesKeepNativeSelectionAndLabels(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		var output bytes.Buffer
		if err := webrender.ExportChoices("format", "File formats", "css", []webview.CatalogChoice{{Value: "css", Label: "CSS variables"}, {Value: "dtcg", Label: "DTCG JSON"}}, multiple).Render(context.Background(), &output); err != nil {
			t.Fatal(err)
		}
		markup := output.String()
		for _, label := range []string{"CSS variables", "DTCG JSON", `value="css"`} {
			if !strings.Contains(markup, label) {
				t.Fatal("missing option label or selected value", markup)
			}
		}
		if !strings.Contains(markup, `checked`) {
			t.Fatal("selected input missing", markup)
		}
		if multiple && (!strings.Contains(markup, `class="md-filter-chip"`) || !strings.Contains(markup, `type="checkbox"`)) {
			t.Fatal(markup)
		}
		if !multiple && !strings.Contains(markup, `type="radio"`) {
			t.Fatal(markup)
		}
	}
}

func TestExportContextOrdersPlatformBeforeAppearance(t *testing.T) {
	server := scopedExportFixture()
	for key, subject := range server.state.Subjects {
		subject.Tokens[0].Values[0].ContextLabels = []string{"Platform: Web", "Theme: Light", "Contrast: Default"}
		server.state.Subjects[key] = subject
	}
	page := server.exportPage(nil, url.Values{})
	if len(page.Contexts) < 3 || page.Contexts[0].Name != "Platform" || page.Contexts[1].Name != "Theme" || page.Contexts[2].Name != "Contrast" {
		t.Fatalf("wrong dependency order: %#v", page.Contexts)
	}
}

func TestImplementationBundlesExampleImages(t *testing.T) {
	files := map[string][]byte{}
	src := webassets.StaticURL("paintings/" + webview.CarouselPaintings[0])
	markup, err := bundleExampleImages(`<img src="`+src+`" alt="Example">`, files)
	if err != nil {
		t.Fatal(err)
	}
	name := "assets/paintings/" + webview.CarouselPaintings[0]
	if !strings.Contains(markup, `src="`+name+`"`) || len(files[name]) == 0 {
		t.Fatal("standalone example lost its image")
	}
	if strings.Contains(markup, "/static/") {
		t.Fatal("Catalog image dependency remains")
	}
}

func TestSingleTokenExportsKeepTheirFormatAndConversionMetadata(t *testing.T) {
	server := scopedExportFixture()
	for _, format := range []string{"css", "dtcg", "source", "report"} {
		query := url.Values{"scope": {"subject"}, "kind": {"component"}, "slug": {"buttons"}, "format": {format}, "context.Theme": {"Light"}}
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/downloads/export?"+query.Encode(), nil))
		if w.Code != 200 || w.Header().Get("Content-Type") == "application/zip" {
			t.Fatalf("%s: %d %s", format, w.Code, w.Body.String())
		}
		if format == "dtcg" {
			var doc map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
				t.Fatal(err)
			}
			metadata := doc["$extensions"].(map[string]any)["dev.tokencatalog.export"].(map[string]any)
			report := metadata["conversionReport"].(map[string]any)
			if report["exported"] != float64(4) || report["context"].(map[string]any)["Theme"] != "Light" {
				t.Fatal(metadata)
			}
		}
	}
}

func TestExportDialogWarnsAboutOmissionsWithoutBuildingExports(t *testing.T) {
	server := scopedExportFixture()
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	subject := server.state.Subjects[key]
	subject.Tokens[0].Values = subject.Tokens[0].Values[:1]
	subject.Tokens[1].Values = []publish.TokenValueView{{Value: json.RawMessage(`{"numeric":null}`)}}
	server.state.Subjects[key] = subject
	query := url.Values{"scope": {"subject"}, "kind": {"component"}, "slug": {"buttons"}, "format": {"css", "dtcg"}, "context.Theme": {"Light", "Dark"}}
	state := server.exportDialog(context.Background(), query)
	if !state.Ready || state.Error != "" {
		t.Fatal(state)
	}
	for _, want := range []string{"1 token value has no defined value", "2 token values cannot be converted", "including conversion reports"} {
		if !strings.Contains(state.Counts, want) {
			t.Fatalf("missing %s: %s", want, state.Counts)
		}
	}
	query.Set("format", "source")
	state = server.exportDialog(context.Background(), query)
	if strings.Contains(state.Counts, "Omitted:") {
		t.Fatalf("raw source download incorrectly warns about conversion: %s", state.Counts)
	}

	omitted := 0
	for _, theme := range []string{"Light", "Dark"} {
		q := cloneExportQuery(query)
		q.Set("context.Theme", theme)
		_, report, _, _, err := server.exportDocument(q)
		if err != nil {
			t.Fatal(err)
		}
		omitted += len(report.Omitted)
	}
	if omitted != 3 {
		t.Fatalf("actual omissions: %d", omitted)
	}
}

func TestNativeDownloadPageOmitsReportControls(t *testing.T) {
	server := scopedExportFixture()
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/download?scope=subject&kind=component&slug=buttons&format=dtcg&context.Theme=Light", nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, unwanted := range []string{`value="report"`, "Conversion report", "View conversion report"} {
		if strings.Contains(w.Body.String(), unwanted) {
			t.Fatalf("native download page still offers %s", unwanted)
		}
	}
}

func TestTokenDownloadsMatchTableAndKeepFullContextsOptional(t *testing.T) {
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	values := []publish.TokenValueView{
		{ContextLabels: []string{"Theme: Light", "Platform: Web"}, Value: json.RawMessage(`{"color":{"red":1}}`), References: []publish.TokenReference{{Name: "md.sys.color.primary"}}},
		{ContextLabels: []string{"Theme: Dark", "Platform: Web"}, Value: json.RawMessage(`{"color":{"red":0}}`)},
	}
	subject := publish.SubjectView{Key: key, Name: "Buttons", Tokens: []publish.TokenView{
		{SourceName: "md.comp.button.outlined.icon.color", DisplayName: "Button outlined icon color", Category: "COLOR", Values: values, Raw: json.RawMessage(`{"internal":"metadata"}`)},
		{SourceName: "md.comp.button.outlined.container.color", Category: "COLOR", Values: values},
		{SourceName: "md.comp.button.filled.icon.color", Category: "COLOR", Values: values},
	}}
	state := publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: subject}, CurrentSnapshot: publish.SnapshotView{CreatedAt: "2026-09-05T18:16:00Z"}}
	handler := New(state, "").Handler()
	for _, scope := range []string{"view", "set", "all"} {
		t.Run(scope, func(t *testing.T) {
			query := url.Values{"scope": {scope}, "set": {"button|color|outlined"}, "q": {"icon"}, "context.Theme": {"Light"}}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/downloads/component/buttons.json?"+query.Encode(), nil))
			if recorder.Code != 200 || !strings.HasPrefix(recorder.Header().Get("Content-Disposition"), "attachment;") {
				t.Fatalf("not a JSON download: %d %s", recorder.Code, recorder.Body.String())
			}
			var document publish.PublicTokenDocument
			if err := json.Unmarshal(recorder.Body.Bytes(), &document); err != nil {
				t.Fatal(err)
			}
			wantTokens, wantValues := 1, 1
			if scope == "set" {
				wantTokens, wantValues = 2, 2
			}
			if scope == "all" {
				wantTokens, wantValues = 3, 2
			}
			if len(document.Tokens) != wantTokens || len(document.Tokens[0].Values) != wantValues {
				t.Fatalf("incorrect export scope: %#v", document)
			}
			if document.Tokens[0].Values[0].References[0].Name != "md.sys.color.primary" || strings.Contains(recorder.Body.String(), "metadata") {
				t.Fatal("export must preserve references without internal metadata")
			}
			if document.CapturedAt != state.CurrentSnapshot.CreatedAt {
				t.Fatal("missing capture timestamp")
			}
		})
	}
	if len(state.Subjects[key].Tokens[0].Values) != 2 {
		t.Fatal("filtered download mutated the publication")
	}
	for _, query := range []string{"scope=invalid", "scope=view&set=missing"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/downloads/component/buttons.json?"+query, nil))
		if recorder.Code != 400 {
			t.Fatalf("invalid export selection accepted: %s", query)
		}
	}
}

func TestDownloadDialogOffersScopesAndFormats(t *testing.T) {
	var output bytes.Buffer
	page := webview.ExportPage{Slug: "buttons", Kind: "component", Set: "button|color|filled", Scope: "subject", CSS: true, Implementations: true}
	ctx := context.WithValue(context.Background(), webview.PageMetadataKey{}, webview.PageMetadata{GeneratedDownloads: true, Export: &page})
	if err := webrender.SiteDownloadMenu().Render(ctx, &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`class="tokens-export-include tokens-export-disclosure"`, `class="tokens-export-context tokens-export-disclosure"`, `<summary>`, `data-dialog-surface`, `aria-describedby="tokens-export-scope"`, "DTCG JSON", "CSS variables", "Source JSON", "Component implementation", "Current filtered table", "Whole table", "Whole collection", `action="/downloads/export"`} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q", want)
		}
	}
	for _, unwanted := range []string{"View conversion report", "Conversion report", `value="report"`} {
		if strings.Contains(output.String(), unwanted) {
			t.Fatalf("report control remains: %s", unwanted)
		}
	}
}

func TestDownloadSplitPresetsAndRegularFooter(t *testing.T) {
	page := webview.ExportPage{Kind: "component", Slug: "buttons", Scope: "subject", Implementations: true}
	ctx := context.WithValue(context.Background(), webview.PageMetadataKey{}, webview.PageMetadata{Export: &page})
	var heading, footer, table bytes.Buffer
	if err := webrender.SiteDownloadButton().Render(ctx, &heading); err != nil {
		t.Fatal(err)
	}
	if err := webrender.SiteDownloadMenu().Render(ctx, &footer); err != nil {
		t.Fatal(err)
	}
	if err := webrender.TableDownloadButton().Render(ctx, &table); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`md-split-button`, `md-filled-button md-split-leading`, `data-export-content="tokens"`, `data-export-content="implementation"`, `data-export-content="custom"`, `Download design tokens`, `Configure download`} {
		if !strings.Contains(heading.String(), want) {
			t.Fatal(want, heading.String())
		}
	}
	if strings.Contains(footer.String(), `md-split-button`) || !strings.Contains(footer.String(), `class="md-button md-filled-button"`) || !strings.Contains(footer.String(), `data-export-open`) {
		t.Fatal("footer must remain a regular filled button")
	}
	if !strings.Contains(table.String(), `md-standard-icon-button`) {
		t.Fatal("table download must be standard")
	}
}

func TestTableContextUsesNativeModalAndInfoControl(t *testing.T) {
	var output bytes.Buffer
	filters := []webview.TokenFilter{{Name: "Theme", Key: "context.Theme", Selected: "Light", Options: []string{"Light", "Dark"}}, {Name: "Platform", Key: "context.Platform", Selected: "Web", Options: []string{"Web", "Android"}}, {Name: "Audience", Key: "context.Audience", Selected: "3P", Options: []string{"1P Baseline", "3P"}}}
	if err := webrender.TableContextDialog(filters, "/components/buttons").Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	markup := output.String()
	for _, want := range []string{`<dialog`, `data-dialog-surface`, `md-basic-header`, `md-dialog-content`, `md-basic-actions`, `tokens-export-disclosure`, `aria-label="What do 1P and 3P mean?"`, `popovertarget="tokens-audience-help"`} {
		if !strings.Contains(markup, want) {
			t.Fatal(want)
		}
	}
	if strings.Index(markup, `context.Platform`) > strings.Index(markup, `context.Theme`) {
		t.Fatal("Platform must precede Theme")
	}
	if strings.Contains(markup, `tokens-audience-help"><summary>`) {
		t.Fatal("audience help must be an info control")
	}
}

func TestTableTooltipUsesPrimitiveTextAnatomy(t *testing.T) {
	var output bytes.Buffer
	if err := webrender.TableTools().Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"Download table", "Hide previews", "Show grid view", "Expand all groups"} {
		if !strings.Contains(output.String(), `role="tooltip"><p>`+label+`</p></div>`) {
			t.Fatalf("missing primitive tooltip text for %s", label)
		}
	}
}

func TestRadioChoiceUsesItsVisibleLabel(t *testing.T) {
	var body bytes.Buffer
	if err := webrender.RadioChoice(webview.ComponentCase{Family: "radio", State: "disabled"}, 0, "Only me").Render(context.Background(), &body); err != nil {
		t.Fatal(err)
	}
	markup := body.String()
	if !strings.Contains(markup, `disabled`) || !strings.Contains(markup, `<span class="md-radio-label">Only me</span></label>`) {
		t.Fatal("disabled radio must keep its visible text inside the shared choice label")
	}
	if strings.Contains(markup, `aria-label=`) {
		t.Fatal("an unrelated accessible name must not override the visible option label")
	}
}

func TestRadioGroupUsesNativeSelection(t *testing.T) {
	c := webview.ComponentCase{Family: "radio", State: "interactive", Selected: true}
	name := c.RadioInput(0)["name"]
	for index := range 3 {
		input := c.RadioInput(index)
		if input["name"] != name {
			t.Fatal("radio choices must share their native group")
		}
		_, checked := input["checked"]
		if checked != (index == 0) {
			t.Fatal("only the first option is initially selected")
		}
		control := c.RadioControl(index)
		for _, attr := range []string{"type", "name", "checked", "disabled"} {
			if _, ok := control[attr]; ok {
				t.Fatalf("native %s must belong to the input", attr)
			}
		}
	}
	if c.RadioControl(0)["data-component-case"] == c.RadioControl(1)["data-component-case"] {
		t.Fatal("radio specimens need unique identifiers")
	}
	c.State = "disabled"
	if c.RadioInput(0)["disabled"] != true {
		t.Fatal("disabled state must disable the native radio")
	}
}

func TestDownloadRadioSharesTheHostAnatomy(t *testing.T) {
	markup := downloadRadio(`"><script>`, true)
	if !strings.Contains(markup, `<span class="md-radio"><input type="radio"`) || !strings.Contains(markup, ` checked`) {
		t.Fatal("download radio host missing")
	}
	if strings.Contains(markup, `<script>`) {
		t.Fatal("radio values must be escaped")
	}
	if strings.Contains(downloadRadio("all", false), ` checked`) {
		t.Fatal("unchecked option must omit checked")
	}
}
