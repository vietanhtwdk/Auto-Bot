package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/fs"
	"log"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"golang.design/x/clipboard"

	"github.com/fsnotify/fsnotify"

	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/bitmap"
	"gopkg.in/ini.v1"
)

type Direction uint8

type MousePos struct {
	X int
	Y int
}

type Action struct {
	Id           int      `json:"id"`
	Name         string   `json:"name,omitempty"`
	Action       string   `json:"action"`
	Data         string   `json:"data"`
	True         int      `json:"true"`
	False        int      `json:"false"`
	Repeat       int      `json:"repeat"`
	Delay        int      `json:"delay"`
	Actions      []Action `json:"actions"`
	Click        bool     `json:"click"`
	TrueClick    bool     `json:"trueClick"`
	ImgTolerance float64  `json:"tolerance,omitempty"`
}

var defaultDelayValue int
var defaultImgToleranceValue float64

var DEFAULT_DELAY_ENTRY = "100"
var DEFAULT_IMG_TOLERANCE_ENTRY = "0.2"

var moveMouseSmoothProcess *exec.Cmd

func runMoveMouseSmooth(x int, y int, low float64, high float64) {
	cmd := exec.Command("./MoveMouseSmooth.exe", strconv.Itoa(x), strconv.Itoa(y), floatToString(low), floatToString(high))
	moveMouseSmoothProcess = cmd
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	cmd.Wait()
}

func (action *Action) UnmarshalJSON(data []byte) error {
	type ActionAlias Action
	actionAlias := &ActionAlias{
		Delay:        defaultDelayValue,
		ImgTolerance: defaultImgToleranceValue,
	}

	err := json.Unmarshal(data, actionAlias)
	if err != nil {
		return err
	}

	*action = Action(*actionAlias)
	return nil
}

func parseInt(str string) int {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(num)
}

func parseFloat(str string) float64 {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return num
}

func floatToString(x float64) string {
	return strconv.FormatFloat(x, 'f', -1, 64)
}

func getCurrentRefreshRate() (int64, error) {
	cmd := exec.Command("wmic", "PATH", "Win32_videocontroller", "get", "currentrefreshrate")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
	return strconv.ParseInt(strings.Split(out.String(), "\n")[1], 10, 64)
}

func runScreenClip() {
	cmd := exec.Command("explorer", "ms-screenclip:")
	err := cmd.Run()
	if err == nil {
		log.Fatal(err)
	}
}

func getImgFromClipboard() image.Image {
	imageByte := clipboard.Read(clipboard.FmtImage)

	img, _, err := image.Decode(bytes.NewReader(imageByte))
	if err != nil {
		log.Fatalln(err)
	}

	// Get the bounds of the image
	bounds := img.Bounds()

	// Create a new image with the first 3 rows removed
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dx()-3, bounds.Dy()))

	// Copy the pixels, excluding the first 3 rows
	draw.Draw(newImg, newImg.Bounds(), img, image.Pt(3, 0), draw.Src)

	return newImg
	// robotgo.SavePng(img, imagePathName+".png")
}

func findImageFromScreen(imageBitMap robotgo.CBitmap, tolerance float64) (int, int) {
	currentScreen := robotgo.CaptureScreen(allScreenBound...)
	defer robotgo.FreeBitmap(currentScreen)
	return bitmap.Find(imageBitMap, currentScreen, tolerance)
}

func findImageFromImage(img robotgo.CBitmap, imgToFind robotgo.CBitmap, tolerance float64) (int, int) {
	return bitmap.Find(imgToFind, img, tolerance)
}

// func setContentToText(c fyne.Canvas) {
// 	black := color.NRGBA{R: 0, G: 180, B: 0, A: 255}
// 	text := canvas.NewText("Text", black)
// 	canvas.
// 	text.TextStyle.Bold = true
// 	c.SetContent(text)
// }

func getDirList(dir string) []fs.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	return entries
}

func getScriptList() []string {
	entries := getDirList("./scripts")
	var entriesName []string

	for _, e := range entries {
		// fmt.Println(e.Type())
		if e.Type().IsDir() {
			entriesName = append(entriesName, e.Name())
		}
	}

	return entriesName
}

func watchScriptsFolderChange(scriptSelect *widget.Select) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
	}

	err = watcher.Add("./scripts")
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		// case _ = <-watcher.Events:
		case <-watcher.Events:
			// fmt.Println("event:", event)
			// if event.Op&fsnotify.Write == fsnotify.Write {
			// 	fmt.Println("modified file:", event.Name)
			// }
			scriptSelect.Options = getScriptList()
		case err := <-watcher.Errors:
			fmt.Println("error:", err)
		}
	}
}

// var AllRunButtons []*widget.Button

// func newRunButton(label string, tapped func()) *widget.Button {
// 	newButton := widget.NewButton(label, tapped)
// 	AllRunButtons = append(AllRunButtons, newButton)
// 	return newButton
// }

// func disableAllRunButton() {
// 	for _, runButton := range AllRunButtons {
// 		runButton.Disable()
// 	}
// }

// func enableAllRunButton() {
// 	for _, runButton := range AllRunButtons {
// 		runButton.Enable()
// 	}
// }

// var green = color.NRGBA{R: 0, G: 180, B: 0, A: 255}
var black = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
var selectedScript string

// var newSelectedScript string

const configPath = "./config.ini"

var mainWindow fyne.Window
var mainApp fyne.App
var cfgGlobal *ini.File

func findAllScreenBounds() []int {
	var monitors [][]int
	screenIndex := 0
	for {
		sx, sy, sw, sh := robotgo.GetDisplayBounds(screenIndex)
		if sw != 0 && sh != 0 {
			monitors = append(monitors, []int{sx, sy, sw, sh})
			screenIndex++
		} else {
			break
		}
	}

	minX := 0
	minY := 0
	maxTotalW := 0
	maxTotalH := 0
	for _, monitor := range monitors {
		if minX > monitor[0] {
			minX = monitor[0]
		}
		if minY > monitor[1] {
			minY = monitor[1]
		}
		if maxTotalW < monitor[2] {
			maxTotalW = monitor[0] + monitor[2]
		}
		if maxTotalH < monitor[3] {
			maxTotalH = monitor[1] + monitor[3]
		}
	}

	// negative coordinates (screen on top left of main screen)
	maxTotalW -= minX
	maxTotalH -= minY

	return []int{minX, minY, maxTotalW, maxTotalH}
}

var allScreenBound = findAllScreenBounds()

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	fmt.Println(allScreenBound)

	cfg, err := ini.Load(configPath)
	if err != nil {
		cfg = ini.Empty()
		cfgGlobal = cfg
		err := cfg.SaveTo(configPath)
		if err != nil {
			// Handle error
			fmt.Println("Error saving INI file:", err)
			return
		}

	}
	cfgGlobal = cfg

	DEFAULT_DELAY_ENTRY = cfg.Section("setting").Key("DefaultDelay").String()
	DEFAULT_IMG_TOLERANCE_ENTRY = cfg.Section("setting").Key("DefaultImageTolerance").String()
	selectedScript = cfg.Section("setting").Key("SelectedScript").String()
	mainApp = app.NewWithID("com.vietanht.autobot")
	// mainApp.Preferences().SetBool("Boolean", true)

	mainWindow = mainApp.NewWindow("Auto bot")
	mainWindow.SetMaster()

	windowWidth, err := cfg.Section("setting").Key("WindowWidth").Int()
	if err != nil {
		windowWidth = 600
	}
	windowHeight, err := cfg.Section("setting").Key("WindowHeight").Int()
	if err != nil {
		windowHeight = 400
	}

	mainWindow.Resize(fyne.Size{Width: float32(windowWidth), Height: float32(windowHeight)})

	mainApp.Settings().SetTheme(theme.LightTheme())

	tabs := container.NewAppTabs(
		container.NewTabItem("Play", makeMainContainer()),
		container.NewTabItem("Record Simple", makeSimpleScriptContainer()),
		container.NewTabItem("Script Editor", makeScriptEditorContainer()),
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	mainWindow.SetContent(tabs)

	currentposwindow := createCurrentMousePosWindow(mainApp)
	currentposwindow.Show()

	mainWindow.Show()
	mainApp.Run()

	defer func() {
		// cfgGlobal.Section("setting").Key("DefaultDelay").SetValue(scriptIntervalEntry.Text)
		// cfgGlobal.Section("setting").Key("DefaultImageTolerance").SetValue(imageToleranceEntry.Text)
		cfgGlobal.Section("setting").Key("WindowWidth").SetValue(strconv.FormatFloat(float64(mainWindow.Canvas().Size().Width), 'f', -1, 32))
		cfgGlobal.Section("setting").Key("WindowHeight").SetValue(strconv.FormatFloat(float64(mainWindow.Canvas().Size().Height), 'f', -1, 32))
		cfgGlobal.Section("setting").Key("SelectedScript").SetValue(selectedScript)

		// fmt.Println(mainWindow.Canvas().Size())

		err := cfgGlobal.SaveTo(configPath)
		if err != nil {
			// Handle error
			fmt.Println("Error saving INI file:", err)
			return
		}
	}()
}
