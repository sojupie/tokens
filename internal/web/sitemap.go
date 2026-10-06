package web

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type sitemapEntry struct {
	Location string `xml:"loc"`
}

type sitemapDocument struct {
	XMLName xml.Name       `xml:"urlset"`
	XMLNS   string         `xml:"xmlns,attr"`
	URLs    []sitemapEntry `xml:"url"`
}

func (server *Server) handleSitemap(response http.ResponseWriter, request *http.Request) {
	if server.publicOrigin == "" {
		http.Error(response, "Configure serve --public-origin to enable the sitemap.", http.StatusServiceUnavailable)
		return
	}
	document := sitemapDocument{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	add := func(path string) {
		document.URLs = append(document.URLs, sitemapEntry{Location: server.publicOrigin + path})
	}
	for _, path := range []string{"/", "/components", "/styles", "/changes"} {
		add(path)
	}
	for _, subject := range sortedSubjects(server.state) {
		add(subjectURL(subject))
		add(subjectChangesURL(subject))
		for _, set := range tokenSets(subject) {
			add(tokenSetURL(subject, set))
		}
	}
	data, err := xml.Marshal(document)
	if err != nil {
		http.Error(response, "Could not build sitemap.", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	response.Header().Set("Cache-Control", "no-cache")
	_, _ = response.Write(append([]byte(xml.Header), data...))
}

func (server *Server) handleRobots(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Header().Set("Cache-Control", "no-cache")
	if server.publicOrigin == "" {
		_, _ = fmt.Fprint(response, "User-agent: *\nDisallow: /\n")
		return
	}
	_, _ = fmt.Fprintf(response, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", server.publicOrigin)
}
