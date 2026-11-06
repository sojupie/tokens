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

// Reuse the catalog select while keeping history options independent of previews.
func HistoryChoices(all string, options []ChangeOption) []CatalogChoice {
	choices := []CatalogChoice{{Value: "", Label: all}}
	for _, option := range options {
		choices = append(choices, CatalogChoice{Value: option.Value, Label: option.Label})
	}
	return choices
}
