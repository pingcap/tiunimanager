package service

import (
	"context"
	"net/http"

	"github.com/pingcap/tiem/library/firstparty/config"
	host2 "github.com/pingcap/tiem/micro-cluster/service/host"
	domain2 "github.com/pingcap/tiem/micro-cluster/service/tenant/domain"

	manager "github.com/pingcap/tiem/micro-cluster/proto"
)

var TiEMManagerServiceName = "go.micro.tiem.manager"

var ManageSuccessResponseStatus = &manager.ManagerResponseStatus{
	Code: 0,
}

type ManagerServiceHandler struct{}

func (*ManagerServiceHandler) Login(ctx context.Context, req *manager.LoginRequest, resp *manager.LoginResponse) error {

	token, err := domain2.Login(req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = &manager.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TokenString = token
	}
	return nil

}

func (*ManagerServiceHandler) Logout(ctx context.Context, req *manager.LogoutRequest, resp *manager.LogoutResponse) error {
	accountName, err := domain2.Logout(req.TokenString)
	if err != nil {
		resp.Status = &manager.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.AccountName = accountName
	}
	return nil

}

func (*ManagerServiceHandler) VerifyIdentity(ctx context.Context, req *manager.VerifyIdentityRequest, resp *manager.VerifyIdentityResponse) error {
	tenantId, accountId, accountName, err := domain2.Accessible(req.GetAuthType(), req.GetPath(), req.GetTokenString())

	if err != nil {
		if _, ok := err.(*domain2.UnauthorizedError); ok {
			resp.Status = &manager.ManagerResponseStatus{
				Code:    http.StatusUnauthorized,
				Message: "未登录或登录失效，请重试",
			}
		} else if _, ok := err.(*domain2.ForbiddenError); ok {
			resp.Status = &manager.ManagerResponseStatus{
				Code:    http.StatusForbidden,
				Message: "无权限",
			}
		} else {
			resp.Status = &manager.ManagerResponseStatus{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TenantId = tenantId
		resp.AccountId = accountId
		resp.AccountName = accountName
	}

	return nil
}

func InitHostLogger(key config.Key) {
	host2.InitLogger(key)
}

func (*ManagerServiceHandler) ImportHost(ctx context.Context, in *manager.ImportHostRequest, out *manager.ImportHostResponse) error {
	return host2.ImportHost(ctx, in, out)
}

func (*ManagerServiceHandler) ImportHostsInBatch(ctx context.Context, in *manager.ImportHostsInBatchRequest, out *manager.ImportHostsInBatchResponse) error {
	return host2.ImportHostsInBatch(ctx, in, out)
}

func (*ManagerServiceHandler) RemoveHost(ctx context.Context, in *manager.RemoveHostRequest, out *manager.RemoveHostResponse) error {
	return host2.RemoveHost(ctx, in, out)
}

func (*ManagerServiceHandler) RemoveHostsInBatch(ctx context.Context, in *manager.RemoveHostsInBatchRequest, out *manager.RemoveHostsInBatchResponse) error {
	return host2.RemoveHostsInBatch(ctx, in, out)
}

func (*ManagerServiceHandler) ListHost(ctx context.Context, in *manager.ListHostsRequest, out *manager.ListHostsResponse) error {
	return host2.ListHost(ctx, in, out)
}

func (*ManagerServiceHandler) CheckDetails(ctx context.Context, in *manager.CheckDetailsRequest, out *manager.CheckDetailsResponse) error {
	return host2.CheckDetails(ctx, in, out)
}

func (*ManagerServiceHandler) AllocHosts(ctx context.Context, in *manager.AllocHostsRequest, out *manager.AllocHostResponse) error {
	return host2.AllocHosts(ctx, in, out)
}

func (*ManagerServiceHandler) GetFailureDomain(ctx context.Context, in *manager.GetFailureDomainRequest, out *manager.GetFailureDomainResponse) error {
	return host2.GetFailureDomain(ctx, in, out)
}
