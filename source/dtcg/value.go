// Package dtcg converts Material's captured values to DTCG 2025.10 tokens.
package dtcg

import (
	"encoding/json"
	"fmt"
	"math"
)

const Extension = "dev.tokencatalog.material"

type Node map[string]any

type Dimension struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func token(kind string, value any) Node { return Node{"$type": kind, "$value": value} }

// Convert uses the Web unit policy: logical DIPS and type-scale POINTS map to
// CSS pixels. The original value remains in the token's source extension.
func Convert(raw json.RawMessage) (Node, error) {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if len(value) != 1 {
		return nil, fmt.Errorf("expected one typed source value")
	}
	var node Node
	var err error
	for kind, body := range value {
		switch kind {
		case "color":
			if err := validateNumbers(body, "red", "green", "blue", "alpha"); err != nil {
				return nil, err
			}
			var c struct {
				Red, Green, Blue float64
				Alpha            *float64
			}
			if err = json.Unmarshal(body, &c); err != nil || c.Alpha == nil {
				return nil, fmt.Errorf("color requires alpha")
			}
			for _, n := range []float64{c.Red, c.Green, c.Blue, *c.Alpha} {
				if n < 0 || n > 1 {
					return nil, fmt.Errorf("color channel outside [0,1]")
				}
			}
			node = token("color", map[string]any{"colorSpace": "srgb", "components": []float64{c.Red, c.Green, c.Blue}, "alpha": *c.Alpha})
		case "length", "fontSize", "lineHeight", "fontTracking", "elevation":
			var d Dimension
			d, err = dimension(body)
			node = token("dimension", d)
		case "fontWeight", "opacity", "numeric", "durationMs":
			var n *float64
			if err = json.Unmarshal(body, &n); err != nil || n == nil {
				return nil, fmt.Errorf("%s requires a number", kind)
			}
			if kind == "durationMs" {
				if *n < 0 {
					return nil, fmt.Errorf("negative duration")
				}
				node = token("duration", map[string]any{"value": *n, "unit": "ms"})
			} else if kind == "fontWeight" {
				if *n < 1 || *n > 1000 {
					return nil, fmt.Errorf("font weight outside [1,1000]")
				}
				node = token("fontWeight", *n)
			} else {
				if kind == "opacity" && (*n < 0 || *n > 1) {
					return nil, fmt.Errorf("opacity outside [0,1]")
				}
				node = token("number", *n)
			}
		case "fontNames":
			var fonts struct{ Values []string }
			if err = json.Unmarshal(body, &fonts); err != nil || len(fonts.Values) == 0 {
				return nil, fmt.Errorf("empty font family")
			}
			node = token("fontFamily", fonts.Values)
		case "axisValue":
			if err := validateNumbers(body, "value"); err != nil {
				return nil, err
			}
			var axis struct {
				Tag   string
				Value float64
			}
			if err = json.Unmarshal(body, &axis); err != nil || axis.Tag == "" {
				return nil, fmt.Errorf("invalid font axis")
			}
			node = token("number", axis.Value)
		case "cubicBezier":
			if err := validateNumbers(body, "x0", "y0", "x1", "y1"); err != nil {
				return nil, err
			}
			var c struct{ X0, Y0, X1, Y1 float64 }
			if err = json.Unmarshal(body, &c); err != nil {
				return nil, err
			}
			if c.X0 < 0 || c.X0 > 1 || c.X1 < 0 || c.X1 > 1 {
				return nil, fmt.Errorf("invalid cubic Bezier x coordinate")
			}
			node = token("cubicBezier", []float64{c.X0, c.Y0, c.X1, c.Y1})
		case "type":
			node, err = typography(body)
		case "shape":
			node, err = shape(body)
		default:
			return nil, fmt.Errorf("source type %q has no DTCG mapping", kind)
		}
		if err != nil {
			return nil, err
		}
		node["$extensions"] = map[string]any{Extension: map[string]any{"sourceType": kind, "sourceValue": json.RawMessage(raw)}}
	}
	return node, nil
}

func dimension(raw json.RawMessage) (Dimension, error) {
	var d struct {
		Value json.RawMessage
		Unit  string
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return Dimension{}, err
	}
	switch d.Unit {
	case "DIPS", "POINTS", "PIXELS":
	default:
		return Dimension{}, fmt.Errorf("unsupported dimension unit %q", d.Unit)
	}
	n := 0.0
	if len(d.Value) > 0 {
		var number *float64
		if json.Unmarshal(d.Value, &number) != nil || number == nil {
			return Dimension{}, fmt.Errorf("invalid dimension number")
		}
		n = *number
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return Dimension{}, fmt.Errorf("non-finite dimension")
	}
	return Dimension{Value: n, Unit: "px"}, nil
}

func typography(raw json.RawMessage) (Node, error) {
	var t struct {
		FontNames                          struct{ Values []string }
		FontSize, FontTracking, LineHeight json.RawMessage
		FontWeight                         *float64
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, err
	}
	size, err := dimension(t.FontSize)
	if err != nil {
		return nil, err
	}
	tracking, err := dimension(t.FontTracking)
	if err != nil {
		return nil, err
	}
	line, err := dimension(t.LineHeight)
	if err != nil {
		return nil, err
	}
	if size.Value <= 0 || line.Value < 0 || t.FontWeight == nil || *t.FontWeight < 1 || *t.FontWeight > 1000 || len(t.FontNames.Values) == 0 {
		return nil, fmt.Errorf("invalid typography")
	}
	return token("typography", map[string]any{"fontFamily": t.FontNames.Values, "fontSize": size, "fontWeight": *t.FontWeight, "letterSpacing": tracking, "lineHeight": line.Value / size.Value}), nil
}

func shape(raw json.RawMessage) (Node, error) {
	var s map[string]json.RawMessage
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	var family string
	if err := json.Unmarshal(s["family"], &family); err != nil {
		return nil, err
	}
	if family == "SHAPE_FAMILY_CIRCULAR" {
		return Node{}, nil
	} // Family is retained in the extension; DTCG dimensions have no percentage unit.
	if family != "SHAPE_FAMILY_ROUNDED_CORNERS" {
		return nil, fmt.Errorf("unsupported shape family %q", family)
	}
	node := Node{}
	for _, corner := range []string{"topLeft", "topRight", "bottomRight", "bottomLeft"} {
		raw, exists := s[corner]
		if !exists {
			raw = s["defaultSize"]
		}
		d, err := dimension(raw)
		if err != nil || d.Value < 0 {
			return nil, fmt.Errorf("invalid %s shape dimension", corner)
		}
		node[corner] = token("dimension", d)
	}
	return node, nil
}

// Protobuf omits zero-valued fields. Explicit null or a non-number is invalid.
func validateNumbers(raw json.RawMessage, fields ...string) error {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return err
	}
	for _, field := range fields {
		if body, exists := values[field]; exists {
			var number *float64
			if json.Unmarshal(body, &number) != nil || number == nil {
				return fmt.Errorf("invalid %s number", field)
			}
		}
	}
	return nil
}
