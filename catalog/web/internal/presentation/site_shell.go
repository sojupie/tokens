package presentation

import (
	"context"

	"github.com/sojupie/tokens/forge/site"
)

func SiteConfig(ctx context.Context) site.Config {
	meta := Metadata(ctx)
	config := site.Config{Active: meta.Path, Agents: meta.AgentDiscovery, Navigation: meta.Navigation}
	for _, tool := range meta.Tools {
		config.Tools = append(config.Tools, site.Destination{Name: tool.Name, URL: tool.URL})
	}
	return config
}
