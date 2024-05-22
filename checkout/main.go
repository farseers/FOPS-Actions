package main

import (
	"context"
	"fmt"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/async"
	"github.com/farseer-go/utils/exec"
	"github.com/farseer-go/utils/file"
	"os"
	"path/filepath"
	"time"
)

func main() {
	go printProgress()
	setupGit()

	mLog := make(map[int][]string)

	// git操作驱动
	device := gitDevice{}
	worker := async.New()
	for index, git := range With.Gits {
		worker.Add(func() {
			result := false
			// 支持重试3次
			for tryCount := 1; tryCount < 4; tryCount++ {
				// 克隆或更新
				g := make(chan string, 1000)
				result = device.CloneOrPull(git, g, context.Background())
				mLog[index] = append(mLog[index], collections.NewListFromChan(g).ToArray()...)

				if result {
					break
				}
				time.Sleep(time.Second)
				mLog[index] = append(mLog[index], fmt.Sprintf("尝试第%d次拉取/n", tryCount+1))
			}

			if !result {
				fmt.Println("拉取出错了")
				os.Exit(-1)
			}
			dest := filepath.Join(DistRoot, git.GetRelativePath())
			mLog[index] = append(mLog[index], "源文件"+git.GetAbsolutePath()+" 复制到 "+dest)
			file.CopyFolder(git.GetAbsolutePath(), dest)
		})
	}

	_ = worker.Wait()
	for _, logs := range mLog {
		for _, log := range logs {
			fmt.Println(log)
		}
		fmt.Println("---------------------------------------------------------")
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
