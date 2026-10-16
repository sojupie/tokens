package publish

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestWriteDatabaseStoresPublicationAndToken(t *testing.T) {
	databasePath := t.TempDir() + "\\tokens.sqlite"
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	_ = database.Close()

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
					Value: []byte(`{"dimension":{"value":560,"unit":"DIPS"}}`),
				}},
				Sources: []TokenSourceView{{PayloadID: "payload-1", SourceRecordID: "record-1"}},
			}},
		}},
	}

	if err := writeDatabase(databasePath, publication); err != nil {
		t.Fatal(err)
	}
	database, err = sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM tokens").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("token count = %d, want 1", count)
	}
	var value string
	if err := database.QueryRow("SELECT value_json FROM token_values").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != `{"dimension":{"value":560,"unit":"DIPS"}}` {
		t.Fatalf("stored value = %s", value)
	}
}
