package main

import "fmt"

func producer(ch chan<- int) {
	defer close(ch)

	for i := 1; i <= 5; i++ {
		ch <- i
	}
}

func main() {
	ch := make(chan int)

	go producer(ch)

	for {
		value, ok := <-ch

		if !ok {
			break
		}
		fmt.Println(value)
	}

	fmt.Println("Done finished.")
}
