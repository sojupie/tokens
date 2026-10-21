package web

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestGlobalHistorySummarizesWholeComponentsWithoutNestedPaging(t *testing.T) {
	state := publish.PublishedState{}
	for i := range 201 {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: "component-00", SnapshotID: "new", Name: fmt.Sprintf("large-token-%03d", i)})
	}
	// Another context for the same token must not count as another item.
	state.Changes = append(state.Changes, state.Changes[0])
	for i := 1; i <= 10; i++ {
		state.Changes = append(state.Changes, publish.ChangeView{Subject: fmt.Sprintf("component-%02d", i), SnapshotID: "new", Name: "small-token"})
	}
	state.Changes = append(state.Changes, publish.ChangeView{Subject: "component-00", SnapshotID: "old", Name: "earlier-token"})
	handler := New(state, "").Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/changes?page=8", nil))
	body := recorder.Body.String()
	if strings.Count(body, `class="md-card md-outlined-card tokens-change-summary"`) != 12 || !strings.Contains(body, "201 items changed") || strings.Contains(body, "large-token-") {
		t.Fatal("summary must batch by observation/component and count distinct items")
	}
	for _, unwanted := range []string{"Load changes", "tokens-pagination", "@materialSymbol", ` selected="false"`} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("overview still contains %s", unwanted)
		}
	}
	if !strings.Contains(body, "/components/component-00/changes?snapshot=new") || !strings.Contains(body, "/components/component-00/changes?snapshot=old") {
		t.Fatal("summary links must select the relevant component and observation")
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/components/component-00/changes?snapshot=new&page=2", nil))
	body = recorder.Body.String()
	if !strings.Contains(body, "large-token-050") || !strings.Contains(body, "large-token-099") || strings.Contains(body, "large-token-100") || strings.Contains(body, "earlier-token") {
		t.Fatal("component page must show bounded details for the chosen observation")
	}
}

func TestMaterialIconsDoNotUseHandDrawnOrCharacterFallbacks(t *testing.T) {
	for _, name := range []string{"search", "eye", "grid", "expand", "dropdown", "folder"} {
		icon := tableIcon(name)
		if !strings.Contains(icon, `<use href="#tokens-table-icon-`) || !strings.Contains(icon, `class="tokens-material-symbol"`) || strings.Contains(icon, "stroke=") {
			t.Fatalf("%s is not the shared Material Symbols SVG", name)
		}
	}
}
