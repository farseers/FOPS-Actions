package main

import (
	"bytes"
	"context"
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
	result, wait := exec.RunShell("git config --global credential.helper store", nil, "", true)
	exec.SaveToChan(progress, result, wait)
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
	result, wait := exec.RunShell("rm -rf "+gitPath, nil, "", true)
	if exitCode := exec.SaveToChan(progress, result, wait); exitCode != 0 {
		progress <- "Git清除失败"
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
	result, wait := exec.RunShellContext(ctx, "timeout 10 git -C "+savePath+" pull --rebase", nil, "", true)
	if exitCode := exec.SaveToChan(progress, result, wait); exitCode != 0 {
		progress <- "Git拉取失败"
		return false
	}
	return true
}

func (device *gitDevice) clone(gitPath string, github string, branch string, ctx context.Context) bool {
	bf := bytes.Buffer{}
	bf.WriteString("timeout 20 git clone --depth=1")
	if branch != "" {
		bf.WriteString(" -b " + branch)
	}
	bf.WriteString(" ")
	bf.WriteString(github)
	bf.WriteString(" ")
	bf.WriteString(gitPath)

	result, wait := exec.RunShellContext(ctx, bf.String(), nil, "", true)
	if exitCode := exec.SaveToChan(progress, result, wait); exitCode != 0 {
		progress <- "Git克隆失败"
		return false
	}
	return true
}

func (device *gitDevice) merge(gitPath string, branch string, ctx context.Context) bool {
	result, wait := exec.RunShellContext(ctx, "timeout 20 git pull origin main && git merge "+branch+" && git push", nil, gitPath, true)
	if exitCode := exec.SaveToChan(progress, result, wait); exitCode != 0 {
		progress <- "合并分支失败"
		return false
	}
	return true
}
