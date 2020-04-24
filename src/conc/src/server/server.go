package server

import (
	"conc/src/pref"
	"conc/src/utils"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
)

// 接收或重写文件
func RecvFile(seek int64, filePath string, fileNum int32, fileSize int64, chksum *[]byte, conn net.Conn, bufferSize int64) (bool, error) {
	// 不停接收文件，直到接收到正确的文件为止
	var isFirstSend = true
	for {
		// 新建文件
		var f *os.File; var err error
		var receivedSize int64 = 0
		if seek == 0 {
			f, err = os.Create(filePath) // f是文件指针
		} else {
			f, err = os.OpenFile(filePath, os.O_WRONLY, os.ModeAppend)
			f.Seek(seek, 0)
			receivedSize = seek
		}
		// 回复没准备好
		if err != nil {
			pref.HdrErr()
			fmt.Println("os.Create: ", err)
			_, err := conn.Write([]byte("ERR <499><Creating file on remote folder failed.>"))
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				return false, err
			}
		}

		// 回复准备好（无论FIL还是OVR都回复YES FIL）
		// 但是，重发时 及 断点续传时不需要回复！
		if isFirstSend && seek == 0 {
			_, err = conn.Write([]byte("YES <FIL><Prepare is done.>"))
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				return false, err
			}
		}

		// 新建一个缓冲区
		buf := make([]byte, bufferSize*4)
		// 读文件内容
		pref.HdrInf()
		fmt.Println("[FIL] Receiving File " + f.Name() + " ...")
		bar := progressbar.NewOptions64(fileSize, progressbar.OptionSetRenderBlankState(true), progressbar.OptionSetWidth(30))
		bar.Set64(seek)
		for receivedSize < fileSize {
			// 注：接收的是发送方发来的信息
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF || errors.Is(err, syscall.ECONNRESET) || err == syscall.EPIPE {
					fmt.Println("\nCONC INFO: Connection is closed.")
				} else {
					fmt.Println("\nCONC ERROR: Read in file transfer: ", err)
				}
				return false, err
			}
			if n == 0 {
				fmt.Println("File transfer over for \"" + f.Name() + "\" because reveived no data")
				break
			}
			f.Write(buf[:n])
			receivedSize += int64(n)
			bar.Add(n)
		}

		// 发送完毕。发回无意义信息一条 14/4/2020
		_, err = conn.Write([]byte("I have received all data :)"))
		if err != nil {
			pref.HdrErr()
			fmt.Println("Write: ", err)
			return false, err
		}

		// 接受EOF
		_, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF || err == syscall.EPIPE {
				pref.HdrInf()
				fmt.Println("Connection is closed.")
			} else {
				pref.HdrErr()
				fmt.Println("Read in file transfer: ", err)
			}
			return false, err
		}

		// 收到EOF指令
		if string(buf[0:3]) == "EOF" {
			// EOF，文件传输结束
			eofFinum := utils.BytesToInt32(buf[5:9])
			// 文件编码不匹配
			if eofFinum != fileNum {
				// 如果文件号码不匹配，说明有问题，拒绝接受
				fmt.Printf("Received EOF for filenum %d and require %d\n", eofFinum, fileNum)
				return false, errors.New("253 file number is not correspond to which was specified")
			}
		}

		fmt.Println()
		// 验证文件
		err = f.Close()
		if fileSize != -1 {
			// 如果不是list.send
			fs := utils.GetFileSize(filePath)
			if fs != fileSize {
				pref.HdrErr()
				fmt.Println("File size of the received data is not correspond! ")
				return false, errors.New("254 File size is not correspond.")
			}
		}

		// 做MD5校验码检验
		fmt.Println("Checking " + f.Name())
		isMD5OK, err := utils.ChkFileChksum(filePath, chksum)
		if err != nil {
			pref.HdrErr()
			fmt.Println("Error occurred while performing MD5 checksum.")
			return false, err
		}
		if isMD5OK == false {
			pref.HdrErr()
			fmt.Println("MD5 checksum not passed.")
			_, err := conn.Write([]byte("ERR <255><" + "MD5 value is not correspond to which you sent first."))
			if err != nil {
				pref.HdrErr()
				fmt.Println("Write: ", err)
				// 等待重发
				return false, err
			}
			// 接收重发的资料
			isFirstSend = false
			continue
		}

		// 返回正确
		pref.HdrInf()
		fmt.Println("[YES] Transmission of file " + f.Name() + " is over.")
		_, err = conn.Write([]byte("YES <EOF><Received " + f.Name() + ">"))
		if err != nil {
			pref.HdrErr()
			fmt.Println("Write: ", err)
			return false, err
		}
		return true, nil
	}
}


// 运行一个服务器实例
//noinspection ALL
func Server(working_d string, addr string, port string, buffer_size int64) {
	var isConnected bool = false
	// 检视working_d是否存在
	exists, isDir, err := utils.GetPathStat(working_d)
	if !exists || !isDir {
		pref.HdrErr()
		fmt.Println("Working directory \"", working_d, "\"does not exist. ")
		return
	}

	// 提示服务器实已经创建
	pref.HdrInf()
	fmt.Println("A server instance has been created.")

	// 创建进行【监听】的套接字
	listener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		pref.HdrErr()
		fmt.Println("Listening: ", err)
		return
	}
	// 发生错误或连接完毕时关闭监听
	defer listener.Close()
	defer fmt.Println("Server is closing down...")

	for { // connect
		pref.HdrInf()
		fmt.Println("Listening " + addr + ":" + port + "...")
		// 创建用于【通讯】的套接字，阻塞等待用户连接
		conn, err := listener.Accept()
		if err != nil {
			pref.HdrErr()
			fmt.Println("Accept: ", err)
			break
		}
		// 发生错误或连接完毕时关闭通讯
		defer conn.Close()
		isConnected = true

		// 制作长度为buffer_size的缓冲区
		buf := make([]byte, buffer_size)
		working_dir := strings.TrimSuffix(working_d, string(os.PathSeparator))

		// 不停读取，直到对方发送文件传输开始(STR)指令
		pref.HdrInf()
		fmt.Println("Waiting STR from client.")
		for { // waiting start
			n, err := conn.Read(buf) // n为读取的字节长
			if err != nil {
				if err == io.EOF {
					pref.HdrInf()
					fmt.Println("Connection is closed.")
					isConnected = false
					break
				} else {
					pref.HdrErr()
					fmt.Println("While waiting for STR: ", err)
					conn.Close()
					isConnected = false
					break
				}
			}
			// 检视STR指令，如果收到的字节数太小就忽略
			if n < 3 || string(buf[0:3]) != "STR" {
				// 若收到不是STR，则文件传输还没能开始，忽略这条讯息
				continue
			}
			break
		}
		if isConnected == false {
			continue	// 跳回监听阶段
		}

		// 回应STR
		_, err = conn.Write([]byte("YES <STR><STR set.>"))
		if err != nil {
			pref.HdrErr()
			fmt.Println("Write: ", err)
		}
		var rootDir string = ""
		// 下面接收其他指令
		for { // STR received.
			pref.HdrInf()
			fmt.Println("[STR] Waiting instructions...")
			n, err := conn.Read(buf) // n为读取的字节长
			if err != nil {
				// 收到EOF，表示对方关闭连接
				if err == io.EOF {
					pref.HdrInf()
					fmt.Println("Connection is closed.")
					isConnected = false
					break
				} else {
					// 其他错误
					pref.HdrWrn()
					fmt.Println("Connection was terminated.")
					isConnected = false
					break
				}
			}
			// 检视收到的指令
			cmd := string(buf[0:3])
			var currFileNum int32 = 0  // 当前处理到的文件编号
			var currFileRelPath string // 当前文件所处的相对目录

			if cmd == "FIL" {          // 收到FIL指令
				fileNum := utils.BytesToInt32(buf[5:9])
				// 获取这个文件大小及checksum
				fileSize := utils.BytesToInt64(buf[11:19])
				var chksum = make([]byte, md5.Size)
				copy(chksum, buf[21:37])
				// 找到从39（文件名的第1个byte）开始的第1个“>”，为文件名定界
				for n = 39; buf[n] != '>'; n++ {
				}
				currFileRelPath = string(buf[41:n])	// 当前文件相对目录

				if fileNum != currFileNum {
					// 返回文件序号不匹配警告（139）
					pref.HdrWrn()
					fmt.Println("139 Ambiguous file order.")
					_, err := conn.Write([]byte("ERR <139><Ambiguous file order>"))
					if err != nil {
						pref.HdrErr()
						fmt.Println("Write: ", err)
						isConnected = false
						break
					}
					_, err = conn.Read(buf)
					if err != nil {
						pref.HdrErr()
						fmt.Println("Read: ", err)
						isConnected = false
						break
					}
					// 重设文件编号
					currFileNum = fileNum
				}
				// 接收客户端回复 ??
				if rootDir == "" {
					// 若rootDir就没有被设置过，则目前模式是传单个文件而非传文件夹
					// 根目录就是working directory
					rootDir = working_dir
				}

				// 判断文件是否已经拥有，先尝试开启
				var exists bool
				// 当前文件绝对目录
				currFileAbsPath := rootDir + string(os.PathSeparator) + currFileRelPath
				stat, err := os.Stat(currFileAbsPath)
				if err != nil {
					// 出错，说明可能没有文件或者出现其他错误
					if os.IsNotExist(err) {
						// 没有文件
						exists = false
					} else {
						// 出错
						pref.HdrErr()
						fmt.Println("Unknown status while checking file existance.")
						isConnected = false
						break
					}
				} else {
					// 未出错，说明存在目录，但不确定是文件夹还是文件
					if stat.IsDir() == true {
						// 是文件夹，则没有文件，因为系统不同意同名文件和文件夹的存在
						exists = false
					} else {
						// 是文件，则说明存在
						exists = true
					}
				}

				// 若存在，则需要和客户端交涉是否需要重写
				var userNeedsOverwrite bool = true
				if exists {
					// 若存在，则判断文件是否相同，返回响应信息
					isSame, err := utils.ChkFileChksum(currFileAbsPath, &chksum)
					if err != nil {
						// 出错
						pref.HdrErr()
						fmt.Println("Unknown return while checking if the existing file is same.")
						isConnected = false
						break
					} else {
						if isSame {
							// 相同文件，返回SAM
							_, err := conn.Write([]byte("SAM < There is a same file on server. >"))
							if err != nil {
								pref.HdrErr()
								fmt.Println("Write: ", err)
								isConnected = false
								break
							}
							// 接收用户需求
							_, err = conn.Read(buf)
							if err != nil {
								if err == io.EOF {
									pref.HdrInf()
									fmt.Println("Connection is closed.")
									isConnected = false
									break
								} else {
									pref.HdrErr()
									fmt.Println("Read in file transfer: ", err)
									isConnected = false
									break
								}
							}
							// 判断用户需求
							if string(buf[0:3]) == "NOV" {
								// NOV，用户不希望复写
								userNeedsOverwrite = false
								// 返回NOV的YES
								_, err := conn.Write([]byte("YES <NOV>"))
								if err != nil {
									pref.HdrErr()
									fmt.Println("Write: ", err)
									isConnected = false
									break
								}
							} // 省略
						} else {
							// 不同文件，返回NOS
							_, err := conn.Write([]byte("NOS < Do you want to overwrite? >"))
							if err != nil {
								pref.HdrErr()
								fmt.Println("Write: ", err)
								isConnected = false
								break
							}
							// 接收用户需求
							_, err = conn.Read(buf)
							if err != nil {
								if err == io.EOF {
									pref.HdrInf()
									fmt.Println("Connection is closed.")
									isConnected = false
									break
								} else {
									pref.HdrErr()
									fmt.Println("Read in file transfer: ", err)
									isConnected = false
									break
								}
							}
							// 判断用户需求
							if string(buf[0:3]) == "NOV" {
								// NOV，用户不希望复写
								userNeedsOverwrite = false
								// 返回NOV的YES
								_, err := conn.Write([]byte("YES <NOV>"))
								if err != nil {
									pref.HdrErr()
									fmt.Println("Write: ", err)
									isConnected = false
									break
								}
							}
						}
					}
				}	// if exists

				// 若不存在文件，或用户需要重写文件，则直接接收文件内容
				if userNeedsOverwrite {
					succ, err := RecvFile(0, currFileAbsPath, fileNum, fileSize, &chksum, conn, buffer_size)
					// 检视接收结果
					if succ != true || err != nil {
						// 文件有问题	或 提前遇到终止 或 brokenpipe
						if io.EOF == err || syscall.EPIPE == err {
							// 退出
							pref.HdrInf()
							fmt.Println("Connection error.")
							isConnected = false
							break
						} else {
							pref.HdrErr()
							fmt.Println("", err)
							errstr := err.Error()
							_, err := conn.Write([]byte("ERR <" + errstr[0:3] + "><" + errstr[4:]))
							if err != nil {
								pref.HdrErr()
								fmt.Println("Write: ", err)
								// 等待重发
								isConnected = false
								break
							}
						}
					}
					// 发送成功的话，本机currFileNum自增1
					currFileNum++
				}	// if userNeedsOverwrite
			} else if cmd == "CHD" {
				// 获取相对于根目录的相对目录
				for n = 6; buf[n] != '>'; n++ {
				}
				rel_path := string(buf[5:n]) // 相对于根目录的路径
				path := rootDir + string(os.PathSeparator) + rel_path
				if rootDir == "" {
					// 第1次传文件夹，还没设置当前根目录（不是工作目录）
					rootDir = working_dir + string(os.PathSeparator) + rel_path
					pref.HdrInf()
					fmt.Println("Assumed root directory as " + rootDir)
					path = rootDir
				}
				// 检查目录是否有存在
				isExist, isDir, _ := utils.GetPathStat(path)
				if !isExist || !isDir {
					// 创建该文件夹
					err = os.Mkdir(path, os.ModePerm)
					if err != nil {
						pref.HdrErr()
						fmt.Println("Failed creating directory "+path+": ", err)
						// 寄回错误信息
						_, err := conn.Write([]byte("ERR <555><Error occured while creating directory>"))
						if err != nil {
							pref.HdrErr()
							fmt.Println("Write: ", err)
							isConnected = false
							break
						}
						return
					} else {
						pref.HdrInf()
						fmt.Println("[CHD] Created directory at " + path)
					}
					// 14/4/2020 go必须用相对于根目录的路径来做！
					// working_dir = path
				} else {
					// 直接切换目录
					pref.HdrInf()
					fmt.Println("[CHD] ", path, ": directory exists. Ignoring.")
					// 14/4/2020 go必须用相对于根目录的路径来做！
					// working_dir = path
				}
				// 返回成功切换目录信息
				_, err := conn.Write([]byte("YES <CHD><Working directory is now at " + rel_path))
				if err != nil {
					pref.HdrErr()
					fmt.Println("Write: ", err)
					isConnected = false
					break
				}
			} else if cmd == "RES" {
				fileNum := utils.BytesToInt32(buf[5:9])
				if fileNum != currFileNum {
					// 重设文件编号，不用返回139
					currFileNum = fileNum
				}
				// 获取这个文件大小及checksum
				fileSize := utils.BytesToInt64(buf[11:19])
				chksum := make([]byte, md5.Size)
				copy(chksum, buf[21:37])
				// 找到从39（文件名的第1个byte）开始的第1个“>”，为文件名定界
				for n = 39; buf[n] != '>'; n++ {
				}
				currFileRelPath = string(buf[41:n])	// 当前文件相对目录
				if rootDir == "" {
					// 若rootDir就没有被设置过，则目前模式是传单个文件而非传文件夹
					// 根目录就是working directory
					rootDir = working_dir
				}

				// 判断文件是否已经拥有，先尝试开启
				var exists bool
				// 当前文件绝对目录
				currFileAbsPath := rootDir + string(os.PathSeparator) + currFileRelPath
				stat, err := os.Stat(currFileAbsPath)
				if err != nil {
					// 出错，说明可能没有文件或者出现其他错误
					if os.IsNotExist(err) {
						// 没有文件
						exists = false
					} else {
						// 出错
						pref.HdrErr()
						fmt.Println("Unknown status while checking file existance.")
						isConnected = false
						break
					}
				} else {
					// 未出错，说明存在目录，但不确定是文件夹还是文件
					if stat.IsDir() == true {
						// 是文件夹，则没有文件，因为系统不同意同名文件和文件夹的存在
						exists = false
					} else {
						// 是文件，则说明存在
						exists = true
					}
				}

				// 如果存在文件，就返回当前的size
				// 如果不存在文件，就返回size为0，诱导客户端重传
				if exists {
					// 制作EPT指令
					ept := make([]byte, 15)
					seek := utils.Int64ToBytes(stat.Size())
					copy(ept[0:], "EPT <")
					copy(ept[5:], seek)
					copy(ept[13:], ">")
					_, err = conn.Write(ept)
					if err != nil {
						pref.HdrErr()
						fmt.Println("Write: ", err)
						isConnected = false
						break
					}
				} else {
					// 制作ERR指令
					_, err = conn.Write([]byte("ERR <957>< File not exist. >"))
					if err != nil {
						pref.HdrErr()
						fmt.Println("Write: ", err)
						isConnected = false
						break
					}
				}

				// 进入接收进程
				var seek int64 = 0
				if stat != nil {
					seek = stat.Size()
				}
				succ, err := RecvFile(seek, currFileAbsPath, fileNum, fileSize, &chksum, conn, buffer_size)
				// 检视接收结果
				if succ != true || err != nil {
					// 文件有问题	或 提前遇到终止
					if io.EOF == err {
						// 退出
						pref.HdrInf()
						fmt.Println("Connection lost.")
						isConnected = false
						break
					} else {
						pref.HdrErr()
						fmt.Println("", err)
						errstr := err.Error()
						_, err := conn.Write([]byte("ERR <" + errstr[0:3] + "><" + errstr[4:]))
						if err != nil {
							pref.HdrErr()
							fmt.Println("Write: ", err)
							isConnected = false
							break
						}
					}
				}
				// 发送成功的话，本机currFileNum自增1
				currFileNum++

			} else if cmd == "FOL" {
				// 传送文件夹信息（已被否决）
				// 检视currFileNum是否为0
				if currFileNum != 0 {
					_, err := conn.Write([]byte("ERR <142><list.send is not allowed at this time.>"))
					if err != nil {
						pref.HdrErr()
						fmt.Println("Write: ", err)
						isConnected = false
						break
					}
					continue
				}
				// 接收chksum
				chksum := buf[5:21]
				// 回复"ok"
				_, err := conn.Write([]byte("YES <FIL><Prepare is done.>"))
				if err != nil {
					pref.HdrErr()
					fmt.Println("Write: ", err)
					isConnected = false
					break
				}
				// 接收文件内容
				succ, err := RecvFile(0, "." + string(os.PathSeparator) + "list.send", 0, -1, &chksum, conn, buffer_size)
				// 检视接收结果
				if succ == false {
					// 文件有问题
					_, err := conn.Write([]byte("ERR <255><File is broken.>"))
					if err != nil {
						pref.HdrErr()
						fmt.Println("Write: ", err)
						isConnected = false
						break
					}
				}
			} else if cmd == "END" {
				// 全部文件传输结束
				pref.HdrInf()
				fmt.Println("[END] This session is over.")
				continue
			}
		} // STR received (waiting start)
		if isConnected == false {
			conn.Close()
			continue // connect
		}
		break
	}

}
