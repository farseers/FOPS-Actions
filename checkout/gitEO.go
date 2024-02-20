package main

import (
	"github.com/farseer-go/fs/exception"
	"github.com/farseer-go/utils/str"
	"net/url"
	"path/filepath"
	"strings"
)

// GitEO git仓库
type GitEO struct {
	Hub      string // git地址
	Branch   string // Git分支
	UserName string // 账户名称
	UserPwd  string // 账户密码
	Path     string // 存储目录
}

// GetAbsolutePath 获取git存储的绝对路径 如："/var/lib/fops/git/fops/"
func (receiver *GitEO) GetAbsolutePath() string {
	return GitRoot + receiver.GetRelativePath()
}

// GetRelativePath 获取git存储的相对路径 如："fops/"
func (receiver *GitEO) GetRelativePath() string {
	if receiver.Path == "" || receiver.Path == "/" {
		receiver.Path = receiver.GetName()
	}
	// 移除前后/
	receiver.Path = strings.TrimPrefix(receiver.Path, "/")
	receiver.Path = strings.TrimSuffix(receiver.Path, "/")
	return receiver.Path + "/"
}

// GetName 获取仓库名称 如："fops"
func (receiver *GitEO) GetName() string {
	git := filepath.Base(receiver.Hub)
	return str.CutRight(git, ".git")
}

// GetAuthHub 获取带账号密码的地址 如："https://steden:123456@github.com/farseer-go/fs.git"
func (receiver *GitEO) GetAuthHub() string {
	parsedURL, err := url.Parse(receiver.Hub)
	exception.ThrowRefuseExceptionfBool(err != nil, "解析 URL 失败:%s", err)

	// 设置用户名和密码
	parsedURL.User = url.UserPassword(receiver.UserName, receiver.UserPwd)

	return parsedURL.String()
}

// GetRawContent 获取github仓库中的内容
func (receiver *GitEO) GetRawContent(filePath string) string {
	// 如："https://steden:123456@github.com/farseers/FOPS.git"
	gitUrl := receiver.GetAuthHub()
	if strings.Contains(gitUrl, "github.com") {
		// 移除.git后缀 https://raw.githubusercontent.com/farseers/FOPS/main/.fops/workflows/build.yml
		if strings.HasSuffix(gitUrl, ".git") {
			gitUrl = gitUrl[:len(gitUrl)-4]
		}
		gitUrl = gitUrl + "/" + receiver.Branch + "/" + filePath
		gitUrl = strings.ReplaceAll(gitUrl, "github.com", "raw.githubusercontent.com")
	}
	return gitUrl
}
