// Download preparation resolves selected data and assets only after the user requests an export.
package server

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	webview "github.com/sojupie/tokens/catalog/web/internal/presentation"
	webrender "github.com/sojupie/tokens/catalog/web/internal/render"
	"github.com/sojupie/tokens/forge/components"
	"github.com/sojupie/tokens/forge/components/ui"
	"github.com/sojupie/tokens/source/publish"
	xhtml "golang.org/x/net/html"
	"html"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// The specimen declares nested CSS modules while rendering. Runtime ownership
// lives in the same manifest as each module's authored styles and token bindings.
func (server *Server) implementationFiles(ctx context.Context, query url.Values) (map[string][]byte, string, error) {
	_, selection, subject, err := server.selectedExport(query)
	if err != nil {
		return nil, "", err
	}
	if query.Get("scope") != "subject" && query.Get("scope") != "collection" {
		return nil, "", fmt.Errorf("component implementations require a whole component or collection")
	}
	for key := range query {
		if strings.HasPrefix(key, "context.") && key != "context.Theme" && key != "context.Contrast" {
			return nil, "", fmt.Errorf("component implementations use the catalog Web context; only Theme and Contrast can change")
		}
	}
	if selection["Theme"] != "Light" && selection["Theme"] != "Dark" {
		return nil, "", fmt.Errorf("unknown Theme context")
	}
	if selection["Contrast"] != "Default" && selection["Contrast"] != "Medium contrast" && selection["Contrast"] != "High contrast" {
		return nil, "", fmt.Errorf("unknown Contrast context")
	}
	if len(server.themeBundles) == 0 {
		return nil, "", fmt.Errorf("component assets unavailable; rebuild token assets")
	}
	subjects := []publish.SubjectView{}
	name := "components"
	if query.Get("scope") == "subject" {
		if subject == nil || subject.Key.Kind != publish.SubjectComponent {
			return nil, "", fmt.Errorf("choose a component")
		}
		subjects = append(subjects, *subject)
		name = strings.ReplaceAll(subject.Key.Slug, "/", "-")
	} else {
		for _, item := range sortedSubjects(server.state) {
			if item.Key.Kind == publish.SubjectComponent {
				subjects = append(subjects, item)
			}
		}
	}
	assets := &webview.ComponentAssets{Modules: map[string]bool{"common": true}, Server: server}
	ctx = ui.WithAssets(context.WithValue(ctx, webview.ComponentAssetsKey{}, assets), ui.Assets(assets.Modules))
	var examples bytes.Buffer
	count := 0
	for _, item := range subjects {
		config := url.Values{"preview.compare": {"interactive"}}
		for key, value := range query {
			if strings.HasPrefix(key, "preview.") {
				config[key] = value
			}
		}
		preview, e := server.buildComponentPreview(item, config)
		if e != nil {
			return nil, "", e
		}
		if preview == nil || len(preview.Cases) == 0 {
			continue
		}
		fmt.Fprintf(&examples, "<section><h2>%s</h2>\n", html.EscapeString(webview.SubjectName(item)))
		if e := webrender.ComponentSpecimen(preview.Cases[0]).Render(ctx, &examples); e != nil {
			return nil, "", e
		}
		examples.WriteString("</section>\n")
		count++
	}
	if count == 0 {
		return nil, "", fmt.Errorf("no implemented component examples in this selection")
	}
	files := map[string][]byte{}
	notices, e := components.Files.ReadFile("static/NOTICE.txt")
	if e != nil {
		return nil, "", e
	}
	files["NOTICE.txt"] = notices
	appCSS, e := webassets.ReadStaticFile("static/app.css")
	if e != nil {
		return nil, "", e
	}
	var fontCSS strings.Builder
	for _, line := range strings.Split(string(appCSS), "\n") {
		if strings.HasPrefix(line, "@font-face") {
			fontCSS.WriteString(line)
			fontCSS.WriteByte('\n')
		}
	}
	for _, font := range []string{"roboto-latin.woff2", "roboto-latin-ext.woff2", "OFL.txt"} {
		data, e := webassets.ReadStaticFile("static/fonts/" + font)
		if e != nil {
			return nil, "", e
		}
		files["fonts/"+font] = data
	}
	files["fonts.css"] = []byte(fontCSS.String())

	cssFiles := []string{"buttons.css", "bundles/common.css"}
	themeFiles := []string{"shared"}
	scripts := []string{"motion.js", "buttons.js"}
	var scriptModules []string

	for _, id := range webview.ComponentModuleOrder() {
		if id == "common" || assets.Modules[id] {
			if id != "common" {
				cssFiles = append(cssFiles, "bundles/"+id+".css")
				themeFiles = append(themeFiles, id)
			}
			scriptModules = append(scriptModules, id)
		}
	}
	moduleScripts, err := components.ScriptsFor(scriptModules...)
	if err != nil {
		return nil, "", err
	}
	scripts = append(scripts, moduleScripts...)
	var tokenCSS, componentCSS strings.Builder
	for _, id := range themeFiles {
		tokenCSS.Write(server.themeBundles[id])
		tokenCSS.WriteByte('\n')
	}
	for _, file := range cssFiles {
		data, e := webassets.ReadStaticFile("static/" + file)
		if e != nil {
			return nil, "", e
		}
		componentCSS.Write(data)
		componentCSS.WriteByte('\n')
	}
	files["tokens.css"] = []byte(tokenCSS.String())
	files["components.css"] = []byte(componentCSS.String())

	if assets.Modules["dialog"] {
		scripts = append(scripts, "dialog-demo.js")
	}
	var scriptHTML strings.Builder
	for _, script := range scripts {
		data, e := webassets.ReadStaticFile("static/" + script)
		if e != nil {
			return nil, "", e
		}
		files[script] = data
		fmt.Fprintf(&scriptHTML, "<script src=%q defer></script>\n", script)
	}
	if slices.Contains(scripts, "loading.js") {
		data, e := webassets.ReadStaticFile("static/loading-shapes.json")
		if e != nil {
			return nil, "", e
		}
		files["loading-shapes.json"] = data
	}
	exampleHTML, err := bundleExampleImages(examples.String(), files)
	if err != nil {
		return nil, "", err
	}
	theme := "light"
	if selection["Theme"] == "Dark" {
		theme = "dark"
	}
	contrast := strings.ToLower(strings.TrimSuffix(selection["Contrast"], " contrast"))
	files["example.css"] = []byte("body{margin:0;padding:24px;background:var(--md-sys-color-surface);color:var(--md-sys-color-on-surface);font-family:Roboto,sans-serif}main{display:grid;gap:32px}section{min-width:0}h1,h2{font:inherit}svg{vertical-align:middle}button,input,select{font:inherit}\n")
	files["index.html"] = []byte(fmt.Sprintf("<!doctype html>\n<html lang=\"en\" data-theme=%q data-contrast=%q><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>%s example</title><link rel=\"stylesheet\" href=\"tokens.css\"><link rel=\"stylesheet\" href=\"components.css\"><link rel=\"stylesheet\" href=\"fonts.css\"><link rel=\"stylesheet\" href=\"example.css\">%s</head><body><main data-component-assets>%s</main></body></html>", theme, contrast, html.EscapeString(name), scriptHTML.String(), exampleHTML))
	files["README.txt"] = []byte("Component implementation from Token Catalog\n\nServe this directory with any static HTTP server, then open index.html. Some browser features (including loading-indicator geometry) require HTTP rather than file://. No build step or external CDN is required.\n\nindex.html contains example markup from the maintained playground specimens. components.css includes authored component rules and generated token bindings. tokens.css contains the catalog's Web theme and its overrides, including all six light/dark and contrast appearances; change data-theme and data-contrast on <html> to switch. It uses the catalog's motion scheme. This implementation theme differs from raw captured token exports. Roboto font files and their license are included. See NOTICE.txt for code and third-party license terms.\n\nThe JavaScript files provide component behavior. Keep their relative paths and script order. Nested components and common controls are included through the maintained module declarations. Example actions do not implement your application logic.\n\nToken data and implementation provenance: https://github.com/sojupie/tokens\nIndependent project. Not affiliated with Google.\n")
	return files, name, nil
}

func (server *Server) handleImplementationExport(w http.ResponseWriter, r *http.Request) {
	files, name, err := server.implementationFiles(r.Context(), r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("format") == "summary" {
		writeJSON(w, http.StatusOK, map[string]any{"files": len(files), "implementation": true})
		return
	}
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	names := make([]string, 0, len(files))
	for file := range files {
		names = append(names, file)
	}
	sort.Strings(names)
	for _, file := range names {
		entry, e := archive.Create(file)
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		if _, e = entry.Write(files[file]); e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
	}
	if err := archive.Close(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name+"-implementation.zip"))
	_, _ = w.Write(output.Bytes())
}

// Copy maintained example images into the archive instead of retaining Catalog URLs.
func bundleExampleImages(markup string, files map[string][]byte) (string, error) {
	z := xhtml.NewTokenizer(strings.NewReader(markup))
	for {
		kind := z.Next()
		if kind == xhtml.ErrorToken {
			break
		}
		if kind != xhtml.StartTagToken && kind != xhtml.SelfClosingTagToken {
			continue
		}
		token := z.Token()
		if token.Data != "img" {
			continue
		}
		for _, a := range token.Attr {
			if a.Key != "src" {
				continue
			}
			u, err := url.Parse(a.Val)
			if err != nil || u.IsAbs() || u.Host != "" {
				continue
			}
			_, relative, ok := strings.Cut(u.Path, "/static/")
			if !ok || !fs.ValidPath(relative) {
				continue
			}
			name := "assets/" + relative
			if _, ok := files[name]; !ok {
				data, err := webassets.ReadStaticFile("static/" + relative)
				if err != nil {
					return "", fmt.Errorf("example image %s: %w", relative, err)
				}
				files[name] = data
			}
			markup = strings.ReplaceAll(markup, html.EscapeString(a.Val), html.EscapeString(name))
		}
	}
	return markup, nil
}

// Repeated formats and context values describe a Cartesian product. Single-file
// requests keep their existing URLs and response types.
func exportSelections(query url.Values) ([]url.Values, error) {
	keys := []string{}
	for _, format := range query["format"] {
		valid := format == "zip" || format == "summary"
		if query.Get("content") != "implementation" {
			valid = format == "css" || format == "dtcg" || format == "source" || format == "report" || format == "summary"
		}
		if !valid {
			return nil, fmt.Errorf("invalid export format %q", format)
		}
	}
	for key, values := range query {
		if key != "format" && !strings.HasPrefix(key, "context.") {
			continue
		}
		seen := map[string]bool{}
		for _, value := range values {
			if value == "" || seen[value] {
				return nil, fmt.Errorf("empty or duplicate %s selection", key)
			}
			seen[value] = true
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	jobs := []url.Values{cloneExportQuery(query)}
	for _, key := range keys {
		if len(jobs)*len(query[key]) > 64 {
			return nil, fmt.Errorf("choose at most 64 combinations per download")
		}
		next := []url.Values{}
		for _, job := range jobs {
			for _, value := range query[key] {
				clone := cloneExportQuery(job)
				clone.Set(key, value)
				next = append(next, clone)
			}
		}
		jobs = next
	}
	return jobs, nil
}

func (server *Server) handleExportBatch(w http.ResponseWriter, r *http.Request, query url.Values) bool {
	jobs, err := exportSelections(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return true
	}
	if len(jobs) < 2 {
		return false
	}
	implementation := query.Get("content") == "implementation"
	if content := query.Get("content"); content != "" && content != "tokens" && !implementation {
		http.Error(w, "content must be tokens or implementation", http.StatusBadRequest)
		return true
	}
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	summaries := []exportSummary{}
	files := 0
	for i, job := range jobs {
		if r.Context().Err() != nil {
			archive.Close()
			return true
		}
		folder := "selection-" + strconv.Itoa(i+1) + "/"

		if implementation {
			var entries map[string][]byte
			entries, _, err = server.implementationFiles(r.Context(), job)
			if err == nil && job.Get("format") != "zip" {
				err = fmt.Errorf("component implementations use ZIP format")
			}
			names := []string{}
			for name := range entries {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				if err != nil {
					break
				}
				entry, e := archive.Create(folder + name)
				err = e
				if err == nil {
					_, err = entry.Write(entries[name])
					files++
				}
			}
		} else {
			document, report, summary, subject, e := server.exportDocument(job)
			err = e
			if err == nil {
				summaries = append(summaries, summary)
				files++
				if job.Get("format") != "report" {
					files++
				}
				if query.Get("summary") != "1" {
					var data []byte
					var filename string
					data, _, filename, err = server.renderTokenFile(job, document, report, summary, subject)
					if err == nil {
						var entry io.Writer
						entry, err = archive.Create(folder + filename)
						if err == nil {
							_, err = entry.Write(data)
						}
					}

					if err == nil && job.Get("format") != "report" {
						reportQuery := cloneExportQuery(job)
						reportQuery.Set("format", "report")
						data, _, filename, err = server.renderTokenFile(reportQuery, document, report, summary, subject)
						if err == nil {
							var entry io.Writer
							entry, err = archive.Create(folder + filename)
							if err == nil {
								_, err = entry.Write(data)
							}
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		archive.Close()
		http.Error(w, err.Error(), exportStatus(err))
		return true
	}
	if query.Get("summary") == "1" {
		archive.Close()
		writeJSON(w, http.StatusOK, struct {
			Files        int             `json:"files"`
			Combinations int             `json:"combinations"`
			Selections   []exportSummary `json:"selections"`
		}{files, len(jobs), summaries})
		return true
	}
	manifest, _ := archive.Create("selections.txt")
	for i, job := range jobs {
		fmt.Fprintf(manifest, "selection-%d: %s\n", i+1, job.Encode())
	}
	if err = archive.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="tokens-selection.zip"`)
	_, _ = w.Write(buffer.Bytes())
	return true
}

func cloneExportQuery(query url.Values) url.Values {
	clone := url.Values{}
	for key, values := range query {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}

// A native attachment has no page-side completion event. Signal when response
// headers are written, after preparation, without buffering the download in JS.
type exportDownloadResponse struct {
	http.ResponseWriter
	token   string
	started bool
}

func (w *exportDownloadResponse) WriteHeader(status int) {
	if w.started {
		return
	}
	w.started = true
	result := "ready"
	if status >= 400 {
		result = "error"
	}
	http.SetCookie(w.ResponseWriter, &http.Cookie{Name: "tokens-download-" + w.token, Value: result, Path: "/", MaxAge: 60, SameSite: http.SameSiteStrictMode})
	w.Header().Set("Cache-Control", "no-store")
	w.ResponseWriter.WriteHeader(status)
}
func (w *exportDownloadResponse) Write(body []byte) (int, error) {
	if !w.started {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
func downloadSignal(token string) bool {
	return token != "" && len(token) <= 64 && strings.Trim(token, "abcdefghijklmnopqrstuvwxyz0123456789-") == ""
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

func tokenDownloadSubject(subject publish.SubjectView, browser webview.TokenBrowser, scope string) publish.SubjectView {
	visible := make(map[string]bool)
	var collect func([]webview.TokenGroup)
	collect = func(groups []webview.TokenGroup) {
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
	for _, token := range browser.Selected.SourceTokens {
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
