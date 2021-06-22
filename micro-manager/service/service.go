package service

import (
	"context"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"github.com/pingcap/ticp/micro-manager/service/tenant/domain"
)

var TiCPManagerServiceName = "go.micro.ticp.manager"

var SuccessResponseStatus = &manager.ManagerResponseStatus {Code:0}
var BizErrorResponseStatus = &manager.ManagerResponseStatus {Code:1}

type ManagerServiceHandler struct {}

func (*ManagerServiceHandler) Login(ctx context.Context, req *manager.LoginRequest, resp *manager.LoginResponse) error {

	token, err := domain.Login(req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = BizErrorResponseStatus
		resp.Status.Message = err.Error()
	} else {
		resp.Status = SuccessResponseStatus
		resp.TokenString = token
	}
	return nil

}

func (*ManagerServiceHandler) Logout(ctx context.Context, req *manager.LogoutRequest, resp *manager.LogoutResponse) error {
	accountName,err := domain.Logout(req.TokenString)
	if err != nil {
		resp.Status = BizErrorResponseStatus
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
		resp.Status = BizErrorResponseStatus
		resp.Status.Message = err.Error()
	} else {
		resp.Status = SuccessResponseStatus
		resp.TenantId = int32(tenantId)
		resp.AccountName = accountName
	}

	return nil
}