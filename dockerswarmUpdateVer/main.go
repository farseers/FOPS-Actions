package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/docker"
	"github.com/farseer-go/fs/core"
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

		isSuccess := false
		// 尝试5次
		for i := 0; i < 10; i++ {
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
			progress <- fmt.Sprintf("尝试第%d次更新", i+1)
			rsp, err := client.Do(newRequest)
			if err != nil {
				progress <- "更新远程fops的仓库版本失败：" + err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			apiRsp := core.NewApiResponseByReader[any](rsp.Body)
			if apiRsp.StatusCode != 200 {
				progress <- fmt.Sprintf("更新远程fops的仓库版本失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
				time.Sleep(10 * time.Second)
				continue
			}

			isSuccess = true
			progress <- "更新成功：" + apiRsp.StatusMessage
			break
		}

		// 3次还是失败时，读取远程docker日志
		if !isSuccess {
			getDockerLog()
			waitProgress()
			time.Sleep(time.Second)
			os.Exit(-1)
		}

		// 等待退出
		waitProgress()
		return
	}

	// 更新到本地
	dockerClient := docker.NewClient()
	// 首次创建还是更新镜像
	if exists, _ := dockerClient.Service.Exists(With.AppName); exists {
		// 更新镜像
		if err := dockerClient.Service.SetImages(With.AppName, With.DockerImage); err != nil {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	} else {
		// 创建容器服务
		err := dockerClient.Service.Create(With.AppName, With.DockerNodeRole, With.AdditionalScripts, With.DockerNetwork, With.DockerReplicas, With.DockerImage, 0, "")
		if err != nil {
			progress <- "创建服务时出错：" + err.Error()
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	}

	progress <- "镜像更新成功：" + With.DockerImage
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
