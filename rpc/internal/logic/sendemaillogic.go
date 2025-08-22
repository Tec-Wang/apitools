package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"apitools/rpc/internal/svc"
	"apitools/rpc/pb"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/gomail.v2"
)

type SendEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailLogic {
	return &SendEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendEmailLogic) SendEmail(in *pb.SendEmailReq) (*pb.SendEmailResp, error) {
	// 参数验证
	if len(in.To) == 0 {
		return &pb.SendEmailResp{
			Code:    400,
			Message: "收件人不能为空",
		}, nil
	}

	if in.Subject == "" {
		return &pb.SendEmailResp{
			Code:    400,
			Message: "邮件主题不能为空",
		}, nil
	}

	if in.Content == "" {
		return &pb.SendEmailResp{
			Code:    400,
			Message: "邮件内容不能为空",
		}, nil
	}

	// 创建邮件
	m := gomail.NewMessage()

	// 设置发件人
	from := in.From
	fromName := in.FromName
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
	m.SetHeader("To", in.To...)

	// 设置抄送
	if len(in.Cc) > 0 {
		m.SetHeader("Cc", in.Cc...)
	}

	// 设置密送
	if len(in.Bcc) > 0 {
		m.SetHeader("Bcc", in.Bcc...)
	}

	// 设置回复邮箱
	if in.ReplyTo != "" {
		m.SetHeader("Reply-To", in.ReplyTo)
	}

	// 设置主题
	m.SetHeader("Subject", in.Subject)

	// 设置邮件优先级
	if in.Priority > 0 {
		switch in.Priority {
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
	contentType := in.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	if strings.ToLower(contentType) == "text/html" {
		m.SetBody("text/html", in.Content)
	} else {
		m.SetBody("text/plain", in.Content)
	}

	// 处理附件
	for _, attachment := range in.Attachments {
		if attachment.FileName != "" && attachment.Content != "" {
			// 解码Base64内容
			decodedContent, err := base64.StdEncoding.DecodeString(attachment.Content)
			if err != nil {
				l.Errorf("解码附件失败: %v", err)
				return &pb.SendEmailResp{
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
	if err := l.svcCtx.Dialer.DialAndSend(m); err != nil {
		l.Errorf("发送邮件失败: %v", err)
		return &pb.SendEmailResp{
			Code:    500,
			Message: fmt.Sprintf("发送邮件失败: %v", err),
		}, nil
	}

	// 生成邮件ID
	emailId := uuid.New().String()
	sendTime := time.Now().Unix()

	l.Infof("邮件发送成功, ID: %s, 收件人: %v", emailId, in.To)

	return &pb.SendEmailResp{
		Code:     200,
		Message:  "邮件发送成功",
		EmailId:  emailId,
		SendTime: sendTime,
	}, nil
}
