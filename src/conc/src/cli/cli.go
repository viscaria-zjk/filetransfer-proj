package cli

import (
	"conc/src/client"
	"conc/src/pref"
	"conc/src/server"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"runtime"
)


// 运行命令行程序
func RunCli() {
	// 读取配置文件
	config, isOK := pref.GlobalConfig.ReadConf()
	if isOK != true {
		return
	}
	lang, isOK := pref.GlobalDesc.ReadDesc()
	if isOK != true {
		return
	}

	//fmt.Println(pref.GlobalConfig.OverwriteWhenSAM, pref.GlobalConfig.ServerListenAddr, pref.GlobalConfig.ServerListenPort, pref.GlobalConfig.ServerBufferSize, pref.GlobalConfig.ErrInfPort, pref.GlobalConfig.MonitorPort)

	// 读取软件所在地址
	var currentDir = func() string {
		dir, err := os.Getwd()
		if err == nil {
			return dir
		} else {
			return "."
		}
	}()

	var usageTxt string
	// 根据系统不同，定义不同的usage
	switch runtime.GOOS {
	case "darwin":
		// macOS
		usageTxt = pref.GlobalDesc.AppUsgTxtMacOS
		break
	case "windows":
		// Windows
		usageTxt = pref.GlobalDesc.AppUsgTxtWin
		break
	default:
		usageTxt = pref.GlobalDesc.AppUsgTxtLinux
		break
	}

	// 定义作者
	aut := func() []cli.Author {
		if pref.GlobalConfig.AppLang == "zh-CN" {
			return []cli.Author {
				{Name: "Li Han", Email: "han@han-li.cn"},
			}
		} else {
			return []cli.Author {
				{Name: "Li Han", Email: "phantef@gmail.com"},
			}
		}
	}()

	// 定义变量用于接收控制台输入的值
	var serverListenAddr string
	var serverListenPort string
	var serverWorkingDir string
	var serverBufferSize int64

	var clientMonitorPort string
	var clientErrInfoPort string

	// 新建命令行程式
	app := cli.NewApp()
	app.Name = "conc"
	app.Usage = lang.AppDesc
	app.UsageText = usageTxt
	app.Authors = aut
	app.Copyright = "Copyright (c) 2020 Han Li Studios."
	app.Version = pref.GlobalConfig.AppVersion
	//重点可以设置一些选项操作
	//第一个是一个字符串的选项，第二个是一个布尔的选项
	app.Commands = []cli.Command {
		// Server
		{
			Name: "server",
			Aliases: []string{"sv"},
			Usage: pref.GlobalDesc.AppServerUsg,
			// 接收服务器监听地址
			Flags: []cli.Flag {
				cli.StringFlag {
					Name:        "listen-addr,a",
					Value:       config.ServerListenAddr,
					Usage:       pref.GlobalDesc.AppServerAddrUsg,
					Destination: &serverListenAddr,
				},
				// 接收服务器监听的端口
				cli.StringFlag {
					Name:        "listen-port,p",
					Value:       config.ServerListenPort,
					Usage:       pref.GlobalDesc.AppServerPortUsg,
					Destination: &serverListenPort,
				},
				cli.StringFlag {
					Name:        "working-dir,w",
					Value:       currentDir,
					Usage:       pref.GlobalDesc.AppServerWkdrUsg,
					Destination: &serverWorkingDir,
				},
				cli.Int64Flag {
					Name:        "buffer-size,b",
					Value:       config.ServerBufferSize,
					Usage:       pref.GlobalDesc.AppServerBuffUsg,
					Destination: &serverBufferSize,
				},
			},

			// 定义动作
			Action: func(c *cli.Context) error {
				server.Server(serverWorkingDir, serverListenAddr, serverListenPort, serverBufferSize)
				return nil
			},
		},

		// client command
		{
			Name: "client",
			Aliases: []string{"cl"},
			Usage: pref.GlobalDesc.AppClientUsg,
			Description: pref.GlobalDesc.AppClientDesc,

			// 接收监视器控制端口、错误信息收集端口
			Flags: []cli.Flag {
				cli.StringFlag {
					Name:        "mon-port, m",
					Value:       config.MonitorPort,
					Usage:       pref.GlobalDesc.AppClientMonUsg,
					Destination: &clientMonitorPort,
				},
				// 接收服务器监听的端口
				cli.StringFlag {
					Name:        "errinf-port, e",
					Value:       config.ErrInfPort,
					Usage:       pref.GlobalDesc.AppClientErrUsg,
					Destination: &clientErrInfoPort,
				},
			},

			// 定义Client的动作
			Action: func(c *cli.Context) error {
				ret := client.NewMonitor(true, clientMonitorPort, clientErrInfoPort)
				pref.HdrInf()
				fmt.Println(lang.InfExit, ret)
				return nil
			},
		},
	}


	// 定义命令行程序主要的工作
	app.Action = func(c *cli.Context) error {
		if c.NArg() == 0 {
			// 打印错误
			pref.HdrErr()
			fmt.Println(lang.ErrNotSuffArgs)
			pref.HdrInf()
			fmt.Println(lang.InfTryConcHelp)
		}
		return nil
	}

	//执行程序
	app.Run(os.Args)
}
