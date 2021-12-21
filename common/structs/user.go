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
 ******************************************************************************/

/*******************************************************************************
 * @File: user.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap/errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Account struct {
	ID        string                 `json:"accountId"`
	TenantID  string                 `json:"tenantId"`
	Name      string                 `json:"accountName"`
	Salt      string                 `json:"accountSalt"`
	FinalHash string                 `json:"accountFinalHash"`
	Status    constants.CommonStatus `json:"accountStatus"`
}

type Tenant struct {
	Name   string
	ID     string
	Type   constants.TenantType
	Status constants.CommonStatus
}

type Token struct {
	TokenString 	string
	AccountName		string
	AccountID		string
	TenantID		string
	TenantName		string
	ExpirationTime  time.Time
}

func (account *Account) GenSaltAndHash(passwd string) error {
	b := make([]byte, 16)
	_, err := cryrand.Read(b)

	if err != nil {
		return err
	}

	salt := base64.URLEncoding.EncodeToString(b)

	finalSalt, err := finalHash(salt, passwd)

	if err != nil {
		return err
	}

	account.Salt = salt
	account.FinalHash = string(finalSalt)

	return nil
}

func (account *Account) CheckPassword(passwd string) (bool, error) {
	if passwd == "" {
		return false, errors.New("password cannot be empty")
	}
	if len(passwd) > 20 {
		return false, errors.New("password is too long")
	}
	s := account.Salt + passwd

	err := bcrypt.CompareHashAndPassword([]byte(account.FinalHash), []byte(s))

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func finalHash(salt string, passwd string) ([]byte, error) {
	if passwd == "" {
		return nil, errors.New("password cannot be empty")
	}
	s := salt + passwd
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)

	return finalSalt, err
}

func (token *Token) IsValid() bool {
	now := time.Now()
	return now.Before(token.ExpirationTime)
}

func (token *Token) Destroy() {
	token.ExpirationTime = time.Now()
}

type UnauthorizedError struct{}

func (*UnauthorizedError) Error() string {
	return "Unauthorized"
}

type ForbiddenError struct{}

func (*ForbiddenError) Error() string {
	return "Access Forbidden"
}