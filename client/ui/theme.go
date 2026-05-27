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
	AppBackground:  color.NRGBA{R: 30, G: 30, B: 30, A: 255},
	Surface:        color.NRGBA{R: 40, G: 40, B: 40, A: 255},
	ElevatedCard:   color.NRGBA{R: 50, G: 50, B: 50, A: 255},
	Border:         color.NRGBA{R: 60, G: 60, B: 60, A: 255},
	PrimaryText:    color.NRGBA{R: 220, G: 220, B: 220, A: 255},
	SecondaryText:  color.NRGBA{R: 180, G: 180, B: 180, A: 255},
	MutedText:      color.NRGBA{R: 120, G: 120, B: 120, A: 255},
	PrimaryAccent:  color.NRGBA{R: 100, G: 150, B: 250, A: 255},
	AccentHover:    color.NRGBA{R: 120, G: 170, B: 255, A: 255},
	SuccessOnline:  color.NRGBA{R: 80, G: 200, B: 120, A: 255},
	WarningDelayed: color.NRGBA{R: 250, G: 180, B: 50, A: 255},
	ErrorFailed:    color.NRGBA{R: 250, G: 80, B: 80, A: 255},

	SidebarFill:            color.NRGBA{R: 35, G: 35, B: 35, A: 255},
	SidebarStroke:          color.NRGBA{R: 70, G: 70, B: 70, A: 255},
	PanelFill:              color.NRGBA{R: 45, G: 45, B: 45, A: 255},
	PanelStroke:            color.NRGBA{R: 80, G: 80, B: 80, A: 255},
	ConversationCardFill:   color.NRGBA{R: 41, G: 49, B: 61, A: 255},
	ConversationCardStroke: color.NRGBA{R: 72, G: 84, B: 101, A: 255},
}
