package main

import (
	"fmt"
	"time"
)

func someFunc(n string) {
	fmt.Println(n)
}

func main() {
	go someFunc("1")
	go someFunc("2")
	go someFunc("3")

	time.Sleep(time.Second * 1)
	fmt.Println("Hi")

}
