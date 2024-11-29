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
		worker.AddGO(func() {
			result := false
			// 支持重试3次
			for tryCount := 1; tryCount < 3; tryCount++ {
				// 克隆或更新
				result = device.CloneOrPull(gitEO, progress, context.Background())
				if result {
					break
				}
				time.Sleep(time.Second * 3)
				progress <- fmt.Sprintf("尝试第%d次拉取\n", tryCount+1)
			}

			if !result {
				fmt.Println("拉取出错了")
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
	_, output := exec.RunShellCommand("which git", nil, "", false)
	for _, o := range output {
		if o == "/usr/bin/git" {
			return
		}
	}

	// 没有安装git
	exec.RunShell("apk add git", progress, nil, "", true)
}
