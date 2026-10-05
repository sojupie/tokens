package tokensnapshot

import (
	"encoding/json"
	"fmt"
	"testing"
)

func contextDiffFixture(prefix string, dark int, reversed bool, reference string) []ExtractedRecord {
	group := prefix + "/contextTagGroups/theme"
	entry := func(tag string, value int) string {
		return fmt.Sprintf(`{"contextTags":[%q],"referenceTree":{"tokenName":"md.comp.button.color","childNodes":[{"tokenName":%q}]},"resolvedValue":{"number":%d}}`, group+"/tags/"+tag, reference, value)
	}
	light, darkEntry := entry("light", 1), entry("dark", dark)
	if reversed {
		light, darkEntry = darkEntry, light
	}
	return []ExtractedRecord{
		{Kind: "contextTagGroup", ID: group, Raw: json.RawMessage(fmt.Sprintf(`{"displayName":"Theme","defaultTag":%q}`, group+"/tags/light"))},
		{Kind: "tag", ID: group + "/tags/light", Raw: json.RawMessage(`{"displayName":"Light"}`)},
		{Kind: "tag", ID: group + "/tags/dark", Raw: json.RawMessage(`{"displayName":"Dark"}`)},
		{Kind: "contextualReferenceTree", SourcePayloadKind: "component", Raw: json.RawMessage(`{"contextualReferenceTree":[` + light + `,` + darkEntry + `]}`)},
	}
}

func TestContextDiffIncludesNonFirstValuesAndStableLabels(t *testing.T) {
	before := contextDiffFixture("old", 2, false, "md.sys.color.primary")
	after := contextDiffFixture("new", 3, true, "md.sys.color.primary")
	diff := diffSemanticRecords(before, after)
	if len(diff.Changed) != 1 || len(diff.Added) != 0 || len(diff.Removed) != 0 {
		t.Fatalf("diff = %#v", diff)
	}
	var value ContextualValue
	if err := json.Unmarshal([]byte(diff.Changed[0].After), &value); err != nil {
		t.Fatal(err)
	}
	if len(value.Context) != 1 || value.Context[0] != "Theme: Dark" || string(value.Value) != `{"number":3}` {
		t.Fatalf("context lost: %#v", value)
	}
	unchanged := diffSemanticRecords(before, contextDiffFixture("new", 2, true, "md.sys.color.primary"))
	if len(unchanged.Changed)+len(unchanged.Added)+len(unchanged.Removed) != 0 {
		t.Fatalf("reorder or ID churn: %#v", unchanged)
	}
}

func TestContextDiffTracksReferenceAndContextRemoval(t *testing.T) {
	before := contextDiffFixture("same", 2, false, "md.sys.color.primary")
	after := contextDiffFixture("same", 2, false, "md.sys.color.secondary")
	diff := diffSemanticRecords(before, after)
	if len(diff.Changed) != 2 {
		t.Fatalf("reference-only changes lost: %#v", diff)
	}
	after = before[:3]
	diff = diffSemanticRecords(before, after)
	if len(diff.Removed) != 2 {
		t.Fatalf("removed contexts lost: %#v", diff)
	}
	diff = diffSemanticRecords(after, before)
	if len(diff.Added) != 2 || diff.Added[0].Key == diff.Added[1].Key {
		t.Fatalf("context identity collapsed: %#v", diff)
	}
}
