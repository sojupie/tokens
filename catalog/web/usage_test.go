package web

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sojupie/tokens/catalog/usage"
)

func TestUsagePrivacyAndMCPFailures(t *testing.T) {
	dir := t.TempDir()
	var logs bytes.Buffer
	store, err := usage.Open(dir, &logs)
	if err != nil {
		t.Fatal(err)
	}
	server := New(seoTestState(), "")
	if err := server.SetPublicOrigin("https://tokens.test/md3"); err != nil {
		t.Fatal(err)
	}
	server.SetUsage(store)
	handler := server.Handler()
	for _, path := range []string{"/md3/components?q=private-search", "/md3/api/v1/health", "/md3/static/prefetch.js", "/md3/unknown-secret"} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("Cookie", "private-cookie")
		handler.ServeHTTP(httptest.NewRecorder(), r)
	}
	r := httptest.NewRequest("GET", "/md3/components", nil)
	r.Header.Set("Sec-Purpose", "prefetch")
	handler.ServeHTTP(httptest.NewRecorder(), r)
	for _, body := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_tokens","arguments":{"kind":"component","slug":"private-missing-subject"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_subjects","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"private-method"}`,
	} {
		r := httptest.NewRequest("POST", "/md3/mcp", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json, text/event-stream")
		handler.ServeHTTP(httptest.NewRecorder(), r)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"private-search", "private-cookie", "private-missing-subject", "unknown-secret", "private-method"} {
		if strings.Contains(logs.String(), private) {
			t.Fatalf("logged private data %s: %s", private, logs.String())
		}
	}
	reopened, err := usage.Open(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	report, err := reopened.Report(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if report.Requests != 5 || report.Prefetches != 1 || report.MCP["tools/call / get_tokens / failed"] != 1 || report.MCP["tools/call / list_subjects / successful"] != 1 || report.MCP["unknown / failed"] != 1 {
		t.Fatalf("wrong usage: %+v", report)
	}
}
