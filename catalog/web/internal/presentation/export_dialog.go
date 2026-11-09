package presentation

import (
	"net/url"
	"slices"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

// The dialog is an HTML client of the same selection/export rules as the public API.
// Its form still submits a native download; only editing choices uses htmx.
type ExportDialogState struct {
	Page                 ExportPage
	Query                url.Values
	Allowed              map[string][]string
	Counts, Error, Label string
	Ready, Resolved      bool
}

func InitialExportDialog(page ExportPage) ExportDialogState {
	query := url.Values{"kind": {page.Kind}, "slug": {page.Slug}, "set": {page.Set}, "q": {page.Query}, "scope": {page.Scope}, "content": {"tokens"}, "format": {"dtcg"}}
	for _, control := range page.Contexts {
		query.Set("context."+control.Name, control.Selected)
	}
	return ExportDialogState{Page: page, Query: query, Label: "Download"}
}

func (state ExportDialogState) Implementation() bool {
	return state.Query.Get("content") == "implementation"
}
func (state ExportDialogState) Choices(name string) []string {
	if state.Allowed == nil {
		return nil
	}
	return state.Allowed[name]
}
func (state ExportDialogState) Choice(name string, choice CatalogChoice) ui.ChoiceProps {
	disabled := false
	switch name {
	case "content":
		disabled = choice.Value == "implementation" && !state.Page.Implementations
	case "scope":
		disabled = state.Implementation() && (choice.Value == "view" || choice.Value == "set" || choice.Value == "subject" && !state.Page.Component)
	case "format":
		disabled = state.Implementation() || choice.Value == "css" && !state.Page.CSS
	default:
		if state.Allowed != nil {
			disabled = !slices.Contains(state.Allowed[name], choice.Value)
		}
	}
	return ui.ChoiceProps{ID: ExportChoiceID(name) + "-" + url.QueryEscape(choice.Value), Name: name, Value: choice.Value, Label: choice.Label, Checked: slices.Contains(state.Query[name], choice.Value), Disabled: disabled}
}
func (state ExportDialogState) ContextSummary() string {
	var labels []string
	for _, control := range state.Page.Contexts {
		if state.Allowed != nil && len(state.Choices("context."+control.Name)) == 0 {
			continue
		}
		if selected := state.Query["context."+control.Name]; len(selected) > 0 {
			labels = append(labels, control.Name+": "+strings.Join(selected, ", "))
		}
	}
	return strings.Join(labels, " · ")
}

// Request targets follow the same context order used to resolve allowed choices.
// Format changes affect only the summary; structural choices refresh dependent options.
func (state ExportDialogState) RequestAttrs(name string) templ.Attributes {
	endpoint := "/downloads/export/options"
	targets := []string{"#tokens-export-submit-pending"}
	dependent := name == "content" || name == "scope"
	if name == "content" {
		targets = append(targets, "#export-choices-scope")
	}
	if dependent {
		targets = append(targets, "#export-choices-format")
	}
	for _, control := range state.Page.Contexts {
		if dependent {
			targets = append(targets, "#"+ExportChoiceID("context."+control.Name))
		}
		if name == "context."+control.Name {
			dependent = true
		}
	}
	if name == "format" {
		endpoint = "/downloads/export/summary"
	}
	attrs := templ.Attributes{"hx-get": endpoint, "hx-trigger": "change", "hx-include": "closest form", "hx-sync": "closest form:replace", "hx-disable": strings.Join(targets, ","), "hx-target": "#tokens-export-fields", "hx-swap": "outerMorph", "hx-indicator": ExportLoadingTarget}
	if name == "format" {
		attrs["hx-swap"] = "none"
	}
	return attrs
}

const ExportLoadingTarget = "#tokens-export-loading-action"

func ExportChoiceID(name string) string {
	return "export-choices-" + strings.ToLower(strings.NewReplacer(".", "-", " ", "-").Replace(name))
}
