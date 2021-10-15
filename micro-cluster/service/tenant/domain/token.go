
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package domain

import (
	commons2 "github.com/pingcap-inc/tiem/micro-cluster/service/tenant/commons"
	"time"
)

type TiEMToken struct {
	TokenString 	string
	AccountName		string
	AccountId		string
	TenantId		string
	TenantName		string
	ExpirationTime  time.Time
}

func (token *TiEMToken) destroy() error {
	token.ExpirationTime = time.Now()
	return TokenMNG.Modify(token)
}

func (token *TiEMToken) renew() error {
	token.ExpirationTime = time.Now().Add(commons2.DefaultTokenValidPeriod)
	return TokenMNG.Modify(token)
}

func (token *TiEMToken) isValid() bool {
	now := time.Now()

	return now.Before(token.ExpirationTime)
}

func createToken(accountId string, accountName string, tenantId string) (TiEMToken, error) {
	token := TiEMToken{
		AccountName: accountName,
		AccountId: accountId,
		TenantId: tenantId,
		ExpirationTime: time.Now().Add(commons2.DefaultTokenValidPeriod),
	}

	tokenString, err := TokenMNG.Provide(&token)
	token.TokenString = tokenString
	return token, err
}