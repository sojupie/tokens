// Page handlers normalize request filters and render publication views without changing captured values.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	"github.com/sojupie/tokens/catalog/web/internal/playground"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/forge/components/ui"
	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
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
		if err := server.renderHTML(response, request, webrender.OverviewPage(server.state.CurrentSnapshot, subjects, thumbnails, renderer)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	styles := request.URL.Path == "/styles"
	query := request.URL.Query().Get("q")
	partial := request.Header.Get("HX-Request") == "true"
	subjectFilter := ""
	if partial {
		subjectFilter = query
	}
	subjects := filterSubjects(sortedSubjects(server.state), subjectFilter, styles)
	thumbnails, err := server.componentThumbnails(subjects)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	if partial {
		response.Header().Set("Vary", "HX-Request, Accept")
		if err := server.renderHTML(response, request, webrender.SubjectCardGrid(subjects, thumbnails, styles, "")); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if err := server.renderHTML(response, request, webrender.SubjectListPage(server.state.CurrentSnapshot, subjects, styles, query, thumbnails)); err != nil {
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
	if setPath == "" && strings.HasSuffix(slug, "/changes/data.json") {
		server.handleSubjectChangeData(response, request, strings.TrimSuffix(slug, "/changes/data.json"))
		return
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
		for _, set := range append(webview.TokenSets(subjects[0]), webview.AllTokenSet(subjects[0])) {
			if webview.TokenSetPath(set) == setPath {
				query.Set("set", set.Key)
				break
			}
		}
		if query.Get("set") == "" {
			if source, ok := legacyTokenSet(subjects[0], setPath); ok {
				target := webview.TokenSetURL(subjects[0], source)
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
	motion := server.state.Subjects[publish.SubjectKey{Kind: publish.SubjectStyle, Slug: "motion/overview"}]
	browser := buildTokenBrowserView(subjects[0], query, !wantsMarkdown(request), motion.Tokens)
	if query.Get("set") != "" && query.Get("set") != browser.Selected.Key {
		http.NotFound(response, request)
		return
	}
	if setPath != "" && webview.TokenSetPath(browser.Selected) != setPath {
		http.NotFound(response, request)
		return
	}
	if browser.TableOnly && wantsMarkdown(request) {
		if err := server.renderTokenTableMarkdown(response, request, browser); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if !browser.TableOnly {
		preview, err := server.buildComponentPreview(subjects[0], query)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, playground.ErrPreviewInput) {
				status = http.StatusBadRequest
			}
			http.Error(response, err.Error(), status)
			return
		}
		if preview != nil {
			preview.Action = request.URL.Path
		}
		browser.Preview = preview
	}
	if request.Header.Get("HX-Request") == "true" && !wantsMarkdown(request) {
		var fragment templ.Component
		switch request.Header.Get("HX-Target") {
		case "tokens-token-browser", "div#tokens-token-browser":
			fragment = webrender.TokenBrowserView(subjects[0], browser)
		case "tokens-context-fields", "div#tokens-context-fields":
			fragment = webrender.TableContextFields(browser.Filters)
		}
		if fragment != nil {
			if err := server.renderHTML(response, request, fragment); err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	preview := browser.Preview
	if preview != nil && request.Header.Get("HX-Request") == "true" && (request.Header.Get("HX-Target") == "component-preview" || request.Header.Get("HX-Target") == "section#component-preview") && !wantsMarkdown(request) {
		if err := server.renderHTML(response, request, webrender.ComponentPreviewView(*preview)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	info := server.basicMetadata(webview.SubjectURL(subjects[0]), query)
	if browser.TableOnly {
		info.TableOnly = true
		info.Title = subjectTableTitle(subjects[0], browser.Selected)
		export := server.exportPage(&subjects[0], query)
		info.Export = &export
		request = request.WithContext(context.WithValue(request.Context(), webview.PageMetadataKey{}, info))
		if err := server.renderHTML(response, request, webrender.TableOnlyPage(info.Title, subjects[0], browser)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	info.Title = subjectTokenTitle(subjects[0])
	info.Heading = info.Title
	if preview != nil {
		info.Title += " and playground"
		info.Heading = info.Title
	}
	info.Description = subjectTokenDescription(subjects[0], preview != nil)
	server.setIllustrationMetadata(&info, webview.SubjectURL(subjects[0]))
	export := server.exportPage(&subjects[0], query)
	info.Export = &export
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
		if !webview.IsDefaultTokenSet(subjects[0], browser.Selected) {
			info.Title = subjectTableTitle(subjects[0], browser.Selected)
			info.Heading = info.Title
			info.Description = subjectTableDescription(subjects[0], browser.Selected)
		}
		if !info.NoIndex {
			info.Canonical = server.publicOrigin + tokenTableCanonicalPath(subjects[0], browser.Selected)
		}
	}
	request = request.WithContext(context.WithValue(request.Context(), webview.PageMetadataKey{}, info))
	if wantsMarkdown(request) && len(browser.Sets) > 0 {
		guide := server.markdownTokenGuide(subjects[0], browser, info.JSONAlternate)
		request = request.WithContext(context.WithValue(request.Context(), markdownTokenGuideKey{}, guide))
	}
	if err := server.renderHTML(response, request, webrender.SubjectPage(subjects[0], subjectChanges(server.state, subjects[0].Key.Slug), browser)); err != nil {
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

type changeSearchIndexDocument struct {
	Snapshots []string `json:"snapshots"`
	Subjects  []string `json:"subjects"`
	Types     []string `json:"types"`
	Items     [][4]any `json:"items"`
}

func (server *Server) changeSearchIndex() ([]byte, string) {
	server.changeSearchIndexOnce.Do(func() {
		doc := changeSearchIndexDocument{
			Snapshots: []string{},
			Subjects:  []string{},
			Types:     []string{},
			Items:     [][4]any{},
		}
		snapshotIdx := map[string]int{}
		subjectIdx := map[string]int{}
		typeIdx := map[string]int{}
		type itemKey struct {
			snapshot int
			subject  int
			kind     string
			name     string
		}
		itemPos := map[itemKey]int{}
		intern := func(table map[string]int, list *[]string, value string) int {
			if idx, ok := table[value]; ok {
				return idx
			}
			idx := len(*list)
			*list = append(*list, value)
			table[value] = idx
			return idx
		}
		for _, change := range server.state.Changes {
			sIdx := intern(snapshotIdx, &doc.Snapshots, change.SnapshotID)
			uIdx := intern(subjectIdx, &doc.Subjects, change.Subject)
			tIdx := intern(typeIdx, &doc.Types, change.Type)
			bit := 1 << tIdx
			key := itemKey{snapshot: sIdx, subject: uIdx, kind: change.Kind, name: change.Name}
			if pos, ok := itemPos[key]; ok {
				doc.Items[pos][2] = doc.Items[pos][2].(int) | bit
				continue
			}
			itemPos[key] = len(doc.Items)
			doc.Items = append(doc.Items, [4]any{sIdx, uIdx, bit, webview.ChangeSearchText(change.Name)})
		}
		data, err := json.Marshal(doc)
		if err != nil {
			data = []byte(`{"snapshots":[],"subjects":[],"types":[],"items":[]}`)
		}
		server.changeSearchIndexData = data
		server.changeSearchIndexVersion = webassets.AssetVersion(data)
	})
	return server.changeSearchIndexData, server.changeSearchIndexVersion
}

func (server *Server) handleChangeSearchIndex(response http.ResponseWriter, request *http.Request) {
	data, _ := server.changeSearchIndex()
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Header().Set("X-Robots-Tag", "noindex")
	serveCachedAsset(response, request, "search-index.json", data)
}

type cachedChangeData struct {
	data    []byte
	version string
}

type subjectChangeDataDocument struct {
	BrowseURL string                `json:"browseUrl,omitempty"`
	Snapshots []subjectSnapshotMeta `json:"snapshots"`
	Types     []subjectTypeMeta     `json:"types"`
	Items     []subjectChangeRecord `json:"items"`
}

type subjectSnapshotMeta struct {
	ID            string `json:"id"`
	DetectedAt    string `json:"detectedAt"`
	ObservedLabel string `json:"observedLabel"`
	PreviousLabel string `json:"previousLabel,omitempty"`
}

type subjectTypeMeta struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type subjectChangeRecord struct {
	Snapshot  int         `json:"s"`
	Type      int         `json:"t"`
	Name      string      `json:"n"`
	Search    string      `json:"q"`
	Context   string      `json:"c,omitempty"`
	Hierarchy []string    `json:"h,omitempty"`
	Values    [][3]string `json:"v,omitempty"`
}

func (server *Server) subjectChangeData(subject string) ([]byte, string) {
	if cached, ok := server.subjectChangeDataCache.Load(subject); ok {
		entry := cached.(cachedChangeData)
		return entry.data, entry.version
	}
	changes := subjectChanges(server.state, subject)
	doc := subjectChangeDataDocument{
		Snapshots: []subjectSnapshotMeta{},
		Types:     []subjectTypeMeta{},
		Items:     make([]subjectChangeRecord, 0, len(changes)),
	}
	for _, item := range server.state.Subjects {
		if item.Key.Slug == subject {
			doc.BrowseURL = server.basePath + webview.SubjectURL(item)
			break
		}
	}
	snapshotDates := make(map[string]string, len(server.state.Snapshots))
	for _, snap := range server.state.Snapshots {
		snapshotDates[snap.ID] = snap.CreatedAt
	}
	snapshotIdx := map[string]int{}
	typeIdx := map[string]int{}
	for _, change := range changes {
		sIdx, ok := snapshotIdx[change.SnapshotID]
		if !ok {
			detected := change.DetectedAt
			if detected == "" {
				detected = snapshotDates[change.SnapshotID]
			}
			prev := snapshotDates[change.PreviousSnapshotID]
			meta := subjectSnapshotMeta{
				ID:            change.SnapshotID,
				DetectedAt:    detected,
				ObservedLabel: webview.ObservationDate(detected),
			}
			if prev != "" {
				meta.PreviousLabel = webview.ObservationDate(prev)
			}
			sIdx = len(doc.Snapshots)
			doc.Snapshots = append(doc.Snapshots, meta)
			snapshotIdx[change.SnapshotID] = sIdx
		}
		tIdx, ok := typeIdx[change.Type]
		if !ok {
			tIdx = len(doc.Types)
			doc.Types = append(doc.Types, subjectTypeMeta{Key: change.Type, Label: webview.ChangeTypeLabel(change.Type)})
			typeIdx[change.Type] = tIdx
		}
		values := webview.ChangedValues(change)
		rows := make([][3]string, len(values))
		for i, v := range values {
			rows[i] = [3]string{v.Field, v.Before, v.After}
		}
		doc.Items = append(doc.Items, subjectChangeRecord{
			Snapshot:  sIdx,
			Type:      tIdx,
			Name:      change.Name,
			Search:    webview.ChangeSearchText(change.Name),
			Context:   webview.ChangeContext(change),
			Hierarchy: webview.ChangeHierarchy(change.Name),
			Values:    rows,
		})
	}
	data, err := json.Marshal(doc)
	if err != nil {
		data = []byte(`{"snapshots":[],"types":[],"items":[]}`)
	}
	entry := cachedChangeData{data: data, version: webassets.AssetVersion(data)}
	actual, _ := server.subjectChangeDataCache.LoadOrStore(subject, entry)
	stored := actual.(cachedChangeData)
	return stored.data, stored.version
}

func (server *Server) handleSubjectChangeData(response http.ResponseWriter, request *http.Request, subject string) {
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
	data, _ := server.subjectChangeData(subject)
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Header().Set("X-Robots-Tag", "noindex")
	serveCachedAsset(response, request, "data.json", data)
}

func filterChangesByScope(changes []publish.ChangeView, snapshot, changeType string) []publish.ChangeView {
	filtered := make([]publish.ChangeView, 0, len(changes))
	for _, change := range changes {
		if snapshot != "" && change.SnapshotID != snapshot {
			continue
		}
		if changeType != "" && change.Type != changeType {
			continue
		}
		filtered = append(filtered, change)
	}
	return filtered
}

func (server *Server) renderChangeHistory(response http.ResponseWriter, request *http.Request, subject string) {
	changes, filters := filterChangeHistory(server.state, subject, request.URL.Query())
	page := requestedPage(request)
	pageChanges, hasPrevious, hasNext := changes, false, false
	if subject != "" {
		pageChanges, page, hasPrevious, hasNext = paginateChanges(changes, page)
	}
	localMode, mounted := "server", pageChanges
	if !wantsMarkdown(request) {
		if subject == "" {
			localMode = "global"
			mounted = server.state.Changes
			if len(server.state.Changes) > 0 {
				_, filters.SearchVersion = server.changeSearchIndex()
			}
		} else {
			allSubject := subjectChanges(server.state, subject)
			if len(allSubject) <= 50 {
				localMode = "all"
				mounted = allSubject
			} else {
				localMode = "paginated"
				_, filters.DataVersion = server.subjectChangeData(subject)
				if filters.Snapshot != "" {
					snapshotScoped := filterChangesByScope(allSubject, filters.Snapshot, "")
					if len(snapshotScoped) <= 50 {
						localMode = "snapshot"
						mounted = snapshotScoped
					} else if filters.Type != "" {
						typeScoped := filterChangesByScope(snapshotScoped, "", filters.Type)
						if len(typeScoped) <= 50 {
							localMode = "query"
							mounted = typeScoped
						}
					}
				} else if filters.Type != "" {
					typeScoped := filterChangesByScope(allSubject, "", filters.Type)
					if len(typeScoped) <= 50 {
						localMode = "query"
						mounted = typeScoped
					}
				}
			}
		}
	}
	filters.LocalMode = localMode
	localMounted := localMode == "global" || localMode == "all" || localMode == "snapshot" || localMode == "query"
	batches := webview.BuildChangeTimeline(mounted, server.state.Snapshots, server.state.Subjects, filters, subject == "", localMounted)
	if request.Header.Get("HX-Request") == "true" && !wantsMarkdown(request) && (request.Header.Get("HX-Target") == "tokens-change-history" || request.Header.Get("HX-Target") == "div#tokens-change-history") {
		if err := server.renderHTML(response, request, webrender.ChangeHistoryView(server.state, subject, request.URL.String(), filters, changeCountSummary(changes), batches, page, len(changes), hasPrevious, hasNext)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		return
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
			view = publish.SubjectView{Key: publish.SubjectKey{Kind: kind, Slug: subject}, Name: webview.DisplayWords(subject)}
		}
		info.Title = subjectChangesTitle(view)
		info.Heading = info.Title
		info.Description = `Review Material Design 3 "` + webview.SubjectName(view) + `" token changes. Compare previous and current values by context and observation date.`
		if page > 1 {
			info.Title += ": Page " + strconv.Itoa(page)
			info.Description += " Page " + strconv.Itoa(page) + " shows older changes."
		}
	}
	request = request.WithContext(context.WithValue(request.Context(), webview.PageMetadataKey{}, info))
	if err := server.renderHTML(response, request, webrender.ChangesPage(server.state, subject, request.URL.String(), filters, changeCountSummary(changes), batches, page, len(changes), hasPrevious, hasNext)); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) renderTokenTableMarkdown(response http.ResponseWriter, request *http.Request, browser webview.TokenBrowser) error {
	var rendered bytes.Buffer
	if err := webview.WriteTokenGroups(request.Context(), &rendered, browser); err != nil {
		return err
	}
	markdown, err := htmlMarkdown(rendered.Bytes(), server.publicOrigin+request.URL.RequestURI(), "")
	if err != nil {
		return err
	}
	response.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	response.Header().Set("Vary", "Accept")
	response.Header().Set("X-Robots-Tag", "noindex, follow")
	_, err = writePageBody(response, request, []byte(markdown))
	return err
}

func (server *Server) shellVersion() string {
	server.shellVersionOnce.Do(func() {
		assets := &webview.ComponentAssets{Modules: map[string]bool{}, Server: server}
		ctx := context.Background()
		ctx = context.WithValue(ctx, webview.PageThemeKey{}, server.pageThemes())
		ctx = context.WithValue(ctx, webview.ThemeVersionKey{}, webassets.AssetVersion([]byte(server.themeCSS)))
		ctx = ui.WithAssets(context.WithValue(ctx, webview.ComponentAssetsKey{}, assets), ui.Assets(assets.Modules))
		var rendered bytes.Buffer
		_ = webrender.ShellAssets().Render(ctx, &rendered)
		server.shellVersionHash = webassets.AssetVersion(rendered.Bytes())
	})
	return server.shellVersionHash
}

func (server *Server) renderHTML(response http.ResponseWriter, request *http.Request, component templ.Component) error {
	response.Header().Set("Vary", "HX-Request, HX-Target, Accept")
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Header().Set("Cache-Control", "no-cache")
	ctx := request.Context()
	if webview.Metadata(ctx).Title == "" {
		ctx = context.WithValue(ctx, webview.PageMetadataKey{}, server.basicMetadata(request.URL.Path, request.URL.Query()))
	}
	ctx = context.WithValue(ctx, webview.PageThemeKey{}, server.pageThemes())
	ctx = context.WithValue(ctx, webview.ThemeVersionKey{}, webassets.AssetVersion([]byte(server.themeCSS)))
	info := webview.Metadata(ctx)
	info.Path = request.URL.Path
	if info.Illustration == nil {
		server.setIllustrationMetadata(&info, request.URL.Path)
	}
	info.Tools = server.toolDestinations()
	info.Navigation = server.navigation
	info.GeneratedDownloads = server.tokenAssetsDir != ""
	if info.Export == nil {
		page := server.exportPage(nil, request.URL.Query())
		info.Export = &page
	}
	info.Boosted = !info.TableOnly && !wantsMarkdown(request) &&
		(request.Header.Get("HX-Boosted") == "true" || request.Header.Get("HX-History-Restore-Request") == "true") &&
		(request.Header.Get("HX-Target") == "" || request.Header.Get("HX-Target") == "body")
	info.ShellVersion = server.shellVersion()
	ctx = context.WithValue(ctx, webview.PageMetadataKey{}, info)
	if webview.Metadata(ctx).NoIndex {
		response.Header().Set("X-Robots-Tag", "noindex, follow")
	}
	assets := &webview.ComponentAssets{Modules: map[string]bool{}, Server: server}
	ctx = ui.WithAssets(context.WithValue(ctx, webview.ComponentAssetsKey{}, assets), ui.Assets(assets.Modules))
	var rendered bytes.Buffer
	if err := component.Render(ctx, &rendered); err != nil {
		return err
	}
	links, err := server.ComponentAssetLinks(assets)
	if err != nil {
		return err
	}
	content := rendered.Bytes()
	if bytes.Contains(content, []byte("</head>")) {

		content = bytes.Replace(content, []byte("</head>"), append(links, []byte("</head>")...), 1)
	} else {
		content = append(links, content...)
	}
	content, err = hoistComponentStyles(content)
	if err != nil {
		return err
	}
	body, err := prefixHTML(content, server.basePath)
	if err != nil {
		return err
	}
	if wantsMarkdown(request) {
		guide, _ := ctx.Value(markdownTokenGuideKey{}).(string)
		markdown, err := htmlMarkdown(body, server.publicOrigin+request.URL.RequestURI(), guide)
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
	filtered := make([]publish.SubjectView, 0, len(subjects))
	for _, subject := range subjects {
		if styles != (subject.Key.Kind != publish.SubjectComponent) {
			continue
		}
		if !webview.SubjectMatchesQuery(subject, query) {
			continue
		}
		filtered = append(filtered, subject)
	}
	return filtered
}

func subjectDownloadURL(subject publish.SubjectView) string {
	parts := strings.Split(subject.Key.Slug, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return "/downloads/" + url.PathEscape(string(subject.Key.Kind)) + "/" + strings.Join(parts, "/") + ".json"
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

// Rounded shapes use the default for unspecified corners. Explicit zero is a
// square corner, including protobuf dimensions whose numeric field is omitted.
func shapePreviewCSS(raw json.RawMessage) string {
	return dtcg.ShapeCSS(raw)
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

// Old name-derived table URLs can redirect only when every token belongs to
// one source table. A merged legacy table has no single canonical replacement.
func legacyTokenSet(subject publish.SubjectView, path string) (webview.TokenSet, bool) {
	legacy := subject
	legacy.Tokens = append([]publish.TokenView(nil), subject.Tokens...)
	sourceSets := make(map[string]string)
	for i, token := range legacy.Tokens {
		if token.Presentation != nil {
			sourceSets[token.SourceName] = token.Presentation.SetName
		}
		legacy.Tokens[i].Presentation = nil
	}
	for _, set := range webview.TokenSets(legacy) {
		if webview.TokenSetPath(set) != path {
			continue
		}
		name := ""
		for _, token := range set.SourceTokens {
			source := sourceSets[token.SourceName]
			if source == "" || name != "" && name != source {
				return webview.TokenSet{}, false
			}
			name = source
		}
		for _, source := range webview.TokenSets(subject) {
			if source.SourceName == name && name != "" {
				return source, true
			}
		}
	}
	return webview.TokenSet{}, false
}

func subjectTokenDescription(subject publish.SubjectView, playground bool) string {
	text := subjectTokenTitle(subject) + ". "
	if playground {
		return text + "Try the interactive playground, compare values by context, follow references and download CSS variables or JSON."
	}
	return text + "Compare values by context, follow references and download CSS variables or JSON."
}

func subjectTokenTitle(subject publish.SubjectView) string {
	return `Material Design 3 "` + webview.SubjectName(subject) + `" tokens`
}

func subjectTableTitle(subject publish.SubjectView, set webview.TokenSet) string {
	if set.All {
		return subjectTokenTitle(subject) + ": All tokens"
	}
	return subjectTokenTitle(subject) + ": " + webview.TokenTableQualifier(subject, set)
}

func subjectTableDescription(subject publish.SubjectView, set webview.TokenSet) string {
	text := `Material Design 3 "` + webview.SubjectName(subject) + `" tokens`
	if !set.All {
		text += " for " + webview.TokenTableQualifier(subject, set)
	}
	return text + ". Compare values by context, follow references and download CSS variables or JSON."
}

func subjectChangesTitle(subject publish.SubjectView) string {
	return `Material Design 3 "` + webview.SubjectName(subject) + `" token changes`
}

func tokenTableCanonicalPath(subject publish.SubjectView, set webview.TokenSet) string {
	if webview.IsDefaultTokenSet(subject, set) || set.All && len(webview.TokenSets(subject)) == 1 {
		return webview.SubjectURL(subject)
	}
	return webview.TokenSetURL(subject, set)
}

func filterChangeHistory(state publish.PublishedState, subject string, query url.Values) ([]publish.ChangeView, webview.ChangeFilters) {
	filters := webview.ChangeFilters{Query: strings.TrimSpace(query.Get("q")), Subject: query.Get("subject"), Snapshot: query.Get("snapshot"), Type: query.Get("type"), FixedSubject: subject != ""}
	if subject != "" {
		filters.Subject = subject
	}
	labels := make(map[string]string)
	for _, item := range state.Subjects {
		labels[item.Key.Slug] = item.Name
	}
	subjects := make(map[string]bool)
	snapshots := make(map[string]bool)
	var matches []publish.ChangeView
	for _, change := range state.Changes {
		if subject != "" && change.Subject != subject {
			continue
		}
		subjects[change.Subject] = true
		snapshots[change.SnapshotID] = true
		if !webview.ChangeMatchesFilters(change, filters) {
			continue
		}
		matches = append(matches, change)
	}
	for slug := range subjects {
		label := labels[slug]
		if label == "" {
			label = webview.DisplayWords(slug)
		}
		filters.Subjects = append(filters.Subjects, webview.ChangeOption{Value: slug, Label: label})
	}
	sort.Slice(filters.Subjects, func(i, j int) bool { return filters.Subjects[i].Label < filters.Subjects[j].Label })
	for i := len(state.Snapshots) - 1; i >= 0; i-- {
		snapshot := state.Snapshots[i]
		if snapshots[snapshot.ID] {
			filters.Snapshots = append(filters.Snapshots, webview.ChangeOption{Value: snapshot.ID, Label: webview.ObservationDate(snapshot.CreatedAt)})
		}
	}
	return matches, filters
}

func changeCountSummary(changes []publish.ChangeView) string {
	counts := make(map[string]int)
	for _, change := range changes {
		counts[change.Type]++
	}
	return fmt.Sprintf("%d added · %d updated · %d removed", counts["added"], counts["changed"], counts["removed"])
}
