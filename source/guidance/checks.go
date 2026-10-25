package guidance

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sojupie/tokens/schemas"
)

type CheckSummary struct {
	DesignSystem    string
	Component       string
	Checks          int
	PendingSections int
	Removed         int
	Withdrawn       int
	Commits         []string
}

func parseChecks(data []byte) (map[string]any, error) {
	schemaData, err := schemas.Files.ReadFile("component-guidance.schema.json")
	if err != nil {
		return nil, err
	}
	var schemaDoc, doc any
	if err = json.Unmarshal(schemaData, &schemaDoc); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	const location = "https://github.com/sojupie/tokens/schemas/component-guidance.schema.json"
	if err = compiler.AddResource(location, schemaDoc); err != nil {
		return nil, err
	}
	schema, err := compiler.Compile(location)
	if err != nil {
		return nil, err
	}
	if err = schema.Validate(doc); err != nil {
		return nil, err
	}
	return doc.(map[string]any), nil
}

// ValidateChecks checks the curation contract, not whether evidence is true or a
// component conforms. Previous is optional; supply the committed file to enforce
// retained IDs and append-only source/fix history across edits.
func ValidateChecks(data, previous []byte) (CheckSummary, error) {
	var summary CheckSummary
	doc, err := parseChecks(data)
	if err != nil {
		return summary, err
	}
	summary.DesignSystem = doc["designSystem"].(string)
	summary.Component = doc["component"].(string)
	sections := map[string]string{}
	urls := map[string]bool{}
	for _, entry := range doc["sources"].([]any) {
		source := entry.(map[string]any)
		url := source["url"].(string)
		if urls[url] {
			return summary, fmt.Errorf("duplicate source %s", url)
		}
		urls[url] = true
		for _, entry := range source["sections"].([]any) {
			section := entry.(map[string]any)
			key := url + "#" + section["id"].(string)
			if _, exists := sections[key]; exists {
				return summary, fmt.Errorf("duplicate section %s", key)
			}
			sections[key] = section["status"].(string)
			if sections[key] == "pending" || sections[key] == "unclear" {
				summary.PendingSections++
			}
		}
	}
	current := map[string]map[string]any{}
	mapped := map[string]bool{}
	commits := map[string]bool{}
	for _, entry := range doc["checks"].([]any) {
		check := entry.(map[string]any)
		id := check["id"].(string)
		summary.Checks++
		fail := func(message string) (CheckSummary, error) { return summary, fmt.Errorf("check %s: %s", id, message) }
		if current[id] != nil {
			return fail("duplicate ID")
		}
		current[id] = check
		history := check["sourceHistory"].([]any)
		removed := false
		for i, entry := range history {
			event := entry.(map[string]any)
			change := event["change"].(string)
			if (i == 0 && change != "added") || (i > 0 && change == "added") || (i > 0 && removed && change != "restored") || (i > 0 && !removed && change == "restored") {
				return fail("invalid sourceHistory lifecycle")
			}
			if i > 0 && event["date"].(string) < history[i-1].(map[string]any)["date"].(string) {
				return fail("sourceHistory dates must be chronological")
			}
			removed = change == "removed" || change == "withdrawn"
		}
		last := history[len(history)-1].(map[string]any)
		key := last["url"].(string) + "#" + last["section"].(string)
		if !urls[last["url"].(string)] {
			return fail("latest source URL is absent from sources")
		}
		if !removed {
			status, exists := sections[key]
			if !exists {
				return fail("latest source section is absent from sources")
			}
			if status == "no-requirements" {
				return fail("active check references a no-requirements section")
			}
			mapped[key] = true
		} else {
			if last["change"] == "removed" {
				summary.Removed++
			} else {
				summary.Withdrawn++
			}
		}
		impl := check["implementation"].(map[string]any)
		rev := int(impl["sourceRevision"].(float64))
		if rev > len(history) || (!removed && rev != len(history)) {
			return fail("implementation assessment references an old or nonexistent source revision; reassess or reset to unassessed")
		}
		status := impl["status"].(string)
		if status == "implemented" {
			hasCode := false
			for _, entry := range impl["evidence"].([]any) {
				if entry.(map[string]any)["kind"] == "code" {
					hasCode = true
				}
			}
			if !hasCode {
				return fail("implemented requires code evidence")
			}
		}
		if status == "unassessed" && (len(impl["evidence"].([]any)) != 0 || impl["assessedCommit"] != nil) {
			return fail("unassessed must not retain an assessment commit or evidence")
		}
		if commit, ok := impl["assessedCommit"].(string); ok {
			commits[commit] = true
		}
		for _, entry := range check["fixes"].([]any) {
			fix := entry.(map[string]any)
			revision := int(fix["sourceRevision"].(float64))
			if revision > len(history) || (history[revision-1].(map[string]any)["change"] == "removed" || history[revision-1].(map[string]any)["change"] == "withdrawn") {
				return fail("fix references a nonexistent or removed source revision")
			}
			commits[fix["commit"].(string)] = true
		}
	}
	for key, status := range sections {
		if status == "mapped" && !mapped[key] {
			return summary, fmt.Errorf("mapped section has no active checks: %s", key)
		}
	}
	if len(previous) > 0 {
		old, err := parseChecks(previous)
		if err != nil {
			return summary, fmt.Errorf("previous checks: %w", err)
		}
		if old["component"] != doc["component"] || old["designSystem"] != doc["designSystem"] {
			return summary, fmt.Errorf("component identity must not change")
		}
		for _, entry := range old["checks"].([]any) {
			prior := entry.(map[string]any)
			id := prior["id"].(string)
			next := current[id]
			if next == nil {
				return summary, fmt.Errorf("check %s was deleted; retain it and append a removed or withdrawn event", id)
			}
			for _, field := range []string{"sourceHistory", "fixes"} {
				a, b := prior[field].([]any), next[field].([]any)
				if len(b) < len(a) || !reflect.DeepEqual(a, b[:len(a)]) {
					return summary, fmt.Errorf("check %s: %s is append-only", id, field)
				}
			}
			changed := prior["statement"] != next["statement"] || prior["condition"] != next["condition"] || prior["basis"] != next["basis"]
			if changed {
				a, b := prior["sourceHistory"].([]any), next["sourceHistory"].([]any)
				if len(b) == len(a) {
					return summary, fmt.Errorf("check %s: changed instruction requires a sourceHistory event", id)
				}
				latest := b[len(b)-1].(map[string]any)["change"]
				if latest != "changed" && latest != "restored" && latest != "corrected" {
					return summary, fmt.Errorf("check %s: changed instruction requires changed, corrected or restored event", id)
				}
			}
		}
	}
	for commit := range commits {
		summary.Commits = append(summary.Commits, commit)
	}
	sort.Strings(summary.Commits)
	return summary, nil
}
