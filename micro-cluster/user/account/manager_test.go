/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package account

import (
	ctx "context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/user/account"
	"github.com/pingcap/tiunimanager/test/mockaccount"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManager_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(&account.User{ID: "user01"}, nil, nil, nil)
		ctx := &gin.Context{}
		ctx.Set(framework.TiUniManager_X_USER_ID_KEY, "admin")
		_, err := manager.CreateUser(framework.NewMicroCtxFromGinCtx(ctx), message.CreateUserReq{
			Name:     "user",
			TenantID: "tenant",
			Email:    "111@pingcap.com",
			Phone:    "1234567",
			Password: "123",
			Nickname: "user01",
		})
		assert.NoError(t, err)
	})

	t.Run("gen password fail", func(t *testing.T) {
		ctx := &gin.Context{}
		ctx.Set(framework.TiUniManager_X_USER_ID_KEY, "admin")
		_, err := manager.CreateUser(framework.NewMicroCtxFromGinCtx(ctx), message.CreateUserReq{
			Name:     "user",
			TenantID: "tenant",
			Email:    "111@pingcap.com",
			Phone:    "1234567",
			Nickname: "user01",
		})
		assert.Error(t, err)
	})

	t.Run("create fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().CreateUser(gomock.Any(), gomock.Any(),
			gomock.Any()).Return(nil, nil, nil, fmt.Errorf("create user error"))
		ctx := &gin.Context{}
		ctx.Set(framework.TiUniManager_X_USER_ID_KEY, "admin")
		_, err := manager.CreateUser(framework.NewMicroCtxFromGinCtx(ctx), message.CreateUserReq{
			Name:     "user",
			TenantID: "tenant",
			Email:    "111@pingcap.com",
			Phone:    "1234567",
			Password: "123",
			Nickname: "user01",
		})
		assert.Error(t, err)
	})
}

func TestManager_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(structs.UserInfo{ID: "user01"}, nil)
		rw.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)

		_, err := manager.DeleteUser(ctx.TODO(), message.DeleteUserReq{
			ID: "user01",
		})
		assert.NoError(t, err)
	})

	t.Run("get fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(structs.UserInfo{}, fmt.Errorf("get fail"))
		_, err := manager.DeleteUser(ctx.TODO(), message.DeleteUserReq{
			ID: "user01",
		})
		assert.Error(t, err)
	})

	t.Run("delete fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(structs.UserInfo{ID: "user01"}, nil)
		rw.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete fail"))

		_, err := manager.DeleteUser(ctx.TODO(), message.DeleteUserReq{
			ID: "user01",
		})
		assert.Error(t, err)
	})
}

func TestManager_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(structs.UserInfo{ID: "user01"}, nil)
		_, err := manager.GetUser(ctx.TODO(), message.GetUserReq{
			ID: "user01",
		})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(structs.UserInfo{}, fmt.Errorf("get fail"))
		_, err := manager.GetUser(ctx.TODO(), message.GetUserReq{
			ID: "user01",
		})
		assert.Error(t, err)
	})
}

func TestManager_QueryUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().QueryUsers(gomock.Any()).Return(map[string]structs.UserInfo{"user01": {ID: "user01"}}, nil)
		_, err := manager.QueryUsers(ctx.TODO(), message.QueryUserReq{})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().QueryUsers(gomock.Any()).Return(nil, fmt.Errorf("query users fail"))
		_, err := manager.QueryUsers(ctx.TODO(), message.QueryUserReq{})
		assert.Error(t, err)
	})
}

func TestManager_UpdateUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateUserProfile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		_, err := manager.UpdateUserProfile(ctx.TODO(), message.UpdateUserProfileReq{
			ID:       "user",
			Email:    "123@pingcap.com",
			Phone:    "1234",
			Nickname: "user01",
		})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateUserProfile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update fail"))
		_, err := manager.UpdateUserProfile(ctx.TODO(), message.UpdateUserProfileReq{
			ID:       "user",
			Email:    "123@pingcap.com",
			Phone:    "1234",
			Nickname: "user01",
		})
		assert.Error(t, err)
	})
}

func TestManager_UpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		_, err := manager.UpdateUserPassword(ctx.TODO(), message.UpdateUserPasswordReq{
			ID:       "user",
			Password: "123",
		})
		assert.NoError(t, err)
	})

	t.Run("gen password fail", func(t *testing.T) {
		_, err := manager.UpdateUserPassword(ctx.TODO(), message.UpdateUserPasswordReq{
			ID: "user",
		})
		assert.Error(t, err)
	})

	t.Run("update password fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update password fail"))
		_, err := manager.UpdateUserPassword(ctx.TODO(), message.UpdateUserPasswordReq{
			ID:       "user",
			Password: "123",
		})
		assert.Error(t, err)
	})
}

func TestManager_CreateTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{},
			errors.NewError(errors.TenantNotExist, "tenant not exist"))
		rw.EXPECT().CreateTenant(gomock.Any(), gomock.Any()).Return(&structs.TenantInfo{}, nil)
		ctx := &gin.Context{}
		ctx.Set(framework.TiUniManager_X_USER_ID_KEY, "admin")
		_, err := manager.CreateTenant(framework.NewMicroCtxFromGinCtx(ctx), message.CreateTenantReq{
			ID:               "tenant01",
			Name:             "tenant",
			Status:           "Normal",
			OnBoardingStatus: "On",
			MaxCluster:       2,
			MaxCPU:           2,
			MaxMemory:        2,
			MaxStorage:       2,
		})
		assert.NoError(t, err)
	})

	t.Run("tenant exist", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{ID: "tenant01"}, nil)
		ctx := &gin.Context{}
		ctx.Set(framework.TiUniManager_X_USER_ID_KEY, "admin")
		_, err := manager.CreateTenant(framework.NewMicroCtxFromGinCtx(ctx), message.CreateTenantReq{
			ID: "tenant01",
		})
		assert.Error(t, err)
	})

	t.Run("create fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{},
			errors.NewError(errors.TenantNotExist, "tenant not exist"))
		rw.EXPECT().CreateTenant(gomock.Any(), gomock.Any()).Return(&structs.TenantInfo{}, fmt.Errorf("create fail"))
		ctx := &gin.Context{}
		ctx.Set(framework.TiUniManager_X_USER_ID_KEY, "admin")
		_, err := manager.CreateTenant(framework.NewMicroCtxFromGinCtx(ctx), message.CreateTenantReq{
			ID:               "tenant01",
			Name:             "tenant",
			Status:           "Normal",
			OnBoardingStatus: "On",
			MaxCluster:       2,
			MaxCPU:           2,
			MaxMemory:        2,
			MaxStorage:       2,
		})
		assert.Error(t, err)
	})
}

func TestManager_DeleteTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{ID: "tenant01"}, nil)
		rw.EXPECT().DeleteTenant(gomock.Any(), gomock.Any()).Return(nil)

		_, err := manager.DeleteTenant(ctx.TODO(), message.DeleteTenantReq{
			ID: "tenant01",
		})
		assert.NoError(t, err)
	})

	t.Run("get fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{}, fmt.Errorf("get fail"))

		_, err := manager.DeleteTenant(ctx.TODO(), message.DeleteTenantReq{
			ID: "tenant01",
		})
		assert.Error(t, err)
	})

	t.Run("delete fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{ID: "tenant01"}, nil)
		rw.EXPECT().DeleteTenant(gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete fail"))

		_, err := manager.DeleteTenant(ctx.TODO(), message.DeleteTenantReq{
			ID: "tenant01",
		})
		assert.Error(t, err)
	})
}

func TestManager_GetTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{ID: "tenant01"}, nil)
		_, err := manager.GetTenant(ctx.TODO(), message.GetTenantReq{
			ID: "tenant01",
		})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().GetTenant(gomock.Any(), gomock.Any()).Return(structs.TenantInfo{}, fmt.Errorf("get fail"))
		_, err := manager.GetTenant(ctx.TODO(), message.GetTenantReq{
			ID: "tenant01",
		})
		assert.Error(t, err)
	})
}

func TestManager_QueryTenants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().QueryTenants(gomock.Any()).Return(map[string]structs.TenantInfo{"tenant01": {ID: "tenant01"}}, nil)
		_, err := manager.QueryTenants(ctx.TODO(), message.QueryTenantReq{})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().QueryTenants(gomock.Any()).Return(nil, fmt.Errorf("query fail"))
		_, err := manager.QueryTenants(ctx.TODO(), message.QueryTenantReq{})
		assert.Error(t, err)
	})
}

func TestManager_UpdateTenantOnBoardingStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateTenantOnBoardingStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		_, err := manager.UpdateTenantOnBoardingStatus(ctx.TODO(),
			message.UpdateTenantOnBoardingStatusReq{ID: "tenant01", OnBoardingStatus: "On"})
		assert.NoError(t, err)
	})

	t.Run("update fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateTenantOnBoardingStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update fail"))
		_, err := manager.UpdateTenantOnBoardingStatus(ctx.TODO(),
			message.UpdateTenantOnBoardingStatusReq{ID: "tenant01", OnBoardingStatus: "On"})
		assert.Error(t, err)
	})
}

func TestManager_UpdateTenantProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateTenantProfile(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		_, err := manager.UpdateTenantProfile(ctx.TODO(),
			message.UpdateTenantProfileReq{ID: "tenant01"})
		assert.NoError(t, err)
	})

	t.Run("update fail", func(t *testing.T) {
		rw := mockaccount.NewMockReaderWriter(ctrl)
		models.SetAccountReaderWriter(rw)

		rw.EXPECT().UpdateTenantProfile(gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update fail"))
		_, err := manager.UpdateTenantProfile(ctx.TODO(),
			message.UpdateTenantProfileReq{ID: "tenant01"})
		assert.Error(t, err)
	})
}
