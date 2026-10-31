package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
)

func renderAPI(t *testing.T, c templ.Component) (string, Assets) {
	t.Helper()
	var out bytes.Buffer
	assets := Assets{}
	if err := c.Render(WithAssets(context.Background(), assets), &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	tree, err := html.Parse(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		seen := map[string]bool{}
		for _, a := range n.Attr {
			if seen[a.Key] {
				t.Fatalf("duplicate %s on %s: %s", a.Key, n.Data, text)
			}
			seen[a.Key] = true
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(tree)
	return text, assets
}
func TestTypedControlsPreserveNativeFormValues(t *testing.T) {
	for _, c := range []templ.Component{TextField(TextFieldProps{ID: "title", Name: "title", Label: "Title", Value: "A < B", Required: true}), CheckboxField(ChoiceProps{ID: "notify", Name: "notify", Label: "Notify me", Value: "yes", Checked: true}), Radio(ChoiceProps{ID: "delivery", Name: "delivery", Label: "Express", Value: "express", Checked: true})} {
		text, _ := renderAPI(t, c)
		if !strings.Contains(text, `name="`) {
			t.Fatal(text)
		}
	}
	text, _ := renderAPI(t, TextField(TextFieldProps{ID: "title", Label: "Title", Value: "A < B", Supporting: "Required", Invalid: true}))
	for _, want := range []string{`for="title"`, `A &lt; B`, `aria-describedby="title-help"`, `id="title-help"`, `aria-invalid="true"`} {
		if !strings.Contains(text, want) {
			t.Fatal(want, text)
		}
	}
}
func TestNestedComponentsRegisterTheirOwnModules(t *testing.T) {
	p := ListItemProps{Label: "Updates", LeadingKind: LeadingAvatar, Avatar: "M", Badge: Badge("small", "", nil), Trailing: SwitchControl(ChoiceProps{ID: "updates", Label: "Updates"})}
	text, assets := renderAPI(t, ListRow(p))
	for _, module := range []string{"common", "list", "badge", "switch"} {
		if !assets[module] {
			t.Fatal("missing", module, assets)
		}
	}
	if strings.Contains(text, "data-unread") {
		t.Fatal("read row marked unread", text)
	}
}
func TestSliderAPIKeepsBoundsAndRangeValues(t *testing.T) {
	start := 15.0
	text, assets := renderAPI(t, Range(SliderProps{ID: "price", Name: "price", Label: "Price", Min: 10, Max: 80, Value: 70, Step: 5, Start: &start}))
	for _, want := range []string{`name="price-start"`, `name="price-end"`, `value="15"`, `value="70"`, `min="10"`, `max="80"`, `step="5"`} {
		if !strings.Contains(text, want) {
			t.Fatal(want, text)
		}
	}
	if !assets["slider"] {
		t.Fatal(assets)
	}
}
func TestTimePickerRendersItsConfiguredValue(t *testing.T) {
	text, assets := renderAPI(t, TimePicker(TimePickerProps{ID: "alarm", Name: "alarm", Hour: 21, Minute: 45, Format24: true}))
	for _, want := range []string{`value="21:45"`, `data-hour="21"`, `data-minute="45"`, `value="21"`, `value="45"`, `data-ring="inner"`, `--_time-angle:270deg;`} {
		if !strings.Contains(text, want) {
			t.Fatal(want, text)
		}
	}
	if !assets["time"] {
		t.Fatal(assets)
	}
}
func TestModalAndSheetOwnAccessiblePanelReferences(t *testing.T) {
	text, _ := renderAPI(t, Modal(DialogProps{ID: "delete", Title: "Delete?", Description: "Can be restored."}))
	for _, want := range []string{`id="delete"`, `aria-labelledby="delete-heading"`, `id="delete-heading"`, `aria-describedby="delete-support"`, `id="delete-support"`} {
		if !strings.Contains(text, want) {
			t.Fatal(want, text)
		}
	}
	text, _ = renderAPI(t, Sheet(SheetProps{ID: "details", Title: "Details", Modal: true}))
	for _, want := range []string{`<dialog`, `aria-controls="details-content"`, `id="details-content"`, `aria-labelledby="details-heading"`, `id="details-heading"`} {
		if !strings.Contains(text, want) {
			t.Fatal(want, text)
		}
	}
}
func TestInteractiveCardsKeepNativeSemantics(t *testing.T) {
	text, _ := renderAPI(t, CardLink("/artwork", CardProps{Headline: "Artwork"}))
	if !strings.Contains(text, `href="/artwork"`) {
		t.Fatal(text)
	}
	text, _ = renderAPI(t, CardButton(CardProps{Headline: "Select", Attrs: templ.Attributes{"class": "md-card", "aria-pressed": "false"}}))
	if strings.Contains(text, "<h2") || strings.Contains(text, "<div") || strings.Contains(text, "md-card md-card") {
		t.Fatal(text)
	}
}
