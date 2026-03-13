package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"github.com/farseer-go/utils/str"
	"github.com/timandy/routine"
)

type gitDevice struct {
}

func (device *gitDevice) GetGitPath(gitHub string) string {
	if gitHub == "" {
		return ""
	}
	var gitName = device.GetName(gitHub)
	return GitRoot + gitName + "/"
}

func (device *gitDevice) GetName(gitHub string) string {
	if gitHub == "" {
		return ""
	}
	git := filepath.Base(gitHub)
	return str.CutRight(git, ".git")
}

func (device *gitDevice) RememberPassword() {
	args := []string{"config", "--global", "credential.helper", "store"}
	wait := exec.RunShell("git", args, nil, "", true)
	wait.WaitToChan(progress)
}

func (device *gitDevice) ExistsGitProject(gitPath string) bool {
	// 如果Git存放的目录不存在，则创建
	if !file.IsExists(GitRoot) {
		file.CreateDir766(GitRoot)
	}
	return file.IsExists(gitPath)
}

func (device *gitDevice) Clear(git GitEO) bool {
	// 获取Git存放的路径
	gitPath := git.GetAbsolutePath()
	err := os.RemoveAll(gitPath)
	if err != nil {
		progress <- "Git清除失败: " + err.Error()
		return false
	}
	return true
}

func (device *gitDevice) CloneOrPull(git GitEO, ctx context.Context) bool {
	if progress == nil {
		progress = make(chan string, 100)
	}

	// 先得到项目Git存放的物理路径
	gitPath := git.GetAbsolutePath()
	var execSuccess bool

	// 目录存在，且为非应用仓库时，使用git pull
	if device.ExistsGitProject(gitPath) && !git.IsApp {
		// git remote update
		execSuccess = device.pull(gitPath, ctx)
	} else {
		file.Delete(gitPath)
		execSuccess = device.clone(gitPath, git.GetAuthHub(), git.Branch, ctx)
		// 如果是应用仓库，则克隆后需要打印当前的CommitId
		if git.IsApp && execSuccess {

		}
	}
	return execSuccess
}

func (device *gitDevice) CloneOrPullAndDependent(lstGit []GitEO, ctx context.Context) bool {
	progress <- "开始拉取git代码"
	var sw sync.WaitGroup
	result := true
	for _, git := range lstGit {
		sw.Add(1)
		g := git
		routine.Go(func() {
			defer sw.Done()
			if !device.CloneOrPull(g, ctx) {
				result = false
			}
		})
	}
	sw.Wait()
	if result {
		progress <- "拉取完成。"
	}
	return result
}

func (device *gitDevice) pull(savePath string, ctx context.Context) bool {
	//exitCode := exec.RunShellContext(ctx, "timeout 10 git -C "+savePath+" pull origin "+branch+":"+branch+" --rebase", progress, nil, "", true)
	args := []string{"10", "git", "-C", savePath, "pull", "--rebase"}
	wait := exec.RunShellContext(ctx, "timeout", args, nil, "", true)
	if exitCode := wait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git拉取失败"
		return false
	}
	return true
}

func (device *gitDevice) clone(gitPath string, github string, branch string, ctx context.Context) bool {
	// 初始化参数切片，主程序是 timeout
	args := []string{"20", "git", "clone", "--depth=1"}
	// 动态根据条件追加参数
	if branch != "" {
		args = append(args, "-b", branch)
	}
	// 追加剩余的目标地址和本地路径
	args = append(args, github, gitPath)

	wait := exec.RunShellContext(ctx, "timeout", args, nil, "", true)
	if exitCode := wait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git克隆失败"
		return false
	}
	return true
}

func (device *gitDevice) merge(gitPath string, branch string, ctx context.Context) bool { // 拼接完整的 Shell 命令字符串
	// 使用 Args 方式调用 shell 执行该脚本
	args := []string{"-c", fmt.Sprintf("timeout 20 git pull origin main && git merge %s && git push", branch)}
	wait := exec.RunShellContext(ctx, "sh", args, nil, gitPath, true)
	if exitCode := wait.WaitToChan(progress); exitCode != 0 {
		progress <- "合并分支失败"
		return false
	}
	return true
}
