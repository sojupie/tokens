package ui

import (
	"context"
	"github.com/a-h/templ"
	"io"
)

// Assets collects component modules during a templ render. The page host resolves
// these module names through forge/components/modules.json and serves the files.
type Assets map[string]bool
type assetsKey struct{}

func WithAssets(ctx context.Context, assets Assets) context.Context {
	return context.WithValue(ctx, assetsKey{}, assets)
}

// Require records dependencies without coupling components to an HTTP server.
func Require(modules ...string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		if assets, ok := ctx.Value(assetsKey{}).(Assets); ok {
			assets["common"] = true
			for _, module := range modules {
				assets[module] = true
			}
		}
		return nil
	})
}
