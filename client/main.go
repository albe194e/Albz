package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	fmt.Println("starting Albz")

	a := app.New()
	w := a.NewWindow("Albz")

	w.SetContent(widget.NewLabel("Hello Albz"))
	w.Resize(fyne.NewSize(400, 600))

	fmt.Println("showing window")
	w.ShowAndRun()

	fmt.Println("window closed")
}