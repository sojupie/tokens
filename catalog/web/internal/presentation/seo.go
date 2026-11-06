package presentation

import (
	"context"

	"github.com/sojupie/tokens/forge/site"
)

type PageMetadata struct {
	Navigation         []site.Section
	Illustration       *Illustration
	ImagePageURL       string
	ImageURL           string
	ImageAlt           string
	Export             *ExportPage
	Tools              []ToolDestination
	GeneratedDownloads bool
	Path               string
	Title              string
	Description        string
	Canonical          string
	NoIndex            bool
	Heading            string
	AgentDiscovery     bool
	JSONAlternate      string
}

type PageMetadataKey struct{}

func Metadata(ctx context.Context) PageMetadata {
	value, _ := ctx.Value(PageMetadataKey{}).(PageMetadata)
	return value
}

func PageTitle(ctx context.Context, fallback string) string {
	if title := Metadata(ctx).Title; title != "" {
		return title
	}
	return fallback
}
