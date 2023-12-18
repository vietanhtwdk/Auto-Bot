package main

import (
	"os"
	"strconv"

	"github.com/go-vgo/robotgo"
)

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

func main() {
	args := os.Args[1:]
	x := parseInt(args[0])
	y := parseInt(args[1])
	switch len(args) {
	case 2:
		robotgo.MoveSmooth(x, y, 0.999999999999999, 1.0)
	case 3:
		robotgo.MoveSmooth(x, y, parseFloat(args[2]), 1.0)
	case 4:
		robotgo.MoveSmooth(x, y, parseFloat(args[2]), parseFloat(args[3]))
	}
}
