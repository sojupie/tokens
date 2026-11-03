package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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

func TestAllComponentFamiliesRenderWithoutCatalog(t *testing.T) {
	date := func(year, month, day int) time.Time {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}
	cases := []struct {
		name, module string
		component    templ.Component
	}{
		{"icon button", "common", IconButton(IconButtonProps{Label: "Favorite", Icon: "star", Variant: "tonal"})},
		{"chip", "common", Chip(ChipProps{Label: "Landscape", Selected: true, Removable: true})},
		{"floating action", "fab", FAB(FABProps{Label: "Create", Icon: "add"})},
		{"extended action", "fab", FAB(FABProps{Label: "Compose", Icon: "edit", Extended: true})},
		{"split action", "split", SplitButton(SplitButtonProps{ID: "send", Label: "Send", Icon: "send"})},
		{"button group", "button-group", GroupItem(GroupItemProps{Label: "Pause", Icon: "pause", Connected: true, Selected: true})},
		{"segmented control", "segmented", Segment(SegmentProps{Name: "view", Value: "week", Label: "Week", Selected: true})},
		{"tab", "tabs", Tab(TabProps{ID: "tab", PanelID: "panel", Label: "Activity", Content: "label", Selected: true})},
		{"navigation bar", "navigation", NavigationRegion(NavigationProps{ID: "nav", Label: "Main", Kind: "bar"})},
		{"navigation rail", "navigation", NavigationRegion(NavigationProps{ID: "rail", Label: "Main", Kind: "rail"})},
		{"navigation drawer", "navigation", Drawer("drawer", true, "Folders")},
		{"app bar", "appbar", AppBar(AppBarProps{Title: "Courses", Variant: "medium-flexible"})},
		{"toolbar", "toolbar", ToolbarSurface(ToolbarProps{Label: "Formatting"})},
		{"menu", "common", Menu("menu", "Document commands", true)},
		{"fab menu", "fab-menu", FABMenu(FABMenuProps{ID: "create", Label: "Create"})},
		{"carousel", "carousel", CarouselItem(CarouselItemProps{Index: 0, Count: 3, Label: "Collection", ImageSrc: "/painting.webp"})},
		{"search", "search", Search(SearchProps{ID: "search", Label: "Search collection", Disclosure: true})},
		{"snackbar", "messages", SnackbarMessage(SnackbarProps{ID: "notice", Text: "Archived"})},
		{"tooltip", "messages", PlainTooltip("tip", "Archive")},
		{"date picker", "date", DatePicker(DatePickerProps{ID: "date", Name: "visit_date", Value: date(2024, 2, 29), Today: date(2024, 2, 28)})},
		{"select", "common", Field(FieldProps{TextFieldProps: TextFieldProps{ID: "choice", Label: "Choose", Name: "choice"}, Kind: "select", Options: []Option{{Value: "one", Label: "One"}}})},
		{"autocomplete", "common", Field(FieldProps{TextFieldProps: TextFieldProps{ID: "find", Label: "Find", Name: "find"}, Kind: "autocomplete", Options: []Option{{Value: "one", Label: "One"}}})},
		{"divider", "common", Divider("middle", nil)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text, assets := renderAPI(t, tc.component)
			if !assets[tc.module] {
				t.Fatal("missing module", tc.module, assets)
			}
			if strings.Contains(text, "data-component-state") || strings.Contains(text, "data-component-case") {
				t.Fatal("playground state in component", text)
			}
		})
	}
}
func TestDatePickerUsesCivilDatesAcrossYearsAndLeapDays(t *testing.T) {
	value := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	end := value.AddDate(0, 0, 2)
	p := DatePickerProps{ID: "booking", Name: "start", EndName: "end", Value: value, End: end, Today: value, Range: true}
	text, _ := renderAPI(t, DatePicker(p))
	for _, want := range []string{`name="start"`, `name="end"`, `value="2024-02-29"`, `value="2024-03-02"`, `data-month="2024-02"`, `data-date="2024-02-29"`} {
		if !strings.Contains(text, want) {
			t.Fatal(want, text)
		}
	}
	rows := p.Rows(value)
	if len(rows) != 5 || rows[0][0].ISO != "2024-01-28" || rows[4][6].ISO != "2024-03-02" {
		t.Fatal(rows)
	}
	for _, row := range rows {
		for _, day := range row {
			if day.ISO == "2024-02-29" && (!day.Today || !day.Start || !day.Selected) {
				t.Fatal(day)
			}
		}
	}
}

func TestDatePickerRendersInitialCalendarHeight(t *testing.T) {
	for _, tc := range []struct {
		month string
		rows  string
	}{{"2026-02-01", "4"}, {"2026-10-01", "5"}, {"2026-05-01", "6"}} {
		t.Run(tc.month, func(t *testing.T) {
			month, err := time.Parse("2006-01-02", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			text, _ := renderAPI(t, DatePicker(DatePickerProps{
				ID: "calendar", Variant: "docked", Inline: true, Month: month,
				PanelAttrs: templ.Attributes{"style": "color:inherit"},
			}))
			if !strings.Contains(text, `style="color:inherit;--_date-rows:`+tc.rows+`"`) {
				t.Fatal("missing initial row count or consumer style", text)
			}
		})
	}
}

func TestConsumerChipAndSelectNeedNoPlaygroundAttributes(t *testing.T) {
	chip, _ := renderAPI(t, Chip(ChipProps{Label: "Saved", Variant: "filter"}))
	if !strings.Contains(chip, `data-family="chip"`) || !strings.Contains(chip, `data-chip-select`) {
		t.Fatal(chip)
	}
	selectHTML, _ := renderAPI(t, SelectField(SelectProps{ID: "platform", Name: "context.Platform", Label: "Platform", Value: "web", Options: []Option{{Value: "web", Label: "Web browser"}, {Value: "ios", Label: "iOS"}, {Value: "android", Label: "Android"}}}))
	for _, want := range []string{`name="context.Platform"`, `data-field-display>Web browser`, `data-field-value="web"`, `aria-controls="platform-options"`} {
		if !strings.Contains(selectHTML, want) {
			t.Fatal(want, selectHTML)
		}
	}
	if strings.Count(selectHTML, `name="context.Platform"`) != 1 || !strings.Contains(selectHTML, `<option value="web" selected>`) {
		t.Fatal("native selection must be the only submitted value", selectHTML)
	}
	if strings.Contains(selectHTML, "md-field-menu-divider") || strings.Contains(selectHTML, "data-component-state") {
		t.Fatal(selectHTML)
	}
	field, _ := renderAPI(t, Field(FieldProps{TextFieldProps: TextFieldProps{ID: "result", Name: "result", Value: "web", Label: "Platform", Invalid: true, Supporting: "Choose a platform."}, Kind: "select", Options: []Option{{Value: "web", Label: "Web browser"}}}))
	for _, want := range []string{`value="Web browser"`, `name="result" value="web"`, `Choose a platform.`} {
		if !strings.Contains(field, want) {
			t.Fatal(want, field)
		}
	}
	if strings.Contains(field, "Enter a valid value.") {
		t.Fatal(field)
	}
}

func TestModalCanUseItsContentRegionAsANativeForm(t *testing.T) {
	text, _ := renderAPI(t, Modal(DialogProps{ID: "download", Title: "Download", FormAttrs: templ.Attributes{"id": "export-form", "action": "/export", "method": "get"}, Actions: Action(ButtonProps{Type: "submit", Attrs: templ.Attributes{"form": "export-form"}})}))
	tree, err := html.Parse(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	forms := 0
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			a := map[string]string{}
			for _, v := range n.Attr {
				a[v.Key] = v.Val
			}
			if a["id"] == "export-form" && a["class"] == "md-dialog-content" && a["action"] == "/export" && a["method"] == "get" {
				forms++
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(tree)
	if forms != 1 || !strings.Contains(text, `form="export-form"`) {
		t.Fatal(text)
	}
}

func TestChipAndGroupConditionalContent(t *testing.T) {
	for _, component := range []templ.Component{
		Chip(ChipProps{Label: "Painting", Variant: "filter"}),
		Chip(ChipProps{Label: "Maya", Variant: "input", Avatar: "M"}),
		GroupItem(GroupItemProps{Label: "Play", Connected: true, Icon: "play_arrow"}),
	} {
		text, _ := renderAPI(t, component)
		if strings.Contains(text, "else") {
			t.Fatalf("conditional rendered as content: %s", text)
		}
	}
	text, _ := renderAPI(t, Chip(ChipProps{Label: "Maya", Variant: "input", Avatar: "M", Icon: "star"}))
	if strings.Contains(text, `data-slot="leading-icon"`) {
		t.Fatalf("avatar and leading icon both rendered: %s", text)
	}
}

func TestMenuAndBadgePreserveBaseClassesForConsumers(t *testing.T) {
	for _, component := range []templ.Component{
		MenuItem(templ.Attributes{"class": "consumer"}, "Save", nil, nil),
		Badge("small", "", templ.Attributes{"class": "consumer"}),
	} {
		text, _ := renderAPI(t, component)
		if !strings.Contains(text, `class="md-`) || !strings.Contains(text, ` consumer"`) {
			t.Fatal(text)
		}
	}
}

func TestBottomSheetHeightAnchorsAreExplicitAndFullHeightIsOptional(t *testing.T) {
	for _, full := range []bool{false, true} {
		text, _ := renderAPI(t, Sheet(SheetProps{ID: "routes", Title: "Routes", Modal: true, AllowFullHeight: full}))
		if !strings.Contains(text, `data-sheet-size="default"`) {
			t.Fatal(text)
		}
		if strings.Contains(text, `data-sheet-allow-full`) != full {
			t.Fatal(text)
		}
	}
	text, _ := renderAPI(t, Sheet(SheetProps{ID: "routes", Title: "Routes", Modal: true, Minimized: true}))
	if !strings.Contains(text, `data-sheet-size="minimized"`) || !strings.Contains(text, `aria-label="Restore sheet"`) {
		t.Fatal(text)
	}
}

func TestRailCompositionOwnsFABAndNestedDependencies(t *testing.T) {
	for _, variant := range []string{"collapsed", "expanded"} {
		text, assets := renderAPI(t, NavigationRail(RailProps{NavigationProps: NavigationProps{Label: "Mail", Variant: variant}, Menu: true, FAB: &FABProps{Label: "Compose", Icon: "edit"}}))
		for _, module := range []string{"navigation", "fab", "common"} {
			if !assets[module] {
				t.Fatalf("%s missing %s", variant, module)
			}
		}
		for _, part := range []string{"md-nav-fab", "data-navigation-expand", "<span>Compose</span>", "md-nav-items"} {
			if !strings.Contains(text, part) {
				t.Fatalf("%s missing %s", variant, part)
			}
		}
	}
}
