package web

import (
	"strings"
	"testing"
)

func TestRadioGroupUsesNativeSelection(t *testing.T) {
	c := componentCase{Family: "radio", State: "interactive", Selected: true}
	name := c.RadioInput(0)["name"]
	for index := range 3 {
		input := c.RadioInput(index)
		if input["name"] != name {
			t.Fatal("radio choices must share their native group")
		}
		_, checked := input["checked"]
		if checked != (index == 0) {
			t.Fatal("only the first option is initially selected")
		}
		control := c.RadioControl(index)
		for _, attr := range []string{"type", "name", "checked", "disabled"} {
			if _, ok := control[attr]; ok {
				t.Fatalf("native %s must belong to the input", attr)
			}
		}
	}
	if c.RadioControl(0)["data-component-case"] == c.RadioControl(1)["data-component-case"] {
		t.Fatal("radio specimens need unique identifiers")
	}
	c.State = "disabled"
	if c.RadioInput(0)["disabled"] != true {
		t.Fatal("disabled state must disable the native radio")
	}
}

func TestDownloadRadioSharesTheHostAnatomy(t *testing.T) {
	markup := downloadRadio(`"><script>`, true)
	if !strings.Contains(markup, `<span class="md-radio"><input type="radio"`) || !strings.Contains(markup, ` checked`) {
		t.Fatal("download radio host missing")
	}
	if strings.Contains(markup, `<script>`) {
		t.Fatal("radio values must be escaped")
	}
	if strings.Contains(downloadRadio("all", false), ` checked`) {
		t.Fatal("unchecked option must omit checked")
	}
}
