package web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
)

// Context labels describe individual captured values, not a global matrix.
// Include referenced tokens because their settings can affect the exported CSS.
func (server *Server) exportContextChoices(query url.Values) (map[string][]string, error) {
	names, defaults, _, err := server.selectedExport(query)
	if err != nil {
		return nil, err
	}
	tokens := map[string][]publish.TokenView{}
	known := map[string]map[string]bool{}
	for _, subject := range server.state.Subjects {
		for _, token := range subject.Tokens {
			tokens[token.SourceName] = append(tokens[token.SourceName], token)
			for _, value := range token.Values {
				for _, label := range value.ContextLabels {
					key, val, ok := strings.Cut(label, ":")
					if !ok {
						continue
					}
					key, val = strings.TrimSpace(key), strings.TrimSpace(val)
					if known[key] == nil {
						known[key] = map[string]bool{}
					}
					known[key][val] = true
				}
			}
		}
	}
	for key, values := range query {
		if !strings.HasPrefix(key, "context.") {
			continue
		}
		name := strings.TrimPrefix(key, "context.")
		for _, value := range values {
			if options, ok := known[name]; ok {
				if !options[value] {
					return nil, fmt.Errorf("unknown %s context %q", name, value)
				}
			} else if expected, ok := dtcg.DefaultContext()[name]; !ok || expected != value {
				return nil, fmt.Errorf("unknown context %q", name)
			}
		}
	}
	queue := []string{}
	for name := range names {
		queue = append(queue, name)
	}
	for i := 0; i < len(queue); i++ {
		for _, token := range tokens[queue[i]] {
			for _, value := range token.Values {
				// References in another platform's value are not dependencies
				// of this selection. Include every selected theme/contrast.
				compatible := true
				for _, label := range value.ContextLabels {
					key, required, ok := strings.Cut(label, ":")
					if !ok {
						continue
					}
					key, required = strings.TrimSpace(key), strings.TrimSpace(required)
					selected := query["context."+key]
					if len(selected) == 0 {
						selected = []string{defaults[key]}
					}
					match := false
					for _, candidate := range selected {
						if candidate == required {
							match = true
							break
						}
					}
					if !match {
						compatible = false
						break
					}
				}
				if !compatible {
					continue
				}
				for _, reference := range value.References {
					if !names[reference.Name] {
						names[reference.Name] = true
						queue = append(queue, reference.Name)
					}
				}
			}
		}
	}
	choices := map[string]map[string]bool{}
	for name := range names {
		for _, token := range tokens[name] {
			for _, value := range token.Values {
				var object map[string]json.RawMessage
				if json.Unmarshal(value.Value, &object) != nil || object["undefined"] != nil {
					continue
				}
				labels := map[string]string{}
				for _, label := range value.ContextLabels {
					k, v, ok := strings.Cut(label, ":")
					if ok {
						labels[strings.TrimSpace(k)] = strings.TrimSpace(v)
					}
				}
				for dimension, candidate := range labels {
					compatible := true
					for other, required := range labels {
						if other == dimension {
							continue
						}
						selected := query["context."+other]
						if len(selected) == 0 {
							selected = []string{defaults[other]}
						}
						match := false
						for _, selectedValue := range selected {
							if selectedValue == required {
								match = true
								break
							}
						}
						if !match {
							compatible = false
							break
						}
					}
					if compatible {
						if choices[dimension] == nil {
							choices[dimension] = map[string]bool{}
						}
						choices[dimension][candidate] = true
					}
				}
			}
		}
	}
	result := map[string][]string{}
	for name, values := range choices {
		for value := range values {
			result[name] = append(result[name], value)
		}
		sort.Strings(result[name])
	}
	return result, nil
}

func (server *Server) activeExportQuery(query url.Values) (url.Values, error) {
	choices, err := server.exportContextChoices(query)
	if err != nil {
		return nil, err
	}
	query = cloneExportQuery(query)
	for key := range query {
		if strings.HasPrefix(key, "context.") && len(choices[strings.TrimPrefix(key, "context.")]) == 0 {
			query.Del(key)
		}
	}
	return query, nil
}
