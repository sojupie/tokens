package dtcg

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sojupie/tokens/source/publish"
)

func TestMaterialConversionPreservesMeaning(t *testing.T) {
	cases := []struct {
		raw, kind string
		want      any
	}{
		{`{"length":{"unit":"DIPS"}}`, "dimension", Dimension{0, "px"}},
		{`{"fontTracking":{"unit":"POINTS"}}`, "dimension", Dimension{0, "px"}},
		{`{"elevation":{"value":3,"unit":"DIPS"}}`, "dimension", Dimension{3, "px"}},
		{`{"durationMs":250}`, "duration", map[string]any{"unit": "ms", "value": float64(250)}},
		{`{"cubicBezier":{"x0":0.2,"y1":1}}`, "cubicBezier", []float64{0.2, 0, 0, 1}},
		{`{"opacity":0}`, "number", float64(0)},
	}
	for _, c := range cases {
		n, err := Convert(json.RawMessage(c.raw))
		if err != nil {
			t.Fatal(err)
		}
		if n["$type"] != c.kind {
			t.Fatal(n)
		}
		actual, _ := json.Marshal(n["$value"])
		want, _ := json.Marshal(c.want)
		if string(actual) != string(want) {
			t.Fatalf("%s: got %s, want %s", c.raw, actual, want)
		}
	}
	n, err := Convert(json.RawMessage(`{"type":{"fontNames":{"values":["Roboto"]},"fontSize":{"value":16,"unit":"POINTS"},"fontWeight":400,"fontTracking":{"unit":"POINTS"},"lineHeight":{"value":24,"unit":"POINTS"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if n["$value"].(map[string]any)["lineHeight"] != 1.5 {
		t.Fatal("line height must be a multiplier", n)
	}
	shape, err := Convert(json.RawMessage(`{"shape":{"family":"SHAPE_FAMILY_ROUNDED_CORNERS","defaultSize":{"value":16,"unit":"DIPS"},"topLeft":{"unit":"DIPS"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if shape["topLeft"].(Node)["$value"].(Dimension).Value != 0 || shape["bottomRight"].(Node)["$value"].(Dimension).Value != 16 {
		t.Fatal(shape)
	}
	for _, raw := range []string{`{"length":{"unit":"DIPS","value":null}}`, `{"length":{"unit":"PERCENT","value":null}}`, `{"color":{}}`, `{"color":{"alpha":2}}`, `{"color":{"alpha":1,"red":null}}`, `{"cubicBezier":{"x0":null}}`, `{"axisValue":{"tag":"wght","value":null}}`, `{"undefined":true}`, `{"svgPath":"M 0,0"}`} {
		if _, err := Convert(json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestExportAliasesContextsAndRootTokens(t *testing.T) {
	value := func(n int, labels ...string) publish.TokenValueView {
		return publish.TokenValueView{ContextLabels: labels, Value: json.RawMessage(`{"length":{"value":` + strconv.Itoa(n) + `,"unit":"DIPS"}}`)}
	}
	v := value(4, "Theme: Light")
	v.References = []publish.TokenReference{{Name: "ref.size"}}
	state := publish.PublishedState{PublicationID: "test", Subjects: map[publish.SubjectKey]publish.SubjectView{{Kind: publish.SubjectComponent, Slug: "test"}: {Tokens: []publish.TokenView{
		{SourceName: "size", Values: []publish.TokenValueView{v, value(8, "Theme: Dark")}},
		{SourceName: "size.child", Values: []publish.TokenValueView{value(2)}},
	}}}}
	document, report, err := Export(state, map[string]string{"Theme": "Light"})
	if err != nil {
		t.Fatal(err)
	}
	root := document["size"].(Node)["$root"].(Node)
	if root["$value"] != "{ref.size}" || report.Aliases != 1 || report.Exported != 3 {
		t.Fatal(document, report)
	}
	if _, exists := document["ref"].(Node)["size"]; !exists {
		t.Fatal("missing alias dependency")
	}
	if _, _, err := Export(state, map[string]string{"Theme": "Purple"}); err == nil {
		t.Fatal("accepted nonexistent theme")
	}
	if _, _, err := Export(state, map[string]string{}); err == nil {
		t.Fatal("arbitrary context selection")
	}
}

func TestExportRealPublication(t *testing.T) {
	database := os.Getenv("TOKEN_CATALOG_THEME_DB")
	if database == "" {
		t.Skip("set TOKEN_CATALOG_THEME_DB to validate the captured publication")
	}
	state, err := publish.Load(database)
	if err != nil {
		t.Fatal(err)
	}
	for _, theme := range []string{"Light", "Dark"} {
		for _, contrast := range []string{"Default", "Medium contrast", "High contrast"} {
			context := DefaultContext()
			context["Theme"] = theme
			context["Contrast"] = contrast
			document, report, err := Export(state, context)
			if err != nil {
				t.Fatal(err)
			}
			if report.Exported < 6000 || report.Aliases < 4000 {
				t.Fatal(report)
			}
			var walk func(Node, []string)
			walk = func(node Node, path []string) {
				if _, exists := node["$value"]; exists {
					originalName := strings.Join(path, ".")
					if originalName == "md.sys.shape.corner.large.end" {
						t.Fatal("shape was flattened to a scalar")
					}
					if node["$type"] == nil {
						t.Fatal("missing type", originalName)
					}
				}
				for key, value := range node {
					if !strings.HasPrefix(key, "$") {
						walk(value.(Node), append(append([]string{}, path...), key))
					} else if key == "$root" {
						walk(value.(Node), path)
					}
				}
			}
			walk(document, nil)
		}
	}
}

func TestPercentageLengthRetainsSourceWithoutInventingDTCGUnit(t *testing.T) {
	raw := json.RawMessage(`{"length":{"unit":"PERCENT","value":50}}`)
	n, err := Convert(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := n["$value"]; exists {
		t.Fatal("percentage became a DTCG dimension")
	}
	meta := n["$extensions"].(map[string]any)[Extension].(map[string]any)
	if string(meta["sourceValue"].(json.RawMessage)) != string(raw) {
		t.Fatal("percentage source changed")
	}
}
