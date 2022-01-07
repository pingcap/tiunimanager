/*
 * Copyright (c)  2022 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*******************************************************************************
 * @File: manager.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/01/06
*******************************************************************************/

package userinfo

import (
	"context"
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/user/account"
	"github.com/pingcap-inc/tiem/models/user/tenant"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

type Manager struct{}

func NewAccountManager() *Manager {
	return &Manager{}
}

func (p *Manager) genSaltAndHash(passwd string) (salt, finalHash string, er error) {
	if passwd == "" {
		return "", "", errors.NewError(errors.TIEM_PARAMETER_INVALID, "password cannot be empty")
	}

	b := make([]byte, 16)
	_, err := cryrand.Read(b)
	if err != nil {
		return "", "", err
	}

	var hash []byte
	salt = base64.URLEncoding.EncodeToString(b)
	hash, err = bcrypt.GenerateFromPassword([]byte(salt+passwd), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return salt, string(hash), nil
}

func (p *Manager) CreateAccount(ctx context.Context, tenant *tenant.Tenant, name, passwd string) (*account.Account, error) {
	if tenant == nil || !tenant.Status.IsValid() {
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := p.FindAccountByName(ctx, name)

	if e == nil && existed != nil {
		return existed, fmt.Errorf("account already exist")
	}
	account := account.Account{Name: name, Entity: common.Entity{Status: strconv.Itoa(int(constants.Valid))}}
	account.GenSaltAndHash(passwd)
	a, _ := models.GetAccountReaderWriter().AddAccount(ctx, tenant.ID, name, account.Salt, account.FinalHash, int8(constants.Valid))
	return a, nil
}

func (p *Manager) FindAccountByName(ctx context.Context, name string) (*account.Account, error) {
	a, err := models.GetAccountReaderWriter().FindAccountByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return a, err
}

func (p *Manager) CreateUser(ctx context.Context, request message.CreateUserReq) (resp message.CreateUserResp, er error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()

	user, err := rw.GetUser(ctx, request.TenantID, request.ID)
	if err == nil && user.ID == request.ID {
		return resp, errors.NewEMErrorf(errors.UserAlreadyExist, "tenantID: %s, userID: %s", request.TenantID, request.ID)
	}
	if err == nil {
		var slat, hash string
		slat, hash, err = p.genSaltAndHash(request.Password)
		if err != nil {
			log.Warningf("craete user error %v,tenantID: %s, userID: %s", err, request.TenantID, request.ID)
			return resp, errors.NewEMErrorf(errors.UserGenSaltAndHashValueFailed, "tenantID: %s, userID: %s", request.TenantID, request.ID)
		}

		_, err = rw.CreateUser(ctx, account.User{ID: request.ID, Name: request.Name, Status: string(constants.UserStatusNormal),
			TenantID: request.TenantID, Email: request.Email, Phone: request.Phone, /* TODO Creator: request.Creator,*/
			Salt: slat, FinalHash: hash,
		})
		if nil != err {
			log.Warningf("craete user %s error %v", request.ID, err)
		}
	}
	return resp, err
}

func (p *Manager) DeleteUser(ctx context.Context, request message.DeleteUserReq) (resp message.DeleteUserResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	_, er := rw.GetUser(ctx, request.TenantID, request.ID)
	if er != nil {
		log.Warningf("delete user error: %v,tenantID: %s, userID: %s", er, request.TenantID, request.ID)
		return resp, errors.NewEMErrorf(errors.UserNotExist, "tenantID: %s, userID:%s", request.TenantID, request.ID)
	}

	er = rw.DeleteUser(ctx, request.TenantID, request.ID)
	if er != nil {
		log.Warningf("delete user error: %v,tenantID: %s,userID: %s", er, request.TenantID, request.ID)
		return resp, errors.NewEMErrorf(errors.DeleteUserFailed, "tenantID: %s,userID:%s", request.TenantID, request.ID)
	}
	return
}

func (p *Manager) GetUser(ctx context.Context, request message.GetUserReq) (resp message.GetUserResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.Info, err = rw.GetUser(ctx, request.TenantID, request.ID)
	if err != nil {
		log.Warningf("get user error: %v,tenantID: %s, userID: %s", err, request.TenantID, request.ID)
		return resp, errors.NewEMErrorf(errors.UserNotExist, "tenantID: %s,userID: %s", request.TenantID, request.ID)
	}
	return resp, err
}

func (p *Manager) QueryUsers(ctx context.Context, request message.QueryUserReq) (resp message.QueryUserResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.UserInfo, err = rw.QueryUsers(ctx)
	if err != nil {
		log.Warningf("query all user error: %v", err)
		return resp, errors.NewEMErrorf(errors.QueryUserScanRowError, "error: %v", err)
	}
	return resp, err
}

func (p *Manager) UpdateUserProfile(ctx context.Context, request message.UpdateUserProfileReq) (resp message.UpdateUserProfileResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	err = rw.UpdateUserProfile(ctx, request.ID, request.Name, request.Email, request.Phone)
	if err != nil {
		log.Warningf("update user profile error: %v,tenantID: %s, userID: %s,name:%s, email:%s, phone:%s", err, request.TenantID, request.ID, request.Name, request.Email, request.Phone)
		return resp, errors.NewEMErrorf(errors.UpdateUserProfileFailed, "error: %v,tenantID: %s, userID: %s,name:%s, email:%s, phone:%s", err, request.TenantID, request.ID, request.Name, request.Email, request.Phone)
	}
	return resp, err
}
