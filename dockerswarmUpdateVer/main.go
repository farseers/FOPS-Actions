package main

import (
	"bytes"
	"context"
	"crypto/tls"
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
		fopsAddr := With.FopsAddr + "apps/updateDockerImage"

		bodyByte, _ := json.Marshal(map[string]any{"appName": With.AppName, "dockerImage": With.DockerImage, "buildNumber": With.BuildNumber, "clusterId": With.RemoteClusterId, "dockerHub": With.DockerHub, "dockerUserName": With.DockerUserName, "dockerUserPwd": With.DockerUserPwd})
		progress <- "开始更新远程fops：" + fopsAddr + " " + string(bodyByte)

		newRequest, _ := http.NewRequest("POST", fopsAddr, bytes.NewReader(bodyByte))
		newRequest.Header.Set("Content-Type", "application/json")

		// 读取配置
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 不验证 HTTPS 证书
				},
			},
		}
		rsp, err := client.Do(newRequest)
		if err != nil {
			progress <- "更新远程fops的仓库版本失败：" + err.Error()
			getDockerLog()
			waitProgress()
			time.Sleep(time.Second)
			os.Exit(-1)
		}

		apiRsp := core.NewApiResponseByReader[any](rsp.Body)
		if apiRsp.StatusCode != 200 {
			progress <- fmt.Sprintf("更新远程fops的仓库版本失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
			getDockerLog()
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

// 更新失败时，获取docker日志
func getDockerLog() {
	fopsAddr := With.FopsAddr + "apps/logs/dockerSwarm"
	bodyByte, _ := json.Marshal(map[string]any{"appName": With.AppName, "tail": 50})
	newRequest, _ := http.NewRequest("POST", fopsAddr, bytes.NewReader(bodyByte))
	newRequest.Header.Set("Content-Type", "application/json")

	// 读取配置
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 不验证 HTTPS 证书
			},
		},
	}
	rsp, err := client.Do(newRequest)
	if err != nil {
		fmt.Println("查询Docker日志失败：" + err.Error())
		return
	}

	apiRsp := core.NewApiResponseByReader[string](rsp.Body)
	if apiRsp.StatusCode != 200 {
		fmt.Printf("查询Docker日志失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
		return
	}
	fmt.Println(apiRsp.Data)
}
