package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/farseer-go/fs/core"
	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/http"
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
		progress <- fmt.Sprintf("尝试第%d次更新", i+1)
		// 读取配置
		apiRsp, statusCode, err := http.PostJson[core.ApiResponse[any]](fopsAddr, nil, bodyByte, 0)
		if err != nil {
			progress <- "创建新的构建失败：" + err.Error()
			time.Sleep(20 * time.Second)
			continue
		}

		if apiRsp.StatusCode != 200 {
			progress <- fmt.Sprintf("创建新的构建失败（%v）：%s", statusCode, apiRsp.StatusMessage)
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
