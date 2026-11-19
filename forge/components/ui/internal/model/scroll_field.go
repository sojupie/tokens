package model

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/a-h/templ"
)

// ScrollFieldProps describes a finite or wrapping list with a selected index.
type ScrollFieldProps struct {
	ID, Name, Label, Size             string
	Motion                            ScrollFieldMotion
	Options                           []string
	Items                             []ScrollFieldOption
	Index                             int
	Wrap, Fill, Disabled, DirectEntry bool
	StartInTextEntry                  bool
	Attrs                             templ.Attributes
}

// ScrollFieldOption separates the displayed label from the submitted value.
type ScrollFieldOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (p ScrollFieldProps) Entries() []ScrollFieldOption {
	if p.Items != nil {
		return p.Items
	}
	items := make([]ScrollFieldOption, len(p.Options))
	for i, option := range p.Options {
		items[i] = ScrollFieldOption{Label: option, Value: option}
	}
	return items
}

func (p ScrollFieldProps) optionCount() int {
	if p.Items != nil {
		return len(p.Items)
	}
	return len(p.Options)
}

func (p ScrollFieldProps) TextEntryOpen() bool {
	return p.DirectEntry && p.StartInTextEntry && p.HasOptions()
}

func (p ScrollFieldProps) HasOptions() bool {
	return p.optionCount() > 0
}

// ScrollFieldMotion controls how the field settles after interaction.
type ScrollFieldMotion string

const (
	ScrollFieldMotionDecelerate ScrollFieldMotion = "decelerate"
	ScrollFieldMotionSpring     ScrollFieldMotion = "spring"
)

func (p ScrollFieldProps) IndexResolved() int {
	return max(0, min(p.Index, p.optionCount()-1))
}

func (p ScrollFieldProps) optionIndex(offset int) int {
	n := p.optionCount()
	if n == 0 {
		return -1
	}
	i := p.IndexResolved() + offset
	if p.Wrap {
		i = (i + n) % n
	}
	if i < 0 || i >= n {
		return -1
	}
	return i
}

func (p ScrollFieldProps) HasOption(offset int) bool {
	return p.optionIndex(offset) >= 0
}

func (p ScrollFieldProps) Option(offset int) ScrollFieldOption {
	i := p.optionIndex(offset)
	if i < 0 {
		return ScrollFieldOption{}
	}
	if p.Items != nil {
		return p.Items[i]
	}
	return ScrollFieldOption{Label: p.Options[i], Value: p.Options[i]}
}

func (p ScrollFieldProps) Root() templ.Attributes {
	options, _ := json.Marshal(p.Entries())
	motion := ScrollFieldMotionDecelerate
	if p.Motion == ScrollFieldMotionSpring {
		motion = p.Motion
	}
	attrs := templ.Attributes{
		"class": "md-scroll-field", "data-scroll-field": "", "data-scroll-options": string(options),
		"data-scroll-index": p.IndexResolved(), "data-scroll-wrap": fmt.Sprint(p.Wrap),
		"data-scroll-size": Fallback(p.Size, "small"), "data-scroll-fill": fmt.Sprint(p.Fill),
		"data-scroll-motion": string(motion),
		"data-scroll-direct": p.TextEntryOpen(),
	}
	if p.ID != "" {
		attrs["id"] = p.ID + "-scroll-field"
	}
	return Attributes(attrs, p.Attrs)
}

func (p ScrollFieldProps) Rotor() templ.Attributes {
	return templ.Attributes{
		"class": "md-scroll-rotor", "type": "button", "role": "spinbutton",
		"id": p.ID, "aria-label": p.Label, "aria-valuemin": 1,
		"aria-valuemax": max(1, p.optionCount()), "aria-valuenow": p.IndexResolved() + 1,
		"aria-valuetext": p.Option(0).Label, "disabled": p.Disabled || !p.HasOptions(),
		"hidden": p.TextEntryOpen(),
	}
}

func (p ScrollFieldProps) InputMode() string {
	for _, option := range p.Entries() {
		if _, err := strconv.ParseFloat(option.Label, 64); err != nil {
			return "text"
		}
	}
	return "numeric"
}
