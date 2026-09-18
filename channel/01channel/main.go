package main

import "fmt"

func main() {
	mychannel := make(chan string)

	go func() {
		mychannel <- "Data"
	}()

	msg := <-mychannel // Main goroutine blocks here; other runnable goroutines can be scheduled --> When the send completes, the main goroutine becomes runnable again.

	fmt.Println(msg)

}
