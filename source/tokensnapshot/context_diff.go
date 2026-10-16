package tokensnapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// ContextualValue retains the compared snapshot's labels and dependencies.
type ContextualValue struct {
	Context            []string         `json:"context"`
	UnknownContextTags []string         `json:"unknownContextTags,omitempty"`
	Default            bool             `json:"default"`
	Value              json.RawMessage  `json:"value"`
	References         []NamedReference `json:"references,omitempty"`
}

type NamedReference struct {
	Name       string           `json:"name"`
	References []NamedReference `json:"references,omitempty"`
}

type comparisonReference struct {
	TokenName  string                `json:"tokenName"`
	ChildNodes []comparisonReference `json:"childNodes"`
}

func comparisonReferences(nodes []comparisonReference) []NamedReference {
	var refs []NamedReference
	for _, node := range nodes {
		children := comparisonReferences(node.ChildNodes)
		if node.TokenName == "" {
			refs = append(refs, children...)
		} else {
			refs = append(refs, NamedReference{Name: node.TokenName, References: children})
		}
	}
	// Sibling order is presentation, not a dependency change.
	sort.Slice(refs, func(i, j int) bool {
		left, _ := json.Marshal(refs[i])
		right, _ := json.Marshal(refs[j])
		return string(left) < string(right)
	})
	return refs
}

func contextualSemanticRecords(record ExtractedRecord, contexts map[string]ResolvedContext) []SemanticRecord {
	var document struct {
		Entries []struct {
			ContextTags   []string            `json:"contextTags"`
			ReferenceTree comparisonReference `json:"referenceTree"`
			ResolvedValue json.RawMessage     `json:"resolvedValue"`
		} `json:"contextualReferenceTree"`
	}
	if json.Unmarshal(record.Raw, &document) != nil {
		return nil
	}
	var records []SemanticRecord
	for _, entry := range document.Entries {
		name := entry.ReferenceTree.TokenName
		if name == "" {
			continue
		}
		labels, isDefault := ResolveContext(entry.ContextTags, contexts)
		// Labels survive upstream tag-ID rotation and context array reordering.
		sort.Strings(labels)
		var unknown []string
		for _, tag := range entry.ContextTags {
			if _, ok := contexts[tag]; !ok {
				unknown = append(unknown, tag)
			}
		}
		sort.Strings(unknown)
		identity, _ := json.Marshal(struct {
			Labels  []string
			Unknown []string
		}{labels, unknown})
		value := entry.ResolvedValue
		if len(value) == 0 {
			value = json.RawMessage(`null`)
		}
		var decoded any
		decoder := json.NewDecoder(bytes.NewReader(value))
		decoder.UseNumber()
		if decoder.Decode(&decoded) != nil {
			continue
		}
		value, _ = json.Marshal(decoded)
		body, _ := json.Marshal(ContextualValue{Context: labels, UnknownContextTags: unknown, Default: isDefault, Value: value, References: comparisonReferences(entry.ReferenceTree.ChildNodes)})
		records = append(records, SemanticRecord{
			Key:  fmt.Sprintf("%s\x00%s\x00%s", record.Kind, name, identity),
			Kind: record.Kind, Name: name, Value: string(body),
			SourcePayloadKind: record.SourcePayloadKind,
			SourceSubject:     semanticSubject(name, record.SourceSubject),
		})
	}
	return records
}
