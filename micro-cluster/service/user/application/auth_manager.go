
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
 *                                                                            *
 ******************************************************************************/

package application

import (
	"context"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/commons"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/ports"
	"time"
)

type AuthManager struct {
	userManager  *UserManager
	tokenHandler ports.TokenHandler
}

func NewAuthManager(userManager  *UserManager, 	tokenHandler ports.TokenHandler) *AuthManager {
	return &AuthManager{userManager : userManager, tokenHandler: tokenHandler}
}

// Login
func (p *AuthManager) Login(ctx context.Context, userName, password string) (tokenString string, err error) {
	account, err := p.userManager.FindAccountByName(ctx, userName)

	if err != nil {
		return
	}

	loginSuccess, err := account.CheckPassword(password)
	if err != nil {
		return
	}

	if !loginSuccess {
		err = &domain.UnauthorizedError{}
		return
	}

	token, err := p.CreateToken(ctx, account.Id, account.Name, account.TenantId)

	if err != nil {
		return
	} else {
		tokenString = token.TokenString
	}

	return
}

// Logout
func (p *AuthManager) Logout(ctx context.Context, tokenString string) (string, error) {
	token, err := p.tokenHandler.GetToken(ctx, tokenString)

	if err != nil {
		return "", &domain.UnauthorizedError{}
	} else if !token.IsValid() {
		return "", nil
	} else {
		accountName := token.AccountName
		token.Destroy()

		err := p.tokenHandler.Modify(ctx, &token)
		if err != nil {
			return "", err
		}

		return accountName, nil
	}
}

var SkipAuth = true

// Accessible
func (p *AuthManager) Accessible(ctx context.Context, pathType string, path string, tokenString string) (tenantId string, accountId, accountName string, err error) {
	if path == "" {
		err = framework.CustomizeMessageError(common.TIEM_PARAMETER_INVALID, "path empty")
		return
	}

	token, err := p.tokenHandler.GetToken(ctx, tokenString)

	if err != nil {
		return
	}

	accountId = token.AccountId
	accountName = token.AccountName
	tenantId = token.TenantId

	if SkipAuth {
		// todo checkAuth switch
		return
	}

	if !token.IsValid() {
		err = &domain.UnauthorizedError{}
		return
	}

	account, err := p.userManager.findAccountAggregation(ctx, accountName)
	if err != nil {
		return
	}

	permission, err := p.userManager.findPermissionAggregationByCode(ctx, tenantId, path)
	if err != nil {
		return
	}

	ok, err := p.checkAuth(account, permission)

	if err != nil {
		return
	}

	if !ok {
		err = &domain.ForbiddenError{}
	}

	return
}

// checkAuth
func (p *AuthManager) checkAuth(account *domain.AccountAggregation, permission *domain.PermissionAggregation) (bool, error) {

	accountRoles := account.Roles

	if len(accountRoles) == 0 {
		return false, nil
	}

	accountRoleMap := make(map[string]bool)

	for _, r := range accountRoles {
		accountRoleMap[r.Id] = true
	}

	allowedRoles := permission.Roles

	if len(allowedRoles) == 0 {
		return false, nil
	}

	for _, r := range allowedRoles {
		if _, exist := accountRoleMap[r.Id]; exist {
			return true, nil
		}
	}

	return false, nil
}

func (p *AuthManager) CreateToken(ctx context.Context, accountId string, accountName string, tenantId string) (domain.TiEMToken, error) {
	token := domain.TiEMToken{
		AccountName: accountName,
		AccountId: accountId,
		TenantId: tenantId,
		ExpirationTime: time.Now().Add(commons.DefaultTokenValidPeriod),
	}

	tokenString, err := p.tokenHandler.Provide(ctx, &token)
	token.TokenString = tokenString
	return token, err
}

func (p *AuthManager) SetTokenHandler(tokenHandler ports.TokenHandler) {
	p.tokenHandler = tokenHandler
}

func (p *AuthManager) SetUserManager(userManager *UserManager) {
	p.userManager = userManager
}
