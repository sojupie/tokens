package publish

import (
	"encoding/json"
	"testing"
)

func TestLoadReconstructsPublishedState(t *testing.T) {
	databasePath := t.TempDir() + "\\tokens.sqlite"
	publication := Publication{
		SchemaVersion: SchemaVersion,
		PublicationID: "publication-1",
		GeneratedAt:   "2026-09-07T00:00:00Z",
		CurrentSnapshot: SnapshotView{
			ID:            "snapshot-1",
			CreatedAt:     "2026-09-07T00:00:00Z",
			ManifestHash:  "manifest-hash",
			CaptureStatus: "complete",
			ReviewStatus:  "provisional",
		},
		Snapshots: []SnapshotView{{
			ID:            "snapshot-1",
			CreatedAt:     "2026-09-07T00:00:00Z",
			ManifestHash:  "manifest-hash",
			CaptureStatus: "complete",
			ReviewStatus:  "provisional",
		}},
		SourcePayloads: []SourcePayloadView{{
			ID:       "payload-1",
			URL:      "https://example.test/TOKEN_TABLE.json",
			Hash:     "payload-hash",
			PageURLs: []string{"https://example.test/components/snackbar/specs"},
		}},
		Subjects: []SubjectView{{
			Key:            SubjectKey{Kind: SubjectComponent, Slug: "snackbar"},
			Name:           "Snackbar",
			SourcePageURLs: []string{"https://example.test/components/snackbar/specs"},
			Tokens: []TokenView{{
				SourceName: "md.comp.snackbar.container.width",
				Category:   "DIMENSION",
				Values: []TokenValueView{{
					ContextTags: []string{"light"},
					Value:       json.RawMessage(`{"dimension":{"value":560,"unit":"DIPS"}}`),
					Sources:     []TokenSourceView{{PayloadID: "payload-1", SourceRecordID: "record-1"}},
				}},
				Sources:   []TokenSourceView{{PayloadID: "payload-1", SourceRecordID: "record-1"}},
				Lifecycle: "active",
			}},
			LifecycleNotices: []LifecycleNoticeView{{
				ID:              "notice-1",
				State:           "deprecated",
				Subject:         "Snackbar width",
				Message:         "Use the updated width.",
				SourcePayloadID: "payload-1",
				SourceRecordID:  "record-1",
			}},
		}},
	}

	if err := writeDatabase(databasePath, publication); err != nil {
		t.Fatal(err)
	}

	state, err := Load(databasePath)
	if err != nil {
		t.Fatal(err)
	}

	if state.PublicationID != publication.PublicationID {
		t.Fatalf("publication ID = %q, want %q", state.PublicationID, publication.PublicationID)
	}
	if state.CurrentSnapshot.ID != "snapshot-1" || len(state.Snapshots) != 1 {
		t.Fatalf("snapshots = %#v", state.Snapshots)
	}
	if len(state.SourcePayloads) != 1 {
		t.Fatalf("source payload count = %d, want 1", len(state.SourcePayloads))
	}

	subject, ok := state.Subjects[SubjectKey{Kind: SubjectComponent, Slug: "snackbar"}]
	if !ok {
		t.Fatal("snackbar subject is missing")
	}
	if len(subject.Tokens) != 1 || len(subject.LifecycleNotices) != 1 {
		t.Fatalf("subject = %#v", subject)
	}
	token := subject.Tokens[0]
	if len(token.Values) != 1 || len(token.Sources) != 1 {
		t.Fatalf("token = %#v", token)
	}
	if string(token.Values[0].Value) != `{"dimension":{"value":560,"unit":"DIPS"}}` {
		t.Fatalf("token value = %s", token.Values[0].Value)
	}
	if len(token.Values[0].Sources) != 1 || token.Values[0].ContextTags[0] != "light" {
		t.Fatalf("token value metadata = %#v", token.Values[0])
	}
}
