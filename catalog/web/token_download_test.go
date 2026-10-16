package web

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

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

func TestTokenDownloadControlExplainsChoices(t *testing.T) {
	var output bytes.Buffer
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}, Name: "Buttons"}
	browser := tokenBrowser{Sets: []tokenSet{{}}, Selected: tokenSet{Key: "button|color|outlined", Label: "Button - Color - Outlined"}, Query: `icon & label`}
	if err := writeTokenDownloadMenu(&output, subject, browser, downloadPreview{}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`method="get"`, `aria-labelledby="tokens-download-title"`, `value="view" checked`, `value="set"`, `value="all"`, `icon &amp; label`, `Download JSON</button>`, `Button - Color - Outlined`, `This table, with your filters`, `This table, with every value`, `All Buttons tokens`, `Ignores search and filters.`} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %s in %s", want, output.String())
		}
	}
	if strings.Contains(output.String(), "<select") {
		t.Fatal("choices should be explained together, not hidden in a scope selector")
	}
	output.Reset()
	if err := writeTokenDownloadMenu(&output, subject, tokenBrowser{}, downloadPreview{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `value="all" checked`) || strings.Contains(output.String(), `value="view"`) {
		t.Fatal("page without a table must offer only the complete download")
	}
}

func TestDownloadPreviewsMatchTheExportScopes(t *testing.T) {
	values := []publish.TokenValueView{
		{ContextLabels: []string{"Theme: Light", "Platform: Web"}, Value: json.RawMessage(`{"color":{"red":1}}`)},
		{ContextLabels: []string{"Theme: Dark", "Platform: Web"}, Value: json.RawMessage(`{"color":{"red":0}}`)},
	}
	key := publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}
	subject := publish.SubjectView{Key: key, Name: "Buttons", Tokens: []publish.TokenView{
		{SourceName: "md.comp.button.outlined.icon.color", DisplayName: "Icon <color>", Category: "COLOR", Values: values},
		{SourceName: "md.comp.button.outlined.container.color", Category: "COLOR", Values: values},
		{SourceName: "md.comp.button.filled.icon.color", Category: "COLOR", Values: values},
	}}
	styleKey := publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "elevation"}
	style := publish.SubjectView{Key: styleKey, Name: "Elevation", Tokens: []publish.TokenView{{SourceName: "md.sys.elevation.level1", Values: values[:1]}}}
	state := publish.PublishedState{Subjects: map[publish.SubjectKey]publish.SubjectView{key: subject, styleKey: style}}
	query := url.Values{"set": {"button|color|outlined"}, "q": {"icon"}, "context.Theme": {"Light"}}
	browser := buildTokenBrowser(subject, query)
	var output bytes.Buffer
	if err := writeTokenDownloadMenu(&output, subject, browser, previewSiteDownload(state)); err != nil {
		t.Fatal(err)
	}
	for scope, counts := range map[string]string{
		"view":       "1 token · 1 value · 1 table",
		"set":        "2 tokens · 4 values · 1 table",
		"all":        "3 tokens · 6 values · 2 tables",
		"everything": "4 tokens · 7 values · 2 components and styles",
	} {
		option := strings.SplitN(output.String(), `name="scope" value="`+scope+`"`, 2)
		if len(option) != 2 || !strings.Contains(strings.SplitN(option[1], `</label>`, 2)[0], counts) {
			t.Fatalf("%s has incorrect preview: %s", scope, output.String())
		}
	}
	if !strings.Contains(output.String(), "Icon &lt;color&gt;") || !strings.Contains(output.String(), "Theme: Light") || !strings.Contains(output.String(), "Theme: Dark, Light") {
		t.Fatal("examples were missing or not escaped")
	}
	query.Set("q", "no matches")
	output.Reset()
	if err := writeTokenDownloadMenu(&output, subject, buildTokenBrowser(subject, query), previewSiteDownload(state)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "0 tokens · 0 values · 1 table") || !strings.Contains(output.String(), "No tokens match this selection.") {
		t.Fatal("empty filtered selection has a misleading preview")
	}
}
