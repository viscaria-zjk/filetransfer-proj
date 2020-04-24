package cli

import (
	"conc_server/pref"
	"conc_server/server"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

// 运行命令行程序
func RunCli() {

	// 定义变量用于接收控制台输入的值
	var workingDir string
	var monitorAddr string
	var monitorPort string

	// 新建命令行程式
	app := cli.NewApp()
	app.Name = "server"
	app.Usage = "The server terminal programme for conc"
	app.UsageText = `server [working-dir] [monitor-addr] [monitor-tcp-port]`
	app.Copyright = "Copyright (c) 2020 Han Li Studios."
	app.Version = "1.0.0"

	// 定义命令行程序主要的工作
	app.Action = func(c *cli.Context) error {
		if c.NArg() < 3 {
			// 打印错误
			pref.HdrErr()
			fmt.Println("No sufficient arguments.")
			pref.HdrInf()
			fmt.Println("Try", os.Args[0], "--help to get arguments hint.")
		} else {
			workingDir = c.Args().Get(0)
			monitorAddr = c.Args().Get(1)
			monitorPort = c.Args().Get(2)

			// 调用程序
			isOK := server.RunConc(workingDir, monitorAddr, monitorPort)
			if isOK == false {
				pref.HdrErr()
				fmt.Println("CONC is not running successfully.")
				return nil
			}

			// 调用清洁程序
			server.Cleanser()
		}
		return nil
	}

	//执行程序
	app.Run(os.Args)
}
