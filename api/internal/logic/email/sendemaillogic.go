package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"apitools/api/internal/svc"
	"apitools/api/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/gomail.v2"
)

type SendEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailLogic {
	return &SendEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendEmailLogic) SendEmail(req *types.SendEmailReq) (resp *types.SendEmailResp, err error) {
	// 参数验证
	if len(req.To) == 0 {
		return &types.SendEmailResp{
			Code:    400,
			Message: "收件人不能为空",
		}, nil
	}

	if req.Subject == "" {
		return &types.SendEmailResp{
			Code:    400,
			Message: "邮件主题不能为空",
		}, nil
	}

	if req.Content == "" {
		return &types.SendEmailResp{
			Code:    400,
			Message: "邮件内容不能为空",
		}, nil
	}

	// 检查邮件发送器是否配置
	if l.svcCtx.Config.Email.Host == "" || l.svcCtx.Config.Email.Port == 0 || l.svcCtx.Config.Email.Username == "" || l.svcCtx.Config.Email.Password == "" {
		return &types.SendEmailResp{
			Code:    500,
			Message: "邮件服务未配置",
		}, nil
	}

	// 创建邮件
	m := gomail.NewMessage()

	// 设置发件人
	from := req.From
	fromName := req.FromName
	if from == "" {
		from = l.svcCtx.Config.Email.From
	}
	if fromName == "" {
		fromName = l.svcCtx.Config.Email.FromName
	}

	if fromName != "" {
		m.SetHeader("From", m.FormatAddress(from, fromName))
	} else {
		m.SetHeader("From", from)
	}

	// 设置收件人
	m.SetHeader("To", req.To...)

	// 设置抄送
	if len(req.Cc) > 0 {
		m.SetHeader("Cc", req.Cc...)
	}

	// 设置密送
	if len(req.Bcc) > 0 {
		m.SetHeader("Bcc", req.Bcc...)
	}

	// 设置回复邮箱
	if req.ReplyTo != "" {
		m.SetHeader("Reply-To", req.ReplyTo)
	}

	// 设置主题
	m.SetHeader("Subject", req.Subject)

	// 设置邮件优先级
	if req.Priority > 0 {
		switch req.Priority {
		case 1:
			m.SetHeader("X-Priority", "1")
			m.SetHeader("X-MSMail-Priority", "High")
		case 5:
			m.SetHeader("X-Priority", "5")
			m.SetHeader("X-MSMail-Priority", "Low")
		default:
			m.SetHeader("X-Priority", "3")
			m.SetHeader("X-MSMail-Priority", "Normal")
		}
	}

	// 设置邮件内容
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	if strings.ToLower(contentType) == "text/html" {
		m.SetBody("text/html", req.Content)
	} else {
		m.SetBody("text/plain", req.Content)
	}

	// 处理附件
	for _, attachment := range req.Attachments {
		if attachment.FileName != "" && attachment.Content != "" {
			// 解码Base64内容
			decodedContent, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				l.Errorf("解码附件失败: %v", err)
				return &types.SendEmailResp{
					Code:    400,
					Message: fmt.Sprintf("附件 %s 解码失败: %v", attachment.FileName, err),
				}, nil
			}

			// 添加附件 - 使用字节数组
			m.Attach(attachment.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(decodedContent)
				return err
			}))
		}
	}

	// 发送邮件
	dialer := gomail.NewDialer(l.svcCtx.Config.Email.Host, l.svcCtx.Config.Email.Port, l.svcCtx.Config.Email.Username, l.svcCtx.Config.Email.Password)
	if err := dialer.DialAndSend(m); err != nil {
		l.Errorf("发送邮件失败: %v", err)
		return &types.SendEmailResp{
			Code:    500,
			Message: fmt.Sprintf("发送邮件失败: %v", err),
		}, nil
	}

	// 生成邮件ID
	emailId := uuid.New().String()
	sendTime := time.Now().Unix()

	l.Infof("邮件发送成功, ID: %s, 收件人: %v", emailId, req.To)

	return &types.SendEmailResp{
		Code:     200,
		Message:  "邮件发送成功",
		EmailId:  emailId,
		SendTime: sendTime,
	}, nil
}
