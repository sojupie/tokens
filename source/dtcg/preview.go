package dtcg

import (
	"encoding/json"
	"strconv"
	"strings"
)

// ShapeCSS uses the same normalized corners as the DTCG export. Circular shapes
// use 50% in the square preview and an infinite radius in generated app CSS.
func ShapeCSS(raw json.RawMessage) string {
	node, err := Convert(raw)
	if err != nil {
		return ""
	}
	metadata := node["$extensions"].(map[string]any)[Extension].(map[string]any)
	if metadata["sourceType"] != "shape" {
		return ""
	}
	var source struct{ Shape struct{ Family string } }
	if json.Unmarshal(raw, &source) != nil {
		return ""
	}
	if source.Shape.Family == "SHAPE_FAMILY_CIRCULAR" {
		return "50%"
	}
	corners := make([]string, 0, 4)
	for _, key := range []string{"topLeft", "topRight", "bottomRight", "bottomLeft"} {
		corner, ok := node[key].(Node)
		if !ok {
			return ""
		}
		d, ok := corner["$value"].(Dimension)
		if !ok {
			return ""
		}
		corners = append(corners, strconv.FormatFloat(d.Value, 'f', -1, 64)+d.Unit)
	}
	if corners[0] == corners[1] && corners[0] == corners[2] && corners[0] == corners[3] {
		return corners[0]
	}
	return strings.Join(corners, " ")
}
