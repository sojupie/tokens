package presentation

import "github.com/a-h/templ"

// PhoneSceneProps describes the application inside the reference phone.
type PhoneSceneProps struct {
	TopBar, BottomRail, BottomBar, Overlay templ.Component
	BottomBarClass                         string
	PaneAtBottom                           bool
}

type PhoneCrop string

const (
	PhoneCropTop           PhoneCrop = "top"
	PhoneCropLeadingDetail PhoneCrop = "leading-detail"
	PhoneCropDetail        PhoneCrop = "detail"
	PhoneCropBottom        PhoneCrop = "bottom"
	PhoneCropAction        PhoneCrop = "action"
	PhoneCropActionMenu    PhoneCrop = "action-menu"
	PhoneCropCenter        PhoneCrop = "center"
	PhoneCropModal         PhoneCrop = "modal"
	PhoneCropModalUpper    PhoneCrop = "modal-upper"
)
