package presentation

import "sort"

func ContextOrder(name string) int {
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

func TableContextFilters(filters []TokenFilter, extra bool) []TokenFilter {
	var selected []TokenFilter
	for _, filter := range filters {
		if (ContextOrder(filter.Name) == 3) == extra {
			selected = append(selected, filter)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool { return ContextOrder(selected[i].Name) < ContextOrder(selected[j].Name) })
	return selected
}

func TableContextOptions(filter TokenFilter) []CatalogChoice {
	choices := make([]CatalogChoice, 0, len(filter.Options))
	for _, option := range filter.Options {
		choices = append(choices, CatalogChoice{option, option})
	}
	return choices
}
