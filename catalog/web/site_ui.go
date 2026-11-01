package web

import "github.com/sojupie/tokens/forge/components/ui"

type catalogChoice struct{ Value, Label string }

func previewChoices(control previewControl) []catalogChoice {
	choices := make([]catalogChoice, 0, len(control.Options))
	for _, value := range control.Options {
		choices = append(choices, catalogChoice{value, control.OptionLabel(value)})
	}
	return choices
}
func choiceLabel(choices []catalogChoice, selected string) string {
	for _, choice := range choices {
		if choice.Value == selected {
			return choice.Label
		}
	}
	return selected
}

func catalogOptions(choices []catalogChoice) []ui.Option {
	options := make([]ui.Option, len(choices))
	for i, c := range choices {
		options[i] = ui.Option{Value: c.Value, Label: c.Label}
	}
	return options
}
