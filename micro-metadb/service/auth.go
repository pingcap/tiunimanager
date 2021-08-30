package service

import (
	"context"
	"github.com/pingcap-inc/tiem/library/framework"
	"time"

	"github.com/pingcap/errors"

	"github.com/pingcap-inc/tiem/library/common"
	proto "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

var SuccessResponseStatus = &proto.DbAuthResponseStatus{Code: 0}

func (handler *DBServiceHandler) FindTenant(cxt context.Context, req *proto.DBFindTenantRequest, resp *proto.DBFindTenantResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindTenant has invalid parameter")
	}
	log := framework.GetLogger()
	log.Debugf("FindTenant by name : %s", req.GetName())
	accountManager := handler.Dao().AccountManager()
	tenant, err := accountManager.FindTenantByName(req.GetName())

	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Tenant = &proto.DBTenantDTO{
			Id:     tenant.ID,
			Name:   tenant.Name,
			Type:   int32(tenant.Type),
			Status: int32(tenant.Status),
		}
	} else {
		resp.Status.Code = common.TIEM_TENANT_NOT_FOUND
		resp.Status.Message = err.Error()
		return errors.Errorf("FindTenantByName,query database failed, name: %s, error: %v", req.GetName(), err)
	}
	return err
}

func (handler *DBServiceHandler) FindAccount(cxt context.Context, req *proto.DBFindAccountRequest, resp *proto.DBFindAccountResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindAccount has invalid parameter")
	}
	log := framework.GetLogger()
	accountManager := handler.Dao().AccountManager()
	account, err := accountManager.Find(req.GetName())
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
		resp.Status.Code = common.TIEM_ACCOUNT_NOT_FOUND
		resp.Status.Message = err.Error()
		return errors.Errorf("FindAccount,query database failed, name: %s, error: %v", req.GetName(), err)
	}
	log.Debugf("find account by name %s, error: %v", req.GetName(), err)

	if req.WithRole {
		roles, err := accountManager.FetchAllRolesByAccount(account.TenantId, account.ID)

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
			resp.Status.Code = common.TIEM_ACCOUNT_NOT_FOUND
			resp.Status.Message = err.Error()
			return errors.Errorf("FindAccount,query database failed, name: %s, tenantId: %s, error: %v", req.GetName(), account.TenantId, err)
		}

		log.Debugf("find account by name %s, tenantId : %s, error: %v", req.GetName(), account.TenantId, err)
	}
	log.Debugf("find account by name %s, error: %v", req.GetName(), err)
	return err
}

func (handler *DBServiceHandler) SaveToken(cxt context.Context, req *proto.DBSaveTokenRequest, resp *proto.DBSaveTokenResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("SaveToken has invalid parameter, req: %v, resp: %v", req, resp)
	}
	log := framework.GetLogger()
	accountManager := handler.Dao().AccountManager()
	_, err := accountManager.AddToken(req.Token.TokenString, req.Token.AccountName, req.Token.AccountId, req.Token.TenantId, time.Unix(req.Token.ExpirationTime, 0))

	log.Debugf("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
		req.GetToken(), req.Token.TenantId, req.Token.AccountName, req.Token.AccountId, err)
	if err == nil {
		resp.Status = SuccessResponseStatus
	} else {
		resp.Status.Code = common.TIEM_ADD_TOKEN_FAILED
		resp.Status.Message = err.Error()
		return errors.Errorf("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
			req.GetToken(), req.Token.TenantId, req.Token.AccountName, req.Token.AccountId, err)
	}
	log.Debugf("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
		req.GetToken(), req.Token.TenantId, req.Token.AccountName, req.Token.AccountId, err)
	return err
}

func (handler *DBServiceHandler) FindToken(cxt context.Context, req *proto.DBFindTokenRequest, resp *proto.DBFindTokenResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindToken has invalid parameter, req: %v, resp: %v", req, resp)
	}
	log := framework.GetLogger()
	accountManager := handler.Dao().AccountManager()
	token, err := accountManager.FindToken(req.GetTokenString())

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
		resp.Status.Code = common.TIEM_TOKEN_NOT_FOUND
		resp.Status.Message = err.Error()
		err = errors.Errorf("FindToKen,query database failed, token: %s, error: %v", req.GetTokenString(), err)
	}
	log.Debugf("FindToKen, token: %s, error: %v", req.GetTokenString(), err)
	return err
}

func (handler *DBServiceHandler) FindRolesByPermission(cxt context.Context, req *proto.DBFindRolesByPermissionRequest, resp *proto.DBFindRolesByPermissionResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindRolesByPermission has invalid parameter req: %v, resp: %v", req, resp)
	}
	accountManager := handler.Dao().AccountManager()
	permissionDO, err := accountManager.FetchPermission(req.TenantId, req.Code)

	if nil == err {
		resp.Permission = &proto.DBPermissionDTO{
			TenantId: permissionDO.TenantId,
			Code:     permissionDO.Code,
			Name:     permissionDO.Name,
			Type:     int32(permissionDO.Type),
			Desc:     permissionDO.Desc,
			Status:   int32(permissionDO.Status),
		}
	} else {
		resp.Status.Code = common.TIEM_QUERY_PERMISSION_FAILED
		resp.Status.Message = err.Error()
		return errors.Errorf("FindRolesByPermission query database failed, tenantId: %s, code: %s, error: %v", req.TenantId, req.Code, err)
	}

	roles, err := accountManager.FetchAllRolesByPermission(req.TenantId, permissionDO.ID)
	if nil == err {
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
	} else {
		resp.Status.Code = common.TIEM_QUERY_PERMISSION_FAILED
		resp.Status.Message = err.Error()
		err = errors.Errorf("FindRolesByPermission query database failed, tenantId: %s, code: %s, error: %v", req.TenantId, req.Code, err)
	}
	return err
}
