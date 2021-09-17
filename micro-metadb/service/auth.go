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

func (handler *DBServiceHandler) FindTenant(ctx context.Context, req *proto.DBFindTenantRequest, resp *proto.DBFindTenantResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindTenant has invalid parameter")
	}
	log := framework.Log()
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
		resp.Status = &proto.DbAuthResponseStatus{
			Code:    common.TIEM_TENANT_NOT_FOUND,
			Message: err.Error(),
		}
		err = errors.Errorf("FindTenantByName,query database failed, name: %s, error: %v", req.GetName(), err)
	}
	if nil == err {
		log.Infof("Find tenant by name :%s successful ,error: %v", req.GetName(), err)
	} else {
		log.Infof("Find tenant by name :%s failed,error: %v", req.GetName(), err)
	}
	return err
}

func (handler *DBServiceHandler) FindAccount(cxt context.Context, req *proto.DBFindAccountRequest, resp *proto.DBFindAccountResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindAccount has invalid parameter")
	}
	log := framework.Log()
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
		resp.Status = &proto.DbAuthResponseStatus{
			Code:    common.TIEM_ACCOUNT_NOT_FOUND,
			Message: err.Error(),
		}

		resp.Status.Message = err.Error()
		err = errors.Errorf("FindAccount,query database failed, name: %s, error: %v", req.GetName(), err)
	}

	if nil != err {
		log.Infof("Find account by name %s successful, withRole: %t, error: %v", req.GetName(), req.GetWithRole(), err)
	} else {
		log.Infof("Find account by name %s failed, withRole: %t, error: %v", req.GetName(), req.GetWithRole(), err)
		return err
	}

	if req.WithRole && nil == err {
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
			resp.Status = &proto.DbAuthResponseStatus{
				Code:    common.TIEM_ACCOUNT_NOT_FOUND,
				Message: err.Error(),
			}
			err = errors.Errorf("FindAccount,query database failed, name: %s, tenantId: %s, error: %v", req.GetName(), account.TenantId, err)
		}
		if nil == err {
			log.Infof("Fetch all roles by account name %s successful, tenantId: %s, error: %v", req.GetName(), account.TenantId, err)
		} else {
			log.Infof("Fetch all roles by account name %s failed, error: %v", req.GetName(), err)
			return err
		}
	}
	return err
}

func (handler *DBServiceHandler) FindAccountById(cxt context.Context, req *proto.DBFindAccountByIdRequest, resp *proto.DBFindAccountByIdResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindAccount has invalid parameter")
	}
	log := framework.Log()
	accountManager := handler.Dao().AccountManager()
	account, err := accountManager.Find(req.GetId())
	if err == nil {
		resp.Status = SuccessResponseStatus
		resp.Account = &proto.DBAccountDTO{
			Id:        account.ID,
			TenantId:  account.TenantId,
			Name:      account.Name,
			Salt:      account.Salt,
			FinalHash: account.FinalHash,
		}
		log.Infof("Find account by id %s successful, error: %v", req.GetId(), err)
	} else {
		resp.Status = &proto.DbAuthResponseStatus{
			Code:    common.TIEM_ACCOUNT_NOT_FOUND,
			Message: err.Error(),
		}

		resp.Status.Message = err.Error()
		err = errors.Errorf("FindAccount,query database failed, id: %s, error: %v", req.GetId(), err)
		log.Errorf("Find account by id %s successful, error: %v", req.GetId(), err)
	}

	return err
}

func (handler *DBServiceHandler) SaveToken(cxt context.Context, req *proto.DBSaveTokenRequest, resp *proto.DBSaveTokenResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("SaveToken has invalid parameter, req: %v, resp: %v", req, resp)
	}
	log := framework.Log()
	accountManager := handler.Dao().AccountManager()
	_, err := accountManager.AddToken(req.Token.TokenString, req.Token.AccountName, req.Token.AccountId, req.Token.TenantId, time.Unix(req.Token.ExpirationTime, 0))

	if err == nil {
		resp.Status = SuccessResponseStatus
	} else {
		resp.Status = &proto.DbAuthResponseStatus{
			Code:    common.TIEM_ADD_TOKEN_FAILED,
			Message: err.Error(),
		}
		err = errors.Errorf("AddToKen,write database failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
			req.GetToken(), req.Token.TenantId, req.Token.AccountName, req.Token.AccountId, err)
	}
	if nil == err {
		log.Infof("AddToken successful, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
			req.GetToken(), req.Token.TenantId, req.Token.AccountName, req.Token.AccountId, err)
	} else {
		log.Infof("AddToKen failed, token: %s, tenantId: %s, accountName: %s, accountId: %s,error: %v",
			req.GetToken(), req.Token.TenantId, req.Token.AccountName, req.Token.AccountId, err)
	}
	return err
}

func (handler *DBServiceHandler) FindToken(cxt context.Context, req *proto.DBFindTokenRequest, resp *proto.DBFindTokenResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindToken has invalid parameter, req: %v, resp: %v", req, resp)
	}
	log := framework.Log()
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
		resp.Status = &proto.DbAuthResponseStatus{
			Code:    common.TIEM_TOKEN_NOT_FOUND,
			Message: err.Error(),
		}
		err = errors.Errorf("FindToKen,query database failed, token: %s, error: %v", req.GetTokenString(), err)
	}
	if nil == err {
		log.Infof("FindToKen successful, token: %s, error: %v", req.GetTokenString(), err)
	} else {
		log.Infof("FindToKen failed, token: %s, error: %v", req.GetTokenString(), err)
	}
	return err
}

func (handler *DBServiceHandler) FindRolesByPermission(cxt context.Context, req *proto.DBFindRolesByPermissionRequest, resp *proto.DBFindRolesByPermissionResponse) error {
	if nil == req || nil == resp {
		return errors.Errorf("FindRolesByPermission has invalid parameter req: %v, resp: %v", req, resp)
	}
	log := framework.Log()
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
		resp.Status = &proto.DbAuthResponseStatus{
			Code:    common.TIEM_QUERY_PERMISSION_FAILED,
			Message: err.Error(),
		}
		err = errors.Errorf("FindRolesByPermission query database failed, tenantId: %s, code: %s, error: %v", req.TenantId, req.Code, err)
	}

	if nil == err {
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
			resp.Status = &proto.DbAuthResponseStatus{
				Code:    common.TIEM_QUERY_PERMISSION_FAILED,
				Message: err.Error(),
			}
			err = errors.Errorf("FindRolesByPermission query database failed, tenantId: %s, code: %s, error: %v", req.TenantId, req.Code, err)
		}
	}

	if nil == err {
		log.Infof("FindRolesByPermission successful, tenantId: %s, code: %s, error: %v", req.GetTenantId(), req.GetCode(), err)
	} else {
		log.Infof("FindRolesByPermission failed, tenantId: %s, code: %s, error: %v", req.GetTenantId(), req.GetCode(), err)
	}
	return err
}
