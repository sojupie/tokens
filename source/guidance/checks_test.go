package guidance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func curationFixture(t *testing.T) ([]byte, map[string]any) {
	t.Helper()
	data, err := os.ReadFile("testdata/curation-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return data, doc
}
func firstCheck(doc map[string]any) map[string]any { return doc["checks"].([]any)[0].(map[string]any) }
func addSourceEvent(check map[string]any, change string) {
	old := check["sourceHistory"].([]any)
	last := old[len(old)-1].(map[string]any)
	next := map[string]any{}
	for key, value := range last {
		next[key] = value
	}
	next["change"] = change
	next["summary"] = "Record " + change
	check["sourceHistory"] = append(old, next)
}

func TestCurationContract(t *testing.T) {
	cases := []struct {
		name, want string
		change     func(map[string]any)
	}{
		{"valid", "", func(map[string]any) {}},
		{"duplicate check", "duplicate ID", func(d map[string]any) { d["checks"] = append(d["checks"].([]any), firstCheck(d)) }},
		{"duplicate source", "duplicate source", func(d map[string]any) { d["sources"] = append(d["sources"].([]any), d["sources"].([]any)[0]) }},
		{"missing section", "absent from sources", func(d map[string]any) {
			firstCheck(d)["sourceHistory"].([]any)[0].(map[string]any)["section"] = "missing"
		}},
		{"mapped without check", "no active checks", func(d map[string]any) { d["checks"] = []any{} }},
		{"incorrect section exclusion", "no-requirements", func(d map[string]any) {
			s := d["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)
			s["status"] = "no-requirements"
			s["notes"] = "Descriptive content."
		}},
		{"silent statement edit", "requires a sourceHistory event", func(d map[string]any) { firstCheck(d)["statement"] = "A different instruction." }},
		{"stale assessment", "old or nonexistent", func(d map[string]any) { c := firstCheck(d); addSourceEvent(c, "changed") }},
		{"reassessed source change", "", func(d map[string]any) {
			c := firstCheck(d)
			addSourceEvent(c, "changed")
			c["statement"] = "A changed instruction."
			c["implementation"].(map[string]any)["sourceRevision"] = float64(2)
		}},
		{"curation correction", "", func(d map[string]any) {
			c := firstCheck(d)
			addSourceEvent(c, "corrected")
			c["condition"] = "A corrected condition."
			c["implementation"].(map[string]any)["sourceRevision"] = float64(2)
		}},
		{"deleted check", "was deleted", func(d map[string]any) {
			d["checks"] = []any{}
			d["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)["status"] = "pending"
		}},
		{"rewritten source history", "append-only", func(d map[string]any) {
			firstCheck(d)["sourceHistory"].([]any)[0].(map[string]any)["summary"] = "Rewrite old history."
		}},
		{"invalid restoration", "invalid sourceHistory lifecycle", func(d map[string]any) { addSourceEvent(firstCheck(d), "restored") }},
		{"withdrawn retained", "", func(d map[string]any) {
			addSourceEvent(firstCheck(d), "withdrawn")
			d["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)["status"] = "pending"
		}},
		{"verified with only code", "", func(d map[string]any) {
			i := firstCheck(d)["implementation"].(map[string]any)
			i["status"] = "verified"
			i["assessedCommit"] = strings.Repeat("b", 40)
			i["evidence"] = []any{map[string]any{"kind": "code", "location": "file.go:10", "result": "Code exists."}}
		}},
		{"unknown fields", "", func(d map[string]any) { firstCheck(d)["vibe"] = "looks right" }},
		{"interpretation needs reason", "", func(d map[string]any) { firstCheck(d)["basis"] = "interpretation" }},
		{"fixed but unverified", "", func(d map[string]any) {
			firstCheck(d)["fixes"] = []any{map[string]any{"commit": strings.Repeat("b", 40), "sourceRevision": float64(1), "summary": "Fix focus traversal."}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before, doc := curationFixture(t)
			tc.change(doc)
			after, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ValidateChecks(after, before)
			schemaFailure := tc.name == "verified with only code" || tc.name == "unknown fields" || tc.name == "interpretation needs reason"
			if tc.want == "" && !schemaFailure {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected validation failure")
			}
			if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}

func TestRemovalRestorationAndFixHistory(t *testing.T) {
	before, doc := curationFixture(t)
	check := firstCheck(doc)
	addSourceEvent(check, "removed")
	doc["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)["status"] = "pending"
	data, _ := json.Marshal(doc)
	summary, err := ValidateChecks(data, before)
	if err != nil || summary.Removed != 1 {
		t.Fatalf("removal: %+v %v", summary, err)
	}
	addSourceEvent(check, "restored")
	check["implementation"].(map[string]any)["sourceRevision"] = float64(3)
	check["fixes"] = []any{map[string]any{"commit": strings.Repeat("b", 40), "sourceRevision": float64(1), "summary": "First fix."}}
	restored, _ := json.Marshal(doc)
	if _, err = ValidateChecks(restored, data); err != nil {
		t.Fatal(err)
	}
	check["fixes"] = []any{}
	erased, _ := json.Marshal(doc)
	if _, err = ValidateChecks(erased, restored); err == nil || !strings.Contains(err.Error(), "fixes is append-only") {
		t.Fatalf("fix erased: %v", err)
	}
}

// Keep every committed component file covered by the normal offline Go suite.
func TestComponentGuidanceFiles(t *testing.T) {
	files, err := filepath.Glob("../../forge/*/guidance/*/checks.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no canonical component guidance files")
	}
	for _, path := range files {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			retired, readErr := os.ReadFile(filepath.Join(filepath.Dir(path), "retired-checks.json"))
			if readErr != nil && !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
			summary, err := ValidateCheckFiles(data, retired, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			expected := filepath.Join("..", "..", "forge", summary.DesignSystem, "guidance", summary.Component, "checks.json")
			if path != expected {
				t.Fatalf("canonical path: got %s, want %s", path, expected)
			}
		})
	}
}

func TestRetirementArchive(t *testing.T) {
	before, doc := curationFixture(t)
	check := firstCheck(doc)
	addSourceEvent(check, "removed")
	archive := map[string]any{"schemaVersion": 1, "designSystem": doc["designSystem"], "component": doc["component"], "checks": []any{check}}
	doc["checks"] = []any{}
	doc["sources"].([]any)[0].(map[string]any)["sections"].([]any)[0].(map[string]any)["status"] = "pending"
	marshal := func(v any) []byte {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	active, retired := marshal(doc), marshal(archive)
	if summary, err := ValidateCheckFiles(active, retired, before, nil); err != nil || summary.Checks != 0 || summary.Removed != 1 {
		t.Fatalf("retirement: %+v %v", summary, err)
	}
	for _, tc := range []struct {
		name   string
		change func()
		want   string
	}{
		{"lost archive", func() { retired = nil }, "was deleted"},
		{"duplicate ID", func() { doc["checks"] = []any{check}; active = marshal(doc) }, "retired checks belong"},
		{"wrong identity", func() { archive["component"] = "other"; retired = marshal(archive) }, "identity"},
		{"modified assessment", func() { check["implementation"].(map[string]any)["notes"] = "Rewritten"; retired = marshal(archive) }, "preserve implementation"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.change()
			if _, err := ValidateCheckFiles(active, retired, before, nil); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
			archive["component"] = doc["component"]
			check["implementation"].(map[string]any)["notes"] = ""
			doc["checks"] = []any{}
			active = marshal(doc)
			retired = marshal(archive)
		})
	}
	// An active and a retired record must not share an ID, even with distinct history.
	if _, err := ValidateCheckFiles(before, retired, nil, nil); err == nil || !strings.Contains(err.Error(), "duplicate ID") {
		t.Fatalf("cross-file duplicate: %v", err)
	}
	// Restore the same ID, keeping every earlier source event and fix.
	addSourceEvent(check, "restored")
	check["implementation"].(map[string]any)["sourceRevision"] = 3
	doc["checks"] = []any{check}
	restored := marshal(doc)
	if _, err := ValidateCheckFiles(restored, nil, active, retired); err != nil {
		t.Fatalf("restoration: %v", err)
	}
	if _, err := ValidateCheckFiles(active, nil, active, retired); err == nil {
		t.Fatal("accepted deleted archived check")
	}
	// Active records cannot live in the archive.
	archive["checks"] = []any{check}
	if _, err := ValidateCheckFiles(active, marshal(archive), nil, nil); err == nil {
		t.Fatal("accepted active archived check")
	}
}

func TestAssessmentBaselineReference(t *testing.T) {
	before, doc := curationFixture(t)
	impl := firstCheck(doc)["implementation"].(map[string]any)
	impl["status"] = "implemented"
	impl["assessedBaseline"] = strings.Repeat("a", 40)
	impl["evidence"] = []any{map[string]any{"kind": "code", "location": "forge/components/static/buttons.js:1", "result": "Concrete behavior inspected."}}
	data, _ := json.Marshal(doc)
	summary, err := ValidateChecks(data, before)
	if err != nil || len(summary.Baselines) != 1 || len(summary.Commits) != 0 {
		t.Fatalf("baseline assessment: %+v %v", summary, err)
	}
	impl["assessedCommit"] = strings.Repeat("b", 40)
	data, _ = json.Marshal(doc)
	if _, err = ValidateChecks(data, before); err == nil {
		t.Fatal("accepted ambiguous commit and baseline references")
	}
	delete(impl, "assessedCommit")
	impl["status"] = "unassessed"
	impl["evidence"] = []any{}
	data, _ = json.Marshal(doc)
	if _, err = ValidateChecks(data, before); err == nil {
		t.Fatal("accepted unassessed with baseline reference")
	}
}
