package web

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestStatusExamplesDescribeContent(t *testing.T) {
	for variant, want := range map[string]string{"labeled": "99+ unread notifications", "dot": "New notification"} {
		var out bytes.Buffer
		if err := specimenBadge(componentCase{Family: "badge", Variant: variant}).Render(context.Background(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `aria-label="`+want+`"`) {
			t.Errorf("%s missing %s", variant, want)
		}
	}
	c := componentCase{Family: "progress", Control: "determinate"}
	if c.ProgressAttributes()["aria-label"] != "Loading preview content" {
		t.Error("Progress name must identify the affected content")
	}
}

func TestSnackbarDocumentsItsFocusShortcut(t *testing.T) {
	c := componentCase{Family: "snackbar", State: "interactive"}
	var out bytes.Buffer
	if err := specimenSnackbar(c).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`aria-describedby="` + c.ID() + `-shortcut"`, `id="` + c.ID() + `-shortcut"`, `Press F6 to focus the snackbar action.`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s", want)
		}
	}
}

func TestBadgePreviewExercisesAnchorsAndOverflowCounts(t *testing.T) {
	page := &componentPreview{Cases: []componentCase{{Family: "badge", Variant: "labeled"}}}
	if err := configureBadgePreview(page, map[string][]string{"preview.content": {"1000"}, "preview.placement": {"trailing"}, "preview.direction": {"rtl"}}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := specimenBadge(page.Cases[0]).Render(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`dir="rtl"`, `data-badge-placement="trailing"`, `1000 unread notifications`, `999+`, `aria-hidden="true"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %s: %s", want, out.String())
		}
	}
	if err := configureBadgePreview(page, map[string][]string{"preview.content": {"arbitrary"}}); err == nil {
		t.Fatal("invalid preview content accepted")
	}
}
