package web

import (
	"context"
	"github.com/sojupie/tokens/forge/site"
	"github.com/sojupie/tokens/source/publish"
	"sort"
	"strings"
)

func siteConfig(ctx context.Context) site.Config {
	meta := metadata(ctx)
	config := site.Config{Active: meta.Path, Agents: meta.AgentDiscovery, Navigation: meta.Navigation}
	for _, tool := range meta.Tools {
		config.Tools = append(config.Tools, site.Destination{Name: tool.Name, URL: tool.URL})
	}
	return config
}

// SiteNavigation is also used when exporting the shell for static tools.
func SiteNavigation(state publish.PublishedState) []site.Section {
	sections := []site.Section{{Name: "Components", Path: "/components"}, {Name: "Styles", Path: "/styles"}}
	for _, subject := range state.Subjects {
		index := 0
		if subject.Key.Kind == publish.SubjectStyle {
			index = 1
		} else if subject.Key.Kind != publish.SubjectComponent {
			continue
		}
		sections[index].Items = append(sections[index].Items, site.Destination{Name: subjectName(subject), URL: subjectURL(subject)})
	}
	for i := range sections {
		sort.Slice(sections[i].Items, func(a, b int) bool {
			return strings.ToLower(sections[i].Items[a].Name) < strings.ToLower(sections[i].Items[b].Name)
		})
	}
	return sections
}
