package ui

import "github.com/a-h/templ"

type ListAppearance string

const (
	Expressive ListAppearance = "expressive"
	Baseline   ListAppearance = "baseline"
)

type ListVariant string

const (
	StandardList  ListVariant = "standard"
	SegmentedList ListVariant = "segmented"
)

type ListLeading string

const (
	NoLeading       ListLeading = "none"
	LeadingIcon     ListLeading = "icon"
	LeadingAvatar   ListLeading = "avatar"
	LeadingImage    ListLeading = "image"
	LeadingVideo    ListLeading = "video"
	LeadingCheckbox ListLeading = "checkbox"
	LeadingRadio    ListLeading = "radio"
)

type ListLines string

const (
	OneLine    ListLines = "one-line"
	TwoLines   ListLines = "two-line"
	ThreeLines ListLines = "three-line"
)

type ListProps struct {
	ID, Label  string
	Ordered    bool
	Appearance ListAppearance
	Variant    ListVariant
	Lines      ListLines
	Leading    ListLeading
	Dividers   string
	Attrs      templ.Attributes
}

func (p ListProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-list", "data-shape": fallback(string(p.Appearance), "expressive"), "data-variant": fallback(string(p.Variant), "standard"), "data-size": fallback(string(p.Lines), "two-line"), "data-width": fallback(string(p.Leading), "none"), "data-divider": fallback(p.Dividers, "none")}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Label != "" {
		a["aria-label"] = p.Label
	}
	return attributes(a, p.Attrs)
}

type ListItemProps struct {
	ID, Label, LabelFor, Supporting, Overline, Metadata, Avatar, ImageSrc string
	LeadingKind                                                           ListLeading
	Leading, Trailing, Badge                                              templ.Component
	SingleLine, Unread, Current, Disabled                                 bool
	Selected                                                              *bool
	Attrs                                                                 templ.Attributes
}

func (p ListItemProps) Attributes() templ.Attributes {
	a := templ.Attributes{"class": "md-list-item", "data-list-item": "", "data-list-label": p.Label, "data-unread": p.Unread}
	if p.ID != "" {
		a["id"] = p.ID
	}
	if p.Selected != nil {
		a["data-selected"] = *p.Selected
	}
	if p.Current {
		a["aria-current"] = "page"
	}
	return attributes(a, p.Attrs)
}

// ListRowFrameProps describes structural slots for reveal, reorder and nested rows.
type ListRowFrameProps struct {
	Attrs, FrameAttrs, ContentAttrs, ChildrenAttrs templ.Attributes
	Native                                         bool
	Before, After, Content, Expanded               templ.Component
}
