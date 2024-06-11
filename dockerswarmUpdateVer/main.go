package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/farseer-go/fs/core"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	go printProgress()

	// 更新到远程
	if With.RemoteClusterId > 0 {
		if !strings.HasSuffix(With.FopsAddr, "/") {
			With.FopsAddr += "/"
		}
		With.FopsAddr += "apps/updateDockerImage"

		avg := map[string]any{"appName": With.AppName, "dockerImage": With.DockerImage, "buildNumber": With.BuildNumber, "clusterId": With.RemoteClusterId, "dockerHub": With.DockerHub, "dockerUserName": With.DockerUserName, "dockerUserPwd": With.DockerUserPwd}
		bodyByte, _ := json.Marshal(avg)
		progress <- "开始更新远程fops：" + With.FopsAddr + " " + string(bodyByte)

		newRequest, _ := http.NewRequest("POST", With.FopsAddr, bytes.NewReader(bodyByte))
		newRequest.Header.Set("Content-Type", "application/json")

		// 读取配置
		client := &http.Client{}
		rsp, err := client.Do(newRequest)
		if err != nil {
			fmt.Println("更新远程fops的仓库版本失败：" + err.Error())
			waitProgress()
			time.Sleep(time.Second)
			os.Exit(-1)
		}

		apiRsp := core.NewApiResponseByReader[any](rsp.Body)
		if apiRsp.StatusCode != 200 {
			fmt.Printf("更新远程fops的仓库版本失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
			waitProgress()
			time.Sleep(time.Second)
			os.Exit(-1)
		}
		progress <- "更新成功：" + apiRsp.StatusMessage

		// 等待退出
		waitProgress()
		return
	}

	// 更新到本地
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
