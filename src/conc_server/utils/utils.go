package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// 检查是否有当前文件的断点
func IsHavingBreakpointOf(remoteHost string, fileName string) (bool, error) {
	// 检查是否有point.bkpt文件
	breakpoint, err := ReadPoint(remoteHost)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, nil
	}
	if breakpoint != 0 {
		return false, nil
	}
	// 检查是否有符合要求的list.send文件
	fi, err := os.Open("list~"+remoteHost+".send")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Not found: ", "list~"+remoteHost+".send")
			return false, nil
		}
		return false, err
	}
	defer fi.Close()
	// 读取第1行
	line, err := bufio.NewReader(fi).ReadString('\n')
	if err != nil {
		return false, err
	}
	if line[0:3] != "FIL" {
		return false, nil
	}
	n := len(line) - 1
	for ; line[n] != '>'; n-- {}
	if line[8:n] != fileName{
		return false, nil
	}
	return true, nil
}

// 删除当前文件的断点
func TryDeleteBreakpointOf(remoteHost string) {
	_ = os.Remove("point~"+remoteHost+".bkpt")
	_ = os.Remove("list~"+remoteHost+".send")
}

// 用MD5算法拿到文件md5
func GetChksum(filePath string) (value []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()
	md5hash := md5.New()
	if _, err = io.Copy(md5hash, f); err != nil {
		return
	}
	value = md5hash.Sum(nil)
	return value, nil
}

// 分析文件及文件MD5是否匹配
func ChkFileChksum(filePath string, chksum *[]byte) (isChksumOK bool, err error) {
	calChksum, err := GetChksum(filePath)
	// 比较md5
	for i := 0; i < md5.Size; i++ {
		if (*chksum)[i] != calChksum[i] {
			return false, nil
		}
	}
	return true, nil
}

func Int32ToBytes(i int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt32(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

// 拿到文件大小
func GetFileSize(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

// 判断文件或文件夹是否存在及判断是不是文件夹
func GetPathStat(path string) (bool, bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return true, stat.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, false, nil
	}
	return false, false, err
}

// 写入断点至断点文件
func RecordPoint(remoteHost string, breakPoint int32) error {
	// 检查文件是否存在，若不存在就新建
	fi, err := func () (*os.File, error) {
		stat, e := os.Stat("point~"+remoteHost+".bkpt")
		if e != nil {
			if os.IsNotExist(e) {
				return os.Create("point~"+remoteHost+".bkpt")
			}
			return nil, e
		}
		if !stat.IsDir() {
			return os.OpenFile("point~"+remoteHost+".bkpt", os.O_WRONLY, os.ModeAppend)
		} else {
			return nil, errors.New("folder exists.")
		}
	}()
	if err != nil {
		return err
	}
	defer fi.Close()

	// 写入文件
	buf := Int32ToBytes(breakPoint)
	_, err = fi.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// 读取断点记录文件记录的断点
func ReadPoint(remoteHost string) (int32, error) {
	fi, err := os.Open("point~"+remoteHost+".bkpt")
	if err != nil {
		return 0, err
	}
	defer fi.Close()

	// 读取其中文件
	var buf = make([]byte, 4)
	_, err = fi.Read(buf)
	if err != nil {
		return 0, err
	}
	return BytesToInt32(buf), nil
}

// 读取list.send文件
func ReadListSend(remoteHost string) ([]string, error) {
	fi, err := os.Open("list~"+remoteHost+".send")
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	// 一行一行读取
	lists := []string{}
	rd := bufio.NewReader(fi)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("CONC ERROR: ", err)
				return nil, err
			}
		}
		lists = append(lists, string(line[:len(line)-1]))
	}
	return lists, nil
}

// 拿到包含根目录的list.send文件及内容
func GetListSend(remoteHost string, path string) (listSend []string, isHavingFolder bool, size int64, e error) {
	fi, err := os.Create("list~"+remoteHost+".send")
	if err != nil {
		return nil, false, 0, err
	}
	defer fi.Close()
	// 拿到文件夹信息
	paths, haveFolder, totalSize, err := GetPaths(path)
	if err != nil {
		return nil, false, 0, err
	}
	if haveFolder {
		// 写入文件夹递归信息
		for _, p := range paths {
			fi.WriteString(p + "\n")
		}
	} else {
		// 写入本文件
		fi.WriteString(paths[0] + "\n")
	}
	return paths, haveFolder, totalSize, nil
}

// 拿到fnames中指定的文件在系统中的位置
func GetPaths(fname string) (paths []string, haveFolder bool, totalSize int64, err error) {
	haveFolder = false
	paths = []string{}
	stat, errStat := os.Stat(fname)
	if errStat != nil {
		err = errStat
		return
	}
	// 根路径是一个文件夹
	totalSize = 0
	if stat.IsDir() {
		haveFolder = true
		// 遍历该文件夹
		currFileNum := 0
		currDirNum := 0
		err = filepath.Walk(fname,
			func(pathName string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				// 发现go好像是按照文件名递归的
				relPath, _ := filepath.Rel(fname, pathName)
				if info.IsDir() {
					paths = append(paths, "DIR <"+strconv.Itoa(currDirNum)+"><"+filepath.ToSlash(relPath)+">")
					currDirNum++
				} else {
					totalSize += info.Size()
					paths = append(paths, "FIL <"+strconv.Itoa(currFileNum)+"><"+filepath.ToSlash(relPath)+">")
					currFileNum++
				}
				return nil
			})
		if err != nil {
			return
		}
	} else {
		paths = append(paths, "FIL <"+strconv.Itoa(0)+"><"+stat.Name()+">")
	}
	return
}
