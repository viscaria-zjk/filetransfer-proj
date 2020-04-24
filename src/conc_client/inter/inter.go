package inter

import (
	"conc_client/utils"
	"net"
)

// 测试连接
func SayHello(conn net.Conn) (isOK bool) {
	buf := make([]byte, 10)
	_, _ = conn.Write([]byte("HELO"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

func GetCurrentSeek(conn net.Conn) (seek int64, isOK bool) {
	buf := make([]byte, 10)
	_, _ = conn.Write([]byte("SEEK"))
	_, err := conn.Read(buf)
	if err != nil {
		return 0, false
	}
	return utils.BytesToInt64(buf[0:8]), true
}

func GetCurrentTotalSize(conn net.Conn) (totalSize int64, isOK bool) {
	buf := make([]byte, 10)
	_, _ = conn.Write([]byte("TSZE"))
	_, err := conn.Read(buf)
	if err != nil {
		return 0, false
	}
	return utils.BytesToInt64(buf[0:8]), true
}

func CallEndProgress(conn net.Conn) (isOK bool) {
	buf := make([]byte, 5)
	_, _ = conn.Write([]byte("ENDP"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

func CallConnect(conn net.Conn) (isOK bool) {
	buf := make([]byte, 5)
	_, _ = conn.Write([]byte("CONN"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

// 发送配置
func SetSendPref(conn net.Conn, sendPath string) (isSetOK bool) {
	buf := make([]byte, 200)
	_, _ = conn.Write([]byte("SEND"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	_, _ = conn.Write(utils.Int32ToBytes(int32(len(sendPath))))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	_, _ = conn.Write([]byte(sendPath))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

// 设为用户模式
func SetTypeClient(conn net.Conn) (isSetOK bool) {
	buf := make([]byte, 5)
	_, _ = conn.Write([]byte("CLNT"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

func SetTypeServer(conn net.Conn) (isSetOK bool) {
	buf := make([]byte, 5)
	_, _ = conn.Write([]byte("SERV"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

// 设置地址
func SetRemoteAddr(conn net.Conn, addr string) (isSetOK bool) {
	buf := make([]byte, 35)
	_, _ = conn.Write([]byte("RMAD"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	_, _ = conn.Write(utils.Int32ToBytes(int32(len(addr))))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	_, _ = conn.Write([]byte(addr))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}

// 设置端口
func SetRemotePort(conn net.Conn, port string) (isSetOK bool) {
	buf := make([]byte, 35)
	_, _ = conn.Write([]byte("RMPO"))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	_, _ = conn.Write(utils.Int32ToBytes(int32(len(port))))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	_, _ = conn.Write([]byte(port))
	_, _ = conn.Read(buf)
	if string(buf[0:2]) != "OK" {
		return false
	}
	return true
}
