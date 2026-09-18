package main

import (
	"fmt"
	"runtime"
	"time"
)

func worker(id int) {
	for i := 0; i < 5; i++ {
		fmt.Println("worker", id, "iteration", i)
	}
}

func main() {
	runtime.GOMAXPROCS(2)

	for i := 0; i < 10; i++ {
		go worker(i)
	}

	time.Sleep(time.Second)
}
