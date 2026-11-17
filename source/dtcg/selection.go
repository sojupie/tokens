package dtcg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/sojupie/tokens/source/publish"
)

// ValueKey identifies a captured value independently of its context or name.
func ValueKey(raw json.RawMessage) string {
	var value any
	canonical := []byte(raw)
	if json.Unmarshal(raw, &value) == nil {
		if data, err := json.Marshal(value); err == nil {
			canonical = data
		}
	}
	sum := sha256.Sum256(canonical)
	return "v" + hex.EncodeToString(sum[:])
}

// ValueCatalog lets the Node build compile every captured value once with the
// same Style Dictionary transforms used for complete CSS exports.
func ValueCatalog(state publish.PublishedState) Node {
	result := Node{}
	for _, subject := range state.Subjects {
		for _, token := range subject.Tokens {
			for _, value := range token.Values {
				key := ValueKey(value.Value)
				if _, exists := result[key]; exists {
					continue
				}
				if node, err := Convert(value.Value); err == nil {
					result[key] = node
				}
			}
		}
	}
	return result
}

// NamedNodes includes source token groups (such as shapes) as whole units.
func NamedNodes(root Node) map[string]Node {
	nodes := map[string]Node{}
	for name, entry := range sourceNodes(root) {
		nodes[name] = entry.node
	}
	return nodes
}

type sourceNode struct {
	node Node
	path []string
}

func sourceNodes(root Node) map[string]sourceNode {
	nodes := map[string]sourceNode{}
	var walk func(Node, []string)
	walk = func(node Node, parts []string) {
		if ext, ok := node["$extensions"].(map[string]any); ok && ext[Extension] != nil {
			name := strings.TrimSuffix(strings.Join(parts, "."), ".$root")
			nodes[name] = sourceNode{node, parts}
			return
		}
		for key, child := range node {
			if !strings.HasPrefix(key, "$") || key == "$root" {
				if child, ok := child.(Node); ok {
					walk(child, append(append([]string{}, parts...), key))
				}
			}
		}
	}
	walk(root, nil)
	return nodes
}

// AliasSourceName translates a DTCG reference back to the captured token name.
func AliasSourceName(node Node) (string, bool) {
	value, ok := node["$value"].(string)
	if !ok {
		return "", false
	}
	start, hasStart := strings.CutPrefix(value, "{")
	target, hasEnd := strings.CutSuffix(start, "}")
	return strings.TrimSuffix(target, ".$root"), hasStart && hasEnd
}

// SelectNames retains requested roots and the alias targets needed to use them.
// It works on a complete context export so a dependency is never lost merely
// because its source token belongs to a different subject or table.
func SelectNames(root Node, selected map[string]bool) (Node, int, int) {
	nodes := sourceNodes(root)
	kept := map[string]bool{}
	var retain func(string)
	retain = func(name string) {
		entry, exists := nodes[name]
		if !exists || kept[name] {
			return
		}
		kept[name] = true
		if target, ok := AliasSourceName(entry.node); ok {
			retain(target)
		}
	}
	for name := range selected {
		retain(name)
	}
	result := Node{"$extensions": root["$extensions"]}
	roots := 0
	for name := range kept {
		entry := nodes[name]
		parent := result
		for _, part := range entry.path[:len(entry.path)-1] {
			child, ok := parent[part].(Node)
			if !ok {
				child = Node{}
				parent[part] = child
			}
			parent = child
		}
		parent[entry.path[len(entry.path)-1]] = entry.node
		if selected[name] {
			roots++
		}
	}
	return result, roots, len(kept) - roots
}

func TokenCount(root Node) int { return countTokens(root) }
