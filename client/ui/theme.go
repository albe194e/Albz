package ui

import "image/color"

type Theme struct {
	AppBackground  color.NRGBA
	GradientStart  color.NRGBA
	GradientMid    color.NRGBA
	GradientEnd    color.NRGBA
	AppShellFill   color.NRGBA
	AppShellStroke color.NRGBA
	Surface        color.NRGBA
	ElevatedCard   color.NRGBA
	Border         color.NRGBA
	Divider        color.NRGBA
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
	AppBackground:  color.NRGBA{R: 30, G: 34, B: 41, A: 255},
	GradientStart:  color.NRGBA{R: 30, G: 34, B: 41, A: 255},
	GradientMid:    color.NRGBA{R: 30, G: 34, B: 41, A: 255},
	GradientEnd:    color.NRGBA{R: 30, G: 34, B: 41, A: 255},
	AppShellFill:   color.NRGBA{A: 0},
	AppShellStroke: color.NRGBA{R: 102, G: 120, B: 164, A: 72},
	Surface:        color.NRGBA{R: 28, G: 35, B: 50, A: 238},
	ElevatedCard:   color.NRGBA{R: 35, G: 43, B: 60, A: 236},
	Border:         color.NRGBA{R: 102, G: 120, B: 164, A: 84},
	Divider:        color.NRGBA{R: 102, G: 120, B: 164, A: 28},
	PrimaryText:    color.NRGBA{R: 239, G: 242, B: 248, A: 255},
	SecondaryText:  color.NRGBA{R: 168, G: 179, B: 202, A: 255},
	MutedText:      color.NRGBA{R: 118, G: 128, B: 150, A: 255},
	PrimaryAccent:  color.NRGBA{R: 92, G: 117, B: 214, A: 255},
	AccentHover:    color.NRGBA{R: 122, G: 146, B: 232, A: 255},
	SuccessOnline:  color.NRGBA{R: 143, G: 224, B: 193, A: 255},
	WarningDelayed: color.NRGBA{R: 246, G: 203, B: 121, A: 255},
	ErrorFailed:    color.NRGBA{R: 234, G: 138, B: 154, A: 255},

	SidebarFill:            color.NRGBA{A: 0},
	SidebarStroke:          color.NRGBA{A: 0},
	PanelFill:              color.NRGBA{A: 0},
	PanelStroke:            color.NRGBA{A: 0},
	ConversationCardFill:   color.NRGBA{R: 39, G: 47, B: 65, A: 228},
	ConversationCardStroke: color.NRGBA{A: 0},
}

const BtnRadius = 12
