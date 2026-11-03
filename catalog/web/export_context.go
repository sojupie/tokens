package web

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/dtcg"
)

// Context labels describe individual captured values, not a global matrix.
// Include referenced tokens because their settings can affect the exported CSS.
func (server *Server) exportContextChoices(query url.Values) (map[string][]string, error) {
	names, defaults, _, err := server.selectedExport(query)
	if err != nil {
		return nil, err
	}
	index := server.exportIndex()
	known := index.choices
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
		for _, value := range index.tokens[queue[i]] {
			// References in another platform's value are not dependencies
			// of this selection. Include every selected theme/contrast.
			compatible := true
			for key, required := range value.labels {
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
			for _, reference := range value.references {
				if !names[reference.Name] {
					names[reference.Name] = true
					queue = append(queue, reference.Name)
				}
			}
		}
	}
	choices := map[string]map[string]bool{}
	for name := range names {
		for _, value := range index.tokens[name] {
			if !value.defined {
				continue
			}
			labels := value.labels
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
