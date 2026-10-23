package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
)

func (server *Server) handleOverview(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" && request.URL.Path != "/components" && request.URL.Path != "/styles" {
		http.NotFound(response, request)
		return
	}
	if request.URL.Path == "/" {
		subjects := landingSubjects(sortedSubjects(server.state))
		thumbnails, err := server.componentThumbnails(subjects)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		renderer := "canvas"
		if request.URL.Query().Get("artwork") == "dom" {
			renderer = "dom"
		}
		if err := server.renderHTML(response, request, overviewPage(server.state.CurrentSnapshot, subjects, thumbnails, renderer)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	styles := request.URL.Path == "/styles"
	query := request.URL.Query().Get("q")
	subjects := filterSubjects(sortedSubjects(server.state), query, styles)
	thumbnails, err := server.componentThumbnails(subjects)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if request.Header.Get("HX-Request") == "true" {
		response.Header().Set("Vary", "HX-Request, Accept")
		if err := server.renderHTML(response, request, subjectCardGrid(subjects, thumbnails, styles)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := server.renderHTML(response, request, subjectListPage(server.state.CurrentSnapshot, subjects, styles, query, thumbnails)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleHTMLSubject(response http.ResponseWriter, request *http.Request) {
	slug := request.PathValue("slug")
	setPath := ""
	if index := strings.LastIndex(slug, "/sets/"); index >= 0 {
		setPath = slug[index+len("/sets/"):]
		slug = slug[:index]
		if setPath == "" {
			http.NotFound(response, request)
			return
		}
	}
	if setPath == "" && strings.HasSuffix(slug, "/changes") {
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
	query := request.URL.Query()
	if setPath != "" {
		query.Del("set")
		for _, set := range tokenSets(subjects[0]) {
			if tokenSetPath(set) == setPath {
				query.Set("set", set.Key)
				break
			}
		}
		if query.Get("set") == "" {
			if source, ok := legacyTokenSet(subjects[0], setPath); ok {
				target := tokenSetURL(subjects[0], source)
				if len(query) > 0 {
					target += "?" + query.Encode()
				}
				http.Redirect(response, request, target, http.StatusPermanentRedirect)
				return
			}
			http.NotFound(response, request)
			return
		}
	}
	preview, err := server.buildComponentPreview(subjects[0], query)
	if preview != nil {
		preview.Action = request.URL.Path
	}
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, errPreviewInput) {
			status = http.StatusBadRequest
		}
		http.Error(response, err.Error(), status)
		return
	}

	browser := buildTokenBrowser(subjects[0], query)
	browser.Preview = preview
	if query.Get("set") != "" && query.Get("set") != browser.Selected.Key {
		http.NotFound(response, request)
		return
	}
	if setPath != "" && tokenSetPath(browser.Selected) != setPath {
		http.NotFound(response, request)
		return
	}
	if preview != nil && request.Header.Get("HX-Request") == "true" && request.Header.Get("HX-Target") == "component-preview" && !wantsMarkdown(request) {
		if err := server.renderHTML(response, request, componentPreviewView(*preview)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	info := server.basicMetadata(subjectURL(subjects[0]), query)
	info.Title = subjectTokenTitle(subjects[0])
	info.Heading = info.Title
	info.Description = subjectTokenDescription(subjects[0])
	downloadQuery := url.Values{"scope": {"all"}}
	if len(browser.Sets) > 0 {
		downloadQuery.Set("scope", "view")
		if browser.Selected.Key != "" {
			downloadQuery.Set("set", browser.Selected.Key)
		}
		if browser.Query != "" {
			downloadQuery.Set("q", browser.Query)
		}
		for _, filter := range browser.Filters {
			downloadQuery.Set(filter.Key, filter.Selected)
		}
	}
	info.JSONAlternate = subjectDownloadURL(subjects[0]) + "?" + downloadQuery.Encode()
	if setPath != "" || query.Get("set") != "" {
		info.Title = subjectTableTitle(subjects[0], browser.Selected)
		info.Heading = info.Title
		info.Description = subjectTableDescription(subjects[0], browser.Selected)
		if !info.NoIndex {
			info.Canonical = server.publicOrigin + tokenTableCanonicalPath(subjects[0], browser.Selected)
		}
	}
	request = request.WithContext(context.WithValue(request.Context(), pageMetadataKey{}, info))
	if err := server.renderHTML(response, request, subjectPage(subjects[0], subjectChanges(server.state, subjects[0].Key.Slug), browser)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) handleHTMLChanges(response http.ResponseWriter, request *http.Request, subject string) {
	changes := subjectChanges(server.state, subject)
	if len(changes) == 0 {
		found := false
		for _, item := range server.state.Subjects {
			if item.Key.Slug == subject && (item.Key.Kind == publish.SubjectComponent) == strings.HasPrefix(request.URL.Path, "/components/") {
				found = true
				break
			}
		}
		if !found {
			http.NotFound(response, request)
			return
		}
	}
	server.renderChangeHistory(response, request, subject)
}

func (server *Server) handleGlobalChanges(response http.ResponseWriter, request *http.Request) {
	server.renderChangeHistory(response, request, "")
}

func (server *Server) renderChangeHistory(response http.ResponseWriter, request *http.Request, subject string) {
	changes, filters := filterChangeHistory(server.state, subject, request.URL.Query())
	page := requestedPage(request)
	pageChanges, hasPrevious, hasNext := changes, false, false
	if subject != "" {
		pageChanges, page, hasPrevious, hasNext = paginateChanges(changes, page)
	}
	query := request.URL.Query()
	query.Set("page", strconv.Itoa(page))
	info := server.basicMetadata(request.URL.Path, query)
	if subject != "" {
		kind := publish.SubjectStyle
		if strings.HasPrefix(request.URL.Path, "/components/") {
			kind = publish.SubjectComponent
		}
		view := server.state.Subjects[publish.SubjectKey{Kind: kind, Slug: subject}]
		if view.Name == "" {
			view = publish.SubjectView{Key: publish.SubjectKey{Kind: kind, Slug: subject}, Name: displayWords(subject)}
		}
		info.Title = subjectChangesTitle(view)
		info.Heading = info.Title
		info.Description = `Review Material Design 3 "` + subjectName(view) + `" token changes. Compare previous and current values by context and observation date.`
		if page > 1 {
			info.Title += ": Page " + strconv.Itoa(page)
			info.Description += " Page " + strconv.Itoa(page) + " shows older changes."
		}
	}
	request = request.WithContext(context.WithValue(request.Context(), pageMetadataKey{}, info))
	if err := server.renderHTML(response, request, changesPage(server.state, subject, request.URL.String(), filters, changeCountSummary(changes), pageChanges, page, len(changes), hasPrevious, hasNext)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) renderHTML(response http.ResponseWriter, request *http.Request, component templ.Component) error {
	response.Header().Set("Vary", "HX-Request, HX-Target, Accept")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Header().Set("Cache-Control", "no-cache")
	ctx := request.Context()
	if metadata(ctx).Title == "" {
		ctx = context.WithValue(ctx, pageMetadataKey{}, server.basicMetadata(request.URL.Path, request.URL.Query()))
	}
	info := metadata(ctx)
	info.Path = request.URL.Path
	info.GeneratedDownloads = server.tokenAssetsDir != ""
	ctx = context.WithValue(ctx, pageMetadataKey{}, info)
	if metadata(ctx).NoIndex {
		response.Header().Set("X-Robots-Tag", "noindex, follow")
	}
	ctx = context.WithValue(ctx, pageThemeKey{}, server.pageThemes(ctx))
	ctx = context.WithValue(ctx, themeVersionKey{}, assetVersion([]byte(server.themeCSS)))
	var rendered bytes.Buffer
	if err := component.Render(ctx, &rendered); err != nil {
		return err
	}
	body, err := prefixHTML(rendered.Bytes(), server.basePath)
	if err != nil {
		return err
	}
	if wantsMarkdown(request) {
		markdown, err := htmlMarkdown(body, server.publicOrigin+request.URL.RequestURI())
		if err != nil {
			return err
		}
		response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, err = writePageBody(response, request, []byte(markdown))
		return err
	}
	_, err = writePageBody(response, request, body)
	return err
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
	labels := make([]string, 0, len(value.ContextLabels))
	for _, contextLabel := range value.ContextLabels {
		contextLabel = strings.TrimPrefix(contextLabel, "Theme: ")
		contextLabel = strings.TrimPrefix(contextLabel, "Platform: ")
		contextLabel = strings.TrimPrefix(contextLabel, "Audience: 1P ")
		if contextLabel == "Contrast: Default" {
			continue
		}
		labels = append(labels, contextLabel)
	}
	label := strings.Join(labels, " · ")
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
	case "dimension", "length", "elevation", "fontSize", "fontTracking", "lineHeight", "defaultSize",
		"topLeft", "topRight", "bottomLeft", "bottomRight":
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
	red, green, blue, ok := colorChannels(color)
	alpha, alphaOK := numberValue(color["alpha"])
	if !ok {
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

// Source color messages omit zero-valued channels. An empty placeholder is
// still unresolved; alpha (or any channel) must establish a concrete color.
func colorChannels(color map[string]any) (float64, float64, float64, bool) {
	concrete := false
	for _, name := range []string{"red", "green", "blue", "alpha"} {
		if _, exists := color[name]; exists {
			concrete = true
		}
	}
	if !concrete {
		return 0, 0, 0, false
	}
	if value, exists := color["alpha"]; exists {
		alpha, ok := numberValue(value)
		if !ok || alpha < 0 || alpha > 1 {
			return 0, 0, 0, false
		}
	}
	channels := [3]float64{}
	for index, name := range []string{"red", "green", "blue"} {
		if value, exists := color[name]; exists {
			channel, ok := numberValue(value)
			if !ok || channel < 0 || channel > 255 {
				return 0, 0, 0, false
			}
			channels[index] = channel
		}
	}
	return channels[0], channels[1], channels[2], true
}

func dimensionValueText(dimension map[string]any) (string, bool) {
	return sourceDimensionText(dimension)
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

func sourceDimensionText(val any) (string, bool) {
	m, ok := val.(map[string]any)
	if !ok {
		return "", false
	}
	unit, ok := m["unit"].(string)
	if !ok || (unit != "DIPS" && unit != "POINTS" && unit != "PIXELS") {
		return "", false
	}
	value := 0.0
	if raw, exists := m["value"]; exists {
		var valid bool
		value, valid = numberValue(raw)
		if !valid {
			return "", false
		}
	}
	return formatNumber(value) + displayUnit(unit), true
}

func shapeValueText(fields map[string]any) string {
	family, _ := fields["family"].(string)
	if family == "SHAPE_FAMILY_CIRCULAR" || strings.EqualFold(family, "circular") {
		return "Full"
	}
	base, hasDefault := sourceDimensionText(fields["defaultSize"])
	keys := []string{"topLeft", "topRight", "bottomRight", "bottomLeft"}
	labels := []string{"Top-left", "Top-right", "Bottom-right", "Bottom-left"}
	var effective [4]string
	var specified [4]bool
	for i, key := range keys {
		effective[i], specified[i] = sourceDimensionText(fields[key])
		if !specified[i] && hasDefault {
			effective[i] = base
		}
	}
	uniform := effective[0] != ""
	for _, value := range effective[1:] {
		uniform = uniform && value == effective[0]
	}
	var parts []string
	if uniform {
		parts = append(parts, effective[0])
	} else {
		if hasDefault {
			parts = append(parts, "Default: "+base)
		}
		for i, value := range effective {
			if specified[i] && (!hasDefault || value != base) {
				parts = append(parts, labels[i]+": "+value)
			}
		}
	}
	if family != "" && family != "SHAPE_FAMILY_ROUNDED_CORNERS" {
		parts = append([]string{displayWords(strings.ToLower(strings.TrimPrefix(family, "SHAPE_FAMILY_")))}, parts...)
	}
	if len(parts) == 0 {
		return "Shape not specified"
	}
	return strings.Join(parts, " · ")
}

// Rounded shapes use the default for unspecified corners. Explicit zero is a
// square corner, including protobuf dimensions whose numeric field is omitted.
func shapePreviewCSS(raw json.RawMessage) string {
	return dtcg.ShapeCSS(raw)
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
	address, err := url.Parse(base)
	if err != nil {
		return base
	}
	query := address.Query()
	query.Del("page")
	if page > 1 {
		query.Set("page", strconv.Itoa(page))
	}
	address.RawQuery = query.Encode()
	return address.String()
}

// The landing page illustrates the catalog; the indexes carry the full lists.
func landingSubjects(subjects []publish.SubjectView) []publish.SubjectView {
	result := []publish.SubjectView{}
	selected := map[publish.SubjectKey]bool{}
	components, styles := 0, 0
	add := func(subject publish.SubjectView) {
		if selected[subject.Key] {
			return
		}
		component := subject.Key.Kind == publish.SubjectComponent
		if component && components >= 4 || !component && styles >= 4 {
			return
		}
		result = append(result, subject)
		selected[subject.Key] = true
		if component {
			components++
		} else {
			styles++
		}
	}
	for _, slug := range []string{"buttons", "cards", "chips", "menus", "typography/type-scale", "motion/overview", "elevation", "shape/corner-radius"} {
		for _, subject := range subjects {
			if subject.Key.Slug == slug {
				add(subject)
			}
		}
	}
	for _, subject := range subjects {
		add(subject)
	}
	return result
}
