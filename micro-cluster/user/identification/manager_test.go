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

package identification

import (
	ctx "context"
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/user/account"
	"github.com/pingcap-inc/tiem/models/user/identification"
	"github.com/pingcap-inc/tiem/test/mockaccount"
	"github.com/pingcap-inc/tiem/test/mockidentification"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func genSaltAndHash(passwd string) (string, string, error) {
	b := make([]byte, 16)
	_, err := cryrand.Read(b)

	if err != nil {
		return "", "", err
	}

	salt := base64.URLEncoding.EncodeToString(b)

	finalHash, err := common.FinalHash(salt, passwd)

	if err != nil {
		return "", "", err
	}
	return salt, string(finalHash), nil
}

func TestManager_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		accountRW := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)

		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)
		salt, hash, err := genSaltAndHash("123")
		assert.NoError(t, err)
		accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(&account.User{
			ID:        "user01",
			Salt:      salt,
			FinalHash: hash,
		}, nil)

		tokenRW.EXPECT().CreateToken(gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return(&identification.Token{}, nil)
		got, err := manager.Login(ctx.TODO(), message.LoginReq{Name: "user01", Password: "123"})
		assert.NoError(t, err)
		assert.Equal(t, got.UserID, "user01")
	})

	t.Run("get user fail", func(t *testing.T) {
		accountRW := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(&account.User{
			ID: "user01",
		}, fmt.Errorf("get user fail"))

		_, err := manager.Login(ctx.TODO(), message.LoginReq{Name: "user01", Password: "123"})
		assert.Error(t, err)
	})

	t.Run("check password fail", func(t *testing.T) {
		accountRW := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)
		salt, hash, err := genSaltAndHash("123")
		assert.NoError(t, err)
		accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(&account.User{
			ID:        "user01",
			Salt:      salt,
			FinalHash: hash,
		}, nil)

		_, err = manager.Login(ctx.TODO(), message.LoginReq{Name: "user01", Password: ""})
		assert.Error(t, err)
	})

	t.Run("password wrong", func(t *testing.T) {
		accountRW := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)

		salt, hash, err := genSaltAndHash("123")
		assert.NoError(t, err)
		accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(&account.User{
			ID:        "user01",
			Salt:      salt,
			FinalHash: hash,
		}, nil)

		_, err = manager.Login(ctx.TODO(), message.LoginReq{Name: "user01", Password: "234"})
		assert.Error(t, err)
	})

	t.Run("create token fail", func(t *testing.T) {
		accountRW := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)

		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)
		salt, hash, err := genSaltAndHash("123")
		assert.NoError(t, err)
		accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(&account.User{
			ID:        "user01",
			Salt:      salt,
			FinalHash: hash,
		}, nil)

		tokenRW.EXPECT().CreateToken(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any()).Return(nil, fmt.Errorf("create token fail"))
		_, err = manager.Login(ctx.TODO(), message.LoginReq{Name: "user01", Password: "123"})
		assert.Error(t, err)
	})
}

func TestManager_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)

		tokenRW.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(&identification.Token{
			UserID:         "user01",
			ExpirationTime: time.Now().Add(constants.DefaultTokenValidPeriod)}, nil)

		got, err := manager.Logout(ctx.TODO(), message.LogoutReq{TokenString: "123"})
		assert.NoError(t, err)
		assert.Equal(t, got.UserID, "user01")
	})

	t.Run("get token fail", func(t *testing.T) {
		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)

		tokenRW.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("get fail"))

		_, err := manager.Logout(ctx.TODO(), message.LogoutReq{TokenString: "123"})
		assert.Error(t, err)
	})

	t.Run("token invalid", func(t *testing.T) {
		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)

		tokenRW.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(&identification.Token{
			UserID:         "user01",
			ExpirationTime: time.Now()}, nil)

		got, err := manager.Logout(ctx.TODO(), message.LogoutReq{TokenString: "123"})
		assert.NoError(t, err)
		assert.Equal(t, got.UserID, "")
	})
}

func TestManager_Accessible(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)

		accountRW := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(accountRW)

		tokenRW.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(&identification.Token{
			UserID:         "user01",
			ExpirationTime: time.Now().Add(constants.DefaultTokenValidPeriod)}, nil)

		accountRW.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(
			structs.UserInfo{CurrentTenantID: "tenant"}, nil)

		got, err := manager.Accessible(ctx.TODO(), message.AccessibleReq{TokenString: "123"})
		assert.NoError(t, err)
		assert.Equal(t, got.UserID, "user01")
	})

	t.Run("get token fail", func(t *testing.T) {
		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)

		tokenRW.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("get fail"))

		_, err := manager.Accessible(ctx.TODO(), message.AccessibleReq{TokenString: "123"})
		assert.Error(t, err)
	})

	t.Run("token invalid", func(t *testing.T) {
		tokenRW := mockidentification.NewMockReaderWriter(ctrl)
		models.SetTokenReaderWriter(tokenRW)

		tokenRW.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(&identification.Token{
			UserID:         "user01",
			ExpirationTime: time.Now()}, nil)

		_, err := manager.Accessible(ctx.TODO(), message.AccessibleReq{TokenString: "123"})
		assert.Error(t, err)
	})
}
