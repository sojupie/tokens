package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestNativeDownloadPageUsesServerResolvedChoices(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	for _, path := range []string{"/download", "/download?content=tokens&scope=collection&format=source"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
		}
		body := response.Body.String()
		for _, want := range []string{`action="/download"`, `formaction="/downloads/export"`, `Apply choices`} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s missing %s", path, want)
			}
		}
		doc, err := html.Parse(strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		for name, value := range map[string]string{"format": "source", "scope": "collection"} {
			found, checked := false, false
			var visit func(*html.Node)
			visit = func(node *html.Node) {
				if node.Type == html.ElementNode && node.Data == "input" {
					attrs := make(map[string]string)
					for _, attr := range node.Attr {
						attrs[attr.Key] = attr.Val
					}
					if strings.HasPrefix(attrs["id"], "native-") && attrs["name"] == name && attrs["value"] == value {
						found = true
						_, checked = attrs["checked"]
					}
				}
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					visit(child)
				}
			}
			visit(doc)
			if !found {
				t.Fatalf("%s missing input %s=%s", path, name, value)
			}
			if name == "format" && strings.Contains(path, "format=source") && !checked {
				t.Fatal("server did not preserve the submitted format")
			}
		}
	}
}
