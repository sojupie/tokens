// Comparison formatting turns semantic differences into readable CLI output without changing the comparison result.
package main

import (
	"encoding/json"

	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sojupie/tokens/source/tokensnapshot"
)

func writeSnapshotCompareReport(writer io.Writer, diff tokensnapshot.SnapshotDiff, verbose bool) {
	fmt.Fprintf(writer, "Compared %s to %s\n", diff.FromSnapshotID, diff.ToSnapshotID)
	fmt.Fprintf(writer, "Material changes: +%d -%d changed %d\n", len(diff.Semantic.Added), len(diff.Semantic.Removed), len(diff.Semantic.Changed))
	fmt.Fprintf(writer, "Source payloads refreshed: %d\n", len(diff.SourcePayloads.Changed))
	fmt.Fprintf(writer, "Pages with changed results: %d\n", len(diff.PageResults.Changed))
	if !verbose {
		return
	}
	printSemanticChanges(writer, diff.Semantic)
}

func printSemanticChanges(writer io.Writer, diff tokensnapshot.SemanticDiff) {
	printSemanticChangeSet(writer, "Added", diff.Added)
	printSemanticChangeSet(writer, "Removed", diff.Removed)
	printSemanticChangeSet(writer, "Changed", diff.Changed)
}

func printSemanticChangeSet(writer io.Writer, label string, changes []tokensnapshot.SemanticChange) {
	if len(changes) == 0 {
		return
	}
	fmt.Fprintf(writer, "%s Material changes (%d):\n", label, len(changes))
	currentSubject := ""
	for _, change := range changes {
		if change.SourceSubject != currentSubject {
			currentSubject = change.SourceSubject
			fmt.Fprintf(writer, "  %s:\n", valueOrDash(currentSubject))
		}
		name := change.Name
		if name == "" {
			name = change.Kind
		}
		kind := semanticKindLabel(change.Kind)
		symbol := map[string]string{"Added": "+", "Removed": "-", "Changed": "~"}[label]
		fmt.Fprintf(writer, "    %s %s %s:\n", symbol, kind, name)
		for _, line := range semanticChangeLines(change, label) {
			fmt.Fprintf(writer, "      %s\n", line)
		}
	}
}

func semanticChangeLines(change tokensnapshot.SemanticChange, label string) []string {
	if label == "Added" {
		return semanticValueLines(change.After)
	}
	if label == "Removed" {
		return semanticValueLines(change.Before)
	}
	return semanticDiffLines(change.Before, change.After)
}

func semanticValueLines(value string) []string {
	if value == "" {
		return []string{"-"}
	}
	if value == `{"undefined":true}` {
		return []string{"undefined"}
	}
	leaves := flattenJSON(value)
	if len(leaves) == 0 {
		return []string{value}
	}
	lines := make([]string, 0, len(leaves))
	for _, key := range sortedKeys(leaves) {
		lines = append(lines, key+": "+leaves[key])
	}
	return lines
}

func semanticDiffLines(before, after string) []string {
	oldLeaves := flattenJSON(before)
	newLeaves := flattenJSON(after)
	keys := make(map[string]string, len(oldLeaves)+len(newLeaves))
	for key := range oldLeaves {
		keys[key] = ""
	}
	for key := range newLeaves {
		keys[key] = ""
	}
	lines := make([]string, 0)
	for _, key := range sortedKeys(keys) {
		oldValue, newValue := oldLeaves[key], newLeaves[key]
		if oldValue == newValue {
			continue
		}
		if oldValue == "" {
			oldValue = "-"
		}
		if newValue == "" {
			newValue = "-"
		}
		lines = append(lines, key+": "+oldValue+" -> "+newValue)
	}
	if len(lines) == 0 {
		return []string{valueOrDash(before) + " -> " + valueOrDash(after)}
	}
	return lines
}

func semanticKindLabel(kind string) string {
	switch kind {
	case "contextualReferenceTree":
		return "resolved value"
	case "component":
		return "component"
	case "tokenSet":
		return "token set"
	default:
		return kind
	}
}

func flattenJSON(value string) map[string]string {
	result := map[string]string{}
	var raw interface{}
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()
	if decoder.Decode(&raw) != nil {
		return result
	}
	flattenJSONValue(raw, "", result)
	return result
}

func flattenJSONValue(value interface{}, prefix string, result map[string]string) {
	switch value := value.(type) {
	case map[string]interface{}:
		if len(value) == 0 {
			result[prefixOrValue(prefix)] = "{}"
			return
		}
		if measurement, ok := value["value"]; ok {
			if unit, hasUnit := value["unit"]; hasUnit {
				result[prefixOrValue(prefix)] = scalarText(measurement) + " " + scalarText(unit)
				return
			}
			if tag, hasTag := value["tag"]; hasTag {
				result[prefixOrValue(prefix)] = scalarText(measurement) + " (" + scalarText(tag) + ")"
				return
			}
		}
		for key, child := range value {
			childPrefix := key
			if prefix != "" {
				childPrefix = prefix + "." + key
			}
			flattenJSONValue(child, childPrefix, result)
		}
	case []interface{}:
		body, _ := json.Marshal(value)
		result[prefixOrValue(prefix)] = string(body)
	default:
		result[prefixOrValue(prefix)] = fmt.Sprint(value)
	}
}

func scalarText(value interface{}) string {
	return fmt.Sprint(value)
}

func prefixOrValue(prefix string) string {
	if prefix == "" {
		return "value"
	}
	return prefix
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func writeComparisonReport(path, report string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create comparison report directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
		return fmt.Errorf("write comparison report: %w", err)
	}
	return nil
}
