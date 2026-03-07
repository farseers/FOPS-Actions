package main

import (
	"github.com/farseer-go/utils/exec"
)

func main() {
	go printProgress()

	setupNpm()

	// 等待退出
	waitProgress()
}

// 安装npm
func setupNpm() {
	wait := exec.RunShell("which", []string{"npm"}, nil, "", false)
	output, _ := wait.WaitToList()
	for _, o := range output.ToArray() {
		if o == "/usr/bin/npm" {
			return
		}
	}

	// 没有安装git
	wait = exec.RunShell("apk", []string{"add", "nodejs", "npm"}, nil, "", true)
	wait.WaitToChan(progress)
}
