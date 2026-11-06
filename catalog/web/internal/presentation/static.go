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
		return "/theme.css?v=" + version
	}
	return "/theme.css"
}
