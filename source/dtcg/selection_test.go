package dtcg

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSelectionKeepsAliasDependenciesAndRootTokens(t *testing.T) {
	value := func(v string) Node {
		return Node{"$type": "number", "$value": v, "$extensions": map[string]any{Extension: map[string]any{"sourceType": "numeric"}}}
	}
	root := Node{}
	for name, node := range map[string]Node{"md.comp.button.opacity": value("{md.sys.opacity}"), "md.sys.opacity": value("{md.ref.opacity}"), "md.ref.opacity": value("1"), "md.ref.opacity.extra": value("2"), "md.unrelated": value("3")} {
		if err := insert(root, strings.Split(name, "."), node); err != nil {
			t.Fatal(err)
		}
	}
	selected, roots, deps := SelectNames(root, map[string]bool{"md.comp.button.opacity": true})
	names := NamedNodes(selected)
	if roots != 1 || deps != 2 || len(names) != 3 || names["md.ref.opacity"] == nil {
		t.Fatalf("lost dependencies: %d %d %#v", roots, deps, names)
	}
	if len(NamedNodes(root)) != 5 {
		t.Fatal("selection changed input document")
	}
}

func TestValueKeyIgnoresJSONWhitespace(t *testing.T) {
	if ValueKey(json.RawMessage(`{"numeric": 1}`)) != ValueKey(json.RawMessage(`{ "numeric":1 }`)) {
		t.Fatal("unstable value key")
	}
}

func TestValueKeyCanonicalizesInferredReferenceMetadata(t *testing.T) {
	original := json.RawMessage(`{"color":{"red":1,"green":0,"blue":0,"alpha":1}}`)
	inferred := json.RawMessage(`{"color":{"alpha":1,"blue":0,"green":0,"red":1}}`)
	if ValueKey(original) != ValueKey(inferred) {
		t.Fatal("inferred references cannot find their compiled CSS value")
	}
	if ValueKey(json.RawMessage(`{"numeric":1}`)) == ValueKey(json.RawMessage(`{"numeric":2}`)) {
		t.Fatal("distinct context values share a key")
	}
}
