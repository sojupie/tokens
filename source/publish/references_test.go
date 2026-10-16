package publish

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestContextualReferencesPreserveBranches(t *testing.T) {
	raw := json.RawMessage(`{"contextualReferenceTree":[{"contextTags":["light"],"referenceTree":{"tokenName":"md.comp.button.shadow","childNodes":[{"tokenName":"md.sys.color.shadow","childNodes":[{"tokenName":"md.ref.palette.neutral0"}]},{"tokenName":"md.sys.elevation.level1"}]},"resolvedValue":{"undefined":true}}]}`)
	entries, err := contextualEntries(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []TokenReference{{Name: "md.sys.color.shadow", References: []TokenReference{{Name: "md.ref.palette.neutral0"}}}, {Name: "md.sys.elevation.level1"}}
	if len(entries) != 1 || !reflect.DeepEqual(entries[0].References, want) {
		t.Fatalf("lost references: %#v", entries)
	}
	if string(entries[0].ResolvedValue) != `{"undefined":true}` {
		t.Fatal("resolved value was modified")
	}
}

func TestReferencesAndNamesSurvivePublication(t *testing.T) {
	references := []TokenReference{{Name: "md.sys.color.shadow", References: []TokenReference{{Name: "md.ref.palette.neutral0"}}}}
	token := TokenView{SourceName: "md.comp.button.shadow", DisplayName: "Button elevated shadow color", Values: []TokenValueView{{Value: json.RawMessage(`{"color":{"alpha":1}}`), References: references}}}
	publication := Publication{SchemaVersion: SchemaVersion, PublicationID: "test", CurrentSnapshot: SnapshotView{ID: "snapshot"}, Snapshots: []SnapshotView{{ID: "snapshot", CreatedAt: "2026-09-05T18:16:00Z"}}, Subjects: []SubjectView{{Key: SubjectKey{Kind: SubjectComponent, Slug: "buttons"}, Tokens: []TokenView{token}}}}
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "tokens.sqlite")
	if err := writeDatabase(databasePath, publication); err != nil {
		t.Fatal(err)
	}
	state, err := Load(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	loaded := state.Subjects[SubjectKey{Kind: SubjectComponent, Slug: "buttons"}].Tokens[0]
	if loaded.DisplayName != token.DisplayName || !reflect.DeepEqual(loaded.Values[0].References, references) {
		t.Fatalf("lost stored names/references: %#v", loaded)
	}
	public := state.Subjects[SubjectKey{Kind: SubjectComponent, Slug: "buttons"}].Public()
	if !reflect.DeepEqual(public.Tokens[0].Values[0].References, references) {
		t.Fatal("public download lost references")
	}
}
