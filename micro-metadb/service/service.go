package service

import (
	"context"
	"github.com/pingcap/ticp/micro-metadb/models"
	proto "github.com/pingcap/ticp/micro-metadb/proto"
	"time"
)

var TiCPMetaDBServiceName = "go.micro.ticp.db"

var SuccessResponseStatus = &proto.ResponseStatus {Code: 0}

type DBServiceHandler struct {}

func (*DBServiceHandler) FindTenant(cxt context.Context, req *proto.FindTenantRequest, resp *proto.FindTenantResponse) error {
	tenant, err := models.FindTenantByName(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Tenant = &proto.TenantDTO{
			Id: int32(tenant.Id),
			Name: tenant.Name,
			Type: int32(tenant.Type),
			Status: int32(tenant.Status),
		}
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	return nil
}
func (*DBServiceHandler) FindAccount(cxt context.Context, req *proto.FindAccountRequest, resp *proto.FindAccountResponse) error {
	account, err := models.FindAccount(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus

		resp.Account = &proto.AccountDTO{
			Id: int32(account.ID),
			TenantId: int32(account.TenantId),
			Name: account.Name,
			Salt: account.Salt,
			FinalHash: account.FinalHash,
		}
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	roles, err := models.FetchAllRolesByAccount(account.TenantId, account.ID)

	if err == nil {
		roleDTOs := make([]*proto.RoleDTO, len(roles), cap(roles))
		for index,role := range roles {
			roleDTOs[index] = &proto.RoleDTO{
				TenantId: int32(role.TenantId),
				Name:     role.Name,
				Status:   int32(role.Status),
				Desc:     role.Desc,
			}
		}
		resp.Account.Roles = roleDTOs
	} else {

	}
	return nil
}
func (*DBServiceHandler) SaveToken(cxt context.Context, req *proto.SaveTokenRequest, resp *proto.SaveTokenResponse) error {
	_, err := models.AddToken(req.Token.TokenString, req.Token.AccountId, req.Token.TenantId, time.Unix(req.Token.ExpirationTime, 0))

	if err == nil {
		resp.Status = SuccessResponseStatus
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}
	return nil
}
func (*DBServiceHandler) FindToken(cxt context.Context, req *proto.FindTokenRequest, resp *proto.FindTokenResponse) error {
	token, err := models.FindToken(req.GetTokenString())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Token = &proto.TokenDTO{
			TokenString: token.TokenString,
			AccountId: token.AccountId,
			TenantId: token.TenantId,
			ExpirationTime: token.ExpirationTime.Unix(),
		}
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	return nil

}
func (*DBServiceHandler) FindRolesByPermission(cxt context.Context, req *proto.FindRolesByPermissionRequest, resp *proto.FindRolesByPermissionResponse) error {
	roles, err := models.FetchAllRolesByPermission(uint(req.TenantId), req.Code)

	if err == nil {
		roleDTOs := make([]*proto.RoleDTO, len(roles), cap(roles))
		for index,role := range roles {
			roleDTOs[index] = &proto.RoleDTO{
				TenantId: int32(role.TenantId),
				Name:     role.Name,
				Status:   int32(role.Status),
				Desc:     role.Desc,
			}
		}

		resp.Status = SuccessResponseStatus
		resp.Roles = roleDTOs
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	return nil
}
