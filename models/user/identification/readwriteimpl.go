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

package identification

import (
	"context"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"github.com/pingcap/errors"
	"gorm.io/gorm"
	"time"
)

type TokenReadWrite struct {
	dbCommon.GormDB
}

func (g *TokenReadWrite) CreateToken(ctx context.Context, tokenString, userID, tenantID string, expirationTime time.Time) (*Token, error) {
	if "" == tokenString || "" == userID || "" == tenantID {
		return nil, errors.Errorf("AddToken has invalid parameter, tokenString: %s, userID: %s, tenantID: %s", tokenString, userID, tenantID)
	}
	token, err := g.GetToken(ctx, tokenString)
	if err == nil && token.TokenString == tokenString {
		token.ExpirationTime = expirationTime
		err = g.DB(ctx).Save(&token).Error
	} else {
		token.TokenString = tokenString
		token.UserID = userID
		token.TenantID = tenantID
		token.ExpirationTime = expirationTime
		token.Status = 0
		err = g.DB(ctx).Create(&token).Error
	}
	return token, err
}

func (g *TokenReadWrite) GetToken(ctx context.Context, tokenString string) (*Token, error) {
	token := &Token{}
	return token, g.DB(ctx).Where("token_string = ?", tokenString).First(token).Error
}

func NewTokenReadWrite(db *gorm.DB) *TokenReadWrite {
	return &TokenReadWrite{
		dbCommon.WrapDB(db),
	}
}
