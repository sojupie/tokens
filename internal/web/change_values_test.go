package web

import (
	"reflect"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
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
