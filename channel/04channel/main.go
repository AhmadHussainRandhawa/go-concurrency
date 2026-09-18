package main

import "fmt"

func producer(ch chan<- int) {
	ch <- 1
	ch <- 2
	ch <- 3

	close(ch) // Closing a channel means: no more values will be sent through this channel.
}

func consumer(ch <-chan int) {
	fmt.Println(<-ch)
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

func main() {
	ch := make(chan int)
	go producer(ch)
	consumer(ch)
}
