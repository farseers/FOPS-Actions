package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
		execSuccess = device.clone(gitPath, git.GetAuthHub(), git.Branch, git.Branch, ctx)
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

func (device *gitDevice) clone_old(gitPath string, github string, branchOrCommitId string, ctx context.Context) bool {
	// 初始化参数切片，主程序是 timeout
	wait := exec.RunShellContext(ctx, "git", []string{"clone", "--depth=1", "-b", branchOrCommitId, github, gitPath}, nil, "", true)
	if exitCode := wait.WaitToChan(progress); exitCode != 0 {
		progress <- "Git克隆失败"
		return false
	}
	return true
}

// 克隆代码逻辑
// branchOrCommitId: 支持传入分支 或 commitID
// defaultBranch: 当branchOrCommitId传入的是分支,检出发现不存在此分支时,则自动回退到defaultBranch分支检出
func (device *gitDevice) clone2(gitPath string, github string, branchOrCommitId string, defaultBranch string, ctx context.Context) bool {
	// 1. 彻底清理并重建目录
	os.RemoveAll(gitPath)
	if err := os.MkdirAll(gitPath, 0755); err != nil {
		progress <- "创建目录失败: " + err.Error()
		return false
	}

	// 2. 初始化 Git 仓库
	initWait := exec.RunShellContext(ctx, "git", []string{"-C", gitPath, "init", "-q"}, nil, "", false)
	if initWait.WaitToChan(progress) != 0 {
		progress <- "Git初始化失败"
		return false
	}

	// 3. 添加远程仓库
	remoteWait := exec.RunShellContext(ctx, "git", []string{"-C", gitPath, "remote", "add", "origin", github}, nil, "", false)
	if remoteWait.WaitToChan(progress) != 0 {
		progress <- "添加远程仓库失败"
		return false
	}

	// 4. 定义 16位 CommitID 识别逻辑
	is16BitCommitID := regexp.MustCompile(`^[0-9a-fA-F]{16}$`).MatchString(branchOrCommitId)

	// 5. 执行 Fetch (核心修复点)
	var fetchArgs []string
	if is16BitCommitID {
		// 场景 A: CommitID - 使用 depth=1 配合通配符，已被证明对你的依赖库有效
		fetchArgs = []string{"-C", gitPath, "fetch", "--depth=1", "origin", "+refs/heads/*:refs/remotes/origin/*", "--tags"}
	} else {
		// 场景 B: 分支/Tag - 去掉 --depth=1，确保 Tree 对象完整，解决无法读取树的问题
		// 使用通配符拉取能更好地支持分支识别
		fetchArgs = []string{"-C", gitPath, "fetch", "origin", "+refs/heads/*:refs/remotes/origin/*", "--tags"}
	}

	fetchWait := exec.RunShellContext(ctx, "git", fetchArgs, nil, "", true)
	lstResult, exitCode := fetchWait.WaitToList()
	progress <- lstResult.First()

	// 6. 异常处理与自动回退
	if exitCode != 0 {
		if ctx.Err() != nil {
			progress <- "Git操作超时"
			return false
		}

		// 触发回退逻辑：非 CommitID、有默认分支
		if !is16BitCommitID && defaultBranch != "" {
			progress <- "目标 [" + branchOrCommitId + "] 不存在或损坏，尝试默认分支: " + defaultBranch

			// 回退也建议不带 depth 以求稳定
			retryWait := exec.RunShellContext(ctx, "git", []string{"-C", gitPath, "fetch", "origin", defaultBranch}, nil, "", true)
			if retryWait.WaitToChan(progress) != 0 {
				progress <- "Git拉取彻底失败"
				return false
			}
			branchOrCommitId = defaultBranch
		} else {
			lstResult.For(func(index int, item *string) {
				if index > 0 {
					progress <- *item
				}
			})
			progress <- "Git拉取失败: 无法获取 " + branchOrCommitId
			return false
		}
	}

	// 7. 检出目标
	// 使用 -f 强制覆盖，确保工作区干净
	checkoutArgs := []string{
		"-c", "advice.detachedHead=false",
		"-C", gitPath,
		"checkout", "-f", branchOrCommitId, "-q",
	}
	checkoutWait := exec.RunShellContext(ctx, "git", checkoutArgs, nil, "", true)

	if checkoutWait.WaitToChan(progress) != 0 {
		progress <- "Git检出失败: 无法定位到 " + branchOrCommitId
		return false
	}

	return true
}

// 1. 主入口：负责初始化环境、判断类型并分发
func (device *gitDevice) clone(gitPath string, github string, branchOrCommitId string, defaultBranch string, ctx context.Context) bool {
	// --- 1. 环境准备 ---
	os.RemoveAll(gitPath)
	if err := os.MkdirAll(gitPath, 0755); err != nil {
		progress <- gitPath + ", 创建目录失败: " + err.Error()
		return false
	}

	// --- 2. 类型判断 ---
	// 修正正则：长度 >=7 视为 CommitID，否则视为分支/Tag
	// 这样 "分支" 会进入分支逻辑，避免被误判为 CommitID 导致歧义报错
	isCommitID := regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`).MatchString(branchOrCommitId)

	if isCommitID {
		return device.cloneByCommitID(gitPath, github, branchOrCommitId, ctx)
	} else {
		return device.cloneByBranchOrTag(gitPath, github, branchOrCommitId, defaultBranch, ctx)
	}
}

// 2. 分支/Tag 拉取逻辑：支持自动回退
func (device *gitDevice) cloneByBranchOrTag(gitPath string, github string, branchName string, defaultBranch string, ctx context.Context) bool {
	wait := exec.RunShellContext(ctx, "git", []string{"clone", "--depth=1", "-b", branchName, github, gitPath}, nil, "", true)
	lstResult, exitCode := wait.WaitToList()

	// 分支不存在
	if exitCode != 0 {
		// 没有找到分支,尝试退回到默认分支
		if lstResult.ContainsAny("Could not find remote") && defaultBranch != "" {
			progress <- fmt.Sprintf("%s, 没有找到%s分支,尝试退回到%s分支", github, branchName, defaultBranch)
			wait = exec.RunShellContext(ctx, "git", []string{"clone", "--depth=1", "-b", defaultBranch, github, gitPath}, nil, "", true)
			lstResult, exitCode := wait.WaitToList()
			progress <- lstResult.First()
			if exitCode == 0 {
				return true
			}
		}
		progress <- "Git克隆失败"
		return false
	}

	progress <- lstResult.First()
	return true
}

// 3. CommitID 拉取逻辑：精确拉取
func (device *gitDevice) cloneByCommitID(gitPath string, github string, commitID string, ctx context.Context) bool {
	// 初始化仓库
	initWait := exec.RunShellContext(ctx, "git", []string{"-C", gitPath, "init", "-q"}, nil, "", false)
	if initWait.WaitToChan(progress) != 0 {
		progress <- gitPath + ", Git初始化失败"
		return false
	}

	// 添加远程仓库
	remoteWait := exec.RunShellContext(ctx, "git", []string{"-C", gitPath, "remote", "add", "origin", github}, nil, "", false)
	if remoteWait.WaitToChan(progress) != 0 {
		progress <- gitPath + ", 添加远程仓库失败"
		return false
	}

	// 1. 拉取
	fallbackWait := exec.RunShellContext(ctx, "git", []string{"-C", gitPath, "fetch", "origin", "+refs/heads/*:refs/remotes/origin/*", "--tags"}, nil, "", true)
	lstResult, exitCode := fallbackWait.WaitToList()
	progress <- lstResult.First()
	if exitCode != 0 {
		progress <- gitPath + ", Git Fetch 失败"
		return false
	}

	// 2. 强制检出 CommitID
	checkoutWait := exec.RunShellContext(ctx, "git", []string{"-c", "advice.detachedHead=false", "-C", gitPath, "checkout", "-f", commitID, "-q"}, nil, "", true)
	lstResult, exitCode = checkoutWait.WaitToList()
	progress <- lstResult.First()
	if exitCode != 0 {
		progress <- gitPath + ", Git检出失败: Commit ID " + commitID + " 不存在"
		return false
	}

	return true
}
