package web

import "testing"

func TestListSemanticsFollowBehavior(t *testing.T) {
	c := componentCase{Family: "list", Control: "select", Selection: "single", State: "disabled"}
	if c.ListItem(0)["aria-pressed"] != "false" || c.ListItem(1)["aria-pressed"] != "true" || c.ListItem(0)["disabled"] != true {
		t.Fatal("single selection and disabled state are not reflected in native controls")
	}
	c.Control = "expand"
	c.Expanded = true
	if c.ListItem(0)["aria-expanded"] != "true" || c.ListItem(1)["aria-expanded"] != "false" {
		t.Fatal("only the first disclosure should start expanded")
	}
	c.Control = "reveal"
	if c.ListUtility(0, true)["class"] != "md-icon-button md-list-reveal-action" {
		t.Fatal("reveal actions must use their own token profile")
	}
}

func TestListPreviewCompositions(t *testing.T) {
	c := componentCase{Family: "list", Control: "select", Variant: "standard", Width: "video", Size: "three-line", State: "enabled"}
	page := &componentPreview{Cases: []componentCase{c}}
	if err := configureListPreview(page, map[string][]string{"preview.media": {"large"}, "preview.trailing": {"checkbox"}, "preview.divider": {"none"}, "preview.text": {"wrapped"}}); err != nil {
		t.Fatal(err)
	}
	got := page.Cases[0]
	if got.listMedia() != "large" || got.listDivider() != "none" || got.ListNativeKind() != "checkbox" {
		t.Fatalf("composition choices lost: %+v", got)
	}
	a := got.ListItem(0)
	if _, ok := a["type"]; ok {
		t.Fatal("native controls must not be nested inside a button")
	}
	if _, ok := a["data-button-toggle"]; ok {
		t.Fatal("native controls must own selection")
	}
	if got.ListControlInput(0)["type"] != "checkbox" {
		t.Fatal("checkbox input missing")
	}
	got.Width = "radio"
	got.ListTrailing = "text"
	if !got.ListSelected(1) || got.ListSelected(0) {
		t.Fatal("radio composition must have one selected row")
	}
	if err := configureListPreview(&componentPreview{Cases: []componentCase{c}}, map[string][]string{"preview.media": {"unknown"}}); err == nil {
		t.Fatal("invalid media size accepted")
	}
}

func TestListBehaviorCompositions(t *testing.T) {
	for _, mode := range []string{"single", "multiple"} {
		page := &componentPreview{Cases: []componentCase{{Family: "list", Control: "select", Selection: mode}}}
		if err := configureListPreview(page, nil); err != nil {
			t.Fatal(err)
		}
		for _, control := range page.Controls {
			if control.Name == "trailing" {
				for _, value := range control.Options {
					if mode == "single" && (value == "checkbox" || value == "switch") || mode == "multiple" && value == "radio" {
						t.Fatalf("invalid %s selection control: %s", mode, value)
					}
				}
			}
		}
	}
	page := &componentPreview{Cases: []componentCase{{Family: "list", Control: "reveal"}}}
	if err := configureListPreview(page, map[string][]string{"preview.trailing": {"icon"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].ListTrailing == "icon" {
		t.Fatal("reveal must not duplicate its utility action with a navigation chevron")
	}
	page = &componentPreview{Cases: []componentCase{{Family: "list", Control: "expand"}}}
	if err := configureListPreview(page, map[string][]string{"preview.inset": {"inset"}, "preview.surface": {"none"}, "preview.expand-icon": {"plain"}}); err != nil {
		t.Fatal(err)
	}
	if page.Cases[0].ListInset != "inset" || page.Cases[0].ListSurface != "none" || page.Cases[0].ListExpandIcon != "plain" {
		t.Fatal("independent expansion options lost")
	}
	if _, ok := (componentCase{Control: "reorder"}).ListItem(0)["aria-pressed"]; ok {
		t.Fatal("reordering must not toggle persistent row selection")
	}
}
