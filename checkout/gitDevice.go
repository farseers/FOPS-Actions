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

func (device *gitDevice) pullBranch(savePath string, branch string, ctx context.Context) bool {
	// 1. 切换到指定分支
	checkoutArgs := []string{"10", "git", "-C", savePath, "checkout", branch}
	checkoutWait := exec.RunShellContext(ctx, "timeout", checkoutArgs, nil, "", true)
	if exitCode := checkoutWait.WaitToChan(progress); exitCode != 0 {
		progress <- "切换分支失败: " + branch
		return false
	}

	// 2. 拉取并更新该分支
	// 注意：明确指定 origin 和 branch 可以确保即使本地没追踪也能拉取成功
	pullArgs := []string{"20", "git", "-C", savePath, "pull", "origin", branch, "--rebase"}
	pullWait := exec.RunShellContext(ctx, "timeout", pullArgs, nil, "", true)
	if exitCode := pullWait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git拉取更新失败"
		return false
	}

	return true
}

func (device *gitDevice) clone_old(gitPath string, github string, branchOrCommitId string, ctx context.Context) bool {
	// 初始化参数切片，主程序是 timeout
	args := []string{"20", "git", "clone", "--depth=1"}
	// 动态根据条件追加参数
	if branchOrCommitId != "" {
		args = append(args, "-b", branchOrCommitId)
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
func (device *gitDevice) clone(gitPath string, github string, branchOrCommitId string, ctx context.Context) bool {
	// 1. 彻底清理并重建目录，确保环境纯净
	os.RemoveAll(gitPath)
	if err := os.MkdirAll(gitPath, 0755); err != nil {
		progress <- "创建目录失败: " + err.Error()
		return false
	}

	// 2. 初始化 Git 仓库 (使用 timeout 10)
	initArgs := []string{"-C", gitPath, "init", "-q"}
	initWait := exec.RunShellContext(ctx, "git", initArgs, nil, "", false)
	if exitCode := initWait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git初始化失败"
		return false
	}

	// 3. 添加远程仓库地址 (使用 timeout 10)
	remoteArgs := []string{"-C", gitPath, "remote", "add", "origin", github}
	remoteWait := exec.RunShellContext(ctx, "git", remoteArgs, nil, "", false)
	if exitCode := remoteWait.WaitToChan(progress); exitCode != 0 {
		progress <- "添加远程仓库失败"
		return false
	}

	// 4. 拉取远程引用 (使用 timeout 30)
	// 关键：fetch 所有分支的头端 (+refs/heads/*)，配合 --depth=1 性能最优
	fetchArgs := []string{"30", "git", "-C", gitPath, "fetch", "--depth=1", "origin", "+refs/heads/*:refs/remotes/origin/*"}
	fetchWait := exec.RunShellContext(ctx, "timeout", fetchArgs, nil, "", true)
	if exitCode := fetchWait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git拉取远程引用失败"
		return false
	}

	// 5. 检出目标代码 (使用 timeout 20)
	// -c advice.detachedHead=false：彻底禁用切换到 Commit ID 时的长篇建议（防止误报）
	// -q：静默模式，只在真正报错时产生输出
	checkoutArgs := []string{"20", "git", "-c", "advice.detachedHead=false", "-C", gitPath, "checkout", branchOrCommitId, "-q"}
	checkoutWait := exec.RunShellContext(ctx, "timeout", checkoutArgs, nil, "", true)
	// 因为 WaitToChan 是指针方法，这里通过变量 checkoutWait 调用
	if exitCode := checkoutWait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git检出失败: 找不到分支或提交 " + branchOrCommitId
		return false
	}

	//progress <- "Git代码同步完成"
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
