package web

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
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

func TestTokenDownloadControlUsesNativeScopeForm(t *testing.T) {
	var output bytes.Buffer
	subject := publish.SubjectView{Key: publish.SubjectKey{Kind: publish.SubjectComponent, Slug: "buttons"}}
	browser := tokenBrowser{Sets: []tokenSet{{}}, Selected: tokenSet{Key: "button|color|outlined"}, Query: `icon & label`}
	if err := writeTokenDownloadMenu(&output, subject, browser); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`method="get"`, `aria-label="Download scope"`, `value="view"`, `value="set"`, `value="all"`, `icon &amp; label`, `Download JSON</button>`} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %s in %s", want, output.String())
		}
	}
	if strings.Contains(output.String(), "<details") {
		t.Fatal("download control should not be a popup")
	}
}
