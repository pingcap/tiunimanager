package service

import (
	"context"
	manager "github.com/pingcap/ticp/micro-manager/proto"
)

var TiCPManagerServiceName = "go.micro.ticp.manager"

var SuccessResponseStatus = &manager.ResponseStatus {Code:0}

type ManagerServiceHandler struct {}

func (*ManagerServiceHandler) Login(ctx context.Context, req *manager.LoginRequest, resp *manager.LoginResponse) error {
	resp.Status = SuccessResponseStatus

	resp.TokenString = "mocktoken"
	return nil
}
func (*ManagerServiceHandler) Logout(ctx context.Context, req *manager.LogoutRequest, resp *manager.LogoutResponse) error {
	resp.Status = SuccessResponseStatus
	resp.AccountName = "peijin"
	return nil
}
func (*ManagerServiceHandler) VerifyIdentity(ctx context.Context, req *manager.VerifyIdentityRequest, resp *manager.VerifyIdentityResponse) error {
	resp.Status = SuccessResponseStatus
	resp.AccountName = "peijin"
	resp.TenantId = 1
	return nil
}