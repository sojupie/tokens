package dtcg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

func DefaultContext() map[string]string {
	return map[string]string{"Theme": "Light", "Contrast": "Default", "Platform": "Web", "Audience": "3P", "Font": "Static", "Motion scheme": "Standard", "Language height": "Small", "iOS weight setting": "Default", "iOS dark theme elevation": "None"}
}

type Omission struct {
	Name        string          `json:"name"`
	Reason      string          `json:"reason"`
	SourceValue json.RawMessage `json:"sourceValue,omitempty"`
}

type Report struct {
	PublicationID string            `json:"publicationId"`
	SnapshotID    string            `json:"snapshotId"`
	Context       map[string]string `json:"context"`
	Exported      int               `json:"exported"`
	MappedNames   int               `json:"mappedNames"`
	ExtensionOnly []string          `json:"extensionOnly"`
	Aliases       int               `json:"aliases"`
	Omitted       []Omission        `json:"omitted"`
}

// Export produces one complete, selected context. Unavailable values and values
// without standard representations are listed explicitly in the report.
func Export(state publish.PublishedState, context map[string]string) (Node, Report, error) {
	report := Report{PublicationID: state.PublicationID, SnapshotID: state.CurrentSnapshot.ID, Context: context, Omitted: []Omission{}}
	options := map[string]map[string]bool{}
	var tokens []publish.TokenView
	keys := make([]publish.SubjectKey, 0, len(state.Subjects))
	for key := range state.Subjects {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, key := range keys {
		tokens = append(tokens, state.Subjects[key].Tokens...)
	}
	for _, t := range tokens {
		for _, v := range t.Values {
			for _, label := range v.ContextLabels {
				k, v, ok := strings.Cut(label, ":")
				if !ok {
					continue
				}
				k = strings.TrimSpace(k)
				v = strings.TrimSpace(v)
				if options[k] == nil {
					options[k] = map[string]bool{}
				}
				options[k][v] = true
			}
		}
	}
	for key, value := range context {
		if _, exists := options[key]; !exists {
			if _, known := DefaultContext()[key]; !known {
				return nil, report, fmt.Errorf("unknown context dimension %q", key)
			}
		}
		if choices, exists := options[key]; exists && !choices[value] {
			return nil, report, fmt.Errorf("unknown %s context %q", key, value)
		}
	}
	values := map[string]Node{}
	inferred := map[string]Node{}
	conflicting := map[string]bool{}
	seen := map[string]bool{}
	for _, t := range tokens {
		v, ok, err := Select(t.Values, context)
		if err != nil {
			return nil, report, fmt.Errorf("%s: %w", t.SourceName, err)
		}
		if !ok {
			if !seen[t.SourceName] {
				report.Omitted = append(report.Omitted, Omission{Name: t.SourceName, Reason: "no defined value in selected context"})
			}
			seen[t.SourceName] = true
			continue
		}
		n, err := Convert(v.Value)
		if err != nil {
			if !seen[t.SourceName] {
				report.Omitted = append(report.Omitted, Omission{Name: t.SourceName, Reason: err.Error(), SourceValue: v.Value})
			}
			seen[t.SourceName] = true
			continue
		}
		if previous, exists := values[t.SourceName]; exists {
			if !equalValue(previous, n) {
				return nil, report, fmt.Errorf("conflicting values for %s in selected context", t.SourceName)
			}
			continue
		}
		seen[t.SourceName] = true
		values[t.SourceName] = n
		// Inference is limited to a single primitive chain, whose resolved value is
		// identical at each step. Branching composite references retain metadata.
		if _, primitive := n["$value"]; primitive && n["$type"] != "typography" {
			refs := v.References
			for len(refs) == 1 {
				ref := refs[0]
				if ref.Name == t.SourceName || ref.Name == "" {
					break
				}
				inferredNode := clone(n)
				if old, exists := inferred[ref.Name]; exists && !equalValue(old, n) {
					conflicting[ref.Name] = true
				}
				inferred[ref.Name] = inferredNode
				refs = ref.References
			}
		}
		meta := n["$extensions"].(map[string]any)[Extension].(map[string]any)
		meta["sourceName"] = t.SourceName
		meta["references"] = v.References
		meta["presentation"] = t.Presentation
		if t.DisplayName != "" {
			n["$description"] = t.DisplayName
		}
		if strings.EqualFold(t.Lifecycle, "deprecated") {
			n["$deprecated"] = true
		}
	}
	for name, node := range inferred {
		if _, exists := values[name]; !exists && !conflicting[name] {
			values[name] = node
			node["$extensions"].(map[string]any)[Extension].(map[string]any)["inferredFromReference"] = true
		}
	}
	// Preserve aliases only when the target is present with the same type/value.
	// The inferred dependencies remain resolved leaves, preventing inferred cycles.
	for _, t := range tokens {
		node, exists := values[t.SourceName]
		if !exists {
			continue
		}
		if _, exists := node["$value"]; !exists || node["$type"] == "typography" {
			continue
		}
		v, ok, _ := Select(t.Values, context)
		if !ok || len(v.References) != 1 {
			continue
		}
		target := v.References[0].Name
		other, exists := values[target]
		if !exists || target == t.SourceName || !equalValue(node, other) {
			continue
		}
		// Only point to a resolved target; a later pass cannot introduce a cycle.
		if text, ok := other["$value"].(string); ok && strings.HasPrefix(text, "{") {
			continue
		}
		node["$value"] = "{" + target + "}"
		report.Aliases++
	}
	root := Node{}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := insert(root, strings.Split(name, "."), values[name]); err != nil {
			return nil, report, err
		}
	}
	report.MappedNames = len(values)
	for _, name := range names {
		n := values[name]
		if _, hasValue := n["$value"]; !hasValue && n["topLeft"] == nil {
			report.ExtensionOnly = append(report.ExtensionOnly, name)
		}
	}
	report.Exported = countTokens(root)
	sort.Slice(report.Omitted, func(i, j int) bool { return report.Omitted[i].Name < report.Omitted[j].Name })
	root["$extensions"] = map[string]any{"dev.tokencatalog.export": map[string]any{"formatVersion": "2025.10", "publicationId": state.PublicationID, "snapshotId": state.CurrentSnapshot.ID, "context": context}}
	return root, report, nil
}

// Select requires a choice for every dimension relevant to a value. Unspecified
// dimensions never cause an arbitrary platform or font to be chosen.
func Select(values []publish.TokenValueView, context map[string]string) (publish.TokenValueView, bool, error) {
	var selected publish.TokenValueView
	found := false
	best := -1
	for _, v := range values {
		var undefined map[string]json.RawMessage
		if json.Unmarshal(v.Value, &undefined) != nil {
			continue
		}
		if _, ok := undefined["undefined"]; ok {
			continue
		}
		compatible := true
		matches := 0
		for _, label := range v.ContextLabels {
			key, value, ok := strings.Cut(label, ":")
			if !ok {
				return selected, false, fmt.Errorf("invalid context label %q", label)
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			actual, chosen := context[key]
			if !chosen {
				return selected, false, fmt.Errorf("select a value for context %q", key)
			}
			if actual != value {
				compatible = false
				break
			}
			matches++
		}
		if !compatible {
			continue
		}
		if found && matches == best && !bytes.Equal(selected.Value, v.Value) {
			return selected, false, fmt.Errorf("ambiguous values in selected context")
		}
		if !found || matches > best {
			selected = v
			found = true
			best = matches
		}
	}
	return selected, found, nil
}

func equalValue(a, b Node) bool {
	av, _ := json.Marshal([]any{a["$type"], a["$value"], a["topLeft"], a["topRight"], a["bottomRight"], a["bottomLeft"]})
	bv, _ := json.Marshal([]any{b["$type"], b["$value"], b["topLeft"], b["topRight"], b["bottomRight"], b["bottomLeft"]})
	var aValue, bValue any
	_ = json.Unmarshal(av, &aValue)
	_ = json.Unmarshal(bv, &bValue)
	av, _ = json.Marshal(aValue)
	bv, _ = json.Marshal(bValue)
	return bytes.Equal(av, bv)
}
func clone(node Node) Node {
	data, _ := json.Marshal(node)
	var n Node
	_ = json.Unmarshal(data, &n)
	return n
}
func insert(root Node, path []string, node Node) error {
	key := path[0]
	if key == "" || strings.HasPrefix(key, "$") || strings.ContainsAny(key, "{}") {
		return fmt.Errorf("invalid DTCG name segment %q", key)
	}
	existing, ok := root[key].(Node)
	if len(path) == 1 {
		if ok {
			if _, leaf := existing["$value"]; leaf {
				return fmt.Errorf("duplicate token path %s", key)
			}
			existing["$root"] = node
		} else {
			root[key] = node
		}
		return nil
	}
	if !ok {
		existing = Node{}
		root[key] = existing
	}
	if _, leaf := existing["$value"]; leaf {
		existing = Node{"$root": existing}
		root[key] = existing
	}
	return insert(existing, path[1:], node)
}

func countTokens(node Node) int {
	count := 0
	if _, ok := node["$value"]; ok {
		count++
	}
	for key, value := range node {
		if !strings.HasPrefix(key, "$") || key == "$root" {
			if child, ok := value.(Node); ok {
				count += countTokens(child)
			}
		}
	}
	return count
}
