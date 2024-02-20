package main

import (
	"context"
	"github.com/farseer-go/utils/file"
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
	if device.CloneOrPull(git, progress, context.Background()) {
		dest := filepath.Join(DistRoot, git.GetRelativePath())
		progress <- "源文件" + git.GetAbsolutePath() + " 复制到 " + dest
		file.CopyFolder(git.GetAbsolutePath(), dest)
	}

	// 等待退出
	waitProgress()
}
