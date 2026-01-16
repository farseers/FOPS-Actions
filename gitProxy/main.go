package main

import (
	"fmt"
	"os"

	"github.com/farseer-go/fs/color"
	"github.com/farseer-go/utils/exec"
)

func main() {
	go printProgress()

	//setupGit()

	if With.Proxy == "" {
		fmt.Println(color.Red("未设置proxy"))
		os.Exit(-1)
	}

	//cmd := fmt.Sprintf("git config --global git config --global https.proxy  %s && git config --global https.https://github.com.proxy %s", With.Proxy, With.Proxy)
	cmd := fmt.Sprintf("git config --global https.proxy %s && git config --global http.proxy %s && git config --global https.sslVerify false && git config --global http.sslVerify false", With.Proxy, With.Proxy)
	result, wait := exec.RunShell(cmd, nil, "", true)
	exec.SaveToChan(progress, result, wait)

	// 等待退出
	waitProgress()
}

// 安装git
func setupGit() {
	output, _ := exec.RunShellCommand("which git", nil, "", false)
	for _, o := range output.ToArray() {
		if o == "/usr/bin/git" {
			return
		}
	}

	// 没有安装git
	result, wait := exec.RunShell("apk add git", nil, "", true)
	exec.SaveToChan(progress, result, wait)
}
