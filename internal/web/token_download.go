package web

import (
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"sort"
	"strings"

	"github.com/sojupie/tokens/internal/publish"
)

type downloadPreview struct {
	Tokens, Values, Tables, Subjects int
	Examples, Contexts               string
}

func previewTokenDownload(subject publish.SubjectView) downloadPreview {
	preview := downloadPreview{Tokens: len(subject.Tokens)}
	var examples []string
	contexts := map[string]bool{}
	for _, token := range subject.Tokens {
		if len(examples) < 2 {
			examples = append(examples, tokenDisplayName(token))
		}
		preview.Values += len(token.Values)
		for _, value := range token.Values {
			for _, label := range value.ContextLabels {
				contexts[label] = true
			}
		}
	}
	preview.Examples = strings.Join(examples, "; ")
	settings := map[string][]string{}
	for label := range contexts {
		name, value, ok := strings.Cut(label, ": ")
		if !ok {
			name, value = "Other settings", label
		}
		settings[name] = append(settings[name], value)
	}
	var names []string
	for _, name := range []string{"Theme", "Platform", "Contrast", "Audience"} {
		if len(settings[name]) > 0 {
			names = append(names, name)
		}
	}
	var otherNames []string
	for name := range settings {
		if name != "Theme" && name != "Platform" && name != "Contrast" && name != "Audience" {
			otherNames = append(otherNames, name)
		}
	}
	sort.Strings(otherNames)
	names = append(names, otherNames...)
	var labels []string
	for _, name := range names {
		if len(labels) == 4 {
			break
		}
		values := settings[name]
		sort.Strings(values)
		label := name + ": "
		if len(values) > 4 {
			label += strings.Join(values[:4], ", ") + fmt.Sprintf(" (+%d more)", len(values)-4)
		} else {
			label += strings.Join(values, ", ")
		}
		labels = append(labels, label)
	}
	if len(labels) > 0 {
		preview.Contexts = "Contexts: " + strings.Join(labels, " · ")
		if len(names) > len(labels) {
			preview.Contexts += " · +" + downloadCount(len(names)-len(labels), "other setting")
		}
	} else if preview.Values > 0 {
		preview.Contexts = "Default values; no context-specific settings."
	}
	return preview
}

func previewSiteDownload(state publish.PublishedState) downloadPreview {
	preview := downloadPreview{Subjects: len(state.Subjects)}
	var names []string
	for _, subject := range sortedSubjects(state) {
		preview.Tokens += len(subject.Tokens)
		for _, token := range subject.Tokens {
			preview.Values += len(token.Values)
		}
		if len(names) < 3 {
			names = append(names, subject.Name)
		}
	}
	preview.Examples = strings.Join(names, "; ")
	preview.Contexts = "Every supported context across all components and styles."
	return preview
}

func downloadCount(count int, noun string) string {
	if count != 1 {
		noun += "s"
	}
	digits := fmt.Sprint(count)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits + " " + noun
}

func writeTokenDownloadMenu(writer io.Writer, subject publish.SubjectView, browser tokenBrowser, site downloadPreview) error {
	menu := `<div class="tokens-token-download"><button class="md-outlined-button" type="button" data-download-open aria-haspopup="dialog">Download JSON</button><dialog aria-labelledby="tokens-download-title"><form method="get" action="` + html.EscapeString(subjectDownloadURL(subject)) + `"><h2 id="tokens-download-title">Download tokens</h2>`
	if browser.Selected.Label != "" {
		menu += `<p class="tokens-download-table">` + html.EscapeString(browser.Selected.Label) + `</p>`
	}
	menu += `<div class="tokens-download-choices"><fieldset><legend>What should the file include?</legend>`
	menu += `<p class="tokens-download-help">A token can have different values for theme, platform and other settings. Select an option to see examples.</p>`
	option := func(value, label, description string, checked bool, preview downloadPreview) {
		menu += `<label class="tokens-download-option"><input type="radio" name="scope" value="` + value + `"`
		if checked {
			menu += ` checked`
		}
		counts := downloadCount(preview.Tokens, "token") + " · " + downloadCount(preview.Values, "value")
		if preview.Tables > 0 {
			counts += " · " + downloadCount(preview.Tables, "table")
		}
		if preview.Subjects > 0 {
			if preview.Subjects == 1 {
				counts += " · 1 component or style"
			} else {
				counts += fmt.Sprintf(" · %d components and styles", preview.Subjects)
			}
		}
		menu += `><span><strong>` + html.EscapeString(label) + `</strong><small>` + html.EscapeString(description) + `</small><small class="tokens-download-counts">` + html.EscapeString(counts) + `</small><span class="tokens-download-preview">`
		if value == "view" && browser.Query != "" {
			menu += `<small>Search: “` + html.EscapeString(browser.Query) + `”</small>`
		}
		if preview.Tokens == 0 {
			menu += `<small>No tokens match this selection.</small>`
		} else if preview.Examples != "" {
			prefix := "Examples: "
			if value == "all" && preview.Tables > 0 {
				prefix = "Tables included: "
			} else if value == "everything" {
				prefix = "Components and styles include: "
			}
			menu += `<small>` + prefix + html.EscapeString(preview.Examples) + `</small>`
		}
		if preview.Contexts != "" {
			menu += `<small>` + html.EscapeString(preview.Contexts) + `</small>`
		}
		menu += `</span></span></label>`
	}
	if len(browser.Sets) > 0 {
		view := previewTokenDownload(tokenDownloadSubject(subject, browser, "view"))
		view.Tables = 1
		set := previewTokenDownload(publish.SubjectView{Tokens: browser.Selected.sourceTokens})
		set.Tables = 1
		option("view", "This table, with your filters", "Matches your search and selected values.", true, view)
		option("set", "This table, with every value", "Keeps this table. Includes every supported context. Ignores search and filters.", false, set)
	}
	allLabel := "All tokens on this page"
	if subject.Name != "" {
		allLabel = "All " + subject.Name + " tokens"
	}
	all := previewTokenDownload(subject)
	all.Tables = len(browser.Sets)
	if len(browser.Sets) > 0 {
		var tables []string
		for _, set := range browser.Sets {
			if len(tables) < 3 {
				tables = append(tables, set.Label)
			}
		}
		all.Examples = strings.Join(tables, "; ")
		if len(browser.Sets) > 3 {
			all.Examples += fmt.Sprintf("; +%d more", len(browser.Sets)-3)
		}
	}
	option("all", allLabel, "Includes every table and every value. Ignores search and filters.", len(browser.Sets) == 0, all)
	option("everything", "All tokens on this site", "Includes all components and styles, with every supported context.", false, site)
	menu += `</fieldset></div>`
	input := func(name, value string) {
		menu += `<input type="hidden" name="` + html.EscapeString(name) + `" value="` + html.EscapeString(value) + `">`
	}
	input("set", browser.Selected.Key)
	input("q", browser.Query)
	for _, filter := range browser.Filters {
		input(filter.Key, filter.Selected)
	}
	menu += `<footer><button class="md-text-button" type="button" data-download-cancel>Cancel</button><button class="md-filled-button" type="submit">Download JSON</button></footer></form></dialog><noscript><style>.tokens-token-download>button{display:none}.tokens-token-download dialog:not([open]){display:block;position:static}</style></noscript></div>`
	_, err := io.WriteString(writer, menu)
	return err
}

// Filtered downloads use the same token set, search, and value selection as the table.
func (server *Server) handleTokenDownload(response http.ResponseWriter, request *http.Request, relativePath string) {
	scope := request.URL.Query().Get("scope")
	if scope == "everything" {
		http.Redirect(response, request, "/downloads/tokens.json?scope=all", http.StatusSeeOther)
		return
	}
	if scope != "view" && scope != "set" && scope != "all" {
		http.Error(response, "download scope must be view, set or all", http.StatusBadRequest)
		return
	}
	for key, subject := range server.state.Subjects {
		if relativePath != string(key.Kind)+"/"+key.Slug+".json" {
			continue
		}
		if scope == "all" {
			response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": strings.ReplaceAll(key.Slug, "/", "-") + ".json"}))
			writeJSON(response, http.StatusOK, subject.PublicDocument(server.state.CurrentSnapshot.CreatedAt))
			return
		}
		query := request.URL.Query()
		if scope == "set" {
			query.Del("q")
		}
		browser := buildTokenBrowser(subject, query)
		if len(browser.Sets) == 0 || (query.Get("set") != "" && query.Get("set") != browser.Selected.Key) {
			http.Error(response, "unknown token set", http.StatusBadRequest)
			return
		}
		export := tokenDownloadSubject(subject, browser, scope)
		filename := strings.ReplaceAll(subject.Key.Slug, "/", "-") + "-" + strings.ReplaceAll(browser.Selected.Key, "|", "-")
		if scope == "set" {
			filename += "-all-contexts"
		} else {
			for _, filter := range browser.Filters {
				filename += "-" + strings.ReplaceAll(strings.ToLower(filter.Selected), " ", "-")
			}
			if browser.Query != "" {
				filename += "-filtered"
			}
		}
		response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename + ".json"}))
		writeJSON(response, http.StatusOK, export.PublicDocument(server.state.CurrentSnapshot.CreatedAt))
		return
	}
	http.NotFound(response, request)
}

func (server *Server) handleCollectionDownload(response http.ResponseWriter, request *http.Request) {
	scope := request.URL.Query().Get("scope")
	if scope == "" {
		scope = "all"
	}
	if scope != "all" && scope != "components" && scope != "styles" {
		http.Error(response, "download scope must be components, styles or all", http.StatusBadRequest)
		return
	}
	subjects := []publish.PublicSubject{}
	for _, subject := range sortedSubjects(server.state) {
		component := subject.Key.Kind == publish.SubjectComponent
		if (scope == "components" && !component) || (scope == "styles" && component) {
			continue
		}
		subjects = append(subjects, subject.Public())
	}
	response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": "tokens-" + scope + ".json"}))
	writeJSON(response, http.StatusOK, struct {
		CapturedAt string                  `json:"capturedAt"`
		Subjects   []publish.PublicSubject `json:"subjects"`
	}{server.state.CurrentSnapshot.CreatedAt, subjects})
}

func tokenDownloadSubject(subject publish.SubjectView, browser tokenBrowser, scope string) publish.SubjectView {
	visible := make(map[string]bool)
	var collect func([]tokenGroup)
	collect = func(groups []tokenGroup) {
		for _, group := range groups {
			for _, row := range group.Tokens {
				visible[row.SourceName] = true
			}
			collect(group.Children)
		}
	}
	collect(browser.Selected.Groups)
	filters := make(map[string]string)
	for _, filter := range browser.Filters {
		filters[filter.Name] = filter.Selected
	}
	export := publish.SubjectView{Key: subject.Key, Name: subject.Name, Tokens: []publish.TokenView{}}
	for _, token := range browser.Selected.sourceTokens {
		if scope == "view" {
			if !visible[token.SourceName] {
				continue
			}
			value, ok := selectTokenValue(token.Values, filters)
			token.Values = []publish.TokenValueView{}
			if ok {
				token.Values = append(token.Values, value)
			}
		}
		export.Tokens = append(export.Tokens, token)
	}
	return export
}
