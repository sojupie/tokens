package usage

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"
)

func TestSVGCountsAndEscaping(t *testing.T) {
	summary := Summary{From: time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), Until: time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC), Requests: 1234, Prefetches: 20, Lost: 2, Kinds: map[string]int64{"navigation": 1000, "download": 234}, Subjects: map[string]int64{`</text><script>alert(1)</script>`: 1}, MCP: map[string]int64{"tools/call / get_tokens / successful": 3, "tools/call / get_tokens / failed": 2, "tools/list / successful": 10}}
	document := summary.SVG()
	decoder := xml.NewDecoder(strings.NewReader(document))
	var texts []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if start, ok := token.(xml.StartElement); ok {
			if start.Name.Local == "script" {
				t.Fatal("injected SVG script")
			}
			if start.Name.Local == "text" {
				var value string
				if err := decoder.DecodeElement(&value, &start); err != nil {
					t.Fatal(err)
				}
				texts = append(texts, value)
			}
		}
	}
	joined := strings.Join(texts, "\n")
	if !strings.Contains(joined, "1,234") || !strings.Contains(joined, "Successful MCP tool calls\n3") || !strings.Contains(joined, "Incomplete totals: 2 records lost.") {
		t.Fatal(joined)
	}
	if !strings.Contains((Summary{}).SVG(), "No requests recorded") {
		t.Fatal("empty report is misleading")
	}
}
