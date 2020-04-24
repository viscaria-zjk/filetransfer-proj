package pref

import (
	"github.com/gookit/color"
)

// 监视器控制端口5656
const MonitorPort string = "5656"

// 错误信息收集端口9898
const ErrInfPort string = "9898"

func HdrErr() {
	color.Red.Print("[ERROR] ")
}
func HdrInf() {
	color.Green.Print("[INFO] ")
}
func HdrWrn() {
	color.Yellow.Print("[WARNING] ")
}
