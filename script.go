package main

import (
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"path/filepath"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

const DefaultScriptPath = "./scripts/"
const DefaultScriptFileName = "script.json"
const OrderScriptName = "order"

var ActionValue = []string{"findImage", "findImageMove", "findImageMoveSmooth", "moveMouse", "moveMouseSmooth", "moveMouseRelative", "click", "dragMouse", "moveSmoothRelative", "actions", "delay"}

func executeActions(repeating *bool, jsonStruct []Action, scriptName string, startStep any) {
	currentStep := jsonStruct[0].Id
	if startStep != nil {
		currentStep = startStep.(int)
	}

	var scriptRelativePath = "./scripts/" + scriptName + "/"
	for *repeating {
		fmt.Println("currentStep", currentStep)
		if currentStep == -1 {
			*repeating = false
			return
		}

		index := slices.IndexFunc(jsonStruct, func(a Action) bool { return a.Id == currentStep })
		actionRepeat := jsonStruct[index].Repeat
		if actionRepeat < 1 {
			actionRepeat = 1
		}
		i := 0
		for i < actionRepeat && *repeating {
			stepSuccess := true
			switch jsonStruct[index].Action {
			case "findImage":
				img, err := robotgo.Read(scriptRelativePath + jsonStruct[index].Data)
				if err != nil {
					fmt.Println(err)
					return
				}
				currentScreen := getSearchArea()
				fx, fy := findImageFromImage(currentScreen, robotgo.ToCBitmap(robotgo.ImgToBitmap(img)), jsonStruct[index].ImgTolerance)
				robotgo.FreeBitmap(currentScreen)
				if fx == -1 || fy == -1 {
					fmt.Println("not found")
					currentStep = jsonStruct[index].False
					stepSuccess = false
					break
				}
				currentStep = jsonStruct[index].True
			case "findImageMove":
				currentScreen := getSearchArea()
				for _, imgName := range strings.Split(jsonStruct[index].Data, " ") {
					img, err := robotgo.Read(scriptRelativePath + imgName)
					if err != nil {
						fmt.Println(err)
						return
					}
					fx, fy := findImageFromImage(currentScreen, robotgo.ToCBitmap(robotgo.ImgToBitmap(img)), jsonStruct[index].ImgTolerance)
					if fx == -1 || fy == -1 {
						fmt.Println("not found")
						currentStep = jsonStruct[index].False
						stepSuccess = false
						continue
					}
					stepSuccess = true
					robotgo.Move(fx, fy)
					currentStep = jsonStruct[index].True
					break
				}
				robotgo.FreeBitmap(currentScreen)
			case "findImageMoveSmooth":
				currentScreen := getSearchArea()
				for _, imgName := range strings.Split(jsonStruct[index].Data, " ") {
					img, err := robotgo.Read(scriptRelativePath + imgName)
					if err != nil {
						fmt.Println(err)
						return
					}
					fx, fy := findImageFromImage(currentScreen, robotgo.ToCBitmap(robotgo.ImgToBitmap(img)), jsonStruct[index].ImgTolerance)
					if fx == -1 || fy == -1 {
						fmt.Println("not found")
						currentStep = jsonStruct[index].False
						stepSuccess = false
						continue
					}
					stepSuccess = true
					runMoveMouseSmooth(fx+searchAreaX, fy+searchAreaY, 0.999999999999999, 1.0)
					currentStep = jsonStruct[index].True
					break
				}
				robotgo.FreeBitmap(currentScreen)
			case "moveMouse":
				strArr := strings.Split(jsonStruct[index].Data, " ")
				fx, err := strconv.ParseInt(strArr[0], 10, 64)
				fy, err := strconv.ParseInt(strArr[1], 10, 64)

				if err != nil {
					fmt.Println(err)
					return
				}
				robotgo.Move(int(fx), int(fy))
				currentStep = jsonStruct[index].True
			case "moveMouseSmooth":
				strArr := strings.Split(jsonStruct[index].Data, " ")
				fx, err := strconv.ParseInt(strArr[0], 10, 64)
				fy, err := strconv.ParseInt(strArr[1], 10, 64)

				if err != nil {
					fmt.Println(err)
					return
				}
				runMoveMouseSmooth(int(fx), int(fy), 0.999999999999999, 1.0)
				currentStep = jsonStruct[index].True
			case "moveMouseRelative":
				strArr := strings.Split(jsonStruct[index].Data, " ")
				fx, err := strconv.ParseInt(strArr[0], 10, 64)
				fy, err := strconv.ParseInt(strArr[1], 10, 64)

				if err != nil {
					fmt.Println(err)
					return
				}
				robotgo.MoveRelative(int(fx), int(fy))
				currentStep = jsonStruct[index].True
			case "click":
				robotgo.Click()
				fmt.Println("clicked")
				currentStep = jsonStruct[index].True
			case "dragMouse":
				strArr := strings.Split(jsonStruct[index].Data, " ")
				fx, err := strconv.ParseInt(strArr[0], 10, 64)
				fy, err := strconv.ParseInt(strArr[1], 10, 64)
				if err != nil {
					fmt.Println(err)
					return
				}
				robotgo.DragSmooth(int(fx), int(fy))
				currentStep = jsonStruct[currentStep].True
			case "moveSmoothRelative":
				strArr := strings.Split(jsonStruct[index].Data, " ")
				fx, err := strconv.ParseInt(strArr[0], 10, 64)
				fy, err := strconv.ParseInt(strArr[1], 10, 64)
				if err != nil {
					fmt.Println(err)
					return
				}
				cx, cy := robotgo.Location()
				runMoveMouseSmooth(cx+int(fx), cy+int(fy), 0.999999999999999, 1.0)
				currentStep = jsonStruct[index].True
			case "actions":
				executeActions(repeating, jsonStruct[index].Actions, scriptName, nil)
				currentStep = jsonStruct[index].True
			case "delay":
				delayTime, err := strconv.ParseInt(jsonStruct[index].Data, 10, 64)
				if err != nil {
					fmt.Println(err)
					return
				}
				time.Sleep(time.Duration(delayTime) * time.Millisecond)
				currentStep = jsonStruct[index].True
			default:
				return
			}
			if stepSuccess {
				if jsonStruct[index].TrueClick {
					fmt.Println("clicked")
					robotgo.Click()
				}
			}
			if jsonStruct[index].Click {
				fmt.Println("clicked")
				robotgo.Click()
			}
			i++
			fmt.Println(time.Now())
			delay := jsonStruct[index].Delay
			fmt.Println("delay", delay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

func loadScript(scriptName string) ([]Action, error) {
	fmt.Println("Selected option:", scriptName)
	content, err := os.ReadFile(DefaultScriptPath + scriptName + "/" + DefaultScriptFileName)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var jsonStruct []Action

	err = json.Unmarshal(content, &jsonStruct)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return jsonStruct, nil
}

func runScript(scriptName string, startStep any) {
	actions, err := loadScript(scriptName)

	if err != nil {
		fmt.Println(err)
		return
	}

	repeating := true
	go func() {
		if startStep != nil {
			executeActions(&repeating, actions, scriptName, startStep)
		} else {
			executeActions(&repeating, actions, scriptName, nil)
		}
	}()

	eventHook := hook.Start()
	var e hook.Event
	// var key string
	for e = range eventHook {

		if e.Kind == hook.KeyDown {
			if e.Keychar == 27 || e.Keychar == 96 {
				mainWindow.Show()
				mainWindow.RequestFocus()
				repeating = false
				if moveMouseSmoothProcess != nil {
					moveMouseSmoothProcess.Process.Kill()
				}
				moveMouseSmoothProcess = nil
				runScriptButton.Enable()
				fmt.Println("ended")
				break
			}
			// key = string(e.Keychar)
		}
	}
	hook.End()
}

func saveScript(scriptName string, actions []Action) {
	scriptPath := fmt.Sprintf("./scripts/%s/script.json", scriptName)
	err := os.MkdirAll(filepath.Dir(scriptPath), 0755)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(scriptPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(actions)
	if err != nil {
		panic(err)
	}
}

func checkThenSaveScript(scriptName string, actions []Action) bool {
	_, err := os.Stat(fmt.Sprintf("./scripts/%s/script.json", scriptName))
	if err != nil {
		saveScript(scriptName, actions)
		return true
	}
	return false
}

func deleteScript(scriptName string) {
	fmt.Println(scriptName)
	err := os.RemoveAll(fmt.Sprintf("./scripts/%s", scriptName))
	if err != nil {
		fmt.Println("Error deleting folder and its contents:", err)
		return
	}
}
