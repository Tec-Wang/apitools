package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	// 邮件配置
	Email EmailConfig
	// GitLab 配置
	GitLab GitLabConfig
}

// 邮件配置
type EmailConfig struct {
	Host     string `json:",optional"`
	Port     int    `json:",default=587"`
	Username string `json:",optional"`
	Password string `json:",optional"`
	From     string `json:",optional"`
	FromName string `json:",optional"`
}

type GitLabConfig struct {
	// 默认 GitLab 服务器地址
	DefaultUrl string
	// 默认访问令牌
	DefaultAccessToken string
	// 请求超时时间（秒）
	TimeoutSeconds int `json:",default=30"`
}
