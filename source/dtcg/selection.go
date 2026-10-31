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
	var walk func(Node, []string)
	walk = func(node Node, parts []string) {
		if ext, ok := node["$extensions"].(map[string]any); ok && ext[Extension] != nil {
			nodes[strings.Join(parts, ".")] = node
			return
		}
		for key, child := range node {
			if key == "$root" {
				if child, ok := child.(Node); ok {
					walk(child, parts)
				}
			} else if !strings.HasPrefix(key, "$") {
				if child, ok := child.(Node); ok {
					walk(child, append(append([]string{}, parts...), key))
				}
			}
		}
	}
	walk(root, nil)
	return nodes
}

// SelectNames retains requested roots and the alias targets needed to use them.
// It works on a complete context export so a dependency is never lost merely
// because its source token belongs to a different subject or table.
func SelectNames(root Node, selected map[string]bool) (Node, int, int) {
	nodes := NamedNodes(root)
	kept := map[string]bool{}
	var retain func(string)
	retain = func(name string) {
		node, exists := nodes[name]
		if !exists || kept[name] {
			return
		}
		kept[name] = true
		if alias, ok := node["$value"].(string); ok && strings.HasPrefix(alias, "{") && strings.HasSuffix(alias, "}") {
			retain(strings.TrimSuffix(strings.TrimPrefix(alias, "{"), "}"))
		}
	}
	for name := range selected {
		retain(name)
	}
	result := Node{"$extensions": root["$extensions"]}
	roots := 0
	for name := range kept {
		_ = insert(result, strings.Split(name, "."), nodes[name])
		if selected[name] {
			roots++
		}
	}
	return result, roots, len(kept) - roots
}

func TokenCount(root Node) int { return countTokens(root) }
