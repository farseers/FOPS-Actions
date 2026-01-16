package main

import (
	"os"

	"github.com/farseer-go/fs/snc"
	"github.com/farseer-go/utils/file"
)

var (
	FopsRoot         string // Fops根目录
	WithJsonPath     string // with.json文件位置
	KubeRoot         string // kubectlConfig配置
	NpmModulesRoot   string // NpmModules
	DistRoot         string //  编译保存的根目录
	GitRoot          string // GIT根目录
	DockerfilePath   string // Dockerfile文件地址
	DockerIgnorePath string // DockerIgnore文件地址
	ShellRoot        string // 生成Shell脚本的存放路径
	ActionsRoot      string // 执行Actions的缓存目录
	With             WithAvg
)

type WithAvg struct {
	AppName           string `json:"appName"`                 // 应用名称（链路追踪）
	BuildId           int64  `json:"buildId"`                 // 构建主键
	BuildNumber       int    `json:"buildNumber"`             // 构建版本号
	FopsAddr          string `json:"fopsAddr"`                // 集群地址
	FScheduleAddr     string `json:"fScheduleAddr"`           // 调度中心地址
	AppAbsolutePath   string `json:"appAbsolutePath"`         // 应用的git根目录
	DockerImage       string `json:"dockerImage"`             // Docker镜像
	DockerfilePath    string `json:"dockerfilePath"`          // Dockerfile路径
	DockerHub         string `json:"dockerHub"`               // 托管地址
	DockerUserName    string `json:"dockerUserName"`          // 账户名称
	DockerUserPwd     string `json:"dockerUserPwd"`           // 账户密码
	DockerNodeRole    string `json:"dockerNodeRole"`          // 容器节点角色 manager or worker
	DockerNetwork     string `json:"dockerNetwork"`           // Docker网络
	DockerReplicas    int    `json:"dockerReplicas"`          // 副本数量
	AdditionalScripts string `json:"dockerAdditionalScripts"` // 首次创建应用时附加脚本

	GitHub      string `json:"gitHub"`      // git地址
	GitBranch   string `json:"gitBranch"`   // Git分支
	GitUserName string `json:"gitUserName"` // 账户名称
	GitUserPwd  string `json:"gitUserPwd"`  // 账户密码
	GitPath     string `json:"gitPath"`     // 存储目录

	Proxy     string `json:"proxy"`     // Git代理
	ClusterId int    `json:"clusterId"` // 集群ID
}

func init() {
	DistRoot = os.Getenv("distRoot")
	GitRoot = os.Getenv("gitRoot")
	FopsRoot = os.Getenv("fopsRoot")
	NpmModulesRoot = os.Getenv("npmModulesRoot")
	KubeRoot = os.Getenv("kubeRoot")
	WithJsonPath = os.Getenv("withjson")

	DockerfilePath = os.Getenv("dockerfilePath")
	DockerIgnorePath = os.Getenv("dockerIgnorePath")
	ShellRoot = os.Getenv("shellRoot")
	ActionsRoot = os.Getenv("actionsRoot")

	withJsonContent := file.ReadString(WithJsonPath)
	//fmt.Println(withJsonContent)
	_ = snc.Unmarshal([]byte(withJsonContent), &With)
	//fmt.Println(With.DockerImage)
}
