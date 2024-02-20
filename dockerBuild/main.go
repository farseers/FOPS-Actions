package main

import (
	"fmt"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"os"
	"strings"
)

func main() {
	go printProgress()

	progress <- "生成Dockerfile。"
	CreateDockerfile()

	progress <- "开始镜像打包。"
	// 打包
	var result = exec.RunShell("docker build -t "+With.DockerImage+" --network=host -f "+DockerfilePath+" "+DistRoot, progress, nil, DistRoot, true)
	if result == 0 {
		progress <- "镜像打包完成。"
	}

	// 等待退出
	waitProgress()

	if result != 0 {
		fmt.Println(flog.Red("镜像打包出错了"))
		os.Exit(-1)
	}
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
		fmt.Println(flog.Red("Dockerfile没有定义"))
		os.Exit(-1)
	}

	// 文件如果存在，则要先移除
	if file.IsExists(DockerfilePath) {
		_ = os.RemoveAll(DockerfilePath)
	}

	// 生成Dockerfile
	file.WriteString(DockerfilePath, dockerfileContent)
}
