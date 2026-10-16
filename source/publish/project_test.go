package publish

import (
	"encoding/json"
	"testing"

	"github.com/sojupie/tokens/source/tokensnapshot"
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

func TestProjectPreservesComponentDisplayNameOverTokenSetName(t *testing.T) {
	buttonsPayloadID := "payload-buttons"
	splitPayloadID := "payload-split"
	sharedSetID := "designSystems/test/tokenSets/btn-color-elevated"
	input := snapshotInput{
		Manifest: tokensnapshot.Manifest{
			SnapshotID:   "snapshot-1",
			CreatedAt:    "2026-09-07T00:00:00Z",
			ReviewStatus: "provisional",
		},
		ManifestHash: "manifest-hash",
		PayloadDetails: tokensnapshot.SourcePayloadDetails{Records: []tokensnapshot.SourcePayloadDetail{
			{
				ID:                splitPayloadID,
				PageURLs:          []string{"https://m3.material.io/components/split-button/specs"},
				SourcePayloadURL:  "https://m3.material.io/_dsm/SPLIT.json",
				SourcePayloadKind: "component",
				SourceArea:        "components",
				SourceSubject:     "split-button",
				SourcePageType:    "specs",
			},
			{
				ID:                buttonsPayloadID,
				PageURLs:          []string{"https://m3.material.io/components/buttons/specs"},
				SourcePayloadURL:  "https://m3.material.io/_dsm/BUTTONS.json",
				SourcePayloadKind: "component",
				SourceArea:        "components",
				SourceSubject:     "buttons",
				SourcePageType:    "specs",
			},
		}},
		Graph: tokensnapshot.Graph{Records: []tokensnapshot.ExtractedRecord{
			// split-button payload has the tokenSet sorted early
			{
				ID:                sharedSetID,
				Kind:              "tokenSet",
				SourcePayloadID:   splitPayloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Button - Color - Elevated",
			},
			{
				ID:                "designSystems/test/components/split-button",
				Kind:              "component",
				SourcePayloadID:   splitPayloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Split buttons",
				Raw:               json.RawMessage(`{"tokenSets":["` + sharedSetID + `"]}`),
			},
			// buttons component arrives later
			{
				ID:                "designSystems/test/components/buttons",
				Kind:              "component",
				SourcePayloadID:   buttonsPayloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Buttons",
				Raw:               json.RawMessage(`{"tokenSets":["` + sharedSetID + `"]}`),
			},
			{
				ID:                sharedSetID + "/tokens/color",
				Kind:              "token",
				SourcePayloadID:   buttonsPayloadID,
				SourcePayloadKind: "component",
				TokenName:         "md.comp.button.elevated.container.color",
				ValueType:         "COLOR",
				Raw:               json.RawMessage(`{"tokenName":"md.comp.button.elevated.container.color"}`),
			},
		}},
	}

	publication, err := project(input)
	if err != nil {
		t.Fatal(err)
	}

	var buttonsSubject *SubjectView
	for i := range publication.Subjects {
		if publication.Subjects[i].Key.Slug == "buttons" {
			buttonsSubject = &publication.Subjects[i]
			break
		}
	}
	if buttonsSubject == nil {
		t.Fatal("buttons subject not found")
	}
	if buttonsSubject.Name != "Buttons" {
		t.Fatalf("buttons subject name = %q, want %q", buttonsSubject.Name, "Buttons")
	}
}

func TestProjectDisambiguatesSearchComponentFromAppBarsPayload(t *testing.T) {
	payloadID := "payload-app-bars-and-search"
	searchSetID := "designSystems/test/tokenSets/search-view"
	input := snapshotInput{
		Manifest: tokensnapshot.Manifest{
			SnapshotID:   "snapshot-1",
			CreatedAt:    "2026-09-07T00:00:00Z",
			ReviewStatus: "provisional",
		},
		ManifestHash: "manifest-hash",
		PayloadDetails: tokensnapshot.SourcePayloadDetails{Records: []tokensnapshot.SourcePayloadDetail{
			{
				ID: payloadID,
				PageURLs: []string{
					"https://m3.material.io/components/app-bars/specs",
					"https://m3.material.io/components/search/specs",
				},
				SourcePayloadURL:  "https://m3.material.io/_dsm/TOKEN_TABLE.search.json",
				SourcePayloadKind: "component",
				SourceArea:        "components",
				SourceSubject:     "app-bars", // crawled first under app-bars
				SourcePageType:    "specs",
			},
		}},
		Graph: tokensnapshot.Graph{Records: []tokensnapshot.ExtractedRecord{
			{
				ID:                "designSystems/test/components/search",
				Kind:              "component",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Search",
				Raw:               json.RawMessage(`{"tokenSets":["` + searchSetID + `"]}`),
			},
			{
				ID:                searchSetID,
				Kind:              "tokenSet",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				DisplayName:       "Search - View",
			},
			{
				ID:                searchSetID + "/tokens/height",
				Kind:              "token",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "component",
				TokenName:         "md.comp.search-view.container.height",
				ValueType:         "DIMENSION",
				Raw:               json.RawMessage(`{"tokenName":"md.comp.search-view.container.height"}`),
			},
		}},
	}

	publication, err := project(input)
	if err != nil {
		t.Fatal(err)
	}

	var searchSubject *SubjectView
	for i := range publication.Subjects {
		if publication.Subjects[i].Key == (SubjectKey{Kind: SubjectComponent, Slug: "search"}) {
			searchSubject = &publication.Subjects[i]
			break
		}
	}
	if searchSubject == nil {
		t.Fatal("search component subject not found")
	}
	if searchSubject.Name != "Search" {
		t.Fatalf("search subject name = %q, want %q", searchSubject.Name, "Search")
	}
	if len(searchSubject.SourcePageURLs) != 1 || searchSubject.SourcePageURLs[0] != "https://m3.material.io/components/search/specs" {
		t.Fatalf("search source page urls = %#v, want only search/specs", searchSubject.SourcePageURLs)
	}
}

func TestProjectMapsStyleTokensToCanonicalSubjects(t *testing.T) {
	payloadID := "payload-styles-motion-and-shape"
	shapeSetID := "designSystems/test/tokenSets/shape"
	input := snapshotInput{
		Manifest: tokensnapshot.Manifest{
			SnapshotID:   "snapshot-1",
			CreatedAt:    "2026-09-07T00:00:00Z",
			ReviewStatus: "provisional",
		},
		ManifestHash: "manifest-hash",
		PayloadDetails: tokensnapshot.SourcePayloadDetails{Records: []tokensnapshot.SourcePayloadDetail{
			{
				ID: payloadID,
				PageURLs: []string{
					"https://m3.material.io/styles/motion/overview/specs",
					"https://m3.material.io/styles/shape/corner-radius-scale",
				},
				SourcePayloadURL:  "https://m3.material.io/_dsm/TOKEN_TYPE_UNSPECIFIED.json",
				SourcePayloadKind: "motion",
				SourceArea:        "styles",
				SourceSubject:     "motion/overview",
				SourcePageType:    "specs",
			},
		}},
		Graph: tokensnapshot.Graph{Records: []tokensnapshot.ExtractedRecord{
			{
				ID:                shapeSetID,
				Kind:              "tokenSet",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "motion",
				TokenSetName:      "md.sys.shape",
				DisplayName:       "Shape",
			},
			{
				ID:                shapeSetID + "/tokens/corner-large",
				Kind:              "token",
				SourcePayloadID:   payloadID,
				SourcePayloadKind: "motion",
				TokenName:         "md.sys.shape.corner.large",
				ValueType:         "SHAPE",
				Raw:               json.RawMessage(`{"tokenName":"md.sys.shape.corner.large"}`),
			},
		}},
	}

	publication, err := project(input)
	if err != nil {
		t.Fatal(err)
	}

	var shapeSubject *SubjectView
	for i := range publication.Subjects {
		if publication.Subjects[i].Key == (SubjectKey{Kind: SubjectStyle, Slug: "shape/corner-radius"}) {
			shapeSubject = &publication.Subjects[i]
			break
		}
	}
	if shapeSubject == nil {
		t.Fatal("shape style subject not found")
	}
	if shapeSubject.Name != "Shape" {
		t.Fatalf("shape subject name = %q, want %q", shapeSubject.Name, "Shape")
	}
	if len(shapeSubject.SourcePageURLs) != 1 || shapeSubject.SourcePageURLs[0] != "https://m3.material.io/styles/shape/corner-radius-scale" {
		t.Fatalf("shape page urls = %#v, want only shape/corner-radius-scale", shapeSubject.SourcePageURLs)
	}
}
