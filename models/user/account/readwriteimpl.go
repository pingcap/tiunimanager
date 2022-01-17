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
	"context"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type AccountReadWrite struct {
	dbCommon.GormDB
}

func NewAccountReadWrite(db *gorm.DB) *AccountReadWrite {
	return &AccountReadWrite{
		dbCommon.WrapDB(db),
	}
}

func (arw *AccountReadWrite) CreateUser(ctx context.Context, user *User) (*structs.UserInfo, error) {
	if "" == user.ID || "" == user.Name {
		framework.LogWithContext(ctx).Errorf("create user %v, parameter invalid", user)
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "create user %v, parameter invalid", user)
	}
	return &structs.UserInfo{ID: user.ID, Name: user.Name, TenantID: user.TenantID}, arw.DB(ctx).Create(user).Error
}

func (arw *AccountReadWrite) DeleteUser(ctx context.Context, tenantID, userID string) error {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Errorf("delete user,tenantID: %s, userID: %s, parameter invalid", tenantID, userID)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "delete user, tenantID: %s, userID: %s, parameter invalid", tenantID, userID)
	}
	return arw.DB(ctx).Where("tenant_id = ? AND id = ?", tenantID, userID).Unscoped().Delete(&User{}).Error
}

func (arw *AccountReadWrite) GetUser(ctx context.Context, tenantID, userID string) (structs.UserInfo, error) {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Errorf("get user profile, tenantID: %s, userID: %s, parameter invalid", tenantID, userID)
		return structs.UserInfo{}, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID,
			"get user profile, tenantID: %s, userID: %s, parameter invalid", tenantID, userID)
	}
	user := &User{}
	err := arw.DB(ctx).First(user, "id = ? AND tenant_id = ?", userID, tenantID).Error
	if err == gorm.ErrRecordNotFound {
		return structs.UserInfo{}, errors.NewEMErrorf(errors.UserNotExist, "query user %s profile not found", userID)
	}
	if err != nil {
		return structs.UserInfo{}, errors.NewEMErrorf(errors.QueryUserScanRowError,
			"query user profile, tenantID: %s, userID:%s, scan data error: %v", tenantID, userID, err)
	}
	return structs.UserInfo{ID: user.ID, Name: user.Name, TenantID: user.TenantID,
		Email: user.Email, Phone: user.Phone, Status: user.Status, CreateAt: user.CreatedAt, UpdateAt: user.UpdatedAt}, err
}

func (arw *AccountReadWrite) GetUserByID(ctx context.Context, userID string) (*User, error) {
	if "" == userID {
		framework.LogWithContext(ctx).Errorf("get user %s, parameter invaild", userID)
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "get user %s, parameter invaild", userID)
	}

	user := &User{}
	err := arw.DB(ctx).First(user, "id = ?", userID).Error
	if err != nil {
		return nil, errors.NewEMErrorf(errors.QueryUserScanRowError, "get user %s error: %v", userID, err)
	}
	return user, nil
}

func (arw *AccountReadWrite) QueryUsers(ctx context.Context) (map[string]structs.UserInfo, error) {
	var info structs.UserInfo
	userInfos := make(map[string]structs.UserInfo)
	SQL := "SELECT id,name,tenant_id,status,email,phone,created_at,updated_at,creator FROM users;"
	rows, err := arw.DB(ctx).Raw(SQL).Rows()
	defer rows.Close()
	if err != nil {
		return nil, errors.NewEMErrorf(errors.TIEM_SQL_ERROR, "query all user error: %v, SQL: %s", err, SQL)
	}
	for rows.Next() {
		err = rows.Scan(&info.ID, &info.Name, &info.TenantID, &info.Status, &info.Email, &info.Phone, &info.CreateAt, &info.UpdateAt, &info.Creator)
		if err == nil {
			_, ok := userInfos[info.ID]
			if !ok {
				userInfos[info.ID] = info
			}
		} else {
			return nil, errors.NewEMErrorf(errors.QueryUserScanRowError, "query all user, scan data error: %v, SQL: %s", err, SQL)
		}
	}
	return userInfos, err
}

func (arw *AccountReadWrite) UpdateUserStatus(ctx context.Context, tenantID, userID string, status string) error {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Errorf("update user status,tenantID: %s, userID: %s, status: %s, parameter invalid", tenantID, userID, status)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID,
			"update user status: tenantID: %s, userID:%s, status:%s, parameter invalid", tenantID, userID, status)
	}
	return arw.DB(ctx).Model(&User{}).Where("tenant_id = ? AND id = ?", tenantID, userID).Update("status", status).Error
}

func (arw *AccountReadWrite) UpdateUserProfile(ctx context.Context, tenantID, userID, email, phone string) error {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Errorf(
			"update user profile,tenantID: %s, userID: %s, email: %s, phone: %s, parameter invalid", tenantID, userID, email, phone)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID,
			"update user profile: tenantID: %s, userID:%s, email: %s, phone: %s, parameter invalid", tenantID, userID, email, phone)
	}
	user := &User{}
	err := arw.DB(ctx).First(user, "id = ? AND tenant_id = ?", userID, tenantID).Error
	if err != nil {
		return err
	}
	return arw.DB(ctx).Model(user).Update("email", email).Update("phone", phone).Error
}

func (arw *AccountReadWrite) UpdateUserPassword(ctx context.Context, tenantID, userID, salt, finalHash string) error {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Errorf(
			"update user password,tenantID: %s, userID: %s, salt: %s, finalHash: %s, parameter invalid", tenantID, userID, salt, finalHash)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID,
			"update user password: tenantID: %s, userID:%s, salt: %s, finalHash: %s, parameter invalid", tenantID, userID, salt, finalHash)
	}
	return arw.DB(ctx).Model(&User{}).Where("tenant_id = ? AND id = ?",
		tenantID, userID).Update("salt", salt).Update("final_hash", finalHash).Error
}

func (arw *AccountReadWrite) CreateTenant(ctx context.Context, tenant *Tenant) (*structs.TenantInfo, error) {
	if "" == tenant.ID || "" == tenant.Name {
		framework.LogWithContext(ctx).Errorf("create tenant %v, parameter invalid", tenant)
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "create tenant %v, parameter invalid", tenant)
	}
	return &structs.TenantInfo{ID: tenant.ID, Name: tenant.Name}, arw.DB(ctx).Create(tenant).Error
}

func (arw *AccountReadWrite) DeleteTenant(ctx context.Context, tenantID string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("delete tenant,tenantID: %s, parameter invalid", tenantID)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "delete tenant, tenantID: %s, parameter invalid", tenantID)
	}
	return arw.DB(ctx).Where("id = ?", tenantID).Unscoped().Delete(&Tenant{}).Error
}

func (arw *AccountReadWrite) GetTenant(ctx context.Context, tenantID string) (structs.TenantInfo, error) {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("get tenant,tenantID: %s, parameter invalid", tenantID)
		return structs.TenantInfo{}, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "get tenant: tenantID: %s, parameter invalid", tenantID)
	}
	info := &Tenant{}
	err := arw.DB(ctx).First(info, "id = ?", tenantID).Error
	if err == gorm.ErrRecordNotFound {
		return structs.TenantInfo{}, errors.NewEMErrorf(errors.TenantNotExist, "query tenant %s profile not found", tenantID)
	}
	if err != nil {
		return structs.TenantInfo{}, errors.NewEMErrorf(errors.QueryTenantScanRowError, "get tenant profile,tenantID: %s, query data error: %v,", tenantID, err)
	}
	return structs.TenantInfo{ID: info.ID, Name: info.Name, Creator: info.Creator, MaxCPU: info.MaxCPU, MaxMemory: info.MaxMemory, MaxStorage: info.MaxStorage,
		OnBoardingStatus: info.OnBoardingStatus, MaxCluster: info.MaxCluster, Status: info.Status, CreateAt: info.CreatedAt, UpdateAt: info.UpdatedAt}, err
}

func (arw *AccountReadWrite) QueryTenants(ctx context.Context) (map[string]structs.TenantInfo, error) {
	var info structs.TenantInfo
	tenants := make(map[string]structs.TenantInfo)
	SQL := "SELECT id,name,creator,status,on_boarding_status,max_cluster,created_at,updated_at,creator,max_cpu,max_memory,max_storage FROM tenants;"
	rows, err := arw.DB(ctx).Raw(SQL).Rows()
	defer rows.Close()
	if err != nil {
		return nil, errors.NewEMErrorf(errors.TIEM_SQL_ERROR, "query all user error: %v, SQL: %s", err, SQL)
	}
	for rows.Next() {
		err = rows.Scan(&info.ID, &info.Name, &info.Creator, &info.Status, &info.OnBoardingStatus, &info.MaxCluster, &info.CreateAt,
			&info.UpdateAt, &info.Creator, &info.MaxCPU, &info.MaxMemory, &info.MaxStorage)
		if err == nil {
			_, ok := tenants[info.ID]
			if !ok {
				tenants[info.ID] = info
			}
		} else {
			return nil, errors.NewEMErrorf(errors.QueryTenantScanRowError, "query all tenant, scan data error: %v, SQL: %s", err, SQL)
		}
	}
	return tenants, err
}

func (arw *AccountReadWrite) UpdateTenantStatus(ctx context.Context, tenantID, status string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("update tenant status,tenantID: %s, status: %s, parameter invalid", tenantID, status)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update tenant status: tenantID: %s, status:%s, parameter invalid", tenantID, status)
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Update("status", status).Error
}

func (arw *AccountReadWrite) UpdateTenantProfile(ctx context.Context, tenantID, name string, maxCluster, maxCPU, maxMemory, maxStorage int32) error {
	if tenantID == "" {
		framework.LogWithContext(ctx).Errorf("update tenant profile,tenantID: %s, name: %s, maxCluster: %d, maxCPU: %d, maxMemory: %d, maxStorage: %d,parameter invalid", tenantID, name, maxCluster, maxCPU, maxMemory, maxStorage)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update tenant profile: tenantID: %s, name: %s, maxCluster: %d, maxCPU: %d, maxMemory: %d, maxStorage: %d, parameter invalid", tenantID, name, maxCluster, maxCPU, maxMemory, maxStorage)
	}

	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).
		Update("name", name).Update("max_cluster", maxCluster).
		Update("max_cpu", maxCPU).Update("max_memory", maxMemory).Update("max_storage", maxStorage).Error
}

func (arw *AccountReadWrite) UpdateTenantOnBoardingStatus(ctx context.Context, tenantID, status string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("update tenant on boarding status,tenantID: %s, status: %s, parameter invalid", tenantID, status)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update tenant on boarding status: tenantID: %s, status:%s, parameter invalid", tenantID, status)
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Update("on_boarding_status", status).Error
}
