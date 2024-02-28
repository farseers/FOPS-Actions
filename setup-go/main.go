package main

import (
	"fmt"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"github.com/farseer-go/utils/http"
	"os"
	"strings"
)

func main() {
	go printProgress()
	if With.GoVersion == "" {
		fmt.Println("GoVersion默认使用：" + flog.Red("go1.22.0"))
		With.GoVersion = "go1.22.0"
	}

	if With.GoDownload == "" {
		//With.GoDownload = "https://go.dev/dl/" + With.GoVersion + ".linux-amd64.tar.gz"
		With.GoDownload = "https://studygolang.com/dl/golang/" + With.GoVersion + ".linux-amd64.tar.gz"
	}

	fmt.Printf("go环境要求为：%s\n", flog.Blue(With.GoVersion))

	// 先判断本地是否有go目录
	if file.IsExists("/usr/local/go") {
		ver := getGoVersion()
		if ver == With.GoVersion {
			fmt.Printf("当前go环境正确：%s\n", flog.Green(ver))
			return
		}
		fmt.Printf("当前go环境不正确：%s\n", flog.Red(ver))
		fmt.Println(flog.Yellow("移除旧目录/usr/local/go"))
		file.Delete("/usr/local/go")
	}

	fmt.Println("开始下载go安装程序:" + With.GoDownload)
	savePath := "/home/"
	fileName := collections.NewList(strings.Split(With.GoDownload, "/")...).Last()
	if err := http.Download(With.GoDownload, savePath+fileName, 0, With.Proxy); err != nil {
		fmt.Println(flog.Red(err.Error()))
		os.Exit(-1)
	}

	fmt.Println("下载完成，准备解压到：/usr/local/go")
	// tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
	if result, _ := exec.RunShellCommand("tar -C /usr/local -xzf "+fileName+"&& rm -rf ./"+fileName, nil, savePath, true); result == 0 {
		fmt.Println("解压完成。")
		exportEnv()
	}

	fmt.Printf("go环境安装完成，版本：%s\n", flog.Blue(getGoVersion()))

	// 等待退出
	waitProgress()
}

func getGoVersion() string {
	_, receiveOutput := exec.RunShellCommand("/usr/local/go/bin/go version", nil, "", false)
	listFromChan := collections.NewList(receiveOutput...)
	fmt.Println(listFromChan.Last())
	vers := strings.Split(listFromChan.Last(), " ")
	var ver string
	if len(vers) > 2 {
		ver = vers[2]
	}
	return ver
}

// 设置环境变量
func exportEnv() {
	// export PATH=$PATH:/usr/local/go/bin
	exec.RunShellCommand("export PATH=$PATH:/usr/local/go/bin", nil, "", true)
	//exec.RunShellCommand("go env -w GO111MODULE=on", nil, "", true)
	//exec.RunShellCommand("go env -w GOPROXY=https://goproxy.cn,direct", nil, "", true)
}
