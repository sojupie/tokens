package web

import (
	"bytes"
	"compress/gzip"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

func acceptsGzip(r *http.Request) bool {
	explicit, wildcard := -1.0, -1.0
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		name, params, err := mime.ParseMediaType(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		q := 1.0
		if value, ok := params["q"]; ok {
			q, err = strconv.ParseFloat(value, 64)
			if err != nil || q < 0 || q > 1 {
				continue
			}
		}
		if strings.EqualFold(name, "gzip") {
			explicit = max(explicit, q)
		} else if name == "*" {
			wildcard = max(wildcard, q)
		}
	}
	if explicit >= 0 {
		return explicit > 0
	}
	return wildcard > 0
}
func addEncodingVary(h http.Header) {
	for _, value := range h.Values("Vary") {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), "Accept-Encoding") {
				return
			}
		}
	}
	h.Add("Vary", "Accept-Encoding")
}
func gzipBytes(body []byte) ([]byte, error) {
	var output bytes.Buffer
	writer := gzip.NewWriter(&output)
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
func writePageBody(w http.ResponseWriter, r *http.Request, body []byte) (int, error) {
	addEncodingVary(w.Header())
	if acceptsGzip(r) {
		encoded, err := gzipBytes(body)
		if err != nil {
			return 0, err
		}
		body = encoded
		w.Header().Set("Content-Encoding", "gzip")
	}
	return w.Write(body)
}
func compressibleAsset(name string) bool {
	return strings.HasSuffix(name, ".css") || strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".svg")
}
