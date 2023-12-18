package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

func recordOrder() {
	// orderClickMousePos = []MousePos{}
	// fmt.Println(len(orderClickMousePos))
	evChan := hook.Start()
	defer hook.End()
	// var key string
	currentMousePosWindow := createCurrentMousePosWindow(mainApp)
	currentMousePosWindow.Show()

	var jsonStruct []Action

	for ev := range evChan {

		if ev.Kind == hook.KeyDown {
			if ev.Keychar == 27 {
				break
			}
			// key = string(e.Keychar)
		} else if ev.Kind == hook.MouseDown && ev.Button == 1 {
			x, y := robotgo.Location()
			fmt.Println(x, "    ", y)
			// orderClickMousePos = append(orderClickMousePos, MousePos{x, y})

			jsonStruct = append(jsonStruct, Action{
				Id:     len(jsonStruct),
				Action: "moveMouseSmooth",
				Data:   fmt.Sprintf("%s %s", strconv.Itoa(x), strconv.Itoa(y)),
				True:   len(jsonStruct) + 1,
				// Repeat
				// Delay
				Click: true,
				// TrueClick
			})
		}
	}

	jsonStruct[len(jsonStruct)-1].True = 0
	os.Mkdir("./scripts/"+OrderScriptName, os.ModePerm)
	file, err := os.Create(fmt.Sprintf("./scripts/%s/script.json", OrderScriptName))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(jsonStruct)
	if err != nil {
		panic(err)
	}

	currentMousePosWindow.Hide()

	fmt.Println("clicked")
	// mainWindow.Show()
}

func saveSimpleScript() {
	inputEntry := widget.NewEntry()

	inputDialog := dialog.NewCustomConfirm("Enter Text", "OK", "Cancel", container.NewVBox(
		widget.NewLabel("Please enter text:"),
		layout.NewSpacer(),
		inputEntry,
	), func(ok bool) {
		if ok {
			// Retrieve user input when OK is pressed
			userInput := inputEntry.Text
			if userInput != "" {
				// Process the user input (e.g., display it)
				// dialog.ShowInformation("User Input", "You entered: "+userInput, mainWindow)
				err := os.Rename("./scripts/"+SimpleScriptName, "./scripts/"+userInput)
				if err != nil {
					dialog.ShowInformation("Error", "Can not save script", mainWindow)
				}
			} else {
				dialog.ShowInformation("Error", "Please enter a valid text", mainWindow)
			}
		}
	}, mainWindow)

	// Show the dialog
	inputDialog.Show()
}

var runScriptButton = widget.NewButton("Run script", func() {
})

func newEmptyScript() {
	inputEntry := widget.NewEntry()

	inputDialog := dialog.NewCustomConfirm("Enter script name", "OK", "Cancel", container.NewVBox(
		widget.NewLabel("Please enter script name:"),
		layout.NewSpacer(),
		inputEntry,
	), func(ok bool) {
		if ok {
			// Retrieve user input when OK is pressed
			userInput := inputEntry.Text
			if userInput != "" {
				if !checkThenSaveScript(userInput, []Action{}) {
					dialog.ShowInformation("Error", "Script exist", mainWindow)
				}
			} else {
				dialog.ShowInformation("Error", "Please enter a valid name", mainWindow)
			}
		}
	}, mainWindow)

	// Show the dialog
	inputDialog.Show()
}
