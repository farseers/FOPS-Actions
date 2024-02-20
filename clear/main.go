package main

import (
	"github.com/farseer-go/utils/file"
)

func main() {
	go printProgress()

	progress <- "初始化环境。"

	// 先删除之前编译的目标文件
	file.ClearFile(DistRoot)

	// 自动创建目录
	progress <- "自动创建目录。"

	if !file.IsExists(FopsRoot) {
		file.CreateDir766(FopsRoot)
	}
	if !file.IsExists(NpmModulesRoot) {
		file.CreateDir766(NpmModulesRoot)
	}
	if !file.IsExists(DistRoot) {
		file.CreateDir766(DistRoot)
	}
	if !file.IsExists(KubeRoot) {
		file.CreateDir766(KubeRoot)
	}
	if !file.IsExists(GitRoot) {
		file.CreateDir766(GitRoot)
	}
	progress <- "前置检查通过。"

	// 等待退出
	waitProgress()
}
