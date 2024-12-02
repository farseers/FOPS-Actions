package main

import (
	"bytes"
	"crypto/tls"

	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/snc"
)

func main() {
	go printProgress()

	// 得到添加构建的URL
	if !strings.HasSuffix(With.FopsAddr, "/") {
		With.FopsAddr += "/"
	}
	fopsAddr := With.FopsAddr + "apps/build/add"

	bodyByte, _ := snc.Marshal(map[string]any{"appName": With.AppName, "workflowsName": With.WorkflowsName, "branchName": With.Branch})
	progress <- "创建新的构建：" + fopsAddr + " " + string(bodyByte)

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
			progress <- "创建新的构建失败：" + err.Error()
			time.Sleep(20 * time.Second)
			continue
		}

		apiRsp := core.NewApiResponseByReader[any](rsp.Body)
		if apiRsp.StatusCode != 200 {
			progress <- fmt.Sprintf("创建新的构建失败（%v）：%s", rsp.StatusCode, apiRsp.StatusMessage)
			time.Sleep(20 * time.Second)
			continue
		}

		isSuccess = true
		break
	}

	// 3次还是失败时，读取远程docker日志
	if !isSuccess {
		waitProgress()
		time.Sleep(time.Second)
		os.Exit(-1)
	}

	progress <- "创建新的构建成功"
	// 等待退出
	waitProgress()
}
