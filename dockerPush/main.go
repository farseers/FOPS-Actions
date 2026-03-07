package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/utils/exec"
)

func main() {
	go printProgress()
	loginDockerHub()

	// 重试5次
	for tryCount := 0; tryCount < 5; tryCount++ {
		// 上传
		wait := exec.RunShell("docker", []string{"push", With.DockerImage}, nil, "", true)
		if exitCode := wait.WaitToChan(progress); exitCode == 0 {
			progress <- "镜像上传完成。"
			waitProgress()
			return
		}
		time.Sleep(3 * time.Second)
		progress <- fmt.Sprintf("尝试第%d次推送\n", tryCount+1)
	}
	// 上传完后，删除本地镜像
	removeImage()

	// 等待退出
	waitProgress()
	fmt.Println("镜像上传出错了")
	os.Exit(-1)
}

func loginDockerHub() {
	// 私有仓库，可以无用户名密码。
	if With.DockerUserName != "" && With.DockerUserPwd != "" {
		dockerHub := With.DockerHub
		if !strings.Contains(With.DockerHub, ".") {
			dockerHub = ""
		}

		// 重试5次
		for tryCount := 0; tryCount < 5; tryCount++ {
			wait := exec.RunShell("docker", []string{"login", dockerHub, "-u", With.DockerUserName, "-p", With.DockerUserPwd}, nil, "", true)
			if exitCode := wait.WaitToChan(progress); exitCode == 0 {
				progress <- "镜像仓库登陆成功。"
				return
			}
			time.Sleep(3 * time.Second)
			progress <- fmt.Sprintf("尝试第%d次登陆\n", tryCount+1)
		}
		// 上传完后，删除本地镜像
		removeImage()
		fmt.Println("镜像仓库登陆失败。")
		os.Exit(-1)
	}
}

func removeImage() {
	// 上传完后，删除本地镜像
	exec.RunShell("docker", []string{"rmi", With.DockerImage}, nil, "", false)
}
