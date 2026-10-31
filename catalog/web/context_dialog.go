package web

import "sort"

func contextOrder(name string) int {
	switch name {
	case "Platform":
		return 0
	case "Theme":
		return 1
	case "Contrast":
		return 2
	}
	return 3
}

func tableContextFilters(filters []tokenFilter, extra bool) []tokenFilter {
	var selected []tokenFilter
	for _, filter := range filters {
		if (contextOrder(filter.Name) == 3) == extra {
			selected = append(selected, filter)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool { return contextOrder(selected[i].Name) < contextOrder(selected[j].Name) })
	return selected
}

func tableContextOptions(filter tokenFilter) []catalogChoice {
	choices := make([]catalogChoice, 0, len(filter.Options))
	for _, option := range filter.Options {
		choices = append(choices, catalogChoice{option, option})
	}
	return choices
}
