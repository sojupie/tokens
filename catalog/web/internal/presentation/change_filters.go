package presentation

type ChangeOption struct {
	Value string
	Label string
}

type ChangeFilters struct {
	Query        string
	Subject      string
	Snapshot     string
	Type         string
	Subjects     []ChangeOption
	Snapshots    []ChangeOption
	FixedSubject bool
}

func HistoryChangeTypes() []CatalogChoice {
	choices := []CatalogChoice{{Value: "", Label: "All changes"}}
	for _, value := range []string{"changed", "added", "removed", "context-added", "context-removed", "coverage-added", "coverage-lost"} {
		choices = append(choices, CatalogChoice{Value: value, Label: ChangeTypeLabel(value)})
	}
	return choices
}

// Reuse the catalog select while keeping history options independent of previews.
func HistoryChoices(all string, options []ChangeOption) []CatalogChoice {
	choices := []CatalogChoice{{Value: "", Label: all}}
	for _, option := range options {
		choices = append(choices, CatalogChoice{Value: option.Value, Label: option.Label})
	}
	return choices
}
