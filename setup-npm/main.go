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
	output, _ := exec.RunShellCommand("which npm", nil, "", false)
	for _, o := range output.ToArray() {
		if o == "/usr/bin/npm" {
			return
		}
	}

	// 没有安装git
	result, wait := exec.RunShell("apk add nodejs npm", nil, "", true)
	exec.SaveToChan(progress, result, wait)
}
