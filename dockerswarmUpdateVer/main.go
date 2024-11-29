package main

import (
	"bytes"
	"crypto/tls"

	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/docker"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/snc"
)

func main() {
	go printProgress()

	// 更新到远程
	if !With.IsLocal {
		if !strings.HasSuffix(With.FopsAddr, "/") {
			With.FopsAddr += "/"
		}
		fopsAddr := With.FopsAddr + "apps/updateDockerImage"

		bodyByte, _ := snc.Marshal(map[string]any{"appName": With.AppName, "dockerImage": With.DockerImage, "updateDelay": With.UpdateDelay, "buildNumber": With.BuildNumber, "dockerHub": With.DockerHub, "dockerUserName": With.DockerUserName, "dockerUserPwd": With.DockerUserPwd})
		progress <- "开始更新远程fops：" + fopsAddr + " " + string(bodyByte)

		isSuccess := false
		// 尝试10次
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
				time.Sleep(20 * time.Second)
				continue
			}

			apiRsp := core.NewApiResponseByReader[any](rsp.Body)
			if apiRsp.StatusCode != 200 {
				progress <- fmt.Sprintf("更新远程fops的仓库版本失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
				time.Sleep(20 * time.Second)
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
	dockerClient.SetChar(progress)
	// 首次创建还是更新镜像
	if exists, _ := dockerClient.Service.Exists(With.AppName); exists {
		// 更新镜像
		if err := dockerClient.Service.SetImages(With.AppName, With.DockerImage, With.UpdateDelay); err != nil {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	} else {
		// 创建容器服务
		err := dockerClient.Service.Create(With.AppName, With.DockerNodeRole, With.AdditionalScripts, With.DockerNetwork, With.DockerReplicas, With.DockerImage, With.LimitCpus, With.LimitMemory)
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
	bodyByte, _ := snc.Marshal(map[string]any{"appName": With.AppName, "tail": 50})
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
