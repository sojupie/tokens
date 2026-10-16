package usage

import (
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SVG renders only aggregate counts, with no scripts or external assets.
func (s Summary) SVG() string {
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 583" role="img" aria-labelledby="title desc"><title id="title">Token Catalog server usage</title><desc id="desc">Request totals, request types and popular subjects.</desc><style>
 text{font-family:system-ui,-apple-system,Segoe UI,sans-serif;fill:#211c2b} .muted{fill:#635d6d} .bg{fill:#faf8ff} .card{fill:#fff;stroke:#e7e0ef} .track{fill:#ece6f4} .bar{fill:#6750a4}
 @media(prefers-color-scheme:dark){text{fill:#eee7f5}.muted{fill:#c9bed7}.bg{fill:#17131e}.card{fill:#211b2a;stroke:#3b3248}.track{fill:#34283f}.bar{fill:#cbb2ff}}
 </style><rect class="bg" width="1000" height="583" rx="20"/>`)
	text := func(x, y, size int, css, value string) {
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-size="%d" class="%s">%s</text>`, x, y, size, css, html.EscapeString(value))
	}
	text(32, 48, 25, "", "How the catalog is used")
	text(32, 76, 13, "muted", s.From.UTC().Format("Jan 2, 2006 15:04")+" – "+s.Until.UTC().Format("Jan 2, 2006 15:04")+" UTC")
	calls := int64(0)
	for key, count := range s.MCP {
		if strings.HasPrefix(key, "tools/call / ") && strings.HasSuffix(key, " / successful") {
			calls += count
		}
	}
	cards := []struct {
		label string
		value int64
	}{{"Server requests", s.Requests}, {"Download requests", s.Kinds["download"]}, {"Markdown requests", s.Kinds["markdown"]}, {"Successful MCP tool calls", calls}}
	for i, card := range cards {
		x := 32 + i*240
		fmt.Fprintf(&b, `<rect class="card" x="%d" y="100" width="224" height="92" rx="12"/>`, x)
		text(x+16, 130, 12, "muted", card.label)
		text(x+16, 169, 30, "", displayCount(card.value))
	}
	chart := func(x int, title string, counts map[string]int64) {
		text(x, 230, 17, "", title)
		keys := make([]string, 0, len(counts))
		for key := range counts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if counts[keys[i]] == counts[keys[j]] {
				return keys[i] < keys[j]
			}
			return counts[keys[i]] > counts[keys[j]]
		})
		if len(keys) == 0 {
			text(x, 271, 13, "muted", "No requests recorded in this period.")
			return
		}
		maximum := counts[keys[0]]
		for i, key := range keys {
			if i == 6 {
				break
			}
			y := 252 + i*37
			label := key
			if len([]rune(label)) > 45 {
				label = string([]rune(label)[:42]) + "…"
			}
			text(x, y+11, 12, "muted", label)
			fmt.Fprintf(&b, `<rect class="track" x="%d" y="%d" width="370" height="8" rx="4"/>`, x, y+18)
			width := float64(0)
			if maximum > 0 {
				width = 370 * float64(counts[key]) / float64(maximum)
			}
			fmt.Fprintf(&b, `<rect class="bar" x="%d" y="%d" width="%.1f" height="8" rx="4"/>`, x, y+18, width)
			text(x+390, y+23, 12, "", displayCount(counts[key]))
		}
	}
	chart(32, "Request types · top six", s.Kinds)
	chart(526, "Popular subjects · successful requests", s.Subjects)
	text(32, 505, 12, "muted", fmt.Sprintf("Prefetches: %d (excluded above) · HTTP errors: %d", s.Prefetches, s.Errors))
	if s.Lost > 0 {
		text(32, 528, 12, "muted", fmt.Sprintf("Incomplete totals: %d records lost.", s.Lost))
	} else {
		text(32, 528, 12, "muted", "Browser/CDN cache hits may not reach the server. Updated weekly.")
	}
	if started, err := time.Parse(time.RFC3339, s.Started); err == nil {
		text(32, 551, 12, "muted", "Tracking since "+started.UTC().Format("Jan 2, 2006 15:04")+" UTC.")
	}
	b.WriteString("</svg>\n")
	return b.String()
}

func displayCount(value int64) string {
	text := strconv.FormatInt(value, 10)
	for i := len(text) - 3; i > 0; i -= 3 {
		text = text[:i] + "," + text[i:]
	}
	return text
}
