package web

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sojupie/tokens/forge/components/ui"
	xhtml "golang.org/x/net/html"
	"html"
	"io/fs"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/sojupie/tokens/forge/components"
	"github.com/sojupie/tokens/source/publish"
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
	assets := &componentAssets{modules: map[string]bool{"common": true}, server: server}
	ctx = ui.WithAssets(context.WithValue(ctx, componentAssetsKey{}, assets), ui.Assets(assets.modules))
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
		fmt.Fprintf(&examples, "<section><h2>%s</h2>\n", html.EscapeString(subjectName(item)))
		if e := componentSpecimen(preview.Cases[0]).Render(ctx, &examples); e != nil {
			return nil, "", e
		}
		examples.WriteString("</section>\n")
		count++
	}
	if count == 0 {
		return nil, "", fmt.Errorf("no implemented component examples in this selection")
	}
	manifestBytes, err := components.Files.ReadFile("modules.json")
	if err != nil {
		return nil, "", err
	}
	var manifest map[string]struct{ Scripts []string }
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, "", err
	}
	files := map[string][]byte{}
	notices, e := components.Files.ReadFile("static/NOTICE.txt")
	if e != nil {
		return nil, "", e
	}
	files["NOTICE.txt"] = notices
	appCSS, e := readStaticFile("static/app.css")
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
		data, e := readStaticFile("static/fonts/" + font)
		if e != nil {
			return nil, "", e
		}
		files["fonts/"+font] = data
	}
	files["fonts.css"] = []byte(fontCSS.String())

	cssFiles := []string{"buttons.css", "bundles/common.css"}
	themeFiles := []string{"shared"}
	scripts := []string{"motion.js", "buttons.js"}
	seen := map[string]bool{"motion.js": true, "buttons.js": true}
	// Common runtime utilities precede every optional module.
	for _, id := range componentModuleOrder() {
		if id == "common" || assets.modules[id] {
			if id != "common" {
				cssFiles = append(cssFiles, "bundles/"+id+".css")
				themeFiles = append(themeFiles, id)
			}
			for _, script := range manifest[id].Scripts {
				if !seen[script] {
					scripts = append(scripts, script)
					seen[script] = true
				}
			}
		}
	}
	var tokenCSS, componentCSS strings.Builder
	for _, id := range themeFiles {
		tokenCSS.Write(server.themeBundles[id])
		tokenCSS.WriteByte('\n')
	}
	for _, file := range cssFiles {
		data, e := readStaticFile("static/" + file)
		if e != nil {
			return nil, "", e
		}
		componentCSS.Write(data)
		componentCSS.WriteByte('\n')
	}
	files["tokens.css"] = []byte(tokenCSS.String())
	files["components.css"] = []byte(componentCSS.String())
	// Sample save/discard actions belong to the example, never the Forge dialog controller.
	if assets.modules["dialog"] {
		scripts = append(scripts, "dialog-demo.js")
	}
	var scriptHTML strings.Builder
	for _, script := range scripts {
		data, e := readStaticFile("static/" + script)
		if e != nil {
			return nil, "", e
		}
		files[script] = data
		fmt.Fprintf(&scriptHTML, "<script src=%q defer></script>\n", script)
	}
	if seen["loading.js"] {
		data, e := readStaticFile("static/loading-shapes.json")
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
				data, err := readStaticFile("static/" + relative)
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
