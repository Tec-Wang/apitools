package work

import (
	"net/http"

	"apitools/api/internal/logic/work"
	"apitools/api/internal/svc"
	"apitools/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GitCodeCommitRecordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GitCommitRecordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := work.NewGitCodeCommitRecordLogic(r.Context(), svcCtx)
		resp, err := l.GitCodeCommitRecord(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
