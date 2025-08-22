package email

import (
	"context"

	"apitools/api/internal/svc"
	"apitools/api/internal/types"
	"apitools/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
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
	// 转换API请求到RPC请求
	rpcReq := &pb.SendEmailReq{
		To:          req.To,
		Cc:          req.Cc,
		Bcc:         req.Bcc,
		Subject:     req.Subject,
		Content:     req.Content,
		ContentType: req.ContentType,
		From:        req.From,
		FromName:    req.FromName,
		ReplyTo:     req.ReplyTo,
		Priority:    int32(req.Priority),
	}

	// 转换附件
	if len(req.Attachments) > 0 {
		rpcReq.Attachments = make([]*pb.EmailAttachment, len(req.Attachments))
		for i, attachment := range req.Attachments {
			rpcReq.Attachments[i] = &pb.EmailAttachment{
				FileName:    attachment.FileName,
				Content:     attachment.Content,
				ContentType: attachment.ContentType,
				Size:        attachment.Size,
			}
		}
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.EmailRpc.SendEmail(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用邮件RPC服务失败: %v", err)
		return &types.SendEmailResp{
			Code:    500,
			Message: "邮件服务不可用",
		}, nil
	}

	// 转换RPC响应到API响应
	resp = &types.SendEmailResp{
		Code:     rpcResp.Code,
		Message:  rpcResp.Message,
		EmailId:  rpcResp.EmailId,
		SendTime: rpcResp.SendTime,
	}

	return resp, nil
}
