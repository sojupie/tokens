package web

import "testing"

func TestFieldControlSemantics(t *testing.T) {
	for _, family := range []string{"field", "select", "autocomplete"} {
		c := componentCase{Family: family, Variant: "filled", State: "disabled", Width: "populated", Invalid: true}
		a := c.FieldInput()
		if a["disabled"] != true || a["aria-invalid"] != "true" {
			t.Fatalf("missing input state: %v", a)
		}
		if family != "field" {
			if a["role"] != "combobox" || a["aria-controls"] != c.ID()+"-options" {
				t.Fatalf("missing popup semantics: %v", a)
			}
		}
		if family == "select" && a["readonly"] != true {
			t.Fatal("select value must be restricted to its options")
		}
		if family == "autocomplete" {
			if _, ok := a["readonly"]; ok {
				t.Fatal("autocomplete must remain editable")
			}
		}
	}
}
