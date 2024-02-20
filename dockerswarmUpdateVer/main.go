package main

import (
	"context"
	"os"
)

func main() {
	go printProgress()

	dockerSwarmDevice := dockerSwarmDevice{}
	// 首次创建还是更新镜像
	if dockerSwarmDevice.ExistsDocker(With.AppName) {
		// 更新镜像
		if !dockerSwarmDevice.SetImages(With.AppName, With.DockerImage, progress, context.Background()) {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	} else {
		// 创建容器服务
		if !dockerSwarmDevice.CreateService(With.AppName, With.DockerNodeRole, With.AdditionalScripts, With.DockerNetwork, With.DockerReplicas, With.DockerImage, progress, context.Background()) {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	}

	// 等待退出
	waitProgress()
}
