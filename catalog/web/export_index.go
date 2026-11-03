package web

import (
	"encoding/json"
	"github.com/sojupie/tokens/source/publish"
	"strings"
)

// Built once from the immutable publication. Request handlers never parse captured
// values or rediscover the collection's context vocabulary.
type exportIndex struct {
	tokens  map[string][]indexedExportValue
	choices map[string]map[string]bool
	names   map[string]bool
}
type indexedExportValue struct {
	labels     map[string]string
	references []publish.TokenReference
	defined    bool
}

func (server *Server) exportIndex() *exportIndex {
	server.exportIndexOnce.Do(func() {
		index := &exportIndex{tokens: map[string][]indexedExportValue{}, choices: map[string]map[string]bool{}, names: map[string]bool{}}
		for _, subject := range server.state.Subjects {
			for _, token := range subject.Tokens {
				index.names[token.SourceName] = true
				for _, value := range token.Values {
					var object map[string]json.RawMessage
					entry := indexedExportValue{labels: map[string]string{}, references: value.References}
					entry.defined = json.Unmarshal(value.Value, &object) == nil && object["undefined"] == nil
					for _, label := range value.ContextLabels {
						name, choice, ok := strings.Cut(label, ":")
						if !ok {
							continue
						}
						name, choice = strings.TrimSpace(name), strings.TrimSpace(choice)
						entry.labels[name] = choice
						if index.choices[name] == nil {
							index.choices[name] = map[string]bool{}
						}
						index.choices[name][choice] = true
					}
					index.tokens[token.SourceName] = append(index.tokens[token.SourceName], entry)
				}
			}
		}
		server.exportSelectionIndex = index
	})
	return server.exportSelectionIndex
}
