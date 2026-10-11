package usage

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Summary struct {
	From, Until                                        time.Time
	Started                                            string
	Requests, Errors, Prefetches, Lost                 int64
	Bytes                                              int64
	Duration, MaxDuration                              float64
	Kinds, Formats, Subjects, Downloads, MCP, Failures map[string]int64
}

func (s *Store) Report(from, until time.Time) (Summary, error) {
	result := Summary{From: from, Until: until, Kinds: map[string]int64{}, Formats: map[string]int64{}, Subjects: map[string]int64{}, Downloads: map[string]int64{}, MCP: map[string]int64{}, Failures: map[string]int64{}}
	_ = s.db.QueryRow("SELECT value FROM metadata WHERE key='started'").Scan(&result.Started)
	rows, err := s.db.Query(`SELECT event,route,subject,format,kind,scope,method,tool,status,success,count,bytes,duration,max_duration FROM totals WHERE hour>=? AND hour<?`, from.UTC().Truncate(time.Hour).Format(time.RFC3339), until.UTC().Format(time.RFC3339))
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var event, route, subject, format, kind, scope, method, tool string
		var status int
		var success bool
		var count, bytes int64
		var duration, maximum float64
		if err := rows.Scan(&event, &route, &subject, &format, &kind, &scope, &method, &tool, &status, &success, &count, &bytes, &duration, &maximum); err != nil {
			return result, err
		}
		switch event {
		case "usage_lost":
			result.Lost += count
		case "mcp_call":
			label := method
			if tool != "" {
				label += " / " + tool
			}
			if success {
				label += " / successful"
			} else {
				label += " / failed"
			}
			result.MCP[label] += count
		case "http_request":
			if kind == "prefetch" {
				result.Prefetches += count
				continue
			}
			result.Requests += count
			result.Bytes += bytes
			result.Duration += duration
			if maximum > result.MaxDuration {
				result.MaxDuration = maximum
			}
			if status >= 400 {
				result.Errors += count
				result.Failures[fmt.Sprintf("%s / HTTP %d", route, status)] += count
			}
			result.Kinds[kind] += count
			if format != "" {
				result.Formats[format] += count
			}
			if subject != "" && success {
				result.Subjects[subject] += count
			}
			if kind == "download" {
				label := subject
				if label == "" {
					label = "collection"
				}
				outcome := "successful"
				if !success {
					outcome = "failed"
				}
				result.Downloads[label+" / "+scope+" / "+outcome] += count
			}
		}
	}
	return result, rows.Err()
}

func (s Summary) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Token Catalog usage\n%s to %s (UTC; hourly totals)\n\n", s.From.UTC().Format("2006-01-02 15:04"), s.Until.UTC().Format("2006-01-02 15:04"))
	if s.Started != "" {
		fmt.Fprintf(&b, "Tracking started: %s\n", s.Started)
	}
	fmt.Fprintf(&b, "Requests excluding prefetches: %d\nHTTP errors: %d\nPrefetch requests: %d\nResponse bytes excluding prefetches: %d\n", s.Requests, s.Errors, s.Prefetches, s.Bytes)
	if s.Requests > 0 {
		fmt.Fprintf(&b, "Mean response time: %.2f ms\nSlowest response: %.2f ms\n", s.Duration/float64(s.Requests), s.MaxDuration)
	}
	if s.Lost > 0 {
		fmt.Fprintf(&b, "WARNING: %d usage records were lost. Totals are incomplete.\n", s.Lost)
	}
	for _, group := range []struct {
		name   string
		counts map[string]int64
	}{{"Request kinds", s.Kinds}, {"Response formats", s.Formats}, {"Popular subjects (successful requests)", s.Subjects}, {"Downloads", s.Downloads}, {"MCP methods and tools", s.MCP}, {"HTTP errors by route", s.Failures}} {
		fmt.Fprintf(&b, "\n%s\n", group.name)
		keys := make([]string, 0, len(group.counts))
		for key := range group.counts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if group.counts[keys[i]] == group.counts[keys[j]] {
				return keys[i] < keys[j]
			}
			return group.counts[keys[i]] > group.counts[keys[j]]
		})
		if len(keys) == 0 {
			b.WriteString("None recorded.\n")
		}
		for i, key := range keys {
			if i == 15 {
				fmt.Fprintf(&b, "%d additional categories omitted.\n", len(keys)-i)
				break
			}
			fmt.Fprintf(&b, "%d  %s\n", group.counts[key], key)
		}
	}
	b.WriteString("\nCounts measure requests, not unique visitors. Health checks and successful static-file requests are excluded. Browser and Cloudflare cache hits may not reach the server. IP addresses, query text, cookies and MCP arguments are not recorded.\n")
	return b.String()
}

func (s *Store) LastDigest() (time.Time, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM metadata WHERE key='last_digest'").Scan(&value)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, value)
}
func (s *Store) SaveDigest(until time.Time) error {
	_, err := s.db.Exec(`INSERT INTO metadata VALUES ('last_digest',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, until.UTC().Format(time.RFC3339))
	return err
}

// ClaimDigest prevents overlapping manual and scheduled sends; a crash leaves
// only a five-minute lease, so the next service retry can recover.
func (s *Store) ClaimDigest() (func(), error) {
	now := time.Now().UnixNano()
	lease := now + int64(5*time.Minute)
	result, err := s.db.Exec(`INSERT INTO metadata VALUES ('digest_lease',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value WHERE CAST(metadata.value AS INTEGER) < ?`, fmt.Sprint(lease), now)
	if err != nil {
		return nil, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if changed == 0 {
		return nil, fmt.Errorf("another usage digest is running")
	}
	return func() {
		_, _ = s.db.Exec("DELETE FROM metadata WHERE key='digest_lease' AND value=?", fmt.Sprint(lease))
	}, nil
}
