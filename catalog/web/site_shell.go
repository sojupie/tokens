package web

import (
	"context"
	"github.com/sojupie/tokens/forge/site"
)

func siteConfig(ctx context.Context) site.Config {
	meta := metadata(ctx)
	config := site.Config{Active: meta.Path, Agents: meta.AgentDiscovery}
	for _, tool := range meta.Tools {
		config.Tools = append(config.Tools, site.Destination{Name: tool.Name, URL: tool.URL})
	}
	return config
}
