package playground

import (
	"fmt"
	"net/url"

	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
)

func configureScrollFieldPreview(page *webview.ComponentPreview, query url.Values) error {
	controls := []webview.PreviewControl{
		{Name: "wrap", Label: "Scrolling", Options: []string{"finite", "looped"}},
		{Name: "direct-entry", Label: "Direct entry", Options: []string{"off", "on"}},
		{Name: "entry", Label: "Mode", Options: []string{"scroll", "input"}, Labels: map[string]string{"input": "Text entry"}},
		{Name: "content", Label: "Values", Options: []string{"numbers", "three-digits", "four-digits", "text"},
			Labels: map[string]string{"numbers": "Two digits"}},
		{Name: "scene", Label: "Context", Options: []string{"standalone", "room"}, Labels: map[string]string{"room": "Room controls"}},
		{Name: "motion", Label: "Motion", Options: []string{"decelerate", "spring"}, Labels: map[string]string{"spring": "Fast spatial spring"}},
	}
	// Preserve the settings of links to the old combined specimens.
	legacy := map[string]map[string]string{
		"wrapping": {"wrap": "looped"}, "text": {"content": "text"},
		"paired": {"scene": "room"},
	}[query.Get("preview.variant")]
	values := make(map[string]string, len(controls))
	for _, control := range controls {
		value := query.Get("preview." + control.Name)
		if value == "" {
			value = legacy[control.Name]
		}
		if value == "" && control.Name == "scene" && page.Interactive {
			value = "room"
		}
		if value == "" && control.Name == "direct-entry" && query.Get("preview.entry") != "" {
			value = "on"
		}
		if value == "" {
			value = control.Options[0]
		}
		if !ContainsPreview(control.Options, value) {
			return fmt.Errorf("%w: %s %s", ErrPreviewInput, control.Name, value)
		}
		control.Selected = value
		values[control.Name] = value
		if control.Name != "entry" || values["direct-entry"] == "on" {
			page.Controls = append(page.Controls, control)
		}
	}
	for i := range page.Cases {
		c := &page.Cases[i]
		c.ScrollContent, c.ScrollScene = values["content"], values["scene"]
		c.ScrollWrap, c.ScrollMotion = values["wrap"] == "looped", values["motion"]
		c.ScrollDirectEntry, c.ScrollTextEntry = values["direct-entry"] == "on", values["entry"] == "input"
	}
	return nil
}

func configureTimePickerPreview(page *webview.ComponentPreview, query url.Values) error {
	for i := range page.Controls {
		if page.Controls[i].Name == "control" {
			page.Controls[i].Labels = map[string]string{
				"value": "Hour or minute toggle", "field": "Text input",
				"scroll": "Scroll field", "dial": "Clock face", "period": "AM/PM button",
			}
		}
	}
	if page.HasInlineTimePicker() {
		options := []string{"none", "filled"}
		container := query.Get("preview.container")
		if container == "" {
			container = options[0]
		}
		if !ContainsPreview(options, container) {
			return fmt.Errorf("%w: container %s", ErrPreviewInput, container)
		}
		page.Controls = append(page.Controls, webview.PreviewControl{
			Name: "container", Label: "Container", Options: options, Selected: container,
		})
		for i := range page.Cases {
			page.Cases[i].TimeContainer = container == "filled"
		}
	}
	if !page.HasTimeScroll() {
		return nil
	}
	options := []string{"looped", "finite"}
	wrap := query.Get("preview.wrap")
	if wrap == "" {
		wrap = "looped"
	}
	if !ContainsPreview(options, wrap) {
		return fmt.Errorf("%w: wrap %s", ErrPreviewInput, wrap)
	}
	motions := []string{"decelerate", "spring"}
	motion := query.Get("preview.motion")
	if motion == "" {
		motion = motions[0]
	}
	if !ContainsPreview(motions, motion) {
		return fmt.Errorf("%w: motion %s", ErrPreviewInput, motion)
	}
	page.Controls = append(page.Controls,
		webview.PreviewControl{Name: "wrap", Label: "Scrolling", Options: options, Selected: wrap},
		webview.PreviewControl{Name: "motion", Label: "Motion", Options: motions, Selected: motion,
			Labels: map[string]string{"spring": "Fast spatial spring"}},
	)
	for i := range page.Cases {
		page.Cases[i].TimeScrollFinite = wrap == "finite"
		page.Cases[i].ScrollMotion = motion
	}
	return nil
}
