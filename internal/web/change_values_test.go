package web

import (
	"reflect"
	"testing"

	"github.com/sojupie/tokens/internal/publish"
)

func TestChangedValuesOnlyShowDifferencesAndKeepUnits(t *testing.T) {
	change := publish.ChangeView{Type: "changed", Before: `{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":500,"fontSize":{"unit":"POINTS","value":14}}}`, After: `{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":600,"fontSize":{"unit":"POINTS","value":16}}}`}
	want := []changeValue{{"Font size", "14pt", "16pt"}, {"Font weight", "500", "600"}}
	if got := changedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("differences = %#v, want %#v", got, want)
	}
	change = publish.ChangeView{Type: "added", After: "48"}
	if got := changedValues(change); !reflect.DeepEqual(got, []changeValue{{"Value", "—", "48"}}) {
		t.Fatalf("added = %#v", got)
	}
}

func TestAddedMeasurementsDistinguishMissingFromZero(t *testing.T) {
	change := publish.ChangeView{Type: "added", After: `{"type":{"fontSize":{"unit":"POINTS","value":0},"fontTracking":{"unit":"POINTS"}}}`}
	want := []changeValue{{"Font size", "—", "0pt"}, {"Letter spacing", "—", "Not specified (pt)"}}
	if got := changedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
}

func TestAddedAndRemovedTypographyUsePropertyRows(t *testing.T) {
	value := `{"type":{"fontNames":{"values":["Roboto Flex"]},"fontSize":{"unit":"POINTS","value":16},"fontWeight":500,"lineHeight":{"unit":"POINTS","value":24},"fontTracking":{"unit":"POINTS","value":0},"variationAxes":{"wght":{"axisValue":500},"opsz":{"axisValue":16}}}}`
	want := []changeValue{
		{"Font family", "—", "Roboto Flex"},
		{"Font size", "—", "16pt"},
		{"Letter spacing", "—", "0pt"},
		{"Font weight", "—", "500"},
		{"Line height", "—", "24pt"},
		{"Font axes · opsz", "—", "16"},
		{"Font axes · wght", "—", "500"},
	}
	if got := changedValues(publish.ChangeView{Type: "added", After: value}); !reflect.DeepEqual(got, want) {
		t.Fatalf("added rows = %#v", got)
	}
	for i := range want {
		want[i].Before, want[i].After = want[i].After, want[i].Before
	}
	if got := changedValues(publish.ChangeView{Type: "removed", Before: value}); !reflect.DeepEqual(got, want) {
		t.Fatalf("removed rows = %#v", got)
	}
}

func TestCompoundPropertyAdditionRetainsUnitsAndOnlyChangedAxes(t *testing.T) {
	change := publish.ChangeView{Type: "changed", Before: `{"type":{"variationAxes":{"wght":{"axisValue":500},"opsz":{"axisValue":16}}}}`, After: `{"type":{"fontSize":{"unit":"POINTS","value":16},"variationAxes":{"wght":{"axisValue":600},"opsz":{"axisValue":16}}}}`}
	want := []changeValue{{"Font size", "—", "16pt"}, {"Font axes · wght", "500", "600"}}
	if got := changedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("changed rows = %#v", got)
	}
}

func TestHistoricalContextAndReferencesAreShownWithoutEnvelopeDump(t *testing.T) {
	change := publish.ChangeView{Kind: "contextualReferenceTree", Type: "changed",
		Before: `{"context":["Theme: Dark","Platform: Web"],"default":false,"value":{"type":{"fontWeight":500}},"references":[{"name":"md.sys.typescale.body-large"}]}`,
		After:  `{"context":["Theme: Dark","Platform: Web"],"default":false,"value":{"type":{"fontWeight":600}},"references":[{"name":"md.sys.typescale.body-medium"}]}`,
	}
	if got := changeContext(change); got != "Theme: Dark · Platform: Web" {
		t.Fatal(got)
	}
	want := []changeValue{{"Font weight", "500", "600"}, {"References", "md.sys.typescale.body-large", "md.sys.typescale.body-medium"}}
	if got := changedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	change.Type = "removed"
	change.After = ""
	if got := changeContext(change); got != "Theme: Dark · Platform: Web" {
		t.Fatal(got)
	}
	change.Before = `{"number":1}`
	if got := changeContext(change); got != "Context not recorded" {
		t.Fatal(got)
	}
}
