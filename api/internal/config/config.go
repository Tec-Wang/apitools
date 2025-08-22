package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	EmailRpc zrpc.RpcClientConf
	// GitLab 配置
	GitLab GitLabConfig
}

type GitLabConfig struct {
	// 默认 GitLab 服务器地址
	DefaultUrl string
	// 默认访问令牌
	DefaultAccessToken string
	// 请求超时时间（秒）
	TimeoutSeconds int `json:",default=30"`
}
