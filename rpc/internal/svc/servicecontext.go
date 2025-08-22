package svc

import (
	"apitools/rpc/internal/config"

	"gopkg.in/gomail.v2"
)

type ServiceContext struct {
	Config config.Config
	Dialer *gomail.Dialer
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化邮件发送器
	dialer := gomail.NewDialer(c.Email.Host, c.Email.Port, c.Email.Username, c.Email.Password)

	// QQ邮箱使用SSL连接
	if c.Email.Port == 465 {
		dialer.SSL = true
	}

	return &ServiceContext{
		Config: c,
		Dialer: dialer,
	}
}
