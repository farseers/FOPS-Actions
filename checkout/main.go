package main

import (
	"context"
	"fmt"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"os"
	"path/filepath"
	"time"
)

func main() {
	go printProgress()

	setupGit()

	// git操作驱动
	device := gitDevice{}

	for index, git := range With.Gits {
		result := false
		// 支持重试3次
		for tryCount := 1; tryCount < 4; tryCount++ {
			// 克隆或更新
			if result = device.CloneOrPull(git, progress, context.Background()); result {
				break
			}
			time.Sleep(time.Second)
			fmt.Printf("尝试第%d次拉取/n", tryCount+1)
		}

		if !result {
			fmt.Println("拉取出错了")
			os.Exit(-1)
		}
		dest := filepath.Join(DistRoot, git.GetRelativePath())
		progress <- "源文件" + git.GetAbsolutePath() + " 复制到 " + dest
		file.CopyFolder(git.GetAbsolutePath(), dest)

		if index+1 < len(With.Gits) {
			progress <- "---------------------------------------------------------"
		}
	}

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
