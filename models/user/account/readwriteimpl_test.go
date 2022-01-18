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

package account

import (
	ctx "context"
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestAccountReadWrite_CreateUser(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		user := &User{Name: "nick"}
		_, _, _, err := testRW.CreateUser(ctx.TODO(), user, "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick01",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got1, got2, got3, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		assert.Equal(t, got1.Name, user.Name)
		assert.Equal(t, got1.ID, got2.UserID)
		assert.Equal(t, got2.LoginName, "user")
		assert.Equal(t, got3.TenantID, "tenant")
		assert.Equal(t, got1.ID, got3.UserID)
		err = testRW.DeleteUser(ctx.TODO(), got1.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_DeleteUser(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.DeleteUser(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick02",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_GetUser(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		_, err := testRW.GetUser(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := testRW.GetUser(ctx.TODO(), "user01")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick03",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		info, err := testRW.GetUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
		assert.Equal(t, got.ID, info.ID)
		assert.Equal(t, len(info.Name), 1)
		assert.Equal(t, info.Name[0], "user")
		assert.Equal(t, len(info.TenantID), 1)
		assert.Equal(t, info.TenantID[0], "tenant")
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_GetUserByName(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		_, err := testRW.GetUserByName(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick13",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		info, err := testRW.GetUserByName(ctx.TODO(), "user")
		assert.NoError(t, err)
		assert.Equal(t, info.ID, got.ID)
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_QueryUsers(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick04",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		infos, err := testRW.QueryUsers(ctx.TODO())
		assert.Equal(t, 1, len(infos))
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_UpdateUserStatus(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateUserStatus(ctx.TODO(), "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick05",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		err = testRW.UpdateUserStatus(ctx.TODO(), got.ID, "Deactivate")
		assert.NoError(t, err)
		info, err := testRW.GetUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
		assert.Equal(t, info.Status, "Deactivate")
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_UpdateUserProfile(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateUserProfile(ctx.TODO(), "", "", "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick06",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		err = testRW.UpdateUserProfile(ctx.TODO(), got.ID, "user06", "email01", "phone01")
		assert.NoError(t, err)
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_UpdateUserPassword(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateUserPassword(ctx.TODO(), "", "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		user := &User{
			Creator:   "admin",
			Name:      "nick07",
			Salt:      "salt",
			FinalHash: "hash",
			Email:     "email",
			Phone:     "123",
			Status:    "Normal",
		}
		got, _, _, err := testRW.CreateUser(ctx.TODO(), user, "user", "tenant")
		assert.NoError(t, err)
		err = testRW.UpdateUserPassword(ctx.TODO(), got.ID, "salt01", "hash01")
		assert.NoError(t, err)
		err = testRW.DeleteUser(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_CreateTenant(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		_, err := testRW.CreateTenant(ctx.TODO(), &Tenant{ID: "", Name: ""})
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant01",
			Creator:          "user",
			Name:             "tenant01",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		got, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)
		assert.Equal(t, got.ID, tenant.ID)
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_DeleteTenant(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.DeleteTenant(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant02",
			Creator:          "user",
			Name:             "tenant02",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		got, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)
		assert.Equal(t, got.ID, tenant.ID)
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_GetTenant(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		_, err := testRW.GetTenant(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant03",
			Creator:          "user",
			Name:             "tenant03",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		_, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)

		got, err := testRW.GetTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
		assert.Equal(t, got.ID, tenant.ID)
		assert.Equal(t, got.Name, tenant.Name)
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_QueryTenants(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant04",
			Creator:          "user",
			Name:             "tenant04",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		_, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)
		got, err := testRW.QueryTenants(ctx.TODO())
		assert.Equal(t, 2, len(got))
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_UpdateTenantStatus(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateTenantStatus(ctx.TODO(), "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant05",
			Creator:          "user",
			Name:             "tenant05",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		_, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)
		err = testRW.UpdateTenantStatus(ctx.TODO(), tenant.ID, "Deactivate")
		assert.NoError(t, err)
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_UpdateTenantProfile(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateTenantProfile(ctx.TODO(), "", "", 0, 0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant06",
			Creator:          "user",
			Name:             "tenant06",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		_, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)
		err = testRW.UpdateTenantProfile(ctx.TODO(), tenant.ID, "tenant", 1, 1, 1, 1)
		assert.NoError(t, err)
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}

func TestAccountReadWrite_UpdateTenantOnBoardingStatus(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateTenantOnBoardingStatus(ctx.TODO(), "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		tenant := &Tenant{
			ID:               "tenant07",
			Creator:          "user",
			Name:             "tenant07",
			Status:           "Normal",
			OnBoardingStatus: "On",
		}
		_, err := testRW.CreateTenant(ctx.TODO(), tenant)
		assert.NoError(t, err)
		err = testRW.UpdateTenantOnBoardingStatus(ctx.TODO(), tenant.ID, "Off")
		assert.NoError(t, err)
		err = testRW.DeleteTenant(ctx.TODO(), tenant.ID)
		assert.NoError(t, err)
	})
}
