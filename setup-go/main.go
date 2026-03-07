package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/color"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"github.com/farseer-go/utils/http"
)

func main() {
	go printProgress()
	if With.GoVersion == "" {
		With.GoVersion = "go1.25.6"
		fmt.Println("GoVersion默认使用：" + color.Red(With.GoVersion))
	}

	if With.GoDownload == "" {
		With.GoDownload = "https://go.dev/dl/" + With.GoVersion + ".linux-amd64.tar.gz"
		//With.GoDownload = "https://studygolang.com/dl/golang/" + With.GoVersion + ".linux-amd64.tar.gz"
	}

	fmt.Printf("go环境要求为：%s\n", With.GoVersion)

	// 先判断本地是否有go目录
	if file.IsExists("/usr/local/go") {
		ver := getGoVersion()
		if ver == With.GoVersion {
			fmt.Printf("当前go环境正确：%s\n", ver)
			return
		}
		fmt.Printf("当前go环境不正确：%s\n", ver)
		fmt.Println(color.Yellow("移除旧目录/usr/local/go"))
		file.Delete("/usr/local/go")
	} else {
		fmt.Print("go程序未安装，将")
	}

	fmt.Println("开始下载go安装程序到:" + With.GoDownload)
	savePath := "/home/"
	fileName := collections.NewList(strings.Split(With.GoDownload, "/")...).Last()

	// 下载
	if _, err := http.Download(With.GoDownload, savePath+fileName, nil, 0, With.Proxy); err != nil {
		fmt.Println(color.Red(err.Error()))
		os.Exit(-1)
	}

	fmt.Println("下载完成，准备解压到：/usr/local/go")
	// tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
	if wait := exec.RunShell("bash", []string{"-c", fmt.Sprintf("tar -C /usr/local -xzf %s && rm -rf ./%s", fileName, fileName)}, nil, savePath, true); wait.Wait() == 0 {
		fmt.Println("解压完成。")
		exportEnv()
	}

	fmt.Printf("go环境安装完成，版本：%s\n", getGoVersion())

	// 等待退出
	waitProgress()
}

func getGoVersion() string {
	wait := exec.RunShell("/usr/local/go/bin/go", []string{"version"}, nil, "", false)
	listFromChan, _ := wait.WaitToList()
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
	wait := exec.RunShell("bash", []string{"-c", "echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile && source /etc/profile"}, nil, "", true)
	wait.Wait()

	wait = exec.RunShell("go", []string{"env", "-w", "GO111MODULE=on"}, nil, "", true)
	wait.Wait()

	wait = exec.RunShell("go", []string{"env", "-w", "GOPROXY=https://goproxy.cn,direct"}, nil, "", true)
	wait.Wait()

}
