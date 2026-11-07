package presentation

type ExportContext struct {
	Name, ID, Selected string
	Choices            []CatalogChoice
}
type ExportPage struct {
	CSS, Implementations          bool
	Kind, Slug, Set, Query, Scope string
	Contexts                      []ExportContext
	Component                     bool
}

func (page ExportPage) Scopes() []CatalogChoice {
	choices := []CatalogChoice{}
	if page.Slug != "" {
		choices = append(choices, CatalogChoice{"view", "Current filtered table"}, CatalogChoice{"set", "Whole table"}, CatalogChoice{"subject", "This component or style"})
	}
	return append(choices, CatalogChoice{"collection", "Whole collection"})
}

func (page *ExportPage) DefaultScope() string {
	if page == nil {
		return "collection"
	}
	return page.Scope
}
