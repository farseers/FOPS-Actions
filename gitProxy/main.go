package main

import (
	"fmt"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/utils/exec"
	"os"
)

func main() {
	go printProgress()

	//setupGit()

	if With.Proxy == "" {
		fmt.Println(flog.Red("未设置proxy"))
		os.Exit(-1)
	}

	//cmd := fmt.Sprintf("git config --global git config --global https.proxy  %s && git config --global https.https://github.com.proxy %s", With.Proxy, With.Proxy)
	cmd := fmt.Sprintf("git config --global https.proxy %s && git config --global http.proxy %s && git config --global https.sslVerify false && git config --global http.sslVerify false", With.Proxy, With.Proxy)
	exec.RunShell(cmd, progress, nil, "", true)

	// 等待退出
	waitProgress()
}

// 安装git
func setupGit() {
	_, output := exec.RunShellCommand("which git", nil, "", false)
	for _, o := range output {
		if o == "/usr/bin/git" {
			return
		}
	}

	// 没有安装git
	exec.RunShell("apk add git", progress, nil, "", true)
}
