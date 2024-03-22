package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/utils/exec"
	"net/http"
	"os"
	"strings"
)

func main() {
	go printProgress()

	loginDockerHub()

	// 上传完后，删除本地镜像
	defer exec.RunShellCommand("docker rmi "+With.DockerImage, nil, "", true)

	// 上传
	var result = exec.RunShell("docker push "+With.DockerImage, progress, nil, "", true)
	if result == 0 {
		progress <- "镜像上传完成。"
	}

	// 需要更新远程fops的仓库版本
	if With.FopsAddr != "" {
		if !strings.HasSuffix(With.FopsAddr, "/") {
			With.FopsAddr += "/"
		}
		With.FopsAddr += "apps/updateDockerImage"

		avg := map[string]any{"AppName": With.AppName, "dockerImage": With.DockerImage, "buildNumber": With.BuildNumber, "clusterId": With.FopsClusterId}
		bodyByte, _ := json.Marshal(avg)
		progress <- "开始更新远程fops：" + With.FopsAddr + " " + string(bodyByte)

		newRequest, _ := http.NewRequest("POST", With.FopsAddr, bytes.NewReader(bodyByte))
		newRequest.Header.Set("Content-Type", "application/json")

		// 读取配置
		client := &http.Client{}
		rsp, err := client.Do(newRequest)
		if err != nil {
			fmt.Println("更新远程fops的仓库版本失败：" + err.Error())
			os.Exit(-1)
		}

		apiRsp := core.NewApiResponseByReader[any](rsp.Body)
		if apiRsp.StatusCode != 200 {
			fmt.Printf("更新远程fops的仓库版本失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
			os.Exit(-1)
		}
		progress <- "更新成功：" + apiRsp.StatusMessage
	}

	// 等待退出
	waitProgress()

	if result != 0 {
		fmt.Println("镜像上传出错了")
		os.Exit(-1)
	}
}

func loginDockerHub() {
	// 私有仓库，可以无用户名密码。
	if With.DockerUserName != "" && With.DockerUserPwd != "" {
		var result = exec.RunShell("docker login "+With.DockerHub+" -u "+With.DockerUserName+" -p "+With.DockerUserPwd, progress, nil, "", true)
		if result != 0 {
			fmt.Println("镜像仓库登陆失败。")
			os.Exit(-1)
		}
	}

	progress <- "镜像仓库登陆成功。"
}
