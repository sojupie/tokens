package publish

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sojupie/tokens/source/tokensnapshot"
)

func TestSchemaThreePublicationRemainsReadable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.sqlite")
	publication := Publication{SchemaVersion: 3, PublicationID: "legacy", CurrentSnapshot: SnapshotView{ID: "snapshot"}, Snapshots: []SnapshotView{{ID: "snapshot", CreatedAt: "2026-10-11T00:00:00Z"}}, Subjects: []SubjectView{{Key: SubjectKey{Kind: SubjectStyle, Slug: "shape"}, Tokens: []TokenView{{SourceName: "md.sys.shape.corner.large", Lifecycle: "active"}}}}}
	if err := writeDatabase(path, publication); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("ALTER TABLE tokens DROP COLUMN presentation_json"); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Subjects) != 1 {
		t.Fatal("legacy publication lost its subject")
	}
}

func TestSourcePresentationSurvivesProjectionStorageAndPublicJSON(t *testing.T) {
	set := "designSystems/test/tokenSets/elevated"
	record := func(kind, id, raw string) tokensnapshot.ExtractedRecord {
		return tokensnapshot.ExtractedRecord{Kind: kind, ID: id, SourcePayloadID: "payload", SourcePayloadKind: "component", Raw: json.RawMessage(raw)}
	}
	component := record("component", "component", `{"tokenSets":["`+set+`"]}`)
	component.DisplayName = "Buttons"
	token := record("token", set+"/tokens/elevation", `{"displayName":"Button elevation","displayGroup":"`+set+`/displayGroups/nested","orderInDisplayGroup":4}`)
	token.TokenName = "md.comp.button.elevated.container.elevation"
	input := snapshotInput{
		PayloadDetails: tokensnapshot.SourcePayloadDetails{Records: []tokensnapshot.SourcePayloadDetail{{ID: "payload", SourcePayloadKind: "component", SourceSubject: "buttons", PageURLs: []string{"https://m3.material.io/components/buttons/specs"}}}},
		Graph: tokensnapshot.Graph{Records: []tokensnapshot.ExtractedRecord{
			component,
			record("tokenSet", set, `{"tokenSetName":"md.comp.button.elevated","displayName":"Button - Color - Elevated","order":2}`),
			record("displayGroup", set+"/displayGroups/enabled", `{"displayName":"Enabled","orderInParentTokenSet":1}`),
			record("displayGroup", set+"/displayGroups/nested", `{"displayName":"Container","parentGroup":"`+set+`/displayGroups/enabled","orderInParentDisplayGroup":3}`),
			token,
		}},
	}
	publication, err := project(input)
	if err != nil {
		t.Fatal(err)
	}
	got := publication.Subjects[0].Tokens[0]
	want := &TokenPresentation{SetName: "md.comp.button.elevated", SetLabel: "Button - Color - Elevated", SetOrder: 2, Groups: []DisplayGroup{{Name: "Enabled", Order: 1}, {Name: "Container", Order: 3}}, Order: 4}
	if !reflect.DeepEqual(got.Presentation, want) {
		t.Fatalf("presentation=%+v, want %+v", got.Presentation, want)
	}
	publication.PublicationID = "test"
	publication.CurrentSnapshot = SnapshotView{ID: "snapshot"}
	publication.Snapshots = []SnapshotView{{ID: "snapshot", CreatedAt: "2026-10-15T00:00:00Z"}}
	path := filepath.Join(t.TempDir(), "tokens.sqlite")
	if err := writeDatabase(path, publication); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	stored := loaded.Subjects[publication.Subjects[0].Key].Public().Tokens[0]
	if !reflect.DeepEqual(stored.Presentation, want) {
		t.Fatalf("stored/public presentation=%+v", stored.Presentation)
	}
	if string(got.Raw) != string(token.Raw) {
		t.Fatal("presentation changed source evidence")
	}
}

func TestComponentRoutingUsesSpecificMatchRegardlessOfPageOrder(t *testing.T) {
	pages := []string{"https://m3.material.io/components/buttons/specs", "https://m3.material.io/components/split-button/specs"}
	for i := 0; i < 2; i++ {
		got := componentSubjectKey(tokensnapshot.ExtractedRecord{DisplayName: "Split buttons"}, SourcePayloadView{PageURLs: pages}, SubjectKey{Kind: SubjectComponent, Slug: "buttons"}, []string{"buttons", "split-button"})
		if got.Slug != "split-button" {
			t.Fatalf("page order %v: assigned %s", pages, got.Slug)
		}
		pages[0], pages[1] = pages[1], pages[0]
	}
	if got := bestComponentSlug([]string{"navigation"}, []string{"navigation-bar", "navigation-rail"}); got != "" {
		t.Fatalf("ambiguous match assigned %s", got)
	}
	if _, ok := styleTokenSetSubject("md.comp.unknown.shape", "Unknown - Shape"); ok {
		t.Fatal("component set classified as a system style")
	}
}
