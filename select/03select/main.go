package main

import (
	"fmt"
	"time"
)

func worker(done <-chan struct{}) {
	for {
		select {
		case <-done:
			fmt.Println("worker: stopping")
			return

		default:
			fmt.Println("worker: working")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	done := make(chan struct{})

	go worker(done)

	time.Sleep(2 * time.Second)

	close(done)

	time.Sleep(1 * time.Second)
}
