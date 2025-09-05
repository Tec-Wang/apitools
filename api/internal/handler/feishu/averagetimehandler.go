package feishu

import (
	"net/http"

	"apitools/api/internal/logic/feishu"
	"apitools/api/internal/svc"
	"apitools/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AverageTimeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AverageTimeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := feishu.NewAverageTimeLogic(r.Context(), svcCtx)
		resp, err := l.AverageTime(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
