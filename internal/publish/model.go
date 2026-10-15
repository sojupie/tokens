package publish

import (
	"encoding/json"
	"strings"
)

const SchemaVersion = 4

type SubjectKind string

const (
	SubjectComponent      SubjectKind = "component"
	SubjectStyle          SubjectKind = "style"
	SubjectFoundation     SubjectKind = "foundation"
	SubjectReferenceSet   SubjectKind = "reference-token-set"
	SubjectSystemTokenSet SubjectKind = "system-token-set"
)

type SubjectKey struct {
	Kind SubjectKind `json:"kind"`
	Slug string      `json:"slug"`
}

func (key SubjectKey) String() string {
	return string(key.Kind) + ":" + key.Slug
}

type Publication struct {
	SchemaVersion   int                 `json:"schemaVersion"`
	PublicationID   string              `json:"publicationId"`
	GeneratedAt     string              `json:"generatedAt"`
	CurrentSnapshot SnapshotView        `json:"currentSnapshot"`
	Snapshots       []SnapshotView      `json:"snapshots"`
	Changes         []ChangeView        `json:"changes,omitempty"`
	Subjects        []SubjectView       `json:"subjects"`
	SourcePayloads  []SourcePayloadView `json:"sourcePayloads"`
}

type SnapshotView struct {
	ID            string `json:"id"`
	CreatedAt     string `json:"createdAt"`
	ManifestHash  string `json:"manifestHash"`
	CaptureStatus string `json:"captureStatus"`
	ReviewStatus  string `json:"reviewStatus"`
}

type ChangeView struct {
	DetectedAt         string `json:"detectedAt"`
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Subject            string `json:"subject,omitempty"`
	SourcePayloadKind  string `json:"sourcePayloadKind,omitempty"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	Before             string `json:"before,omitempty"`
	After              string `json:"after,omitempty"`
	PreviousSnapshotID string `json:"previousSnapshotId"`
	SnapshotID         string `json:"snapshotId"`
}

type SubjectView struct {
	Key              SubjectKey            `json:"key"`
	Name             string                `json:"name"`
	SourcePageURLs   []string              `json:"sourcePageUrls"`
	Tokens           []TokenView           `json:"tokens"`
	LifecycleNotices []LifecycleNoticeView `json:"lifecycleNotices"`
}

// PublicSubject is the small token document exposed to API and download users.
// SubjectView remains the internal publication model and keeps source lineage.
type PublicSubject struct {
	Kind   SubjectKind   `json:"kind"`
	Slug   string        `json:"slug"`
	Name   string        `json:"name"`
	Tokens []PublicToken `json:"tokens"`
}

type PublicTokenDocument struct {
	CapturedAt string        `json:"capturedAt"`
	Kind       SubjectKind   `json:"kind"`
	Slug       string        `json:"slug"`
	Name       string        `json:"name"`
	Tokens     []PublicToken `json:"tokens"`
}

type PublicToken struct {
	Presentation *TokenPresentation `json:"presentation,omitempty"`
	Name         string             `json:"name"`
	DisplayName  string             `json:"displayName,omitempty"`
	Type         string             `json:"type,omitempty"`
	Values       []PublicTokenValue `json:"values"`
	Lifecycle    string             `json:"lifecycle,omitempty"`
}

type PublicTokenValue struct {
	Context    []string         `json:"context,omitempty"`
	Default    bool             `json:"default,omitempty"`
	Value      json.RawMessage  `json:"value"`
	References []TokenReference `json:"references,omitempty"`
}

func (subject SubjectView) Public() PublicSubject {
	public := PublicSubject{
		Kind:   subject.Key.Kind,
		Slug:   subject.Key.Slug,
		Name:   subject.Name,
		Tokens: make([]PublicToken, 0, len(subject.Tokens)),
	}
	for _, token := range subject.Tokens {
		publicToken := PublicToken{
			Name:         token.SourceName,
			Presentation: token.Presentation,
			DisplayName:  token.DisplayName,
			Type:         token.Category,
			Values:       make([]PublicTokenValue, 0, len(token.Values)),
		}
		if !strings.EqualFold(token.Lifecycle, "active") {
			publicToken.Lifecycle = token.Lifecycle
		}
		for _, value := range token.Values {
			publicToken.Values = append(publicToken.Values, PublicTokenValue{
				Context:    append([]string(nil), value.ContextLabels...),
				Default:    value.IsDefaultContext,
				Value:      append(json.RawMessage(nil), value.Value...),
				References: value.References,
			})
		}
		public.Tokens = append(public.Tokens, publicToken)
	}
	return public
}

func (subject SubjectView) PublicDocument(capturedAt string) PublicTokenDocument {
	public := subject.Public()
	return PublicTokenDocument{
		CapturedAt: capturedAt,
		Kind:       public.Kind,
		Slug:       public.Slug,
		Name:       public.Name,
		Tokens:     public.Tokens,
	}
}

type TokenView struct {
	Presentation *TokenPresentation `json:"presentation,omitempty"`
	SourceName   string             `json:"sourceName"`
	DisplayName  string             `json:"displayName,omitempty"`
	Category     string             `json:"category,omitempty"`
	Values       []TokenValueView   `json:"values"`
	Raw          json.RawMessage    `json:"raw,omitempty"`
	Lifecycle    string             `json:"lifecycle,omitempty"`
	Sources      []TokenSourceView  `json:"sources"`
}

type TokenValueView struct {
	ContextTags      []string          `json:"contextTags,omitempty"`
	ContextLabels    []string          `json:"contextLabels,omitempty"`
	IsDefaultContext bool              `json:"isDefaultContext,omitempty"`
	Value            json.RawMessage   `json:"value"`
	References       []TokenReference  `json:"references,omitempty"`
	Raw              json.RawMessage   `json:"raw,omitempty"`
	Sources          []TokenSourceView `json:"sources"`
}

// TokenReference preserves named dependencies, including branching composite values.
// It excludes upstream record IDs and revision metadata.
type TokenReference struct {
	Name       string           `json:"name"`
	References []TokenReference `json:"references,omitempty"`
}

type TokenSourceView struct {
	PayloadID      string `json:"payloadId"`
	SourceRecordID string `json:"sourceRecordId"`
}

type LifecycleNoticeView struct {
	ID              string `json:"id"`
	State           string `json:"state"`
	Subject         string `json:"subject"`
	Message         string `json:"message"`
	SourcePayloadID string `json:"sourcePayloadId"`
	SourceRecordID  string `json:"sourceRecordId"`
}

type SourcePayloadView struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Hash          string   `json:"hash"`
	Kind          string   `json:"kind,omitempty"`
	Area          string   `json:"area,omitempty"`
	SourceSubject string   `json:"sourceSubject,omitempty"`
	PageType      string   `json:"pageType,omitempty"`
	PageURLs      []string `json:"pageUrls"`
}
