package service

import (
	"context"
	"time"

	"github.com/pingcap/ticp/micro-metadb/models"
	proto "github.com/pingcap/ticp/micro-metadb/proto"
)

var SuccessResponseStatus = &proto.DbAuthResponseStatus{Code: 0}

func (*DBServiceHandler) FindTenant(cxt context.Context, req *proto.DBFindTenantRequest, resp *proto.DBFindTenantResponse) error {
	tenant, err := models.FindTenantByName(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Tenant = &proto.DBTenantDTO{
			Id:     tenant.ID,
			Name:   tenant.Name,
			Type:   int32(tenant.Type),
			Status: int32(tenant.Status),
		}
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	return nil
}
func (*DBServiceHandler) FindAccount(cxt context.Context, req *proto.DBFindAccountRequest, resp *proto.DBFindAccountResponse) error {
	account, err := models.FindAccount(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus

		resp.Account = &proto.DBAccountDTO{
			Id:        account.ID,
			TenantId:  account.TenantId,
			Name:      account.Name,
			Salt:      account.Salt,
			FinalHash: account.FinalHash,
		}
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	if req.WithRole {
		roles, err := models.FetchAllRolesByAccount(account.TenantId, account.ID)

		if err == nil {
			roleDTOs := make([]*proto.DBRoleDTO, len(roles), cap(roles))
			for index, role := range roles {
				roleDTOs[index] = &proto.DBRoleDTO{
					TenantId: role.TenantId,
					Name:     role.Name,
					Status:   int32(role.Status),
					Desc:     role.Desc,
				}
			}
			resp.Account.Roles = roleDTOs
		} else {
			resp.Status.Code = 1
			resp.Status.Message = err.Error()
		}
	}

	return nil
}
func (*DBServiceHandler) SaveToken(cxt context.Context, req *proto.DBSaveTokenRequest, resp *proto.DBSaveTokenResponse) error {
	_, err := models.AddToken(req.Token.TokenString, req.Token.AccountName, req.Token.AccountId, req.Token.TenantId, time.Unix(req.Token.ExpirationTime, 0))

	if err == nil {
		resp.Status = SuccessResponseStatus
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}
	return nil
}
func (*DBServiceHandler) FindToken(cxt context.Context, req *proto.DBFindTokenRequest, resp *proto.DBFindTokenResponse) error {
	token, err := models.FindToken(req.GetTokenString())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Token = &proto.DBTokenDTO{
			TokenString:    token.TokenString,
			AccountId:      token.AccountId,
			AccountName:    token.AccountName,
			TenantId:       token.TenantId,
			ExpirationTime: token.ExpirationTime.Unix(),
		}
	} else {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
	}

	return nil

}
func (*DBServiceHandler) FindRolesByPermission(cxt context.Context, req *proto.DBFindRolesByPermissionRequest, resp *proto.DBFindRolesByPermissionResponse) error {
	permissionDO, err := models.FetchPermission(req.TenantId, req.Code)

	if err != nil {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
		return nil
	}

	resp.Permission = &proto.DBPermissionDTO{
		TenantId: permissionDO.TenantId,
		Code:     permissionDO.Code,
		Name:     permissionDO.Name,
		Type:     int32(permissionDO.Type),
		Desc:     permissionDO.Desc,
		Status:   int32(permissionDO.Status),
	}

	roles, err := models.FetchAllRolesByPermission(req.TenantId, permissionDO.ID)

	if err != nil {
		resp.Status.Code = 1
		resp.Status.Message = err.Error()
		return nil
	} else {
		roleDTOs := make([]*proto.DBRoleDTO, len(roles), cap(roles))
		for index, role := range roles {
			roleDTOs[index] = &proto.DBRoleDTO{
				TenantId: role.TenantId,
				Name:     role.Name,
				Status:   int32(role.Status),
				Desc:     role.Desc,
			}
		}

		resp.Status = SuccessResponseStatus
		resp.Roles = roleDTOs
	}

	return nil
}
