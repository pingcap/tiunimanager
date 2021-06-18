package service

import (
	"context"
	"github.com/pingcap/ticp/micro-db/models"
	db "github.com/pingcap/ticp/micro-db/proto"
	"time"
)

var TiCPDbServiceName = "go.micro.ticp.db"

type DBHandler struct {}

var SuccessResponseStatus = &db.ResponseStatus {Code:0}

func (*DBHandler) FindTenant(cxt context.Context, req *db.FindTenantRequest, resp *db.FindTenantResponse) error {
	tenant, err := models.FindTenantByName(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Tenant = &db.TenantDTO{
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
func (*DBHandler) FindAccount(cxt context.Context, req *db.FindAccountRequest, resp *db.FindAccountResponse) error {
	account, err := models.FindAccount(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus

		resp.Account = &db.AccountDTO{
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
		roleDTOs := make([]*db.RoleDTO, len(roles), cap(roles))
		for index,role := range roles {
			roleDTOs[index] = &db.RoleDTO{
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
func (*DBHandler) SaveToken(cxt context.Context, req *db.SaveTokenRequest, resp *db.SaveTokenResponse) error {
	_, err := models.AddToken(req.Token.TokenString, req.Token.AccountId, req.Token.TenantId, time.Unix(req.Token.ExpirationTime, 0))

	if err == nil {
		resp.Status = SuccessResponseStatus
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}
	return nil
}
func (*DBHandler) FindToken(cxt context.Context, req *db.FindTokenRequest, resp *db.FindTokenResponse) error {
	token, err := models.FindToken(req.GetTokenString())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Token = &db.TokenDTO{
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
func (*DBHandler) FindRolesByPermission(cxt context.Context, req *db.FindRolesByPermissionRequest, resp *db.FindRolesByPermissionResponse) error {
	roles, err := models.FetchAllRolesByPermission(uint(req.TenantId), req.Code)

	if err == nil {
		roleDTOs := make([]*db.RoleDTO, len(roles), cap(roles))
		for index,role := range roles {
			roleDTOs[index] = &db.RoleDTO{
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
