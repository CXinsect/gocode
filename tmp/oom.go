package main

import (
	"fmt"
	"time"
)

var i int = 0

func oom_kill_process() {
	if i%5 == 0 && i != 10 {
		fmt.Println("out of memory")
	} else if i == 10 {
		i = 0
	}
	i++
}
func main() {
	for {
		oom_kill_process()
		time.Sleep(time.Second * 10)
	}

}
