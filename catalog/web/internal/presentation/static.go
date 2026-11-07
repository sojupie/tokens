package presentation

import (
	"context"
)

type ThemeVersionKey struct{}

func ThemeURL(ctx context.Context) string {
	if themes, ok := ctx.Value(PageThemeKey{}).(PageThemesContext); ok && themes.Shared != "" {
		return themes.Shared
	}
	if version, ok := ctx.Value(ThemeVersionKey{}).(string); ok {
		return StylesheetURL(ctx, "/theme.css?v="+version)
	}
	return StylesheetURL(ctx, "/theme.css")
}

// StylesheetURL selects a server's delivery profile without changing public
// component exports or embedding a property-name dictionary in the browser.
func StylesheetURL(ctx context.Context, original string) string {
	if assets, ok := ctx.Value(ComponentAssetsKey{}).(*ComponentAssets); ok && assets.Server != nil {
		return assets.Server.StylesheetURL(original)
	}
	return original
}
