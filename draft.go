package main

import (
	"fmt"
)

type I interface {
	Go() 
}

type Float float64

type Vertex struct {
	X, Y int
}

func (f *Float)	Change(v float64) {
	*f = Float(v)
}

var (
	x int = 1
	meg string = "This"
)



func main() {
	fmt.Println("This is GO lang")
	num := Float(10.1)
	num.Change(1.0)
	fmt.Println(num)
	arr := make([]int32, 0, 5)
	fmt.Printf("%T, len = %d, cap = %d, %v", arr, len(arr), cap(arr), arr)
}