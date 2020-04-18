package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func main() {
	runCmdWithPipe()
}

func runCmdWithPipe() {
	cmd1 := exec.Command("ps", "aux")
	cmd2 := exec.Command("grep", "apipe")

	// 匿名管道
	var outputBuf1 bytes.Buffer
	cmd1.Stdout = &outputBuf1
	if err := cmd1.Start(); err != nil {
		return
	}
	if err := cmd1.Wait(); err != nil {
		return
	}

	cmd2.Stdin = &outputBuf1
	var outputBuf2 bytes.Buffer
	cmd2.Stdout = &outputBuf2
	if err := cmd2.Start(); err != nil {
		return
	}
	if err := cmd2.Wait(); err != nil {
		return
	}

	fmt.Printf("%s\n", outputBuf2.String())
}