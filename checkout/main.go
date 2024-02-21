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

	for index, git := range With.Gits {
		// 克隆或更新
		result := device.CloneOrPull(git, progress, context.Background())
		if result {
			dest := filepath.Join(DistRoot, git.GetRelativePath())
			progress <- "源文件" + git.GetAbsolutePath() + " 复制到 " + dest
			file.CopyFolder(git.GetAbsolutePath(), dest)
		} else {
			fmt.Println("拉取出错了")
			os.Exit(-1)
		}
		if index+1 < len(With.Gits) {
			progress <- "---------------------------------------------------------"
		}
	}

	// 等待退出
	waitProgress()
}
