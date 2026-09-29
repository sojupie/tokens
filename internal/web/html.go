package web

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/sojupie/m3-source/internal/publish"
)

func (server *Server) handleOverview(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" && request.URL.Path != "/components" && request.URL.Path != "/styles" {
		http.NotFound(response, request)
		return
	}
	if request.URL.Path == "/" {
		if err := renderHTML(response, request, overviewPage(server.state.CurrentSnapshot, sortedSubjects(server.state))); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	styles := request.URL.Path == "/styles"
	query := request.URL.Query().Get("q")
	subjects := filterSubjects(sortedSubjects(server.state), query, styles)
	if request.Header.Get("HX-Request") == "true" {
		response.Header().Set("Vary", "HX-Request")
		if err := renderHTML(response, request, subjectList(subjects)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := renderHTML(response, request, subjectListPage(server.state.CurrentSnapshot, subjects, styles, query)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleHTMLSubject(response http.ResponseWriter, request *http.Request) {
	slug := request.PathValue("slug")
	if strings.HasSuffix(slug, "/changes") {
		server.handleHTMLChanges(response, request, strings.TrimSuffix(slug, "/changes"))
		return
	}
	subjects := make([]publish.SubjectView, 0, 1)
	for _, subject := range server.state.Subjects {
		if subject.Key.Slug != slug {
			continue
		}
		if strings.HasPrefix(request.URL.Path, "/components/") && subject.Key.Kind != publish.SubjectComponent {
			continue
		}
		if strings.HasPrefix(request.URL.Path, "/styles/") && subject.Key.Kind == publish.SubjectComponent {
			continue
		}
		subjects = append(subjects, subject)
	}
	if len(subjects) == 0 {
		http.NotFound(response, request)
		return
	}
	if len(subjects) > 1 {
		http.Error(response, "subject slug is ambiguous", http.StatusConflict)
		return
	}
	if request.Header.Get("HX-Request") == "true" {
		if sourceName := request.URL.Query().Get("variants"); sourceName != "" {
			values := tokenVariantValues(subjects[0], sourceName)
			if len(values) == 0 {
				http.NotFound(response, request)
				return
			}
			response.Header().Set("Vary", "HX-Request")
			if err := renderHTML(response, request, tokenVariants(values)); err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	query := request.URL.Query().Get("q")
	groups := subjectTokenGroups(subjects[0], query)
	if request.Header.Get("HX-Request") == "true" {
		response.Header().Set("Vary", "HX-Request")
		if err := renderHTML(response, request, subjectTokenGroupsView(subjects[0], groups)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := renderHTML(response, request, subjectPage(server.state.CurrentSnapshot, subjects[0], subjectChanges(server.state, subjects[0].Key.Slug), groups, query)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleHTMLChanges(response http.ResponseWriter, request *http.Request, subject string) {
	changes := subjectChanges(server.state, subject)
	if len(changes) == 0 {
		http.NotFound(response, request)
		return
	}
	page := requestedPage(request)
	pageChanges, page, hasPrevious, hasNext := paginateChanges(changes, page)
	if err := renderHTML(response, request, changesPage(server.state.CurrentSnapshot, subject, request.URL.Path, pageChanges, page, len(changes), hasPrevious, hasNext)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleGlobalChanges(response http.ResponseWriter, request *http.Request) {
	page := requestedPage(request)
	pageChanges, page, hasPrevious, hasNext := paginateChanges(server.state.Changes, page)
	if err := renderHTML(response, request, changesPage(server.state.CurrentSnapshot, "", request.URL.Path, pageChanges, page, len(server.state.Changes), hasPrevious, hasNext)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func renderHTML(response http.ResponseWriter, request *http.Request, component templ.Component) error {
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(request.Context(), response)
}

func sortedSubjects(state publish.PublishedState) []publish.SubjectView {
	subjects := make([]publish.SubjectView, 0, len(state.Subjects))
	for _, subject := range state.Subjects {
		subjects = append(subjects, subject)
	}
	sort.Slice(subjects, func(i, j int) bool {
		if subjects[i].Key.Kind == subjects[j].Key.Kind {
			return subjects[i].Key.Slug < subjects[j].Key.Slug
		}
		return subjects[i].Key.Kind < subjects[j].Key.Kind
	})
	return subjects
}

func filterSubjects(subjects []publish.SubjectView, query string, styles bool) []publish.SubjectView {
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := make([]publish.SubjectView, 0, len(subjects))
	for _, subject := range subjects {
		if styles != (subject.Key.Kind != publish.SubjectComponent) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(subject.Key.Slug), query) && !strings.Contains(strings.ToLower(subject.Name), query) {
			continue
		}
		filtered = append(filtered, subject)
	}
	return filtered
}

func subjectListPath(styles bool) string {
	if styles {
		return "/styles"
	}
	return "/components"
}

func subjectURL(subject publish.SubjectView) string {
	section := "styles"
	if subject.Key.Kind == publish.SubjectComponent {
		section = "components"
	}
	parts := strings.Split(subject.Key.Slug, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return "/" + section + "/" + strings.Join(parts, "/")
}

func subjectChangesURL(subject publish.SubjectView) string {
	return subjectURL(subject) + "/changes"
}

func subjectDownloadURL(subject publish.SubjectView) string {
	parts := strings.Split(subject.Key.Slug, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return "/downloads/" + url.PathEscape(string(subject.Key.Kind)) + "/" + strings.Join(parts, "/") + ".json"
}

func subjectKindLabel(kind publish.SubjectKind) string {
	if kind == publish.SubjectComponent {
		return "Component"
	}
	if kind == publish.SubjectStyle {
		return "Style"
	}
	return string(kind)
}

func tokenContextLabel(value publish.TokenValueView) string {
	label := strings.Join(value.ContextLabels, " · ")
	if value.IsDefaultContext || len(value.ContextTags) == 0 {
		if label == "" {
			return "Default"
		}
		return "Default · " + label
	}
	if label == "" {
		return "Other context"
	}
	return label
}

func snapshotDate(value string) string {
	createdAt, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return createdAt.Local().Format("2 Jan 2006")
}

func tokenValueText(value json.RawMessage) string {
	var decoded any
	if json.Unmarshal(value, &decoded) != nil {
		return string(value)
	}
	return formattedTokenValue(decoded)
}

func formattedTokenValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	case nil:
		return "No value published"
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			values = append(values, formattedTokenValue(item))
		}
		return strings.Join(values, ", ")
	case map[string]any:
		if len(typed) == 1 {
			for name, value := range typed {
				if formatted, ok := formattedNamedValue(name, value); ok {
					return formatted
				}
				return displayWords(name) + ": " + formattedTokenValue(value)
			}
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fields := make([]string, 0, len(keys))
		for _, key := range keys {
			value := formattedTokenValue(typed[key])
			if formatted, ok := formattedNamedValue(key, typed[key]); ok {
				value = formatted
			}
			fields = append(fields, displayWords(key)+": "+value)
		}
		return strings.Join(fields, " · ")
	}
	return "No value published"
}

func formattedNamedValue(name string, value any) (string, bool) {
	fields, ok := value.(map[string]any)
	switch name {
	case "color":
		if color, ok := value.(map[string]any); ok {
			return colorValueText(color)
		}
	case "dimension", "length", "elevation", "fontSize", "fontTracking", "lineHeight", "defaultSize":
		if ok {
			if text, ok := dimensionValueText(fields); ok {
				return text, true
			}
			unit, _ := fields["unit"].(string)
			if unit != "" {
				return "No value specified (" + displayUnit(unit) + ")", true
			}
			return "No numeric value specified", true
		}
	case "fontNames":
		if names, ok := fontNamesValue(value); ok {
			return names, true
		}
	case "fontWeight":
		if weight, ok := numberValue(value); ok {
			return formatNumber(weight), true
		}
	case "opacity":
		if opacity, ok := numberValue(value); ok {
			return formatNumber(opacity*100) + "%", true
		}
	case "shape":
		if ok {
			return shapeValueText(fields), true
		}
	case "type":
		if ok {
			return typographyValueText(fields), true
		}
	}
	return "", false
}

func colorValueText(color map[string]any) (string, bool) {
	red, redOK := numberValue(color["red"])
	green, greenOK := numberValue(color["green"])
	blue, blueOK := numberValue(color["blue"])
	alpha, alphaOK := numberValue(color["alpha"])
	if !redOK || !greenOK || !blueOK {
		return "", false
	}
	if red <= 1 && green <= 1 && blue <= 1 {
		red *= 255
		green *= 255
		blue *= 255
	}
	red = tidyColorChannel(red)
	green = tidyColorChannel(green)
	blue = tidyColorChannel(blue)
	if !alphaOK {
		alpha = 1
	}
	alpha = tidyColorChannel(alpha)
	return fmt.Sprintf("rgba(%s, %s, %s, %s)",
		formatNumber(red), formatNumber(green), formatNumber(blue), formatNumber(alpha)), true
}

func dimensionValueText(dimension map[string]any) (string, bool) {
	value, valueOK := numberValue(dimension["value"])
	unit, unitOK := dimension["unit"].(string)
	if !valueOK || !unitOK {
		return "", false
	}
	return formatNumber(value) + displayUnit(unit), true
}

func typographyValueText(fields map[string]any) string {
	parts := make([]string, 0, 6)
	if names, ok := fontNamesValue(fields["fontNames"]); ok {
		parts = append(parts, names)
	}
	if size, ok := formattedNamedValue("fontSize", fields["fontSize"]); ok {
		parts = append(parts, size)
	}
	if weight, ok := numberValue(fields["fontWeight"]); ok {
		parts = append(parts, "weight "+formatNumber(weight))
	}
	if lineHeight, ok := formattedNamedValue("lineHeight", fields["lineHeight"]); ok {
		parts = append(parts, "line height "+lineHeight)
	}
	if tracking, ok := formattedNamedValue("fontTracking", fields["fontTracking"]); ok {
		parts = append(parts, "tracking "+tracking)
	}
	known := map[string]bool{"fontNames": true, "fontSize": true, "fontWeight": true, "lineHeight": true, "fontTracking": true}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if known[key] {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, displayWords(key)+": "+formattedTokenValue(fields[key]))
	}
	return strings.Join(parts, " · ")
}

func fontNamesValue(value any) (string, bool) {
	fields, ok := value.(map[string]any)
	if !ok {
		return "", false
	}
	values, ok := fields["values"].([]any)
	if !ok || len(values) == 0 {
		return "", false
	}
	names := make([]string, 0, len(values))
	for _, item := range values {
		if name, ok := item.(string); ok && name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", false
	}
	return strings.Join(names, ", "), true
}

func shapeValueText(fields map[string]any) string {
	parts := make([]string, 0, len(fields))
	if family, ok := fields["family"].(string); ok {
		family = strings.TrimPrefix(family, "SHAPE_FAMILY_")
		parts = append(parts, displayWords(strings.ToLower(family)))
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if key != "family" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, displayWords(key)+": "+formattedTokenValue(map[string]any{key: fields[key]}))
	}
	if len(parts) == 0 {
		return "Shape not specified"
	}
	return strings.Join(parts, " · ")
}

func displayUnit(unit string) string {
	switch strings.TrimSpace(unit) {
	case "DIPS":
		return "dp"
	case "POINTS":
		return "pt"
	case "PIXELS":
		return "px"
	default:
		return strings.TrimSpace(unit)
	}
}

func tidyColorChannel(value float64) float64 {
	nearest := math.Round(value)
	if math.Abs(value-nearest) < 0.001 {
		return nearest
	}
	return value
}

func numberValue(value any) (float64, bool) {
	number, ok := value.(float64)
	return number, ok
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func subjectChanges(state publish.PublishedState, subject string) []publish.ChangeView {
	changes := make([]publish.ChangeView, 0)
	for _, change := range state.Changes {
		if change.Subject == subject {
			changes = append(changes, change)
		}
	}
	return changes
}

func changePageTitle(subject string) string {
	if subject == "" {
		return "Change log"
	}
	return subject + " changes"
}

func changeTypeLabel(changeType string) string {
	switch changeType {
	case "added":
		return "Added"
	case "removed":
		return "Removed"
	case "changed":
		return "Changed"
	default:
		return changeType
	}
}

func changeSummaryValue(value string) string {
	if value == "" {
		return "—"
	}
	return tokenValueText(json.RawMessage(value))
}

func requestedPage(request *http.Request) int {
	page, err := strconv.Atoi(request.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func paginateChanges(changes []publish.ChangeView, page int) (pageChanges []publish.ChangeView, actualPage int, hasPrevious, hasNext bool) {
	const pageSize = 50
	pageCount := (len(changes) + pageSize - 1) / pageSize
	if pageCount == 0 {
		return nil, 1, false, false
	}
	if page > pageCount {
		page = pageCount
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(changes) {
		end = len(changes)
	}
	return changes[start:end], page, page > 1, page < pageCount
}

func changesPageURL(base string, page int) string {
	if page <= 1 {
		return base
	}
	return base + "?page=" + strconv.Itoa(page)
}
