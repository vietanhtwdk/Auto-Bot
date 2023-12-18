package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/go-vgo/robotgo"
)

const SimpleScriptName = "SimpleScript"

func createCurrentMousePosWindow(myApp fyne.App) fyne.Window {
	mousePosWindow := myApp.NewWindow("Mouse pos")
	textX := canvas.NewText("0", black)
	textY := canvas.NewText("0", black)
	go func() {
		for {
			x, y := robotgo.Location()
			textX.Text = strconv.Itoa(x)
			textY.Text = strconv.Itoa(y)
			textX.Refresh()
			textY.Refresh()
		}
	}()
	mousePosWindow.SetContent(container.New(layout.NewHBoxLayout(), textX, textY))
	return mousePosWindow
}

// func makeAppTabsTab(_ fyne.Window) fyne.CanvasObject {
// 	tabs := container.NewAppTabs(
// 		container.NewTabItem("Main", widget.NewLabel("Main")),
// 		container.NewTabItem("Script maker", widget.NewLabel("Script maker")),
// 		// container.NewTabItem("Tab 3", widget.NewLabel("Content of tab 3")),
// 	)
// 	for i := 4; i <= 12; i++ {
// 		tabs.Append(container.NewTabItem(fmt.Sprintf("Tab %d", i), widget.NewLabel(fmt.Sprintf("Content of tab %d", i))))
// 	}
// 	locations := makeTabLocationSelect(tabs.SetTabLocation)
// 	return container.NewBorder(locations, nil, nil, nil, tabs)
// }
