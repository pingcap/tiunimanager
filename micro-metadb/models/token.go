
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

package models

import (
	"context"
	"github.com/pingcap/errors"
	"gorm.io/gorm"
	"time"
)

type Token struct {
	gorm.Model

	TokenString    string    `gorm:"size:255"`
	AccountId      string    `gorm:"size:255"`
	AccountName    string    `gorm:"size:255"`
	TenantId       string    `gorm:"size:255"`
	Status         int8      `gorm:"size:255"`
	ExpirationTime time.Time `gorm:"size:255"`
}

func (m *DAOAccountManager) AddToken(ctx context.Context, tokenString, accountName string, accountId, tenantId string, expirationTime time.Time) (token *Token, err error) {
	if "" == tokenString /*|| "" == accountName || "" == accountId || "" == tenantId */ {
		return nil, errors.Errorf("AddToken has invalid parameter, tokenString: %s, accountName: %s, tenantId: %s, accountId: %s", tokenString, accountName, tenantId, accountId)
	}
	token, err = m.FindToken(ctx, tokenString)
	if err == nil && token.TokenString == tokenString {
		token.ExpirationTime = expirationTime
		err = m.DB(ctx).Save(&token).Error
	} else {
		token.TokenString = tokenString
		token.AccountId = accountId
		token.TenantId = tenantId
		token.ExpirationTime = expirationTime
		token.AccountName = accountName
		token.Status = 0
		err = m.DB(ctx).Create(&token).Error
	}
	return token, err
}

func (m *DAOAccountManager) FindToken(ctx context.Context, tokenString string) (token *Token, err error) {
	token = &Token{}
	return token, m.DB(ctx).Where("token_string = ?", tokenString).First(token).Error
}
