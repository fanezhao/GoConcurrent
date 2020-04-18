package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

}

func handleSignal() {
	sigRecv1 := make(chan os.Signal, 1)
	sigs1 := []os.Signal{syscall.SIGINT, syscall.SIGQUIT}
	fmt.Printf("Set notification for %s... [sigRecv1]\n", sigs1)
	signal.Notify(sigRecv1, sigs1...)
	sigRecv2 := make(chan os.Signal, 1)
	sigs2 := []os.Signal{syscall.SIGQUIT}
	fmt.Printf("Set notification for %s... [sigRecv2]\n", sigs1)
	signal.Notify(sigRecv2, sigs2...)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for sig := range sigRecv1 {
			fmt.Printf("Received a signal from sigRecv1: %s\n", sig)
		}
		fmt.Printf("End. [sigRecv1]\n")
	}()
	go func() {
		defer wg.Done()
		for sig := range sigRecv1 {
			fmt.Printf("Received a signal from sigRecv2: %s\n", sig)
		}
		fmt.Printf("End. [sigRecv2]\n")
	}()
	fmt.Println("Wait for 2 seconds... ")
	time.Sleep(2 * time.Second)
	fmt.Printf("Stop notification...")
	signal.Stop(sigRecv1)
	close(sigRecv1)
	fmt.Printf("done. [sigRecv1]\n")
	wg.Wait()
}