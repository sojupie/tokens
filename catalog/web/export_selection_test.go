package web

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
)

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
	if len(archive.File) != 5 {
		t.Fatalf("expected four exports and manifest, got %d", len(archive.File))
	}
	light, dark, css, dtcg := 0, 0, 0, 0
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
		if strings.HasSuffix(file.Name, ".tokens.json") {
			dtcg++
		}
	}
	if light != 1 || dark != 1 || css != 2 || dtcg != 2 {
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
		if err := exportChoices("format", "File formats", "css", []catalogChoice{{"css", "CSS variables"}, {"dtcg", "DTCG JSON"}}, multiple).Render(context.Background(), &output); err != nil {
			t.Fatal(err)
		}
		markup := output.String()
		for _, label := range []string{"CSS variables", "DTCG JSON", `value="css" checked`} {
			if !strings.Contains(markup, label) {
				t.Fatal("missing option label or selected value", markup)
			}
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
