package switchover

import (
	"context"

	"github.com/pingcap-inc/tiem/message/cluster"
)

type CDCManagerAPI interface {
	Create(ctx context.Context, request cluster.CreateChangeFeedTaskReq) (resp cluster.CreateChangeFeedTaskResp, err error)
	Delete(ctx context.Context, request cluster.DeleteChangeFeedTaskReq) (resp cluster.DeleteChangeFeedTaskResp, err error)
	Pause(ctx context.Context, request cluster.PauseChangeFeedTaskReq) (resp cluster.PauseChangeFeedTaskResp, err error)
	Resume(ctx context.Context, request cluster.ResumeChangeFeedTaskReq) (resp cluster.ResumeChangeFeedTaskResp, err error)
	Detail(ctx context.Context, request cluster.DetailChangeFeedTaskReq) (resp cluster.DetailChangeFeedTaskResp, err error)
	Query(ctx context.Context, request cluster.QueryChangeFeedTaskReq) (resps []cluster.QueryChangeFeedTaskResp, total int, err error)
}
