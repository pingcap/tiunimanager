package service

import (
	"context"
	"time"

	"github.com/pingcap/errors"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	proto "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

var SuccessResponseStatus = &proto.DbAuthResponseStatus{Code: 0}

func (handler *DBServiceHandler) FindTenant(cxt context.Context, req *proto.DBFindTenantRequest, resp *proto.DBFindTenantResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindTenant has invalid parameter")
	}
	log := handler.Log()
	log.Debug("FindTenant by name : %s", req.GetName())
	pt := handler.Dao().Tables()[models.TABLE_NAME_TENANT].(*models.Tenant)
	tenant, err := pt.FindTenantByName(handler.Dao().Db(),req.GetName())

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
	log := handler.Log()
	accountTbl := handler.Dao().Tables()[models.TABLE_NAME_ACCOUNT].(*models.Account)
	framework.Assert(nil != accountTbl)
	account, err := accountTbl.Find(handler.Dao().Db(),req.GetName())
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
	log.Debug("find account by name %s, error: %v", req.GetName(), err)

	if req.WithRole {
		roleTable := handler.Dao().Tables()[models.TABLE_NAME_ROLE].(*models.Role)
		roles, err := roleTable.FetchAllRolesByAccount(handler.Dao().Db(), account.TenantId, account.ID)

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
			return errors.Errorf("FindAccount,query database failed, name: %s, tenantId: %s, error: %v", req.GetName(), account.TenantId,err)
		}

		log.Debug("find account by name %s, tenantId : %s, error: %v", req.GetName(), account.TenantId,err)
	}
	log.Debug("find account by name %s, error: %v", req.GetName(), err)
	return err
}

func (handler *DBServiceHandler) SaveToken(cxt context.Context, req *proto.DBSaveTokenRequest, resp *proto.DBSaveTokenResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("SaveToken has invalid parameter, req: %v, resp: %v", req, resp)
	}
	log := handler.Log()
	db := handler.Dao().Db()
	tokenTbl := handler.Dao().Tables()[models.TABLE_NAME_TOKEN].(*models.Token)
	framework.Assert(nil != tokenTbl)
	_, err := tokenTbl.AddToken(db,req.Token.TokenString, req.Token.AccountName, req.Token.AccountId, req.Token.TenantId, time.Unix(req.Token.ExpirationTime, 0))

	log.Debug("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
		req.GetToken(), req.Token.TenantId, req.Token.AccountName,req.Token.AccountId,err)
	if err == nil {
		resp.Status = SuccessResponseStatus
	} else {
		resp.Status.Code = common.TIEM_ADD_TOKEN_FAILED
		resp.Status.Message = err.Error()
		return errors.Errorf("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
			req.GetToken(), req.Token.TenantId, req.Token.AccountName,req.Token.AccountId,err)
	}
	log.Debug("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
		req.GetToken(), req.Token.TenantId, req.Token.AccountName,req.Token.AccountId,err)
	return err
}

func (handler *DBServiceHandler) FindToken(cxt context.Context, req *proto.DBFindTokenRequest, resp *proto.DBFindTokenResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindToken has invalid parameter, req: %v, resp: %v", req, resp)
	}
	log := handler.Log()
	db := handler.Dao().Db()
	tokenTbl := handler.Dao().Tables()[models.TABLE_NAME_TOKEN].(*models.Token)
	framework.Assert(nil != tokenTbl)
	token, err := tokenTbl.FindToken(db,req.GetTokenString())

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
	log.Debug("FindToKen, token: %s, error: %v", req.GetTokenString(), err)
	return err
}

func (handler *DBServiceHandler) FindRolesByPermission(cxt context.Context, req *proto.DBFindRolesByPermissionRequest, resp *proto.DBFindRolesByPermissionResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindRolesByPermission has invalid parameter req: %v, resp: %v", req, resp)
	}
	db := handler.Dao().Db()
	permissionTbl := handler.Dao().Tables()[models.TABLE_NAME_PERMISSION].(*models.Permission)
	roleTbl := handler.Dao().Tables()[models.TABLE_NAME_ROLE].(*models.Role)
	permissionDO, err := permissionTbl.FetchPermission(db, req.TenantId, req.Code)

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

	roles, err := roleTbl.FetchAllRolesByPermission(db,req.TenantId, permissionDO.ID)
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