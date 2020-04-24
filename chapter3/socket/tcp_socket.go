package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SERVER_NETWORK = "tcp"
	SERVER_ADDRESS = "127.0.0.1:8085"
	DELIMITER      = '\t' // 约定的数据边界
)

var wg sync.WaitGroup

func printLog(role string, sn int, format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf("%s[%d]: %s", role, sn, fmt.Sprintf(format, args...))
}

// printServerLog 服务端日志记录
func printServerLog(format string, args ...interface{}) {
	printLog("Server", 0, format, args...)
}

// printClientLog 客户端日志记录
func printClientLog(sn int, format string, args ...interface{}) {
	printLog("Client", sn, format, args...)
}

func strToInt32(str string) (int32, error) {
	num, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("\"%s\" is not integer", str)
	}
	if num > math.MaxInt32 || num < math.MinInt32 {
		return 0, fmt.Errorf("%d is not 32-bit integer", num)
	}
	return int32(num), nil
}

func cbrt(param int32) float64 {
	return math.Cbrt(float64(param))
}

// read 从连接中读取一段以数据分界符为结尾的数据
func read(conn net.Conn) (string, error) {
	var dataBuffer bytes.Buffer
	// readBytes 的长度初始化为1的目的是：防止从连接值中读出多余的数据从而对后续读取操作造成影响。
	// 我从连接上每读取一个字节的数据，都需要检查它是否为数据分界符。
	// 如果不是，就继续读取；如果是，就停止读取并返回结果。
	readBytes := make([]byte, 1)
	for {
		_, err := conn.Read(readBytes)
		if err != nil {
			return "", err
		}
		readByte := readBytes[0]
		if readByte == DELIMITER {
			break
		}
		dataBuffer.WriteByte(readByte)
	}
	return dataBuffer.String(), nil
}

// write 向连接发送数据
func write(conn net.Conn, content string) (int, error) {
	var buffer bytes.Buffer
	buffer.WriteString(content)
	buffer.WriteByte(DELIMITER)
	return conn.Write(buffer.Bytes())
}

// serverGo 服务端程序
func serverGo() {
	var listener net.Listener
	// 监听端口
	listener, err := net.Listen(SERVER_NETWORK, SERVER_ADDRESS)
	if err != nil {
		printServerLog("Listen Error: %s", err)
		return
	}
	// 保证函数结束的时候关闭监听器
	defer listener.Close()
	printServerLog("Got listener for the server. (local address: %s)", listener.Addr())
	for {
		// 阻塞直到新连接到来
		conn, err := listener.Accept()
		if err != nil {
			printServerLog("Accept Error: %s", err)
			continue
		}
		printServerLog("Established a connection with a client application. (remote address: %s)",
			conn.RemoteAddr())

		// 启动一个新的goroutine，来并发执行handleConn函数。
		// 在服务端程序中，这通常是非常有必要的。
		// 为了快速、独立的处理一个已经建立的每一个连接，就应该尽量让这些处理程序并发执行。
		// 否则，当处理已建立的第一个连接时，后续连接就只能排除等待，这相当于完全串行处理众多连接，这样做效率非常低下。
		// 并且，只要其中的一个连接由于某种原因阻塞了，后续所有连接就都无法处理了，这肯定是不行的。
		go handleConn(conn)
	}
}

// handleConn 处理已经建立的连接
func handleConn(conn net.Conn) {
	defer func() {
		// 函数执行完毕的时候一定要关闭连接，也同时关闭了闲置连接
		conn.Close()
		wg.Done()
	}()

	// 通过循环不断尝试从已经建立的连接上读取数据
	for {
		// 关闭闲置连接设定，超时10s
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		// 从连接中读取数据
		strReq, err := read(conn)
		if err != nil {
			if err == io.EOF {
				printServerLog("The connection is closed by another side.")
			} else {
				printServerLog("Read Error: %s", err)
			}
			// 读取超时，结束for loop使用defer函数完成连接关闭
			break
		}
		printServerLog("Received request: %s.", strReq)
		// 转成Int23类型的值
		intReq, err := strToInt32(strReq)
		if err != nil {
			n, err := write(conn, err.Error())
			printServerLog("Sent error message (written %d bytes): %s.", n, err)
			continue
		}
		// 求立方根
		floatResp := cbrt(intReq)
		respMsg := fmt.Sprintf("The cube root of %d is %f.", intReq, floatResp)
		// 发送数据
		n, err := write(conn, respMsg)
		if err != nil {
			printServerLog("Write Error: %s", err)
		}
		printServerLog("Sent response (written %d bytes): %s.", n, respMsg)
	}
}

// clientGo 客户端程序
func clientGo(id int) {
	defer wg.Done()
	// 请求建立TCP连接，2s超时
	conn, err := net.DialTimeout(SERVER_NETWORK, SERVER_ADDRESS, 2*time.Second)
	if err != nil {
		printClientLog(id, "Dial Error: %s", err)
		return
	}
	defer conn.Close()
	printClientLog(id, "Connected to server. (remote address: %s, local address: %s)",
		conn.RemoteAddr(), conn.LocalAddr())
	time.Sleep(200 * time.Millisecond)
	requestNum := 5
	for i := 0; i < requestNum; i++ {
		req := rand.Int31()
		// 向连接发送数据
		n, err := write(conn, fmt.Sprintf("%d", req))
		if err != nil {
			printClientLog(id, "Write Error: %s", err)
			continue
		}
		printClientLog(id, "Sent request (written %d bytes): %d.", n, req)
	}
	for i := 0; i < requestNum; i++ {
		// 从连接中读取数据
		strResp, err := read(conn)
		if err != nil {
			if err == io.EOF {
				printClientLog(id, "The connection is closed by another side.")
			} else {
				printClientLog(id, "Read Error: %s", err)
			}
			break
		}
		printClientLog(id, "Received response: %s.", strResp)
	}
}

func main() {
	wg.Add(2)
	go serverGo()
	// 通过睡眠，确保服务端先于客户端运行
	time.Sleep(500 * time.Millisecond)
	// 客户端程序在服务端开始运行并已准备好接收新连接的时候运行
	go clientGo(1)
	wg.Wait()
}
