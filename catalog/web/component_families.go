package web

func componentFamily(slug string) string {
	return map[string]string{"buttons": "button", "cards": "card", "icon-buttons": "icon", "chips": "chip", "radio-button": "radio", "text-fields": "field", "menus": "menu", "dialogs": "dialog", "divider": "divider", "checkbox": "checkbox", "switch": "switch", "floating-action-button": "fab", "extended-fab": "extended-fab", "badges": "badge", "snackbar": "snackbar", "tooltips": "tooltip", "navigation-bar": "nav-bar", "navigation-rail": "nav-rail", "navigation-drawer": "nav-drawer", "tabs": "tabs", "segmented-buttons": "segmented", "sliders": "slider", "progress-indicators": "progress", "loading-indicator": "loading", "app-bars": "appbar", "toolbars": "toolbar", "bottom-sheets": "bottom-sheet", "side-sheets": "side-sheet", "split-button": "split", "button-groups": "button-group", "fab-menu": "fab-menu", "lists": "list", "carousel": "carousel", "search": "search", "date-pickers": "date", "time-pickers": "time"}[slug]
}
