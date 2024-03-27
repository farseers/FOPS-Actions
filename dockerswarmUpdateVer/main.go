package main

import (
	"context"
	"os"
)

func main() {
	go printProgress()

	swarmDevice := dockerSwarmDevice{}
	// 首次创建还是更新镜像
	if swarmDevice.ExistsDocker(With.AppName) {
		// 更新镜像
		if !swarmDevice.SetImages(With.AppName, With.DockerImage, progress, context.Background()) {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	} else {
		// 创建容器服务
		if !swarmDevice.CreateService(With.AppName, With.DockerNodeRole, With.AdditionalScripts, With.DockerNetwork, With.DockerReplicas, With.DockerImage, progress, context.Background()) {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	}

	// 等待退出
	waitProgress()
}
