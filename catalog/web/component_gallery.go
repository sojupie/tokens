package web

import (
	"fmt"
	"strings"
)

// componentCase configures Catalog specimens; it is not a public component API.
type componentCase struct {
	// instanceID scopes nested controls without changing their token configuration.
	CarouselAction  string `json:"carouselAction,omitempty"`
	instanceID      string
	SliderIcon      string `json:"sliderIcon,omitempty"`
	SliderStops     string `json:"sliderStops,omitempty"`
	SliderDisplay   string `json:"sliderDisplay,omitempty"`
	SliderStart     string `json:"sliderStart,omitempty"`
	SliderEnd       string `json:"sliderEnd,omitempty"`
	SliderStep      string `json:"sliderStep,omitempty"`
	MenuMode        string `json:"menuMode,omitempty"`
	SegmentDensity  string `json:"segmentDensity,omitempty"`
	SearchContent   string `json:"searchContent,omitempty"`
	SearchElevation string `json:"searchElevation,omitempty"`
	DialogHeadline  string `json:"dialogHeadline,omitempty"`
	DateRange       bool   `json:"dateRange,omitempty"`
	ListMedia       string `json:"listMedia,omitempty"`
	ListDivider     string `json:"listDivider,omitempty"`
	ListTrailing    string `json:"listTrailing,omitempty"`
	ListText        string `json:"listText,omitempty"`
	ListInset       string `json:"listInset,omitempty"`
	ListSurface     string `json:"listSurface,omitempty"`
	ListExpandIcon  string `json:"listExpandIcon,omitempty"`
	NavLayout       string `json:"navLayout,omitempty"`
	NavCollapse     string `json:"navCollapse,omitempty"`
	NavAlign        string `json:"navAlign,omitempty"`
	NavHeader       string `json:"navHeader,omitempty"`
	NavIndicator    string `json:"navIndicator,omitempty"`
	NavBadge        string `json:"navBadge,omitempty"`
	Expanded        bool   `json:"expanded,omitempty"`
	Orientation     string `json:"orientation,omitempty"`
	Control         string `json:"control,omitempty"`
	Selection       string `json:"selection,omitempty"`
	Family          string `json:"family"`
	Variant         string `json:"variant"`
	Size            string `json:"size"`
	Shape           string `json:"shape"`
	Width           string `json:"width"`
	State           string `json:"state"`
	Selected        bool   `json:"selected"`
	Elevated        bool   `json:"elevated"`
	Invalid         bool   `json:"invalid"`
}

func (c componentCase) ID() string {
	if c.instanceID != "" {
		return c.instanceID
	}
	id := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%t-%t", c.Family, c.Variant, c.Size, c.Shape, c.Width, c.State, c.Selected, c.Elevated, c.Invalid)
	if c.DialogHeadline == "hidden" {
		id += "-no-headline"
	}
	if c.DateRange {
		id += "-range"
	}
	if c.Expanded {
		id += "-expanded"
	}
	if (c.Family == "list" || c.Family == "date" || c.Family == "time") && c.Selection != "" {
		id += "-" + c.Selection
	}
	if c.Control != "" {
		id += "-" + c.Control
	}
	if c.Orientation != "" {
		id += "-" + c.Orientation
	}
	if c.SegmentDensity != "" && c.SegmentDensity != "0" {
		id += "-density" + c.SegmentDensity
	}
	return id
}

func (c componentCase) Label() string {
	if c.Family == "button" {
		return c.button().Label()
	}
	parts := []string{c.Family}
	if c.Variant != "" {
		parts = append(parts, c.Variant)
	}
	if c.Size != "" {
		parts = append(parts, c.Size, c.Shape, c.Width+" width")
	}
	if c.Family == "chip" {
		if c.Elevated {
			parts = append(parts, "elevated")
		} else {
			parts = append(parts, "flat")
		}
	}
	if c.Family == "icon" || c.Family == "chip" || c.Family == "radio" || c.Family == "menu" {
		if c.Selected {
			parts = append(parts, "selected")
		} else {
			parts = append(parts, "unselected")
		}
	}
	if c.Invalid {
		parts = append(parts, "error")
	}
	return strings.Join(append(parts, c.State), " · ")
}

func (c componentCase) button() buttonCase {
	return buttonCase{Variant: c.Variant, Size: c.Size, Shape: c.Shape, Selection: c.Selection, State: c.State}
}
