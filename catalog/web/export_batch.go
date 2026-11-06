package web

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

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
		// The manifest records exact values; folder names never include user input.
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
					// An explicitly requested report is already the selection's metadata.
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
