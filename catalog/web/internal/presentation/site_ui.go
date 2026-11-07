package presentation

import "github.com/sojupie/tokens/forge/components/ui"

type CatalogChoice struct{ Value, Label string }

func PreviewCompareTabIndex(selected bool) string {
	if selected {
		return "0"
	}
	return "-1"
}

func PreviewChoices(control PreviewControl) []CatalogChoice {
	choices := make([]CatalogChoice, 0, len(control.Options))
	for _, value := range control.Options {
		choices = append(choices, CatalogChoice{value, control.OptionLabel(value)})
	}
	return choices
}
func ChoiceLabel(choices []CatalogChoice, selected string) string {
	for _, choice := range choices {
		if choice.Value == selected {
			return choice.Label
		}
	}
	return selected
}

func CatalogOptions(choices []CatalogChoice) []ui.Option {
	options := make([]ui.Option, len(choices))
	for i, c := range choices {
		options[i] = ui.Option{Value: c.Value, Label: c.Label}
	}
	return options
}
