package main

import (
	"fmt"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/utils/exec"
	"os"
)

func main() {
	go printProgress()

	if With.Proxy == "" {
		fmt.Println(flog.Red("未设置proxy"))
		os.Exit(-1)
	}

	cmd := fmt.Sprintf("git config --global http.https://github.com.proxy %s && git config --global https.https://github.com.proxy %s", With.Proxy, With.Proxy)
	exec.RunShell(cmd, progress, nil, "", true)

	// 等待退出
	waitProgress()
}
