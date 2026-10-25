package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAppbarSearchCompositionUsesScopedSearch(t *testing.T) {
	a := componentCase{Family: "appbar", Variant: "small", Control: "search", State: "interactive"}
	b := a
	b.Variant = "centered"
	child := a.AppbarSearch()
	if child.Family != "search" || child.Variant != "adaptive" || child.Control != "contained" || child.ID() == b.AppbarSearch().ID() || child.ID() == a.ID() {
		t.Fatal("nested search lost its configuration or instance scope")
	}
	var out bytes.Buffer
	if err := specimenAppbar(a).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{`data-search-open`, `aria-controls="` + child.ID() + `-view"`, `class="md-search-input"`, `data-search-result="Buttons"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing composed search markup %s", want)
		}
	}
	a.Control = "title"
	out.Reset()
	if err := specimenAppbar(a).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), `class="md-search-input"`) {
		t.Fatal("ordinary app bar rendered unrelated search content")
	}
}
