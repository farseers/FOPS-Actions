package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/docker"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/exception"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/http"
)

// 定义一个全局复用的 Client（只需要初始化一次）
// var httpClient = &http.Client{
// 	Timeout: 10 * time.Second, // 必须设置总超时
// 	Transport: &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 不验证 HTTPS 证书
// 		MaxIdleConns:    100,
// 		IdleConnTimeout: 90 * time.Second,
// 	},
// }

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
			progress <- fmt.Sprintf("尝试第%d次更新", i+1)
			// 读取配置
			apiRsp, statusCode, err := http.PostJson[core.ApiResponse[any]](fopsAddr, nil, bodyByte, 0)
			if err != nil {
				progress <- "更新远程fops的仓库版本失败：" + err.Error()
				time.Sleep(20 * time.Second)
				continue
			}
			if apiRsp.StatusCode != 200 {
				progress <- fmt.Sprintf("更新远程fops的仓库版本失败（%v）：%s", statusCode, apiRsp.StatusMessage)
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
	// 首次创建还是更新镜像
	if exists := dockerClient.Service.Exists(With.AppName); exists {
		// 更新镜像
		wait := dockerClient.Service.SetImages(With.AppName, With.DockerImage, With.UpdateDelay)
		if exitCode := wait.WaitToChan(progress); exitCode != 0 {
			// 等待退出
			waitProgress()
			os.Exit(-1)
		}
	} else {
		// 准备配置文件
		lastVersion, err := dockerClient.Config.GetLastVersion(With.AppName)
		exception.ThrowRefuseExceptionError(err)

		// 创建容器服务
		wait := dockerClient.Service.Create(With.AppName, With.DockerNodeRole, With.AdditionalScripts, With.DockerNetwork, With.DockerReplicas, With.DockerImage, With.LimitCpus, With.LimitMemory, docker.ConfigTarget{
			Name:   lastVersion.Spec.Name,
			Target: "/app/config.yaml",
		})
		if exitCode := wait.WaitToChan(progress); exitCode != 0 {
			progress <- "创建服务时出错"
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

	// 读取配置
	apiRsp, statusCode, err := http.PostJson[core.ApiResponse[string]](fopsAddr, nil, bodyByte, 0)
	if err != nil {
		fmt.Println("查询Docker日志失败：" + err.Error())
		return
	}
	if apiRsp.StatusCode != 200 {
		fmt.Printf("查询Docker日志失败（%v）：%s", statusCode, apiRsp.StatusMessage)
		return
	}
	fmt.Println(apiRsp.Data)
}
