package main

import (
	"fmt"
	"time"
)

func main() {
	result := make(chan string)

	go func() {
		time.Sleep(3 * time.Second)
		result <- "SUCCESS"
	}()

	select {
	case msg := <-result:
		fmt.Println(msg)

	case <-time.After(2 * time.Second):
		fmt.Println("Timeout")

	}
}
