/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 ******************************************************************************/

/*******************************************************************************
 * @File: manager.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/01/06
*******************************************************************************/

package account

import (
	"context"
	"fmt"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/user/account"
)

type Manager struct{}

func NewAccountManager() *Manager {
	return &Manager{}
}

func (p *Manager) CreateUser(ctx context.Context, request message.CreateUserReq) (message.CreateUserResp, error) {
	resp := message.CreateUserResp{}
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	user := &account.User{
		DefaultTenantID: request.TenantID,
		Name:            request.Nickname,
		Status:          string(constants.UserStatusNormal),
		Email:           request.Email,
		Phone:           request.Phone,
		Creator:         framework.GetUserIDFromContext(ctx),
	}
	err := user.GenSaltAndHash(string(request.Password))
	if err != nil {
		log.Errorf("user %s generate salt and hash error: %v", request.Name, err)
		return resp, errors.NewErrorf(errors.UserGenSaltAndHashValueFailed,
			"user %s generate salt and hash error: %v", request.Name, err)
	}

	_, _, _, err = rw.CreateUser(ctx, user, request.Name)
	if err != nil {
		log.Errorf("create user %s error: %v", request.Name, err)
		return resp, err
	}

	return resp, nil
}

func (p *Manager) DeleteUser(ctx context.Context, request message.DeleteUserReq) (message.DeleteUserResp, error) {
	resp := message.DeleteUserResp{}
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()

	_, err := rw.GetUser(ctx, request.ID)
	if err != nil {
		log.Errorf("get user %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.UserNotExist,
			"get user %s error: %v", request.ID, err)
	}

	err = rw.DeleteUser(ctx, request.ID)
	if err != nil {
		log.Errorf("delete user %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.DeleteUserFailed,
			"delete user %s error: %v", request.ID, err)
	}
	return resp, nil
}

func (p *Manager) GetUser(ctx context.Context, request message.GetUserReq) (resp message.GetUserResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.User, err = rw.GetUser(ctx, request.ID)
	if err != nil {
		log.Errorf("get user %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.UserNotExist,
			"get user %s error: %v", request.ID, err)
	}
	return resp, err
}

func (p *Manager) QueryUsers(ctx context.Context, request message.QueryUserReq) (resp message.QueryUserResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.Users, err = rw.QueryUsers(ctx)
	if err != nil {
		log.Errorf("query all users error: %v", err)
		return resp, errors.NewErrorf(errors.QueryUserScanRowError, "query all users error: %v", err)
	}
	return resp, err
}

func (p *Manager) UpdateUserProfile(ctx context.Context, request message.UpdateUserProfileReq) (resp message.UpdateUserProfileResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	err = rw.UpdateUserProfile(ctx, request.ID, request.Nickname, request.Email, request.Phone)
	if err != nil {
		errMsg := fmt.Sprintf("update user %s profile ( nickname: %s email: %s phone: %s) error: %v",
			request.ID, request.Nickname, request.Email, request.Phone, err)
		log.Errorf(errMsg)
		return resp, errors.NewErrorf(errors.UpdateUserProfileFailed, errMsg)
	}
	return resp, err
}

func (p *Manager) UpdateUserPassword(ctx context.Context, request message.UpdateUserPasswordReq) (resp message.UpdateUserPasswordResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()

	user := &account.User{}
	err = user.GenSaltAndHash(string(request.Password))
	if err != nil {
		log.Errorf("user %s generate salt and hash error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.UserGenSaltAndHashValueFailed,
			"user %s generate salt and hash error: %v", request.ID, err)
	}
	err = rw.UpdateUserPassword(ctx, request.ID, user.Salt, user.FinalHash.Val)
	if err != nil {
		errMsg := fmt.Sprintf("update user %s password error: %v", request.ID, err)
		log.Errorf(errMsg)
		return resp, errors.NewErrorf(errors.UpdateUserProfileFailed, errMsg)
	}
	return resp, err
}

func (p *Manager) CreateTenant(ctx context.Context, request message.CreateTenantReq) (message.CreateTenantResp, error) {
	resp := message.CreateTenantResp{}
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()

	tenant, err := rw.GetTenant(ctx, request.ID)
	if err == nil && tenant.ID == request.ID {
		return resp, errors.NewErrorf(errors.TenantAlreadyExist, "tenant %s has exist", request.ID)
	} else if err.(errors.EMError).GetCode() == errors.TenantNotExist {
		_, err = rw.CreateTenant(ctx,
			&account.Tenant{
				ID:               request.ID,
				Name:             request.Name,
				Status:           request.Status,
				OnBoardingStatus: request.OnBoardingStatus,
				MaxCluster:       request.MaxCluster,
				MaxCPU:           request.MaxCPU,
				MaxMemory:        request.MaxMemory,
				MaxStorage:       request.MaxStorage,
				Creator:          framework.GetUserIDFromContext(ctx),
			})
		if nil != err {
			log.Errorf("create tenant %s error: %v", request.ID, err)
			return resp, err
		}
	}

	return resp, err
}

func (p *Manager) DeleteTenant(ctx context.Context, request message.DeleteTenantReq) (resp message.DeleteTenantResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	_, err = rw.GetTenant(ctx, request.ID)
	if err != nil {
		log.Errorf("delete tenant %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.TenantNotExist, "delete tenant %s error: %v", request.ID, err)
	}

	err = rw.DeleteTenant(ctx, request.ID)
	if err != nil {
		log.Errorf("delete tenant %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.DeleteTenantFailed, "delete tenant %s error: %v", request.ID, err)
	}
	return resp, err
}

func (p *Manager) GetTenant(ctx context.Context, request message.GetTenantReq) (resp message.GetTenantResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.Info, err = rw.GetTenant(ctx, request.ID)
	if err != nil {
		log.Errorf("get tenant %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.TenantNotExist, "get tenant %s error: %v", request.ID, err)
	}
	return resp, err
}

func (p *Manager) QueryTenants(ctx context.Context, request message.QueryTenantReq) (resp message.QueryTenantResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.Tenants, err = rw.QueryTenants(ctx)
	if err != nil {
		log.Errorf("query all tenants error: %v", err)
		return resp, errors.NewErrorf(errors.QueryTenantScanRowError, "query all tenants error: %v", err)
	}
	return resp, err
}

func (p *Manager) UpdateTenantOnBoardingStatus(ctx context.Context, request message.UpdateTenantOnBoardingStatusReq) (resp message.UpdateTenantOnBoardingStatusResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	err = rw.UpdateTenantOnBoardingStatus(ctx, request.ID, request.OnBoardingStatus)
	if err != nil {
		log.Errorf("update tenant %s on boarding status error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.UpdateTenantOnBoardingStatusFailed,
			"update tenant %s on boarding status error: %v", request.ID, err)
	}
	return resp, err
}

func (p *Manager) UpdateTenantProfile(ctx context.Context, request message.UpdateTenantProfileReq) (resp message.UpdateTenantProfileResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	err = rw.UpdateTenantProfile(ctx, request.ID, request.Name, request.MaxCluster, request.MaxCPU, request.MaxMemory, request.MaxStorage)
	if err != nil {
		errMsg := fmt.Sprintf("update tenant %s profile (Name: %s, MaxCluster: %d, MaxCPU: %d, MaxMemory: %d, MaxStorage: %d) error: %v",
			request.ID, request.Name, request.MaxCluster, request.MaxCPU, request.MaxMemory, request.MaxStorage, err)
		log.Errorf(errMsg)
		return resp, errors.NewErrorf(errors.UpdateTenantOnBoardingStatusFailed, errMsg)
	}
	return resp, err
}
