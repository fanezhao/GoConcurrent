package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigRecv := make(chan os.Signal, 1)
	sigs := []os.Signal{syscall.SIGINT, syscall.SIGQUIT}
	signal.Notify(sigRecv, sigs...)
	// 永远都不会停止
	for sig := range sigRecv {
		// 接收到信号不做任何处理，直接输出，相当于忽略信号，因此永远不会停止，除非kill掉
		fmt.Printf("%s\n", sig)
	}
}
