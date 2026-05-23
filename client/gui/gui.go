package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func Init() {
	fmt.Println("starting Albz")

	a := app.New()
	w := a.NewWindow("Albz")

	// Left conversation container
	sidebar := container.New(
		layout.NewCustomPaddedLayout(12, 12, 12, 12),
		container.NewGridWrap(
			fyne.NewSize(220, 600),
			widget.NewList(
				func() int {
					return 10
				},
				func() fyne.CanvasObject {
					return widget.NewLabel("Conversation")
				},
				func(i widget.ListItemID, o fyne.CanvasObject) {
					o.(*widget.Label).SetText(fmt.Sprintf("Conversation %d", i+1))
				},
			),
		),
	)

	// Right content container ( Messeages, inputs, etc )
	contentArea := container.New(
		layout.NewCustomPaddedLayout(12, 12, 12, 12),
		widget.NewLabel("Content Area"),
	)

	// Main layout
	content := container.NewBorder(
		nil,
		nil,
		sidebar,
		contentArea,
	)

	background := canvas.NewRectangle(AppBackground)
	w.SetContent(container.NewMax(background, content))
	w.Resize(fyne.NewSize(500, 700))

	fmt.Println("showing window")
	w.ShowAndRun()

	fmt.Println("window closed")
}
