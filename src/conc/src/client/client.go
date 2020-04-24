package client

import (
	"conc/src/pref"
	"conc/src/utils"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

const localPort = "5656"
const localErrPort = "9898"

// 发送RES指令（和FIL指令相当），返回断点所在位置
func sendRES(fileNum int32, fileSize int64, checksum []byte, fileRelPath string, conn net.Conn) (int64, error) {
	// 准备发送RES指令
	lenOfInstr := 42 + len(fileRelPath)
	var instr = make([]byte, lenOfInstr)
	fileN := utils.Int32ToBytes(fileNum)
	fileS := utils.Int64ToBytes(fileSize)

	// 制作指令
	copy(instr[0:], "RES <")
	copy(instr[5:], fileN)
	copy(instr[9:], "><")
	copy(instr[11:], fileS)
	copy(instr[19:], "><")
	copy(instr[21:], checksum)
	copy(instr[37:], "><")
	copy(instr[41:lenOfInstr], fileRelPath+">")

	// 发送指令
	_, err := conn.Write(instr)
	buf := make([]byte, 1024)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Write: ", err)
		return 0, err
	}
	// 接收传回的信息
	n, err := conn.Read(buf)
	if n <= 3 {
		return 0, errors.New("ambiguous return instruction")
	}
	ret := string(buf[0:3])
	if ret == "EPT" {
		// EPT，存在该文件，再读8个字节找到文件位置
		return utils.BytesToInt64(buf[5:13]), nil
	} else {
		// 文件无效，再读出错误代码
		ret = string(buf[5:8])
		// 如果ret是957，表示这个文件根本不存在，您直接传就好了
		if ret == "957" {
			return 0, nil
		}
		// 其他情况均视为错误
		return 0, errors.New(ret + " error from remote")
	}
	// TODO - 不搞139警告
}

// 发送CHD指令
func sendCHD(relDir string, conn net.Conn) error {
	// 发送CHD指令
	var instr = make([]byte, 7+len(relDir))
	copy(instr[0:], "CHD <")
	copy(instr[5:], relDir+">")
	_, err := conn.Write(instr)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Write: ", err)
		return err
	}

	// 检视CHD回应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if n <= 3 {
		return errors.New("ambiguous return instruction")
	}
	ret := string(buf[0:3])
	if ret == "YES" {
		// 并无错误
		return nil
	} else {
		// 再读出错误代码
		ret = string(buf[5:8])
		if ret == "555" {
			// 555 Error occured while creating directory
			pref.HdrErr()
			fmt.Println("555 Error occured while creating directory.")
			return errors.New("Remote server refuses to create new directory.")
		} else {
			pref.HdrErr()
			fmt.Println("Unhandled error from remote: " + ret)
			return errors.New("Unhandled error from remote")
		}
	}
}

// 发送END指令
func sendEND(sentSize int64, conn net.Conn) error {
	// 发送END指令
	var instr = make([]byte, 15)
	send_s := utils.Int64ToBytes(sentSize)
	copy(instr[0:], "END <")
	copy(instr[5:12], send_s)
	copy(instr[13:14], ">")
	_, err := conn.Write(instr)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Write: ", err)
		return err
	}
	return nil
}

// 发送STR指令
func sendSTR(conn net.Conn) error {
	// 发送STR指令
	_, err := conn.Write([]byte("STR <File transfer is started!>"))
	if err != nil {
		pref.HdrErr()
		fmt.Println("Write: ", err)
		return err
	}
	// 接收回应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if n <= 3 {
		return errors.New("Ambiguous returned instruction")
	}
	ret := string(buf[0:3])
	if ret == "YES" {
		// STR启动完成
		return nil
	} else {
		// 再读出错误代码
		fmt.Println(string(buf[0:10]))
		return errors.New(ret + " error from remote")
	}

}

// 发送EOF指令
func sendEOF(fileNum int32, conn net.Conn) error {
	// 发送EOF指令
	var instr = make([]byte, 11)
	file_n := utils.Int32ToBytes(fileNum)
	copy(instr[0:], "EOF <")
	copy(instr[5:], file_n)
	copy(instr[9:], ">")
	_, err := conn.Write(instr)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Write: ", err)
		return err
	}
	return nil
}

// 发送FIL指令，返回服务器的返回值：0 ERR 1 YES 2 SAM 3 NOS
func sendFIL(fileNum int32, fileSize int64, checksum []byte, fileRelPath string, conn net.Conn) (byte, error) {
	// 准备发送FIL指令
	len_of_instr := 42 + len(fileRelPath)
	var instr = make([]byte, len_of_instr)
	file_n := utils.Int32ToBytes(fileNum)
	file_s := utils.Int64ToBytes(fileSize)

	// 制作指令
	copy(instr[0:], "FIL <")
	copy(instr[5:], file_n)
	copy(instr[9:], "><")
	copy(instr[11:], file_s)
	copy(instr[19:], "><")
	copy(instr[21:], checksum)
	copy(instr[37:], "><")
	copy(instr[41:len_of_instr], fileRelPath+">")

	// 发送指令
	_, err := conn.Write(instr)
	buf := make([]byte, 1024)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Write: ", err)
		return 0, err
	}
	// 接收YES指令/NO指令
	n, err := conn.Read(buf)
	if n <= 3 {
		return 0, errors.New("Ambiguous return instruction")
	}
	ret := string(buf[0:3])
	pref.HdrInf()
	fmt.Println("Check if file exists...")
	if ret == "YES" {
		// FIL回应成功，传输可以开始
		return 1, nil
	} else if ret == "SAM" {
		// 回应SAM，表示远端已经存在该文件，但文件一样
		return 2, nil
	} else if ret == "NOS" {
		// 回应NOS，表示远端已经存在该文件，且文件不一样
		return 3, nil
	} else {
		// 再读出错误代码
		ret = string(buf[5:8])
		if ret == "139" {
			// TODO - 139错误属于警告：file no有问题
			pref.HdrErr()
			fmt.Println("Received 139 warning. Ignoring...")
			// 回复139 21/4/2020更新
			conn.Write([]byte("YES <ERR>"))
		} else {
			return 0, errors.New(ret + " error from remote")
		}
	}
	// TODO - 如果收到的是139等警告，就再继续接收信息
	n, err = conn.Read(buf)
	if n <= 3 {
		return 0, errors.New("Weird return instruction")
	}
	ret = string(buf[0:3])
	fmt.Println("Received", ret, "beyond 139 warning.")
	if ret == "YES" {
		// FIL回应成功，传输可以开始
		return 1, nil
	} else if ret == "SAM" {
		// 回应SAM，表示远端已经存在该文件，但文件一样
		return 2, nil
	} else if ret == "NOS" {
		// 回应NOS，表示远端已经存在该文件，且文件不一样
		return 3, nil
	} else {
		// 再读出错误代码
		ret = string(buf[5:8])
		return 0, errors.New(ret + " error from remote")
	}
}

// 发送一个资料夹
func SendOneFolder(task *Task) (int64, int64, error) {
	task.mon.ClientTaskAbsPath = strings.TrimSuffix(task.mon.ClientTaskAbsPath, string(os.PathSeparator))
	// 检查该文件夹，获取list.send文件
	remoteHost := strings.Replace(task.TargetConn.RemoteAddr().String(), ":", "_", 1)
	paths, haveFolder, folderSize, err := utils.GetListSend(remoteHost, task.mon.ClientTaskAbsPath)
	if paths == nil || err != nil || !haveFolder {
		// 递归错误
		return 0, folderSize, errors.New("007 Error fetching file info or list.send error")
	}
	// 假定STR命令已经有效
	info, err := os.Stat(task.mon.ClientTaskAbsPath)
	if err != nil {
		// 递归错误
		return 0, folderSize, errors.New("002 Error opening root path")
	}
	// 获取根目录名字
	rootDirName := info.Name()
	// 创建该目录及切换工作路径
	err = sendCHD(rootDirName, task.TargetConn)
	if err != nil {
		// 发送CHD出现错误，退出
		return 0, folderSize, errors.New("006 Error sending CHD")
	}

	task.mon.ReportInfo("002 " + info.Name())
	task.mon.CurrentTaskSize = folderSize
	// 从paths中依次读取
	var sentSize int64 = 0
	// 从根目录下的第1条记录开始
	for _, p := range paths[1:] {
		i := 6
		j := len(p) - 1
		for ; p[i] != '>'; i++ { // 找第1个“>”符号，以拿出数字
		}
		for ; p[j] != '>'; j-- { // 找最后1个“>”符号，以拿出目录
		}
		int_num, _ := strconv.Atoi(p[5:i]) // 拿到文件/文件夹编号
		dir_fil := p[0:3]
		rel_path := p[i+2 : j] // 拿到相对于根目录的路径
		if dir_fil == "DIR" {
			// 是文件夹记录，切换到文件夹
			err = sendCHD(rel_path, task.TargetConn)
			if err != nil {
				// 发送CHD出现错误，退出
				return 0, folderSize, errors.New("006 Error sending CHD")
			}
			// 14/4/2020
			//path = rootPath + string(os.PathSeparator) + rel_path
		} else if dir_fil == "FIL" {
			// 是文件记录，发送文件
			ss, err := SendOneFile(true, task.mon.ClientTaskAbsPath+string(os.PathSeparator)+rel_path, rel_path, task, int32(int_num))
			if err != nil {
				// 发送FIL出现错误，退出
				return 0, folderSize, errors.New("006 Error sending FIL inside folder")
			}
			sentSize += ss
		} else {
			// 发送list.send文件结构出现错误，退出
			return 0, folderSize, errors.New("008 Invalid list.send")
		}
	}
	task.mon.ReportInfo("003 File transfer is over.")
	// 发送完毕，并无错误
	pref.HdrInf()
	fmt.Println("Sent", task.mon.ClientTaskAbsPath)
	return sentSize, folderSize, nil

}

// 以绝对路径断点续传1个文件
// 不取缔fileAbsolutePath的原因是：为了让在发送文件夹的时候也能调用这个程序
func ResumeSendOneFile(fileAbsolutePath string, relPath string, task *Task, fileNum int32) (int64, error) {
	info, err := os.Stat(fileAbsolutePath)
	if err != nil {
		pref.HdrErr()
		fmt.Println("File is not accessible: ", err)
		return 0, errors.New("002 No such file or directory / permission denied" + err.Error())
	}

	// 以只读方式打开文件
	f, err := os.Open(fileAbsolutePath)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Appointing root path: ", err)
		return 0, errors.New("002 No such file or directory / permission denied" + err.Error())
	}
	defer f.Close()
	// 拿取文件MD5
	pref.HdrInf()
	fmt.Println("Analyzing file " + info.Name() + "...")
	chksum, err := utils.GetChksum(fileAbsolutePath)
	pref.HdrInf()
	fmt.Println("file MD5 got.")
	if err != nil {
		pref.HdrErr()
		fmt.Println("Error occured while analyzing file: ", err)
		return 0, errors.New("010 Error while fetching MD5")
	}
	// 准备发送RES指令
	pref.HdrInf()
fmt.Println("Sending RES...")
	seek, err := sendRES(fileNum, info.Size(), chksum, relPath, task.TargetConn)
	if err != nil {
		// 发送RES出现错误，退出
		pref.HdrErr()
		fmt.Println("Error on RES: ", err)
		return 0, errors.New("006 Error while sending RES")
	}

	// 如果返回的断点比本文件还要大，则认为seek为0，重传
	if seek > info.Size() {
		seek = 0
	}

	// 跳到响应位置续传，如果seek > 0
	if seek > 0 {
		return resumeFileProcess(seek, fileNum, f, &info, task)
	} else {
		return sendFileProcess(false, fileNum, f, &info, task)
	}
}

// 以绝对路径发送1个文件
func SendOneFile(isSendingFolder bool, fileAbsolutePath string, relPath string, task *Task, fileNum int32) (int64, error) {
	info, err := os.Stat(fileAbsolutePath)
	if err != nil {
		pref.HdrErr()
		fmt.Println("File is not accessible: ", err)
		return 0, err
	}
	// 制作list.send - debug on 21/4/2020
	if !isSendingFolder {
		remoteHost := strings.Replace(task.TargetConn.RemoteAddr().String(), ":", "_", 1)
		_, _, _, err = utils.GetListSend(remoteHost, fileAbsolutePath)
	}
	if err != nil {
		pref.HdrErr()
		fmt.Println("list.send created error: ", err)
		return 0, err
	}

	// 以只读方式打开文件
	f, err := os.Open(fileAbsolutePath)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Appointing root path: ", err)
		return 0, err
	}
	defer f.Close()
	// 拿取文件MD5
	pref.HdrInf()
	fmt.Println("Analyzing file " + info.Name() + "...")
	chksum, err := utils.GetChksum(fileAbsolutePath)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Error occured while analyzing file: ", err)
		return 0, err
	}
	// 准备发送FIL指令
	pref.HdrInf()
	fmt.Println("Sending FIL...")
	ret, err := sendFIL(fileNum, info.Size(), chksum, relPath, task.TargetConn)
	if err != nil {
		// 发送FIL出现错误，退出
		pref.HdrErr()
		fmt.Println("Error on FIL: ", err)
		return 0, err
	}

	// 检视ret：0 ERR（已经handle） 1 YES 2 SAM 3 NOS
	if ret == 2 {
		// 检查设置
		if pref.GlobalConfig.OverwriteWhenSAM == false {
			// 用户要求：若存在相同文件则不要复写
			_, err := task.TargetConn.Write([]byte("NOV < Please don't overwrite. >"))
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				return 0, err
			}
			// 接收NOV的服务器回应
			buf := make([]byte, 100)
			_, err = task.TargetConn.Read(buf)
			if err != nil {
				pref.HdrErr()
				fmt.Println("Read: ", err)
				return 0, err
			}
			// 发单个文件时也假装进度条
			if isSendingFolder == false {
				task.mon.ReportInfo("002 " + info.Name())
				task.mon.CurrentTaskProgress = info.Size()
				task.mon.ReportInfo("003 File transfer is over.")
			} else {
				// 发文件夹时假装进度条也在滚动
				task.mon.CurrentTaskProgress += info.Size()
			}
			return 0, nil
		}
		// 用户要求：若存在相同文件也要重写
		_, err := task.TargetConn.Write([]byte("OVR < Feel free to overwrite. >"))
		if err != nil {
			pref.HdrErr()
			fmt.Println("Write: ", err)
			return 0, err
		}
	} else if ret == 3 {
		// 透过外壳询问用户
		var want bool
		want, err = wouldOverwriteWhenRecvNOS(true, task.mon, info.Name())
		if err != nil {
			return 0, err
		}
		if want == false {
			// 用户要求不要复写
			_, err := task.TargetConn.Write([]byte("NOV < Please don't overwrite. >"))
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				return 0, err
			}
			// 接收服务器回应
			buf := make([]byte, 100)
			_, err = task.TargetConn.Read(buf)
			if err != nil {
				pref.HdrErr()
				fmt.Println("Read: ", err)
				return 0, err
			}
			if isSendingFolder == false {
				task.mon.ReportInfo("002 " + info.Name())
				task.mon.CurrentTaskProgress = info.Size()
				task.mon.ReportInfo("003 File transfer is over.")
			}
			return 0, nil
		}
		// 用户要求"我要重写"
		_, err := task.TargetConn.Write([]byte("OVR < Feel free to overwrite. >"))
		if err != nil {
			pref.HdrErr()
			fmt.Println("Write: ", err)
			return 0, err
		}
	}

	// 对于请求重写或者请求发送，服务器会返回一条YES <FIL>
	// 其中，若是请求发送得到立即应答的，上边已经接收过了，不要再接收
	if ret != 1 {
		buf := make([]byte, 100)
		_, err = task.TargetConn.Read(buf)
		if err != nil {
			pref.HdrErr()
			fmt.Println("Read: ", err)
			return 0, err
		}
		if string(buf[0:3]) == "YES" {
			// 发送或重写文件
			return sendFileProcess(isSendingFolder, fileNum, f, &info, task)
		} else {
			return 0, errors.New("unexpected error received: " + string(buf[0:3]))
		}
	}
	// 对于直接发送，不用重写的
	return sendFileProcess(isSendingFolder, fileNum, f, &info, task)
}

// 询问NOS操作，例如可以访问用户配置之类
func wouldOverwriteWhenRecvNOS(remoteControl bool, monitor *Monitor, fileName string) (bool, error) {
	if remoteControl == false {
		pref.HdrInf()
		fmt.Println("There is file that has same name on server. Overwrite? (y/n)")
		var inp string
		_, _ = fmt.Scan(&inp)
		if inp[0] == 'y' || inp[0] == 'Y' {
			// 好的
			return true, nil
		} else {
			return false, nil
		}
	} else {
		question := "\"" + fileName + "\" is already exists but has different contents. Replace? "
		answer, err := monitor.Ask(len(question), question)
		if err != nil {
			return false, err
		}
		return answer, nil
	}
}

// 续传文件主进程
func resumeFileProcess(seek int64, fileNum int32, f *os.File, info *os.FileInfo, task *Task) (int64, error) {
	// 缓冲区
	buf := make([]byte, 1024*4)
	// 循环发送文件，直到发送成功
	pref.HdrInf()
	fmt.Println("Sending File...")

	// 通知monitor
	task.mon.CurrentTaskFileName = (*info).Name()
	task.mon.CurrentTaskFileNum = fileNum
	task.mon.CurrentTaskSize = (*info).Size()

	var isFirstTry = true
	for {
		task.mon.ReportInfo("002 " + (*info).Name())
		// 回到文件头或者断点（仅当第1次尝试）
		var err error
		if isFirstTry == true {
			_, err = f.Seek(seek, 0)
			isFirstTry = false
		} else {
			_, err = f.Seek(0, 0)
		}
		if err != nil {
			return 0, err
		}
		// 新建进度条
		bar := progressbar.NewOptions64((*info).Size(), progressbar.OptionSetRenderBlankState(true), progressbar.OptionSetWidth(30))
		bar.Set64(seek)
		task.mon.CurrentTaskProgress += int64(seek)
		// 新建计时
		start := time.Now()
		// 循环读取文件
		for {
			n, err := f.Read(buf)
			if err != nil {
				// 纯属已经读完文件
				if err == io.EOF {
					break
				} else {
					return 0, err
				}
			}
			// 给服务器发送内容
			n, err = task.TargetConn.Write(buf[:n])
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				return 0, errors.New("003 Connection failed while transferring")
			}
			bar.Add(n)
			task.mon.CurrentTaskProgress += int64(n)
		}
		// 14/4/2020更新：为避免粘包，传输完毕时等待对方回应无意义信息，才发回EOF
		n, err := task.TargetConn.Read(buf)
		if err != nil {
			pref.HdrErr()
			fmt.Println("", err)
			return 0, errors.New("003 Connection failed while transferring")
		}

		// 睡10ms再发EOF，确保对面电脑准备好接收了
		time.Sleep(10 * time.Millisecond)
		err = sendEOF(fileNum, task.TargetConn)
		if err != nil {
			return 0, errors.New("006 Send instructions failed")
		}
		color.Green.Print("\nCONC INFO: ")
		fmt.Println("Transfer file: " + (*info).Name() + " is over.")
		task.mon.ReportInfo("003 Current file transfer is over. ")
		// 统计发送时间
		rate := (float64((*info).Size()) / 1048576) / (float64(time.Since(start).Milliseconds()) / 1000)
		pref.HdrErr()
		fmt.Printf("Speed: %.2f MB/s\n", rate)

		// 检视EOF回应
		n, err = task.TargetConn.Read(buf)
		if n <= 3 {
			return 0, errors.New("ER 011 Weird instruction.")
		}
		ret := string(buf[0:3])
		if ret == "YES" {
			// 并无错误，发送完毕可退出，退出前删除两个记录文件
			remoteHost := strings.Replace(task.TargetConn.RemoteAddr().String(), ":", "_", 1)
			utils.TryDeleteBreakpointOf(remoteHost)
			return (*info).Size(), nil
		} else {
			// 再读出错误代码
			ret = string(buf[5:8])
			if ret == "255" {
				// 255 File is broken. 解决方案：重发
				pref.HdrErr()
				fmt.Println("255 File is broken. Trying resend.")
				task.mon.ReportInfo("005 File is broken, trying resend.")
				continue
			} else {
				pref.HdrErr()
				fmt.Println("Unhandled error from remote: " + ret)
				return 0, errors.New("ER 012 Unhandled error from remote server")
			}
		}
	}
}

// 发送或重写文件主进程
func sendFileProcess(isSendingFolder bool, fileNum int32, f *os.File, info *os.FileInfo, task *Task) (int64, error) {
	// 缓冲区
	buf := make([]byte, 1024*4)
	// 记录断点
	remoteHost := strings.Replace(task.TargetConn.RemoteAddr().String(), ":", "_", 1)
	err := utils.RecordPoint(remoteHost, fileNum)
	if err != nil {
		return 0, err
	}
	// 通知monitor
	task.mon.CurrentTaskFileName = (*info).Name()
	task.mon.CurrentTaskFileNum = fileNum
	// 循环发送文件，直到发送成功
	pref.HdrInf()
	fmt.Println("Sending File...")
	for {
		if isSendingFolder == false {
			task.mon.CurrentTaskSize = (*info).Size()
			task.mon.CurrentTaskProgress = 0
			task.mon.ReportInfo("002 " + (*info).Name())
			fmt.Println("002 sent.")
		}
		// 回到文件头
		_, err = f.Seek(0, 0)
		if err != nil {
			return 0, err
		}
		// 新建进度条
		bar := progressbar.NewOptions64((*info).Size(), progressbar.OptionSetRenderBlankState(true), progressbar.OptionSetWidth(30))
		// 新建计时
		start := time.Now()
		// 循环读取文件
		for {
			n, err := f.Read(buf)
			if err != nil {
				// 纯属已经读完文件
				if err == io.EOF {
					break
				} else {
					pref.HdrErr()
					fmt.Println("", err)
					return 0, errors.New("ER 013 I/O Error while sending file")
				}
			}
			// 给服务器发送内容
			n, err = task.TargetConn.Write(buf[:n])
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				return 0, errors.New("ER 003 I/O Connection failed while transferring")
			}
			bar.Add(n)
			task.mon.CurrentTaskProgress += int64(n)
		}
		// 14/4/2020更新：为避免粘包，传输完毕时等待对方回应无意义信息，才发回EOF
		n, err := task.TargetConn.Read(buf)
		if err != nil {
			pref.HdrErr()
			fmt.Println("", err)
			return 0, errors.New("ER 003 I/O Connection failed while transferring")
		}

		// 睡10ms再发EOF，确保对面电脑准备好接收了
		time.Sleep(10 * time.Millisecond)
		err = sendEOF(fileNum, task.TargetConn)
		if err != nil {
			return 0, errors.New("ER 006 Send instruction EOF failed")
		}
		color.Green.Print("\n"+"CONC INFO: ")
		fmt.Println(" Transfer file: " + (*info).Name() + " is over.")
		if isSendingFolder == false {
			task.mon.ReportInfo("003 Current file transfer is over. ")
		}
		// 统计发送时间
		rate := (float64((*info).Size()) / 1048576) / (float64(time.Since(start).Milliseconds()) / 1000)
		color.Green.Print("\n"+"CONC INFO: ")
		fmt.Printf("Speed: %.2f MB/s\n", rate)

		// 检视EOF回应
		n, err = task.TargetConn.Read(buf)
		if n <= 3 {
			return 0, errors.New("ER 012 Unhandled error from remote server")
		}
		ret := string(buf[0:3])
		if ret == "YES" {
			// 并无错误，发送完毕可退出，退出前删除两个记录文件
			utils.TryDeleteBreakpointOf(remoteHost)
			return (*info).Size(), nil
		} else {
			// 再读出错误代码
			ret = string(buf[5:8])
			if ret == "255" {
				// 255 File is broken. 解决方案：重发
				pref.HdrErr()
				fmt.Println("255 File is broken. Trying resend.")
				task.mon.ReportInfo("005 File is broken. Trying resend")
				continue
			} else {
				pref.HdrErr()
				fmt.Println("Unhandled error from remote: " + ret)
				return 0, errors.New("ER 012 Unhandled error from remote")
			}
		}
	}
}

type Task struct {
	isInitialized bool
	TaskName string
	TargetConn         net.Conn // 对方通讯套接字
	mon *Monitor
}

func (task *Task) New(mon *Monitor) error {
	// 主动连接服务器
	addr_port := mon.TargetAddr + ":" + mon.TargetPort
	pref.HdrInf()
	fmt.Println("Dialing " + addr_port)
	conn, err := net.Dial("tcp", addr_port)
	if err != nil {
		return errors.New("001 Connect failed panic")
	}
	task.TargetConn = conn
	task.mon = mon
	task.isInitialized = true
	return nil
}

func (ses *Task) SetTaskPath() error {
	if ses.mon != nil {
		info, err := os.Stat(ses.mon.ClientTaskAbsPath)
		if err != nil {
			return errors.New("002 No such file or directory")
		}

		ses.mon.IsTaskDir = info.IsDir()
		ses.TaskName = info.Name()
		ses.mon.CurrentTaskSize = info.Size()
	}
	return nil
}

func (tsk *Task) DoTask(mon *Monitor) error {

	// 让客户端实例按照monitor的设置来配置目录
	err := tsk.SetTaskPath()
	if err != nil {
		mon.ReportError(err.Error())
	}

	// 发送STR指令(o)
	err = sendSTR(tsk.TargetConn)
	if err != nil {
		mon.ReportError("006 " + err.Error())
	}

	var sentSize int64 = 0
	start := time.Now() // 开始计时
	if tsk.mon.IsTaskDir {
		// 发送文件夹
		sentSize, _, err = SendOneFolder(tsk)
		if err != nil {
			// 错误代码在中间已经输出
			mon.ReportError(err.Error())
		}
		// 删掉list.send
		_ = os.Remove("list~" + tsk.mon.TargetAddr + "_" + tsk.mon.TargetPort + ".send")
	} else {
		// 发送文件(FIL、EOF)
		isHavingB, err := utils.IsHavingBreakpointOf(tsk.mon.TargetAddr + "_" + tsk.mon.TargetPort, tsk.TaskName)
		if err != nil {
			mon.ReportError("009 " + err.Error())
		}
		if isHavingB {
			sentSize, err = ResumeSendOneFile(tsk.mon.ClientTaskAbsPath, tsk.TaskName, tsk, 0)
			if err != nil {
				mon.ReportError(err.Error())
			}
		} else {
			sentSize, err = SendOneFile(false, tsk.mon.ClientTaskAbsPath, tsk.TaskName, tsk, 0)
			if err != nil {
				mon.ReportError(err.Error())
			}
		}
	}

	// 仅当发送了有效文件时使用
	if sentSize != 0 {
		rate := (float64(sentSize) / 1048576) / (float64(time.Since(start).Milliseconds()) / 1000)
		pref.HdrInf()
		fmt.Printf("Total used time: %s. Avg. speed: %.2f MB/s\n", time.Since(start).String(), rate)
	}

	// 发送END指令
	err = sendEND(tsk.mon.CurrentTaskSize, tsk.TargetConn)
	if err != nil {
		return err
	}

	pref.HdrInf()
	fmt.Println("Send file procedure is over.")
	return nil
}

//// 运行一客户端实例。但是：客户端实例不能拥有太多功能！
//// 仅供连接建立、管理资源使用。所有功能应该透过函数使用！
//func Client(addr string, port string) {
//	var tsk Task
//	err := tsk.New(addr, port)
//	if err != nil {
//		fmt.Println("task create err =", err)
//		return
//	}
//
//	// 提示输入命令
//	fmt.Println("Choose file to be transferred: ")
//	var path string
//	fmt.Scan(&path)
//
//	// 获取文件/文件夹名 info.Name()
//	err = tsk.SetTaskPath(path)
//	if err != nil {
//		fmt.Println("os.Stat err =", err)
//	}
//
//	// 发送STR指令(o)
//	err = sendSTR(tsk.TargetConn)
//	if err != nil {
//		pref.HdrInf()
//	fmt.Println(	"Exit because error: ", err)
//		return
//	}
//
//	var sentSize int64 = 0
//	start := time.Now() // 开始计时
//	if tsk.isTaskDir {
//		// 发送文件夹
//		sentSize, _, err = SendOneFolder(path, &tsk)
//		if err != nil {
//			pref.HdrInf()
//	fmt.Println(		"Exit because error: ", err)
//			return
//		}
//		// 删掉list.send
//		_ = os.Remove("list~"+addr + "_" + port + ".send")
//	} else {
//		// 发送文件(FIL、EOF)
//		isHavingB, err := utils.IsHavingBreakpointOf(addr + "_" + port, tsk.taskName)
//		if err != nil {
//			pref.HdrInf()
//	fmt.Println(		"Exit when finding bkpt: ", err)
//			return
//		}
//		if isHavingB {
//			sentSize, err = ResumeSendOneFile(path, tsk.taskName, &tsk, 0)
//			if err != nil {
//				pref.HdrInf()
//	fmt.Println(			"Exit because error: ", err)
//				return
//			}
//		} else {
//			sentSize, err = SendOneFile(path, tsk.taskName, &tsk, 0)
//			if err != nil {
//				pref.HdrInf()
//	fmt.Println(			"Exit because error: ", err)
//				return
//			}
//		}
//	}
//
//	// 仅当发送了有效文件时使用
//	if sentSize != 0 {
//		rate := (float64(sentSize) / 1048576) / (float64(time.Since(start).Milliseconds()) / 1000)
//		fmt.Printf(pref.HdrInf, "Total used time: %s. Avg. speed: %.2f MB/s\n", time.Since(start).String(), rate)
//	}
//
//	// 发送END指令
//	err = sendEND(tsk.taskSize, tsk.TargetConn)
//	if err != nil {
//		pref.HdrInf()
//	fmt.Println(	"Exit because error: ", err)
//		return
//	}
//
//	// 一次客户端实例完毕
//	pref.HdrInf()
//fmt.Println(	"Send file procedure is over.")
//}
