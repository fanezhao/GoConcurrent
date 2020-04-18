package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

func main() {
	cmd0 := exec.Command("echo", "-n", "My first command comes from golang.")

	stdout0, err := cmd0.StdoutPipe()
	defer stdout0.Close()
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}

	if err := cmd0.Start(); err != nil {
		fmt.Printf("the err is %s\n", err)
		return
	}

	// output0(stdout0)
	// output1(stdout0)
	output2(stdout0)
}

func output0(r io.Reader) {
	output0 := make([]byte, 30)
	n, err := r.Read(output0)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", output0[:n])
}

func output1(r io.Reader) {
	var outputBuf0 bytes.Buffer
	for {
		tempOutput := make([]byte, 5)
		n, err := r.Read(tempOutput)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}
		if n > 0 {
			outputBuf0.Write(tempOutput[:n])
		}
	}
	fmt.Printf("%s\n", outputBuf0.String())
}

func output2(r io.Reader) {
	outputBuf0 := bufio.NewReader(r)
	output0, isPrefix, err := outputBuf0.ReadLine()
	if err != nil {
		return
	}
	fmt.Printf("%v, %s\n", isPrefix, string(output0))
}