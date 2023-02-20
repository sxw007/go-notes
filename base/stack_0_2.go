package main

import "fmt"

func b() {
	i := 0         // 因为下面的 fmt.Println() 接收的是 interface{}，i会逃逸到堆上
	fmt.Println(i) // func Println(a ...interface{}) (n int, err error) {...}
}

func main() {
	b()
}
