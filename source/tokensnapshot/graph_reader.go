package tokensnapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sojupie/tokens/source/crawl"
)

// Decode array entries separately: Decoder.Decode on the top-level graph
// buffers the whole document, even when its input is a gzip reader.
func decodeGraph(reader io.Reader, validate, retain bool) (Graph, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return Graph{}, fmt.Errorf("graph must be a JSON object: %v", err)
	}
	header := make(map[string]any)
	var graph Graph
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return Graph{}, err
		}
		name := token.(string)
		if _, exists := header[name]; exists {
			return Graph{}, fmt.Errorf("duplicate graph field %s", name)
		}
		switch name {
		case "pages", "observations", "sourcePayloads", "records":
			token, err := decoder.Token()
			if err != nil || token != json.Delim('[') {
				return Graph{}, fmt.Errorf("graph %s must be an array", name)
			}
			header[name] = []any{}
			if retain {
				switch name {
				case "pages":
					graph.Pages = []string{}
				case "observations":
					graph.Observations = []crawl.Observation{}
				case "sourcePayloads":
					graph.SourcePayloads = []GraphPayload{}
				case "records":
					graph.Records = []ExtractedRecord{}
				}
			}
			var schema *jsonschema.Schema
			if validate {
				schema, err = compileJSONSchema("generated-token-graph.schema.json", "#/properties/"+name+"/items")
				if err != nil {
					return Graph{}, err
				}
			}
			for index := 0; decoder.More(); index++ {
				var body json.RawMessage
				if err := decoder.Decode(&body); err != nil {
					return Graph{}, err
				}
				if validate {
					value, err := graphItemValue(name, body)
					if err != nil {
						return Graph{}, err
					}
					if err := schema.Validate(value); err != nil {
						return Graph{}, fmt.Errorf("graph %s[%d]: %w", name, index, err)
					}
				}
				if retain {
					// Preserve raw JSON values, not formatting from the snapshot.
					var compact bytes.Buffer
					if err := json.Compact(&compact, body); err != nil {
						return Graph{}, err
					}
					if err := appendGraphItem(&graph, name, compact.Bytes()); err != nil {
						return Graph{}, err
					}
				}
			}
			if _, err := decoder.Token(); err != nil {
				return Graph{}, err
			}
		default:
			var value any
			if err := decoder.Decode(&value); err != nil {
				return Graph{}, err
			}
			header[name] = value
		}
	}
	if _, err := decoder.Token(); err != nil {
		return Graph{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("trailing JSON after graph")
		}
		return Graph{}, err
	}
	// Array item rules were checked above. Empty arrays let the root schema
	// check required fields and scalar rules without keeping another graph.
	body, err := json.Marshal(header)
	if err != nil {
		return Graph{}, err
	}
	if validate {
		if err := validateJSONSchema("generated-token-graph.schema.json", body); err != nil {
			return Graph{}, err
		}
	}
	var metadata Graph
	if err := json.Unmarshal(body, &metadata); err != nil {
		return Graph{}, err
	}
	graph.SchemaVersion, graph.SnapshotID, graph.CrawlID = metadata.SchemaVersion, metadata.SnapshotID, metadata.CrawlID
	return graph, nil
}

func graphItemValue(name string, body []byte) (any, error) {
	if name == "records" {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil {
			return nil, err
		}
		if fields == nil {
			return nil, nil
		}
		// The schema deliberately imposes no constraints on upstream raw JSON.
		// Keep its presence, but do not expand it into a second object tree.
		if _, exists := fields["raw"]; exists {
			fields["raw"] = json.RawMessage("null")
		}
		var err error
		body, err = json.Marshal(fields)
		if err != nil {
			return nil, err
		}
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	err := decoder.Decode(&value)
	return value, err
}

func appendGraphItem(graph *Graph, name string, body []byte) error {
	switch name {
	case "pages":
		var value string
		if err := json.Unmarshal(body, &value); err != nil {
			return err
		}
		graph.Pages = append(graph.Pages, value)
	case "observations":
		var value crawl.Observation
		if err := json.Unmarshal(body, &value); err != nil {
			return err
		}
		graph.Observations = append(graph.Observations, value)
	case "sourcePayloads":
		var value GraphPayload
		if err := json.Unmarshal(body, &value); err != nil {
			return err
		}
		graph.SourcePayloads = append(graph.SourcePayloads, value)
	case "records":
		var value ExtractedRecord
		if err := json.Unmarshal(body, &value); err != nil {
			return err
		}
		graph.Records = append(graph.Records, value)
	}
	return nil
}
