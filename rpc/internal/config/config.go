package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Email EmailConfig `json:",optional"`
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
