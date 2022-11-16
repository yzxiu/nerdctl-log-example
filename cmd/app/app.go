package main

import (
	"fmt"
	"time"
)

func main() {
	i := 5
	for i > 0 {
		fmt.Println(time.Now())
		time.Sleep(time.Second)
		i--
	}
}
