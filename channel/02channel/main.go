package main

import "fmt"

func main() {
	jobs := make(chan int, 3)

	go func() {
		jobs <- 1
		jobs <- 2
		jobs <- 3
		jobs <- 4
	}()

	fmt.Println(<-jobs)
	fmt.Println(<-jobs)
	fmt.Println(<-jobs)
	fmt.Println(<-jobs)
}
