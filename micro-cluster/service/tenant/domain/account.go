
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
	cryrand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Id        string
	TenantId  string
	Name      string
	Salt      string
	FinalHash string
	Status    CommonStatus
}

func (account *Account) genSaltAndHash(passwd string) error {
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

func (account *Account) checkPassword(passwd string) (bool, error) {
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

func (account *Account) persist() error {
	return RbacRepo.AddAccount(account)
}

// CreateAccount 创建账号
func CreateAccount(tenant *Tenant, name, passwd string) (*Account, error) {
	if tenant == nil || !tenant.Status.IsValid() {
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := findAccountByName(name)

	if e == nil && existed != nil {
		return existed, fmt.Errorf("account already exist")
	}

	account := Account{Name: name, Status: Valid}

	account.genSaltAndHash(passwd)
	account.persist()

	return &account, nil
}

// findAccountByName 根据名称获取账号
func findAccountByName(name string) (*Account, error) {
	a, err := RbacRepo.LoadAccountByName(name)
	if err != nil {
		return nil, err
	}

	return &a, err
}

// assignRoles 给账号赋予角色
func (account *Account) assignRoles(roles []Role) error {
	bindings := make([]RoleBinding, len(roles), len(roles))

	for index, r := range roles {
		bindings[index] = RoleBinding{Account: account, Role: &r, Status: Valid}
	}
	return RbacRepo.AddRoleBindings(bindings)
}
