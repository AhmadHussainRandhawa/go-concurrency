package main

import (
	"fmt"
	"time"
)

func worker(charChannel <-chan string, done <-chan struct{}) {
	for {
		select {
		case result, ok := <-charChannel:
			if !ok {
				fmt.Println("channel closed")
				return
			}

			fmt.Println("worker received:", result)

		case <-done:
			fmt.Println("worker received stop signal")
			return
		}
	}
}

func main() {
	char := []string{"a", "b", "c"}
	charChannel := make(chan string, 3)
	done := make(chan struct{})

	for _, s := range char {
		charChannel <- s
	}

	close(charChannel)

	go worker(charChannel, done)

	time.Sleep(1 * time.Second)

	close(done)

	time.Sleep(500 * time.Millisecond)
}
