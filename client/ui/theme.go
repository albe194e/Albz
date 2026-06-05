package ui

import "image/color"

type Theme struct {
	AppBackground  color.NRGBA
	Surface        color.NRGBA
	ElevatedCard   color.NRGBA
	Border         color.NRGBA
	PrimaryText    color.NRGBA
	SecondaryText  color.NRGBA
	MutedText      color.NRGBA
	PrimaryAccent  color.NRGBA
	AccentHover    color.NRGBA
	SuccessOnline  color.NRGBA
	WarningDelayed color.NRGBA
	ErrorFailed    color.NRGBA

	SidebarFill            color.NRGBA
	SidebarStroke          color.NRGBA
	PanelFill              color.NRGBA
	PanelStroke            color.NRGBA
	ConversationCardFill   color.NRGBA
	ConversationCardStroke color.NRGBA
}

var CurrentTheme = DefaultTheme

var DefaultTheme = Theme{
	AppBackground:  color.NRGBA{R: 12, G: 14, B: 17, A: 255},
	Surface:        color.NRGBA{R: 22, G: 26, B: 31, A: 255},
	ElevatedCard:   color.NRGBA{R: 29, G: 34, B: 40, A: 255},
	Border:         color.NRGBA{R: 52, G: 60, B: 70, A: 255},
	PrimaryText:    color.NRGBA{R: 236, G: 234, B: 229, A: 255},
	SecondaryText:  color.NRGBA{R: 176, G: 184, B: 190, A: 255},
	MutedText:      color.NRGBA{R: 118, G: 127, B: 136, A: 255},
	PrimaryAccent:  color.NRGBA{R: 68, G: 92, B: 98, A: 255},
	AccentHover:    color.NRGBA{R: 84, G: 112, B: 119, A: 255},
	SuccessOnline:  color.NRGBA{R: 101, G: 164, B: 131, A: 255},
	WarningDelayed: color.NRGBA{R: 184, G: 148, B: 88, A: 255},
	ErrorFailed:    color.NRGBA{R: 176, G: 100, B: 100, A: 255},

	SidebarFill:            color.NRGBA{R: 16, G: 19, B: 23, A: 255},
	SidebarStroke:          color.NRGBA{R: 45, G: 52, B: 61, A: 255},
	PanelFill:              color.NRGBA{R: 25, G: 29, B: 35, A: 255},
	PanelStroke:            color.NRGBA{R: 53, G: 61, B: 71, A: 255},
	ConversationCardFill:   color.NRGBA{R: 31, G: 36, B: 43, A: 255},
	ConversationCardStroke: color.NRGBA{R: 66, G: 75, B: 86, A: 255},
}

const BtnRadius = 12
