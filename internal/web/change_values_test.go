package web

import (
	"reflect"
	"testing"

	"github.com/sojupie/m3-source/internal/publish"
)

func TestChangedValuesOnlyShowDifferencesAndKeepUnits(t *testing.T) {
	change := publish.ChangeView{Type: "changed", Before: `{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":500,"fontSize":{"unit":"POINTS","value":14}}}`, After: `{"type":{"fontNames":{"values":["Roboto"]},"fontWeight":600,"fontSize":{"unit":"POINTS","value":16}}}`}
	want := []changeValue{{"Type · Font Size", "14pt", "16pt"}, {"Type · Font Weight", "500", "600"}}
	if got := changedValues(change); !reflect.DeepEqual(got, want) {
		t.Fatalf("differences = %#v, want %#v", got, want)
	}
	change = publish.ChangeView{Type: "added", After: "48"}
	if got := changedValues(change); !reflect.DeepEqual(got, []changeValue{{"Value", "—", "48"}}) {
		t.Fatalf("added = %#v", got)
	}
}
