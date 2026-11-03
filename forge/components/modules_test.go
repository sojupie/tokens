package components

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestScriptResolutionPreservesOrderAndRejectsUnknownModules(t *testing.T) {
	got, err := ScriptsFor("common", "button-group", "button-group", "card")
	want := []string{"runtime/utils.js", "runtime/chips.js", "fields.js", "menus.js", "runtime/button-groups.js"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v; want %v", got, err, want)
	}
	if _, err := ScriptsFor("missing"); err == nil {
		t.Fatal("unknown module accepted")
	}
}

func TestManifestScriptsAreEmbedded(t *testing.T) {
	data, err := Files.ReadFile("modules.json")
	if err != nil {
		t.Fatal(err)
	}
	var modules map[string]struct{ Scripts []string }
	if err := json.Unmarshal(data, &modules); err != nil {
		t.Fatal(err)
	}
	for name := range modules {
		scripts, err := ScriptsFor(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, script := range scripts {
			if _, err := Files.ReadFile("static/" + script); err != nil {
				t.Errorf("%s: %s: %v", name, script, err)
			}
		}
	}
}
