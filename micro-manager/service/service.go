package service

import (
	"context"
	"net/http"

	manager "github.com/pingcap/ticp/micro-manager/proto"
	"github.com/pingcap/ticp/micro-manager/service/host"
	"github.com/pingcap/ticp/micro-manager/service/tenant/domain"
)

var TiCPManagerServiceName = "go.micro.ticp.manager"

var SuccessResponseStatus = &manager.ManagerResponseStatus{
	Code: 0,
}

type ManagerServiceHandler struct{}

func (*ManagerServiceHandler) Login(ctx context.Context, req *manager.LoginRequest, resp *manager.LoginResponse) error {

	token, err := domain.Login(req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = &manager.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = SuccessResponseStatus
		resp.TokenString = token
	}
	return nil

}

func (*ManagerServiceHandler) Logout(ctx context.Context, req *manager.LogoutRequest, resp *manager.LogoutResponse) error {
	accountName, err := domain.Logout(req.TokenString)
	if err != nil {
		resp.Status = &manager.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = SuccessResponseStatus
		resp.AccountName = accountName
	}
	return nil

}

func (*ManagerServiceHandler) VerifyIdentity(ctx context.Context, req *manager.VerifyIdentityRequest, resp *manager.VerifyIdentityResponse) error {
	tenantId, accountName, err := domain.Accessible(req.GetAuthType(), req.GetPath(), req.GetTokenString())

	if err != nil {
		if _, ok := err.(*domain.UnauthorizedError); ok {
			resp.Status = &manager.ManagerResponseStatus{
				Code:    http.StatusUnauthorized,
				Message: "未登录或登录失效，请重试",
			}
		} else if _, ok := err.(*domain.ForbiddenError); ok {
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
		resp.Status = SuccessResponseStatus
		resp.TenantId = int32(tenantId)
		resp.AccountName = accountName
	}

	return nil
}

func (*ManagerServiceHandler) ImportHost(ctx context.Context, in *manager.ImportHostRequest, out *manager.ImportHostResponse) error {
	return host.ImportHost(ctx, in, out)
}

func (*ManagerServiceHandler) ImportHostsInBatch(ctx context.Context, in *manager.ImportHostsInBatchRequest, out *manager.ImportHostsInBatchResponse) error {
	return host.ImportHostsInBatch(ctx, in, out)
}

func (*ManagerServiceHandler) RemoveHost(ctx context.Context, in *manager.RemoveHostRequest, out *manager.RemoveHostResponse) error {
	return host.RemoveHost(ctx, in, out)
}

func (*ManagerServiceHandler) RemoveHostsInBatch(ctx context.Context, in *manager.RemoveHostsInBatchRequest, out *manager.RemoveHostsInBatchResponse) error {
	return host.RemoveHostsInBatch(ctx, in, out)
}

func (*ManagerServiceHandler) ListHost(ctx context.Context, in *manager.ListHostsRequest, out *manager.ListHostsResponse) error {
	return host.ListHost(ctx, in, out)
}

func (*ManagerServiceHandler) CheckDetails(ctx context.Context, in *manager.CheckDetailsRequest, out *manager.CheckDetailsResponse) error {
	return host.CheckDetails(ctx, in, out)
}

func (*ManagerServiceHandler) AllocHosts(ctx context.Context, in *manager.AllocHostsRequest, out *manager.AllocHostResponse) error {
	return host.AllocHosts(ctx, in, out)
}
