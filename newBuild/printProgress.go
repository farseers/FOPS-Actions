package main

import (
	"fmt"
	"time"
)

// 接收执行时的消息
var progress = make(chan string, 1000)

// 输出消息
func printProgress() {
	for msg := range progress {
		fmt.Println(msg)
	}
}

// 等待消息清空
func waitProgress() {
	for len(progress) > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}
