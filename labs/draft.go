package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	file, err := os.OpenFile("temp_0_1.json", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("cannot open")
	}
	defer file.Close()

	m := make(map[string]int)
	m["this"] = 1
	enc := json.NewEncoder(file)
	if err := enc.Encode(&m); err != nil {
		fmt.Println("err")
	}
}
