package presentation_test

import (
	"reflect"
	"testing"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	"github.com/sojupie/tokens/source/publish"
)

func TestChangedValuesOnlyShowDifferencesAndKeepUnits(t *testing.T) {
	change := publish.ChangeView{Type: "changed", Before: `{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":500,"fontSize":{"unit":"POINTS","value":14}}}`, After: `{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":600,"fontSize":{"unit":"POINTS","value":16}}}`}
	want := []webview.ChangeValue{{"Font size", "14pt", "16pt"}, {"Font weight", "500", "600"}}
	if got := webview.ChangedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("differences = %#v, want %#v", got, want)
	}
	change = publish.ChangeView{Type: "added", After: "48"}
	if got := webview.ChangedValues(change); !reflect.DeepEqual(got, []webview.ChangeValue{{"Value", "—", "48"}}) {
		t.Fatalf("added = %#v", got)
	}
}

func TestAddedMeasurementsUseSourceZeroDefaults(t *testing.T) {
	change := publish.ChangeView{Type: "added", After: `{"type":{"fontSize":{"unit":"POINTS","value":0},"fontTracking":{"unit":"POINTS"}}}`}
	want := []webview.ChangeValue{{"Font size", "—", "0pt"}, {"Letter spacing", "—", "0pt"}}
	if got := webview.ChangedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
}

func TestAddedAndRemovedTypographyUsePropertyRows(t *testing.T) {
	value := `{"type":{"fontNames":{"values":["Roboto Flex"]},"fontSize":{"unit":"POINTS","value":16},"fontWeight":500,"lineHeight":{"unit":"POINTS","value":24},"fontTracking":{"unit":"POINTS","value":0},"variationAxes":{"wght":{"axisValue":500},"opsz":{"axisValue":16}}}}`
	want := []webview.ChangeValue{
		{"Font family", "—", "Roboto Flex"},
		{"Font size", "—", "16pt"},
		{"Letter spacing", "—", "0pt"},
		{"Font weight", "—", "500"},
		{"Line height", "—", "24pt"},
		{"Font axes · opsz", "—", "16"},
		{"Font axes · wght", "—", "500"},
	}
	if got := webview.ChangedValues(publish.ChangeView{Type: "added", After: value}); !reflect.DeepEqual(got, want) {
		t.Fatalf("added rows = %#v", got)
	}
	for i := range want {
		want[i].Before, want[i].After = want[i].After, want[i].Before
	}
	if got := webview.ChangedValues(publish.ChangeView{Type: "removed", Before: value}); !reflect.DeepEqual(got, want) {
		t.Fatalf("removed rows = %#v", got)
	}
}

func TestCompoundPropertyAdditionRetainsUnitsAndOnlyChangedAxes(t *testing.T) {
	change := publish.ChangeView{Type: "changed", Before: `{"type":{"variationAxes":{"wght":{"axisValue":500},"opsz":{"axisValue":16}}}}`, After: `{"type":{"fontSize":{"unit":"POINTS","value":16},"variationAxes":{"wght":{"axisValue":600},"opsz":{"axisValue":16}}}}`}
	want := []webview.ChangeValue{{"Font size", "—", "16pt"}, {"Font axes · wght", "500", "600"}}
	if got := webview.ChangedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("changed rows = %#v", got)
	}
}

func TestHistoricalContextAndReferencesAreShownWithoutEnvelopeDump(t *testing.T) {
	change := publish.ChangeView{Kind: "contextualReferenceTree", Type: "changed",
		Before: `{"context":["Theme: Dark","Platform: Web"],"default":false,"value":{"type":{"fontWeight":500}},"references":[{"name":"md.sys.typescale.body-large"}]}`,
		After:  `{"context":["Theme: Dark","Platform: Web"],"default":false,"value":{"type":{"fontWeight":600}},"references":[{"name":"md.sys.typescale.body-medium"}]}`,
	}
	if got := webview.ChangeContext(change); got != "Theme: Dark · Platform: Web" {
		t.Fatal(got)
	}
	want := []webview.ChangeValue{{"Font weight", "500", "600"}, {"References", "md.sys.typescale.body-large", "md.sys.typescale.body-medium"}}
	if got := webview.ChangedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	change.Type = "removed"
	change.After = ""
	if got := webview.ChangeContext(change); got != "Theme: Dark · Platform: Web" {
		t.Fatal(got)
	}
	change.Before = `{"number":1}`
	if got := webview.ChangeContext(change); got != "Context not recorded" {
		t.Fatal(got)
	}
}
