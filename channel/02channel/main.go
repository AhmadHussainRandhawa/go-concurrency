package main

import (
	"fmt"
	"sync"
)

func worker(n int, myChannel chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	myChannel <- n
}

func main() {
	myChannel := make(chan int, 3)
	var wg sync.WaitGroup

	wg.Add(3)

	go worker(1, myChannel, &wg)
	go worker(2, myChannel, &wg)
	go worker(3, myChannel, &wg)

	wg.Wait()

	msg := <-myChannel // block, if no wg.wait
	fmt.Println(msg)

	msg1 := <-myChannel
	fmt.Println(msg1)

	msg2 := <-myChannel
	fmt.Println(msg2)
}
