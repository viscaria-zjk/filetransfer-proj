package client

import (
	"conc/src/pref"
	"conc/src/utils"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"net"
	"os"
	"os/signal"
)

type MonType byte

var errorReplyPipe chan int = make(chan int)
var exitPipe chan int = make(chan int)

// 异常信号
const (
	SIG_LOSTCONN int32 = 0	// 提示连接丢失
	SIG_ERR int32 = 1		// 提示未知错误
	SIG_PANIC int32 = 2		// 提出严重错误
	SIG_TIMEOUT int32 = 3	// 提出超时信号
)

// 客户端种类
const (
	MON_CLIENT MonType = 0
	MON_SERVER MonType = 1
)

// 接收指令协程的回传
const (
	TUN_ERR          int = -1
	TUN_SET_CLIENT   int = 0
	TUN_SET_SERVER   int = 1
	TUN_CLOSE_CONN   int = 2
	TUN_SEND         int = 3
	TUN_SEEK         int = 4
	TUN_CURRFILENAME int = 5
	TUN_TARGETHOST   int = 6
	TUN_TARGETPORT   int = 7
	TUN_TOTALSIZE	 int = 8
	TUN_EXITPROG	 int = 9
	TUN_CONNECT		 int = 10
	TUN_HELLO		 int = 11
)

const (
	ASK_OVRWHENNOS int32 = 1
)

// 一个Monitor类
type Monitor struct {
	// 标识
	MonitorType        MonType  // 判断是服务器监控还是用户监控
	IsTransferring     bool     // 判断是否正在传输
	IsTransferComplete bool     // 判断传输任务是否完成
	IsTaskDir          bool     // 判断是否正传输文件夹
	ClientTaskAbsPath  string   // 若是用户端，则指出任务的本机目录
	ClientTask         *Task    // 客户端的任务对象
	ServerWorkingDir   string   // 若是服务器，则指出服务器工作目录
	ServerBuffer       int32    // 若是服务器，则指出服务器缓冲区
	TargetAddr         string   // 目标主机
	TargetPort         string   // 目标端口
	MonitorConn        net.Conn // 和程序通讯套接字
	ErrConn            net.Conn // 和程序传递错误套接字

	// 状态
	CurrentTaskFileName string // 当前任务正传输文件名
	CurrentTaskFileNum  int32  // 当前文件编号
	CurrentTaskSize     int64  // 当前任务（文件、文件夹）的总大小
	CurrentTaskProgress int64  // 当前任务（文件、文件夹）已传输的总大小
}


// 通知新建一个客户端传输任务，返回任务是否执行成功
func (monitor *Monitor) NewClientTask() error {
	if monitor.MonitorType != MON_CLIENT {
		return errors.New("005 Type is not fit")
	}
	// 新建一个客户端实例
	monitor.ClientTask = new(Task)
	err := monitor.ClientTask.New(monitor)
	if err != nil {
		return err
	}
	// 若err = nil，则报告连接成功
	monitor.ReportInfo("001 Connected to remote host.")

	return nil
}

func (monitor *Monitor) DoClientTask() {
	err := monitor.ClientTask.DoTask(monitor)
	if err != nil {
		monitor.ReportError(err.Error())
	} else {
		monitor.ReportComplete()
	}
}


func (monitor *Monitor) monitorErrChannel(pipe chan int, exit chan int) {
	buf := make([]byte, 30)
	for {
		select {
		case <- exit:
			return
		default:
			_, _ = monitor.ErrConn.Read(buf)
			if string(buf[0:2]) == "EX" {
				// EX信号标示退出程式
				monitor.cleanser()
			} else if string(buf[0:2]) == "OK" {
				// 其他讯息，例如OK
				pipe <- 1
			} else if string(buf[0:2]) == "NO" {
				pipe <- 0
			}
			continue
		}


	}
}

// 清除函数
func (mon *Monitor) cleanser() {
	// 强制关闭连接
	exitPipe <- 1
	pref.HdrInf()
	fmt.Println("User forces to exit. Cleaning caches... ")
	if mon != nil {
		if mon.MonitorConn != nil {
			mon.MonitorConn.Close()
		}
		if mon.ErrConn != nil {
			mon.ErrConn.Close()
		}
	}
	// 强制关闭程序
	os.Exit(1)
}

func (monitor *Monitor) ReportError(errstr string) {
	// 向上级报告错误
	buf := make([]byte, len(errstr) + 3)
	copy(buf[0:], "ER ")
	copy(buf[3:], errstr)
	fmt.Println(string(buf))
	_, _ = monitor.ErrConn.Write(buf)
	// Read由协程完成
	<- errorReplyPipe
}

func (monitor *Monitor) ReportInfo(infoStr string) {
	// 向上级报告信息
	buf := make([]byte, len(infoStr) + 3)
	copy(buf[0:], "IF ")
	copy(buf[3:], infoStr)
	fmt.Println(string(buf))
	_, _ = monitor.ErrConn.Write(buf)
	// Read由协程完成
	<- errorReplyPipe
}

func (monitor *Monitor) ReportComplete() {
	// 向上级报告完成
	_, _ = monitor.ErrConn.Write([]byte("CP "))
	// Read由协程完成
	<- errorReplyPipe
	exitPipe <- 1 // 通知错误处理协程退出
}

func (monitor *Monitor) Ask(quesLen int, quesString string) (answer bool, err error) {
	// 向上级询问信息
	buf := make([]byte, 10)
	copy(buf[0:], "QU ")
	copy(buf[3:], utils.Int32ToBytes(int32(quesLen)))
	_, _ = monitor.ErrConn.Write(buf)
	<- errorReplyPipe
	_, _ = monitor.ErrConn.Write([]byte(quesString))
	reply := <- errorReplyPipe
	// 返回0（不）1（好）
	if reply == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

// 新建客户端或服务器端监控器
// 本地提供一个检视器TCP端口
func NewMonitor(isClient bool, monPort string, errPort string) (ret int32) {
	// 新建一个监控器
	var mon Monitor
	// 各种连接
	var errConn, monConn net.Conn

	// 识别Ctrl+C退出
	c := make(chan os.Signal,1)
	signal.Notify(c,os.Interrupt)
	go func() {
		for _ = range c {
			cleanser(mon)
		}
	}()

	pref.HdrInf()
	fmt.Println("Waiting error collector to connect on " + ":"+errPort+ "...")
	// 1. 尝试等待错误收集端的连接
	errListener, err := net.Listen("tcp", ":"+errPort)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Error occ when creating monlistener for err collect: ", err)
		return 0
	}
	defer errListener.Close()

	// 尝试接收连接
	errConn, err = errListener.Accept()
	if err != nil {
		pref.HdrErr()
		fmt.Println("Accept error: ", err)
		return 0
	}
	// 发生错误或连接完毕时关闭通讯
	defer errConn.Close()

	pref.HdrInf()
	fmt.Println("Waiting monitor to connect on " + "localhost:"+monPort+ "...")
	// 2. 尝试等待主控的连接
	monlistener, err := net.Listen("tcp", ":"+monPort)
	if err != nil {
		return 0
	}
	defer monlistener.Close()

	// 尝试接收连接
	monConn, err = monlistener.Accept()
	if err != nil {
		return 0
	}
	defer monConn.Close()

	if isClient {
		mon.MonitorType = MON_CLIENT
	} else {
		mon.MonitorType = MON_SERVER
	}
	mon.MonitorConn = monConn
	mon.ErrConn = errConn

	// 设置错误及信息监听器
	go mon.monitorErrChannel(errorReplyPipe, exitPipe)

	pref.HdrInf()
	fmt.Println("All set. Awaiting instructions from monitor...")
	// 新建协程，等待传入用户命令，及完成各种查询动作
	tun := make(chan int)	// 新建一个和协程通讯的管道
	buf := make([]byte, 1024)	// 新建缓冲器
	// 不停地检测5656端口是否有命令回复
	for {
		go acceptInst(monConn, tun, &buf)
		// 接收传来的指令信息
		inst := <- tun
		// 回复协程（我收到了）
		tun <- 1
		switch inst {
		case TUN_ERR:
			break
		case TUN_SET_CLIENT: // 0
			mon.MonitorType = MON_CLIENT
			break
		case TUN_SET_SERVER:  // 0
			mon.MonitorType = MON_SERVER
			break
		case TUN_SEND:// 0
			// 要求acceptInst自动接收目录地址
			infoLen := <- tun
			tun <- 1
			<- tun
			// fmt.Println("Set task path: "+string(buf[0:infoLen]) + ", start send.")
			mon.ClientTaskAbsPath = string(buf[0:infoLen])
			// 然后开始任务
			go mon.DoClientTask()
			break
		case TUN_CONNECT: // 0
			// 要求连接（即newTask）
			err = mon.NewClientTask()
			if err != nil {
				mon.ReportError(err.Error())
			}
			break
		case TUN_SEEK: //
			// 查询发送进度，就返回进度
			_, _ = monConn.Write(utils.Int64ToBytes(mon.CurrentTaskProgress))
			break
		case TUN_CURRFILENAME: //
			// 查询目前文件名，先返回文件名长度，再返回文件名
			_, _ = monConn.Write(utils.Int32ToBytes(int32(len(mon.ClientTask.TaskName))))
			_, _ = monConn.Read(buf)
			_, _ = monConn.Write([]byte(mon.ClientTask.TaskName))
			break
		case TUN_TARGETHOST: //
			// 设置目标主机
			infoLen := <- tun
			tun <- 1
			<- tun
			pref.HdrInf()
			fmt.Println("Host addr was set to: ", string(buf[0:infoLen]))
			mon.TargetAddr = string(buf[0:infoLen])
			break
		case TUN_TARGETPORT: //
			// 设置目标端口
			infoLen := <- tun
			tun <- 1
			<- tun
			pref.HdrInf()
			fmt.Println("Host port was set to: ", string(buf[0:infoLen]))
			mon.TargetPort = string(buf[0:infoLen])
			break
		case TUN_TOTALSIZE:
			// 查询目前任务的总size
			_, _ = monConn.Write(utils.Int64ToBytes(mon.CurrentTaskSize))
			break
		case TUN_EXITPROG: //
			// 退出程式的唯一合法方法
			return 1
		case TUN_HELLO: //
			// 打招呼
			// fmt.Println("Host says: Hello")
			break
		default:
			// ignored
			break
		}
	}
}

func cleanser(monitor Monitor) {
	color.Green.Print("\nCONC INFO: ")
	fmt.Println("User attempts to exit. Cleaning caches...")
	// 关闭和服务器的连接
	if monitor.ClientTask != nil {
		if monitor.ClientTask.TargetConn != nil {
			monitor.ClientTask.TargetConn.Close()
		}
	}
	// 关闭和monitor的连接
	if monitor.MonitorConn != nil {
		monitor.MonitorConn.Close()
	}
	// 关闭和monitor错误收集器的连接
	if monitor.ErrConn != nil {
		monitor.ErrConn.Close()
	}
	os.Exit(1)
}

// 和主程序交流的协程（5656端口）
func acceptInst(conn net.Conn, tun chan int, buff *[]byte) {
	// 尝试读取
	n, err := conn.Read(*buff)
	// 如果读取错误或是读取到EOF就关闭监听器
	if err != nil || n < 4 {
		tun <- TUN_ERR
		<- tun
	}
	// 检查接收到的数据
	instr := string((*buff)[0:4])
	if instr == "SEND" {
		// 设置目标主机
		tun <- TUN_SEND
		<- tun
		n, err = conn.Write([]byte("OK"))
		// 自动接收infoLen
		n, err = conn.Read(*buff)
		tun <- int(utils.BytesToInt32((*buff)[0:4]))
		n, err = conn.Write([]byte("OK"))
		<- tun
		// 自动接收目录
		n, err = conn.Read(*buff)
		n, err = conn.Write([]byte("OK"))
		tun <- 1
	} else if instr == "CONN" {
		// 查询发送进度
		tun <- TUN_CONNECT
		<- tun
		// 回复ok
		n, err = conn.Write([]byte("OK"))
	} else if instr == "CLNT" {
		// 设置为用户端
		tun <- TUN_SET_CLIENT
		<- tun
		// 回复ok
		n, err = conn.Write([]byte("OK"))
	} else if instr == "SERV" {
		// 设置为服务器
		tun <- TUN_SET_SERVER
		<- tun
		// 回复ok
		n, err = conn.Write([]byte("OK"))
	} else if instr == "SEEK" {
		// 查询发送进度
		tun <- TUN_SEEK
		<- tun
	} else if instr == "CUFN" {
		// 目前任务文件名
		tun <- TUN_CURRFILENAME
		<- tun
	} else if instr == "RMAD" {
		// 设置目标主机
		tun <- TUN_TARGETHOST
		<- tun
		n, err = conn.Write([]byte("OK"))
		// 自动接收infoLen
		n, err = conn.Read(*buff)
		tun <- int(utils.BytesToInt32((*buff)[0:4]))
		n, err = conn.Write([]byte("OK"))
		<- tun
		// 自动接收hostname
		n, err = conn.Read(*buff)
		n, err = conn.Write([]byte("OK"))
		tun <- 1
	} else if instr == "RMPO" {
		// 设置目标主机
		tun <- TUN_TARGETPORT
		<- tun
		n, err = conn.Write([]byte("OK"))
		// 自动接收infoLen
		n, err = conn.Read(*buff)
		tun <- int(utils.BytesToInt32((*buff)[0:4]))
		n, err = conn.Write([]byte("OK"))
		<- tun
		// 自动接收hostname
		n, err = conn.Read(*buff)
		n, err = conn.Write([]byte("OK"))
		tun <- 1
	} else if instr == "TSZE" {
		// 查询目前任务的总大小
		tun <- TUN_TOTALSIZE
		<- tun
	} else if instr == "ENDP" || instr == "EXit" {
		// monitor退出
		tun <- TUN_EXITPROG
		n, err = conn.Write([]byte("OK"))
		color.Green.Print("\nCONC INFO: ")
		fmt.Println("User attempts to exit but no connections were established. Cleaning caches...")
		<- tun
	} else if instr == "HELO" {
		tun <- TUN_HELLO
		<- tun
		n, err = conn.Write([]byte("OK"))
	}

}