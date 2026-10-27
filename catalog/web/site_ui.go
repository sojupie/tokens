package web

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
func exportFormatChoices() []catalogChoice {
	return []catalogChoice{{"tokens.tokens.json", "DTCG JSON"}, {"tokens.css", "CSS variables"}, {"report.json", "Conversion report"}}
}
func exportContextChoices() []catalogChoice {
	return []catalogChoice{{"light-default", "Light · Default contrast"}, {"light-medium", "Light · Medium contrast"}, {"light-high", "Light · High contrast"}, {"dark-default", "Dark · Default contrast"}, {"dark-medium", "Dark · Medium contrast"}, {"dark-high", "Dark · High contrast"}}
}
