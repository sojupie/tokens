package web

import "testing"

func TestTimePickerProfiles(t *testing.T) {
	c := componentCase{Family: "time", Variant: "clock", Width: "24h", Shape: "hour", Control: "value", Selection: "unselected", State: "pressed"}
	if len(c.TimeTicks()) != 24 || c.TimeTicks()[0].Label != "00" || c.TimeTicks()[12].Label != "12" {
		t.Fatal("24-hour dial must provide outer and inner rings")
	}
	if c.TimeValue("minute")["data-component-state"] != "pressed" {
		t.Fatal("unselected preview must target the inactive value")
	}
	if _, ok := c.TimeValue("hour")["data-component-state"]; ok {
		t.Fatal("selected value must retain its native state")
	}
	c.Shape = "minute"
	if len(c.TimeTicks()) != 12 || c.TimeTicks()[2].Label != "10" {
		t.Fatal("minute dial must label five-minute intervals")
	}
	c.State = "interactive"
	if _, ok := c.TimePanel()["open"]; ok {
		t.Fatal("interactive picker must use showModal")
	}
	for _, name := range []string{"keyboard", "schedule"} {
		if materialSymbolSVG(name) == "" {
			t.Fatalf("missing mode icon %s", name)
		}
	}
}
