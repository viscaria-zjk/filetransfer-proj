package main

import (
	"conc_server/cli"
	"conc_server/server"
	"os"
	"os/signal"
)

func main() {

	// 检测用户是否有点按ctrl+c，点按了就启动清除程式
	c := make(chan os.Signal,1)
	signal.Notify(c,os.Interrupt)
	go func() {
		for _ = range c {
			server.Cleanser()
		}
	}()

	// 运行客户端程序
	cli.RunCli()
}
