package main

import (
	"fmt"
	"github.com/farseer-go/utils/exec"
	"os"
	"strings"
)

func main() {
	go printProgress()

	loginDockerHub()

	// 上传完后，删除本地镜像
	defer exec.RunShellCommand("docker rmi "+With.DockerImage, nil, "", true)

	// 上传
	var result = exec.RunShell("docker push "+With.DockerImage, progress, nil, "", true)
	if result == 0 {
		progress <- "镜像上传完成。"
	}

	// 等待退出
	waitProgress()

	if result != 0 {
		fmt.Println("镜像上传出错了")
		os.Exit(-1)
	}
}

func loginDockerHub() {
	// 私有仓库，可以无用户名密码。
	if With.DockerUserName != "" && With.DockerUserPwd != "" {
		dockerHub := With.DockerHub
		if !strings.Contains(With.DockerHub, ".") {
			dockerHub = ""
		}
		var result = exec.RunShell("docker login "+dockerHub+" -u "+With.DockerUserName+" -p "+With.DockerUserPwd, progress, nil, "", true)
		if result != 0 {
			fmt.Println("镜像仓库登陆失败。")
			os.Exit(-1)
		}
	}

	progress <- "镜像仓库登陆成功。"
}
