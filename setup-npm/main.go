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
	_, output := exec.RunShellCommand("which npm", nil, "", false)
	for _, o := range output {
		if o == "/usr/bin/npm" {
			return
		}
	}

	// 没有安装git
	exec.RunShell("apk add nodejs npm", progress, nil, "", true)
}
