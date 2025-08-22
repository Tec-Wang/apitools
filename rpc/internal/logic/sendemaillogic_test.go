package logic

import (
	"context"
	"time"

	"apitools/rpc/internal/svc"
	"apitools/rpc/pb"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendEmailLogicTest struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendEmailLogicTest(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailLogicTest {
	return &SendEmailLogicTest{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendEmailLogicTest) SendEmail(in *pb.SendEmailReq) (*pb.SendEmailResp, error) {
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

	// 模拟邮件发送成功（不实际发送）
	emailId := uuid.New().String()
	sendTime := time.Now().Unix()

	l.Infof("模拟邮件发送成功, ID: %s, 收件人: %v", emailId, in.To)

	return &pb.SendEmailResp{
		Code:     200,
		Message:  "邮件发送成功（测试模式）",
		EmailId:  emailId,
		SendTime: sendTime,
	}, nil
}

