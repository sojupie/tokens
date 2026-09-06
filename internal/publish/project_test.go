package publish

import (
	"encoding/json"
	"testing"

	"github.com/sojupie/m3-source/internal/tokensnapshot"
)

func TestProjectUsesComponentTokenSetOwnership(t *testing.T) {
	payloadID := "payload-component"
	setID := "designSystems/test/tokenSets/snackbar"
	tokenID := setID + "/tokens/width"
	input := snapshotInput{
		Manifest: tokensnapshot.Manifest{
			SnapshotID:   "snapshot-1",
			CreatedAt:    "2026-09-07T00:00:00Z",
			ReviewStatus: "provisional",
		},
		ManifestHash: "manifest-hash",
		PayloadDetails: tokensnapshot.SourcePayloadDetails{Records: []tokensnapshot.SourcePayloadDetail{{
			ID:                payloadID,
			PageURLs:          []string{"https://m3.material.io/components/snackbar/specs"},
			SourcePayloadURL:  "https://m3.material.io/_dsm/TOKEN_TABLE.json",
			SourcePayloadKind: "component",
			SourceArea:        "components",
			SourceSubject:     "snackbar",
			SourcePageType:    "specs",
		}}},
		Graph: tokensnapshot.Graph{Records: []tokensnapshot.ExtractedRecord{
			{
				ID:                "designSystems/test/components/snackbar",
				Kind:              "component",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Snackbar",
				Raw:               json.RawMessage(`{"tokenSets":["designSystems/test/tokenSets/snackbar"]}`),
			},
			{
				ID:                setID,
				Kind:              "tokenSet",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Snackbar",
			},
			{
				ID:                tokenID,
				Kind:              "token",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				TokenName:         "md.comp.snackbar.container.width",
				ValueType:         "DIMENSION",
				Raw:               json.RawMessage(`{"tokenName":"md.comp.snackbar.container.width"}`),
			},
			{
				ID:                tokenID,
				Kind:              "contextualReferenceTree",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				Raw:               json.RawMessage(`{"contextualReferenceTree":[{"referenceTree":{"tokenName":"md.comp.snackbar.container.width"},"resolvedValue":{"dimension":{"value":560,"unit":"DIPS"}}}]}`),
			},
		}},
	}

	publication, err := project(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(publication.Subjects) != 1 {
		t.Fatalf("subjects = %d, want 1", len(publication.Subjects))
	}
	subject := publication.Subjects[0]
	if subject.Key != (SubjectKey{Kind: SubjectComponent, Slug: "snackbar"}) {
		t.Fatalf("subject key = %#v", subject.Key)
	}
	if len(subject.Tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(subject.Tokens))
	}
	if len(subject.Tokens[0].Values) != 1 {
		t.Fatalf("token values = %d, want 1", len(subject.Tokens[0].Values))
	}
	if got, want := string(subject.Tokens[0].Values[0].Value), `{"dimension":{"value":560,"unit":"DIPS"}}`; got != want {
		t.Fatalf("resolved value = %s, want %s", got, want)
	}
}
