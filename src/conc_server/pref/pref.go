package pref

import (
	"github.com/gookit/color"
)

func HdrErr() {
	color.Red.Print("[ERROR] ")
}
func HdrInf() {
	color.Green.Print("[INFO] ")
}
func HdrWrn() {
	color.Yellow.Print("[WARNING] ")
}
