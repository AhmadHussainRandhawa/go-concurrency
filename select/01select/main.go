package main

import "fmt"

func main() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	ch1 <- "Hello from ch 1"
	ch2 <- "Hello from ch 2"

	// select selects a communication case that is ready.

	select {
	case msg := <-ch1:
		fmt.Println(msg)

	case msg := <-ch2:
		fmt.Println(msg)

	default:
		fmt.Println("Nothing Available")
	}

}
