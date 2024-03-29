package main

import (
	"bytes"
	"context"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"github.com/farseer-go/utils/str"
	"github.com/timandy/routine"
	"path/filepath"
	"sync"
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

func (device *gitDevice) RememberPassword(progress chan string) {
	exec.RunShell("git config --global credential.helper store", progress, nil, "", true)
}

func (device *gitDevice) ExistsGitProject(gitPath string) bool {
	// 如果Git存放的目录不存在，则创建
	if !file.IsExists(GitRoot) {
		file.CreateDir766(GitRoot)
	}
	return file.IsExists(gitPath)
}

func (device *gitDevice) Clear(git GitEO, progress chan string) bool {
	// 获取Git存放的路径
	gitPath := git.GetAbsolutePath()
	exitCode := exec.RunShell("rm -rf "+gitPath, progress, nil, "", true)
	if exitCode != 0 {
		progress <- "Git清除失败"
		return false
	}
	return true
}

func (device *gitDevice) CloneOrPull(git GitEO, progress chan string, ctx context.Context) bool {
	if progress == nil {
		progress = make(chan string, 100)
	}

	// 先得到项目Git存放的物理路径
	gitPath := git.GetAbsolutePath()
	var execSuccess bool

	// 存在则使用pull
	if device.ExistsGitProject(gitPath) {
		execSuccess = device.pull(gitPath, progress, ctx)
	} else {
		execSuccess = device.clone(gitPath, git.GetAuthHub(), git.Branch, progress, ctx)
	}
	return execSuccess
}

func (device *gitDevice) CloneOrPullAndDependent(lstGit []GitEO, progress chan string, ctx context.Context) bool {
	progress <- "开始拉取git代码"
	var sw sync.WaitGroup
	result := true
	for _, git := range lstGit {
		sw.Add(1)
		g := git
		routine.Go(func() {
			defer sw.Done()
			if !device.CloneOrPull(g, progress, ctx) {
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

func (device *gitDevice) pull(savePath string, progress chan string, ctx context.Context) bool {
	exitCode := exec.RunShellContext(ctx, "git -C "+savePath+" pull --rebase", progress, nil, "", true)
	if exitCode != 0 {
		progress <- "Git拉取失败"
		return false
	}
	return true
}

func (device *gitDevice) clone(gitPath string, github string, branch string, progress chan string, ctx context.Context) bool {
	bf := bytes.Buffer{}
	bf.WriteString("git clone --depth=1")
	if branch != "" {
		bf.WriteString(" -b " + branch)
	}
	bf.WriteString(" ")
	bf.WriteString(github)
	bf.WriteString(" ")
	bf.WriteString(gitPath)

	exitCode := exec.RunShellContext(ctx, bf.String(), progress, nil, "", true)
	if exitCode != 0 {
		progress <- "Git克隆失败"
		return false
	}
	return true
}
