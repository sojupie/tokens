package publish

import "github.com/sojupie/tokens/source/tokensnapshot"

type resolvedContext = tokensnapshot.ResolvedContext

func contextDescriptions(records []tokensnapshot.ExtractedRecord) map[string]resolvedContext {
	return tokensnapshot.ContextDescriptions(records)
}

func resolveContext(tags []string, descriptions map[string]resolvedContext) ([]string, bool) {
	return tokensnapshot.ResolveContext(tags, descriptions)
}
