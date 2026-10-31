package ui

import "github.com/a-h/templ"

type DialogProps struct {
	ID, TitleID, Title, Description                       string
	Alert, FullScreen, HideTitle, Dividers, BottomActions bool
	Icon, Actions, HeaderActions                          templ.Component
	Attrs, CloseAttrs, ContentAttrs                       templ.Attributes
}

func (p DialogProps) titleID() string { return fallback(p.TitleID, p.ID+"-heading") }
func (p DialogProps) Attributes() templ.Attributes {
	class, variant, role := "md-dialog", "basic", "dialog"
	if p.FullScreen {
		class, variant = "md-full-screen-dialog", "full-screen"
	}
	if p.Alert {
		role = "alertdialog"
	}
	a := templ.Attributes{"class": class, "id": p.ID, "role": role, "data-dialog-surface": "", "data-variant": variant}
	if p.HideTitle {
		a["aria-label"] = p.Title
	} else {
		a["aria-labelledby"] = p.titleID()
	}
	if p.Description != "" {
		a["aria-describedby"] = p.ID + "-support"
	}
	return attributes(a, p.Attrs)
}
