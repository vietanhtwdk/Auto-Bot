package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"golang.org/x/exp/slices"
)

func makeSimpleScriptContainer() *fyne.Container {
	imageWidthEntry := widget.NewEntry()
	imageHeightEntry := widget.NewEntry()

	recordImageWidth := cfgGlobal.Section("setting").Key("RecordImageWidth").String()
	if recordImageWidth == "" {
		recordImageWidth = "20"
	}
	recordImageHeight := cfgGlobal.Section("setting").Key("RecordImageHeight").String()
	if recordImageHeight == "" {
		recordImageHeight = "20"
	}

	imageWidthEntry.OnChanged = func(s string) {
		cfgGlobal.Section("setting").Key("RecordImageWidth").SetValue(s)
	}

	imageHeightEntry.OnChanged = func(s string) {
		cfgGlobal.Section("setting").Key("RecordImageHeight").SetValue(s)
	}

	imageWidthEntry.Text = recordImageWidth
	imageHeightEntry.Text = recordImageHeight

	imageDirectionSelect := widget.NewSelect([]string{}, func(selected string) {
		fmt.Println("Selected option:", selected)
	})
	imageDirectionSelect.PlaceHolder = "Select image direction"
	imageDirectionSelect.Options = []string{"Top-Left", "Top", "Top-Right", "Left", "Center", "Right", "Bottom-Left", "Bottom", "Bottom-Right"}
	startRecordButton := widget.NewButton("Start record", func() {
		os.RemoveAll("./scripts/" + SimpleScriptName)
		os.Mkdir("./scripts/"+SimpleScriptName, os.ModePerm)
		evChan := hook.Start()
		defer hook.End()

		imageWidth := parseInt(imageWidthEntry.Text)
		imageHeight := parseInt(imageHeightEntry.Text)
		var jsonStruct []Action

		for ev := range evChan {
			// print(ev.Kind)
			if ev.Kind == hook.KeyDown {
				if ev.Keychar == 27 {
					break
				}
				// key = string(e.Keychar)
			} else if ev.Kind == hook.MouseDown && ev.Button == 1 {
				x, y := robotgo.Location()
				// fmt.Println(x, "    ", y)
				// // orderClickMousePos = append(orderClickMousePos, MousePos{x, y})

				switch imageDirectionSelect.Selected {
				case "Top-Left":
					//
				case "Top":
					x -= imageWidth / 2
				case "Top-Right":
					x -= imageWidth
				case "Left":
					y -= imageHeight / 2
				case "Center":
					x -= imageWidth / 2
					y -= imageHeight / 2
				case "Right":
					x -= imageWidth
					y -= imageHeight / 2
				case "Bottom-Left":
					y -= imageHeight
				case "Bottom":
					x -= imageWidth / 2
					y -= imageHeight
				case "Bottom-Right":
					x -= imageWidth
					y -= imageHeight
				}

				clickImage := robotgo.CaptureScreen(x, y, imageWidth, imageHeight)
				robotgo.Save(robotgo.ToImage(clickImage), "./scripts/"+SimpleScriptName+"/"+strconv.Itoa(len(jsonStruct))+".png")

				jsonStruct = append(jsonStruct, Action{
					Id:     len(jsonStruct),
					Action: "findImageMoveSmooth",
					// Data:   fmt.Sprintf("%s %s", strconv.Itoa(x), strconv.Itoa(y)),
					Data:  strconv.Itoa(len(jsonStruct)) + ".png",
					True:  len(jsonStruct) + 1,
					False: len(jsonStruct),
					// Repeat
					// Delay
					// Click:     true,
					TrueClick: true,
				})
			}
		}

		if len(jsonStruct) == 0 {
			return
		}
		jsonStruct[len(jsonStruct)-1].True = 0

		saveScript(SimpleScriptName, jsonStruct)
	})
	// go func() {
	// 	for {
	// 		x, y := robotgo.Location()
	// 		textX.Text = strconv.Itoa(x)
	// 		textY.Text = strconv.Itoa(y)
	// 		textX.Refresh()
	// 		textY.Refresh()
	// 	}
	// }()
	saveSimpleScriptButton := widget.NewButton("Save as", saveSimpleScript)
	return container.NewVBox(container.New(layout.NewHBoxLayout(), startRecordButton, canvas.NewText("width", black), imageWidthEntry, canvas.NewText("height", black), imageHeightEntry, imageDirectionSelect), saveSimpleScriptButton)
}

func makeMainContainer() *fyne.Container {
	scriptStepSelect := widget.NewSelect([]string{}, func(selected string) {
		fmt.Println("Selected option:", selected)
	})
	scriptStepSelect.PlaceHolder = "Select start step"

	scriptSelect := widget.NewSelect([]string{}, func(selected string) {
		fmt.Println("Selected option:", selected)

		content, err := os.ReadFile("./scripts/" + selected + "/script.json")

		if err != nil {
			fmt.Println(err)
			return
		}

		var jsonStruct []Action

		err = json.Unmarshal(content, &jsonStruct)
		if err != nil {
			fmt.Println(err)
			return
		}

		var scriptActionIndexs []string

		for _, action := range jsonStruct {
			scriptActionIndexs = append(scriptActionIndexs, strconv.Itoa(action.Id))
		}

		scriptStepSelect.Options = scriptActionIndexs
		selectedScript = selected
	})
	scriptSelect.PlaceHolder = "Select script"

	scriptIntervalEntry := widget.NewEntry()
	scriptIntervalEntry.Validator = func(text string) error {
		_, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		return nil
	}
	scriptIntervalEntry.Text = DEFAULT_DELAY_ENTRY
	defaultDelayValue = parseInt(scriptIntervalEntry.Text)
	scriptIntervalEntry.OnChanged = func(s string) {
		defaultDelayValue = parseInt(s)
		cfgGlobal.Section("setting").Key("DefaultDelay").SetValue(scriptIntervalEntry.Text)
	}

	imageToleranceEntry := widget.NewEntry()
	imageToleranceEntry.Validator = func(text string) error {
		_, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		return nil
	}
	imageToleranceEntry.Text = DEFAULT_IMG_TOLERANCE_ENTRY
	defaultImgToleranceValue = parseFloat(imageToleranceEntry.Text)
	imageToleranceEntry.OnChanged = func(s string) {
		defaultImgToleranceValue = parseFloat(s)
		cfgGlobal.Section("setting").Key("DefaultImageTolerance").SetValue(imageToleranceEntry.Text)
	}

	scriptSelect.Options = getScriptList()

	go watchScriptsFolderChange(scriptSelect)

	if slices.Contains(scriptSelect.Options, selectedScript) {
		scriptSelect.Selected = selectedScript
	}

	// runScriptButton := widget.NewButton("Run script", func() {
	// 	if scriptRunning {
	// 		return
	// 	}
	// 	// TODO mainWindow.Minimize()
	// 	scriptName := scriptSelect.Selected
	// 	startStep := parseInt(scriptStepSelect.Selected)
	// 	if len(scriptStepSelect.Selected) > 0 {
	// 		go runScript(scriptName, startStep)
	// 	} else {
	// 		go runScript(scriptName, nil)
	// 	}
	// })

	runScriptButton.OnTapped = func() {
		// TODO mainWindow.Minimize()
		scriptName := scriptSelect.Selected
		startStep := parseInt(scriptStepSelect.Selected)
		if len(scriptStepSelect.Selected) > 0 {
			go runScript(scriptName, startStep)
		} else {
			go runScript(scriptName, nil)
		}
		runScriptButton.Disable()
	}

	stopButton := widget.NewButton("Stop", func() {
		robotgo.KeyTap("esc")
	})
	hbox5 := container.NewGridWithColumns(4, widget.NewLabel("Default delay"), scriptIntervalEntry, widget.NewLabel("Default img tolerance"), imageToleranceEntry)

	vbox := container.New(layout.NewVBoxLayout(), container.NewHBox(scriptSelect, scriptStepSelect), hbox5, makeSearchLocationContainer(), container.NewHBox(runScriptButton, stopButton))

	return vbox
}

var (
	searchAreaX  int
	searchAreaY  int
	searchAreaX2 int
	searchAreaY2 int
)

func makeSearchLocationContainer() *fyne.Container {
	var imgSearchX = widget.NewEntry()
	var imgSearchY = widget.NewEntry()
	var imgSearchX2 = widget.NewEntry()
	var imgSearchY2 = widget.NewEntry()
	imgSearchX.PlaceHolder = "X1"
	imgSearchX.OnChanged = func(s string) {
		searchAreaX = parseInt(s)
		fmt.Println(searchAreaX)
	}
	imgSearchX.SetText(strconv.Itoa((allScreenBound[0])))
	imgSearchY.PlaceHolder = "Y1"
	imgSearchY.OnChanged = func(s string) {
		searchAreaY = parseInt(s)
	}
	imgSearchY.SetText(strconv.Itoa((allScreenBound[1])))
	imgSearchX2.PlaceHolder = "X2"
	imgSearchX2.OnChanged = func(s string) {
		searchAreaX2 = parseInt(s)
	}
	imgSearchX2.SetText(strconv.Itoa((allScreenBound[0] + allScreenBound[2])))

	imgSearchY2.PlaceHolder = "Y2"
	imgSearchY2.OnChanged = func(s string) {
		searchAreaY2 = parseInt(s)
	}
	imgSearchY2.SetText(strconv.Itoa((allScreenBound[1] + allScreenBound[3])))

	setAllScreenSearchButton := widget.NewButton("All screen", func() {
		imgSearchX.SetText(strconv.Itoa((allScreenBound[0])))
		imgSearchY.SetText(strconv.Itoa((allScreenBound[1])))
		imgSearchX2.SetText(strconv.Itoa((allScreenBound[0] + allScreenBound[2])))
		imgSearchY2.SetText(strconv.Itoa((allScreenBound[1] + allScreenBound[3])))
	})

	return container.NewBorder(nil, nil, widget.NewLabel("Search area: "), setAllScreenSearchButton, container.NewGridWithColumns(4, imgSearchX, imgSearchY, imgSearchX2, imgSearchY2))
}

func getSearchArea() robotgo.CBitmap {
	fmt.Println(searchAreaX, searchAreaY, searchAreaX2-searchAreaX, searchAreaY2-searchAreaY)
	return robotgo.CaptureScreen(searchAreaX, searchAreaY, searchAreaX2-searchAreaX, searchAreaY2-searchAreaY)
}

var listSelected widget.ListItemID

var loadedActions []Action

func makeScriptEditorContainer() *fyne.Container {
	var hsplit *container.Split
	var con *fyne.Container
	list := widget.NewList(
		func() int {
			return len(loadedActions)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Id"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(strconv.Itoa(loadedActions[id].Id))
		},
	)

	scriptSelect := widget.NewSelect([]string{}, func(selected string) {
		fmt.Println("Selected option:", selected)

		if selected == "" {
			return
		}

		content, err := os.ReadFile("./scripts/" + selected + "/script.json")

		if err != nil {
			fmt.Println(err)
			return
		}

		var jsonStruct []Action

		err = json.Unmarshal(content, &jsonStruct)
		if err != nil {
			fmt.Println(err)
			return
		}

		loadedActions = jsonStruct
		// list.UnselectAll()
		list.Refresh()
		hsplit.Trailing = widget.NewLabel("Select an action")
		con.Refresh()
	})
	list.OnSelected = func(id widget.ListItemID) {
		listSelected = id
		hsplit.Trailing = makeActionEditorContainer(list, scriptSelect.Selected)
		hsplit.Refresh()
	}
	scriptSelect.PlaceHolder = "Select script"
	scriptSelect.Options = getScriptList()
	go watchScriptsFolderChange(scriptSelect)
	addActionButton := widget.NewButton("Add action", func() {
		loadedActions = append(loadedActions, Action{})
		list.Refresh()
	})
	hsplit = container.NewHSplit(container.NewBorder(nil, addActionButton, nil, nil, list), widget.NewLabel("Select an action"))
	hsplit.Offset = 0.2

	newScriptButton := widget.NewButton("New script", newEmptyScript)
	saveScriptButton := widget.NewButton("Save script", func() {
		saveScript(scriptSelect.Selected, loadedActions)
		list.Refresh()
	})
	// saveScriptButton.Disable()
	deleteScriptButton := widget.NewButton("Delete script", func() {
		confirmDialog := dialog.NewConfirm("Confirm", fmt.Sprintf("Are you sure to delete %s", scriptSelect.Selected), func(response bool) {
			if response {
				deleteScript(scriptSelect.Selected)
				loadedActions = []Action{}
				scriptSelect.ClearSelected()
				list.Refresh()
				hsplit.Trailing = widget.NewLabel("Select an action")
				hsplit.Refresh()
			}
		}, mainWindow)
		confirmDialog.Show()
	})
	// deleteScriptButton.Disable()
	con = container.NewBorder(container.NewVBox(widget.NewLabel("Script editor"), container.NewBorder(nil, nil, newScriptButton, container.NewHBox(saveScriptButton, deleteScriptButton), scriptSelect)), nil, nil, nil, hsplit)

	return con
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateRandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func makeActionEditorContainer(list *widget.List, scriptName string) *fyne.Container {
	idEntry := widget.NewEntry()
	idEntry.OnChanged = func(s string) {
		loadedActions[listSelected].Id = parseInt(s)
	}
	nameEntry := widget.NewEntry()
	nameEntry.OnChanged = func(s string) {
		loadedActions[listSelected].Name = s
	}
	actionEntry := widget.NewSelectEntry(ActionValue)
	actionEntry.PlaceHolder = "Type or select"

	dataEntry := widget.NewEntry()
	dataEntry.OnChanged = func(s string) {
		loadedActions[listSelected].Data = s
	}
	trueEntry := widget.NewEntry()
	trueEntry.OnChanged = func(s string) {
		loadedActions[listSelected].True = parseInt(s)
	}
	falseEntry := widget.NewEntry()
	falseEntry.OnChanged = func(s string) {
		loadedActions[listSelected].False = parseInt(s)
	}
	repeatEntry := widget.NewEntry()
	repeatEntry.OnChanged = func(s string) {
		loadedActions[listSelected].Repeat = parseInt(s)
	}
	delayEntry := widget.NewEntry()
	delayEntry.OnChanged = func(s string) {
		loadedActions[listSelected].Delay = parseInt(s)
	}
	clickRadio := widget.NewRadioGroup([]string{"true", "false"}, func(s string) { fmt.Println("selected", s) })
	clickRadio.Horizontal = true
	clickRadio.Required = true
	clickRadio.Selected = "False"
	clickRadio.OnChanged = func(s string) {
		loadedActions[listSelected].Click, _ = strconv.ParseBool(s)
	}
	trueClickRadio := widget.NewRadioGroup([]string{"true", "false"}, func(s string) { fmt.Println("selected", s) })
	trueClickRadio.Horizontal = true
	trueClickRadio.Required = true
	trueClickRadio.Selected = "True"
	trueClickRadio.OnChanged = func(s string) {
		loadedActions[listSelected].TrueClick, _ = strconv.ParseBool(s)
	}
	imgToleranceEntry := widget.NewEntry()
	imgToleranceEntry.OnChanged = func(s string) {
		loadedActions[listSelected].ImgTolerance = parseFloat(s)
	}

	addPhotoButton := widget.NewButton("Add photo", func() {
		runScreenClip()
		for {
			cmd := exec.Command("cmd", "/C", "tasklist | findstr ScreenClippingHost")
			_, err := cmd.Output()
			if err != nil {
				fmt.Println("Error executing command:", err)
				break
			}
		}
		img := getImgFromClipboard()
		imageName := generateRandomString(4) + ".png"
		robotgo.SavePng(img, "./scripts/"+scriptName+"/"+imageName)
		// robotgo.SaveJpeg(img, "./scripts/"+scriptName+"/"+generateRandomString(4)+".jpeg")
		if dataEntry.Text == "" {
			dataEntry.SetText(imageName)
		} else {
			dataEntry.SetText(dataEntry.Text + " " + imageName)
		}

	})

	actionEntry.OnChanged = func(s string) {
		loadedActions[listSelected].Action = s
		if s == "findImageMoveSmooth" || s == "findImageMove" {
			addPhotoButton.Enable()
		} else {
			addPhotoButton.Disable()
		}
	}

	dataWidget := container.NewBorder(nil, nil, nil, addPhotoButton, dataEntry)
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Id", Widget: idEntry, HintText: ""},
			{Text: "Name", Widget: nameEntry},
			{Text: "Action", Widget: actionEntry},
			{Text: "Data", Widget: dataWidget},
			{Text: "True", Widget: trueEntry},
			{Text: "False", Widget: falseEntry},
			{Text: "Repeat", Widget: repeatEntry},
			{Text: "Delay", Widget: delayEntry},
			{Text: "Click", Widget: clickRadio},
			{Text: "TrueClick", Widget: trueClickRadio},
			{Text: "ImgTolerance", Widget: imgToleranceEntry},
		},
		// SubmitText: "Save",
		// CancelText: "Delete",
		// OnCancel: func() {
		// 	loadedActions = append(loadedActions[:listSelected], loadedActions[listSelected+1:]...)
		// 	fmt.Println(len(loadedActions))
		// 	list.Refresh()
		// },
		// OnSubmit: func() {
		// 	saveScript(scriptName, loadedActions)
		// 	list.Refresh()
		// },
	}

	idEntry.SetText(strconv.Itoa(loadedActions[listSelected].Id))
	nameEntry.SetText(loadedActions[listSelected].Name)
	actionEntry.SetText(loadedActions[listSelected].Action)
	dataEntry.SetText(loadedActions[listSelected].Data)
	trueEntry.SetText(strconv.Itoa(loadedActions[listSelected].True))
	falseEntry.SetText(strconv.Itoa(loadedActions[listSelected].False))
	repeatEntry.SetText(strconv.Itoa(loadedActions[listSelected].Repeat))
	delayEntry.SetText(strconv.Itoa(loadedActions[listSelected].Delay))
	clickRadio.SetSelected(strconv.FormatBool(loadedActions[listSelected].Click))
	trueClickRadio.SetSelected(strconv.FormatBool(loadedActions[listSelected].TrueClick))
	imgToleranceEntry.SetText(floatToString(loadedActions[listSelected].ImgTolerance))

	addBeforeButton := widget.NewButton("Add before", func() {
		loadedActions = append(loadedActions[:listSelected], append([]Action{{}}, loadedActions[listSelected:]...)...)
		list.Select(listSelected + 1)
		list.Refresh()
	})
	addAfterButton := widget.NewButton("Add after", func() {
		loadedActions = append(loadedActions[:listSelected+1], append([]Action{{}}, loadedActions[listSelected+1:]...)...)
		list.Refresh()
	})
	deleteButton := widget.NewButton("Delete", func() {
		loadedActions = append(loadedActions[:listSelected], loadedActions[listSelected+1:]...)
		list.Refresh()
	})
	// saveButton := widget.NewButton("Save", func() {})

	return container.NewBorder(nil, container.NewGridWithColumns(3, addBeforeButton, addAfterButton, deleteButton), nil, nil, form)
}
