package main

import (
	"context"
	"fmt"
	"github.com/farseer-go/utils/file"
	"os"
	"path/filepath"
)

func main() {
	go printProgress()

	// git操作驱动
	device := gitDevice{}
	git := GitEO{
		Hub:      With.GitHub,
		Branch:   With.GitBranch,
		UserName: With.GitUserName,
		UserPwd:  With.GitUserPwd,
		Path:     With.GitPath,
	}
	// 克隆或更新
	progress <- "开始拉取git:" + git.Hub
	result := device.CloneOrPull(git, progress, context.Background())
	if result {
		dest := filepath.Join(DistRoot, git.GetRelativePath())
		progress <- "源文件" + git.GetAbsolutePath() + " 复制到 " + dest
		file.CopyFolder(git.GetAbsolutePath(), dest)
	}

	// 等待退出
	waitProgress()

	if !result {
		fmt.Println("拉取出错了")
		os.Exit(-1)
	}
}
