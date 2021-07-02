package main

import (
	"fmt"
)

type I interface {
	Go() 
}

type Float float64

func (f *Float)	Change(v float64) {
	*f = Float(v)
}



func main() {
	fmt.Println("This is GO lang")
	num := Float(10.1)
	num.Change(1.0)
	fmt.Println(num)
}