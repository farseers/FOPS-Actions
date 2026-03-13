package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/fs/color"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
)

func main() {
	go printProgress()

	progress <- "生成Dockerfile。"
	CreateDockerfile()

	// 打包
	progress <- "开始镜像打包。"

	// 1. 初始化基础参数列表
	args := []string{"build", "-t", With.DockerImage, "--network=host", "-f", DockerfilePath}
	// 2. 根据条件追加 --build-arg 参数
	if With.Proxy != "" {
		args = append(args,
			"--build-arg", fmt.Sprintf("HTTP_PROXY=%s", With.Proxy),
			"--build-arg", fmt.Sprintf("HTTPS_PROXY=%s", With.Proxy),
			"--build-arg", "NO_PROXY=localhost,127.0.0.1",
		)
	}
	// 3. 最后追加构建上下文路径 (Context)
	args = append(args, DistRoot)

	// 重试5次
	for tryCount := 0; tryCount < 5; tryCount++ {
		wait := exec.RunShell("docker", args, nil, DistRoot, true)
		if exitCode := wait.WaitToChan(progress); exitCode == 0 {
			progress <- "镜像打包完成。"
			waitProgress()
			return
		}
		time.Sleep(1 * time.Second)
		progress <- fmt.Sprintf("尝试第%d次重新打包\n", tryCount+1)
	}
	// 等待退出
	fmt.Println("镜像打包出错了")
	waitProgress()
	os.Exit(-1)
}

func CreateDockerfile() {
	// 如果没有自定义，则使用应用仓库根目录的Dockerfile文件
	if With.DockerfilePath == "" {
		With.DockerfilePath = "Dockerfile"
	} else {
		// 自定义Dockerfile路径
		if strings.HasPrefix(With.DockerfilePath, "/") {
			With.DockerfilePath = With.DockerfilePath[1:]
		} else if strings.HasPrefix(With.DockerfilePath, "./") {
			With.DockerfilePath = With.DockerfilePath[2:]
		}
	}

	dockerfileContent := file.ReadString(With.AppAbsolutePath + With.DockerfilePath)
	if dockerfileContent == "" {
		fmt.Println(color.Red("Dockerfile没有定义"))
		os.Exit(-1)
	}

	// 文件如果存在，则要先移除
	if file.IsExists(DockerfilePath) {
		_ = os.RemoveAll(DockerfilePath)
	}

	// 生成Dockerfile
	file.WriteString(DockerfilePath, dockerfileContent)
}
