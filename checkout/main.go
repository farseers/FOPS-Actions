package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/farseer-go/fs/async"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
)

func main() {
	go printProgress()
	//setupGit()

	// git操作驱动
	device := gitDevice{}
	worker := async.New()
	for index := 0; index < len(With.Gits); index++ {
		gitEO := With.Gits[index]
		gitPath := gitEO.GetAbsolutePath()
		authHub := gitEO.GetAuthHub()
		worker.AddGO(func() {
			execSuccess := false
			// 支持重试3次
			for tryCount := 1; tryCount < 3; tryCount++ {
				file.Delete(gitPath)
				execSuccess = device.clone(gitPath, authHub, gitEO.Branch, context.Background())
				if execSuccess {
					break
				}
				time.Sleep(time.Second * 3)
				progress <- fmt.Sprintf("尝试第%d次拉取: %s\n", tryCount+1, gitEO.Hub)
			}

			if !execSuccess {
				fmt.Println("拉取出错了: " + gitEO.Hub)
				os.Exit(-1)
			}

			dest := filepath.Join(DistRoot, gitEO.GetRelativePath())
			progress <- "源文件" + gitEO.GetAbsolutePath() + " 复制到 " + dest
			file.CopyFolder(gitEO.GetAbsolutePath(), dest)
		})
	}
	_ = worker.Wait()

	// 等待退出
	waitProgress()
}

// 安装git
func setupGit() {
	wait := exec.RunShell("which", []string{"git"}, nil, "", false)
	output, _ := wait.WaitToList()
	for _, o := range output.ToArray() {
		if o == "/usr/bin/git" {
			return
		}
	}

	// 没有安装git
	wait = exec.RunShell("apk", []string{"add", "git"}, nil, "", true)
	wait.WaitToChan(progress)
}
