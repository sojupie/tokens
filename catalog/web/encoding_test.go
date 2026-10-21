package web

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageEncodingNegotiation(t *testing.T) {
	body := strings.Repeat("Material component preview and token values. ", 100)
	for _, test := range []struct {
		accept     string
		compressed bool
	}{{"", false}, {"gzip, deflate, br", true}, {"gzip;q=0", false}, {"*;q=1, gzip;q=0", false}, {"gzip;q=0.5", true}} {
		r := httptest.NewRequest("GET", "/components", nil)
		r.Header.Set("Accept-Encoding", test.accept)
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "HX-Request, Accept")
		if _, err := writePageBody(w, r, []byte(body)); err != nil {
			t.Fatal(err)
		}
		result := w.Body.String()
		if test.compressed {
			reader, err := gzip.NewReader(w.Body)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			reader.Close()
			result = string(decoded)
			if w.Header().Get("Content-Encoding") != "gzip" {
				t.Fatal("missing encoding")
			}
		} else if w.Header().Get("Content-Encoding") != "" {
			t.Fatal("compressed an excluded representation")
		}
		if result != body {
			t.Fatal("encoding changed content")
		}
		if strings.Join(w.Header().Values("Vary"), ",") != "HX-Request, Accept,Accept-Encoding" {
			t.Fatal("lost representation negotiation")
		}
	}
}
func TestEncodedAssetsRetainConditionalAndRangeRequests(t *testing.T) {
	body := []byte(strings.Repeat(".md-button { color: purple; }", 100))
	get := func(method, etag, byteRange string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/static/test.css", nil)
		r.Header.Set("Accept-Encoding", "gzip")
		if etag != "" {
			r.Header.Set("If-None-Match", etag)
		}
		if byteRange != "" {
			r.Header.Set("Range", byteRange)
		}
		w := httptest.NewRecorder()
		serveCachedAsset(w, r, "test.css", body)
		return w
	}
	full := get("GET", "", "")
	if full.Code != 200 || full.Header().Get("Content-Encoding") != "gzip" || full.Body.Len() >= len(body) {
		t.Fatal("asset was not compressed")
	}
	cached := get("GET", full.Header().Get("ETag"), "")
	if cached.Code != 304 || cached.Body.Len() != 0 {
		t.Fatal("conditional request changed")
	}
	head := get("HEAD", "", "")
	if head.Code != 200 || head.Body.Len() != 0 || head.Header().Get("Content-Length") != full.Header().Get("Content-Length") {
		t.Fatal("HEAD differs from GET representation")
	}
	partial := get("GET", "", "bytes=0-3")
	if partial.Code != 206 || partial.Header().Get("Content-Encoding") != "" || partial.Body.String() != string(body[:4]) {
		t.Fatal("range request changed")
	}
}
