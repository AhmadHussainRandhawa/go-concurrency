package main

import (
	"fmt"
	"sync"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Worker Finished:", id)

}
func main() {
	var wg sync.WaitGroup

	wg.Add(3)

	go worker(1, &wg)
	go worker(2, &wg)
	go worker(3, &wg)

	wg.Wait()

	fmt.Println("All workers are done.")
}
