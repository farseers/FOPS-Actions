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
	// 没有安装npm
	if len(output) == 0 || <-output != "/usr/bin/npm" {
		exec.RunShell("apk add nodejs npm && npm install -g cnpm --registry=https://registry.npm.taobao.org", progress, nil, "", true)
	}
}
