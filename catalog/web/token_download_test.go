package web

import (
	"bytes"
	"context"
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

func TestDownloadDialogOffersScopesAndFormats(t *testing.T) {
	var output bytes.Buffer
	page := exportPage{Slug: "buttons", Kind: "component", Set: "button|color|filled", Scope: "subject", CSS: true, Implementations: true}
	ctx := context.WithValue(context.Background(), pageMetadataKey{}, pageMetadata{GeneratedDownloads: true, Export: &page})
	if err := siteDownloadMenu().Render(ctx, &output); err != nil {
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
	page := exportPage{Kind: "component", Slug: "buttons", Scope: "subject", Implementations: true}
	ctx := context.WithValue(context.Background(), pageMetadataKey{}, pageMetadata{Export: &page})
	var heading, footer, table bytes.Buffer
	if err := siteDownloadButton().Render(ctx, &heading); err != nil {
		t.Fatal(err)
	}
	if err := siteDownloadMenu().Render(ctx, &footer); err != nil {
		t.Fatal(err)
	}
	if err := tableDownloadButton().Render(ctx, &table); err != nil {
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
	filters := []tokenFilter{{Name: "Theme", Key: "context.Theme", Selected: "Light", Options: []string{"Light", "Dark"}}, {Name: "Platform", Key: "context.Platform", Selected: "Web", Options: []string{"Web", "Android"}}, {Name: "Audience", Key: "context.Audience", Selected: "3P", Options: []string{"1P Baseline", "3P"}}}
	if err := tableContextDialog(filters, "/components/buttons").Render(context.Background(), &output); err != nil {
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
	if err := tableTools().Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"Download table", "Hide previews", "Show grid view", "Expand all groups"} {
		if !strings.Contains(output.String(), `role="tooltip"><p>`+label+`</p></div>`) {
			t.Fatalf("missing primitive tooltip text for %s", label)
		}
	}
}
