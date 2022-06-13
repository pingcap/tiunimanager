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
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
	"time"
)

type AccountReadWrite struct {
	dbCommon.GormDB
}

func NewAccountReadWrite(db *gorm.DB) *AccountReadWrite {
	return &AccountReadWrite{
		dbCommon.WrapDB(db),
	}
}

func (arw *AccountReadWrite) CreateUser(ctx context.Context, user *User, name string) (*User, *UserLogin, *UserTenantRelation, error) {
	if "" == name {
		framework.LogWithContext(ctx).Errorf("create user %v, parameter invalid", user)
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "create user %v, parameter invalid", user)
	}
	// query tenant
	_, err := arw.GetTenant(ctx, user.DefaultTenantID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("not found tenant %s", user.DefaultTenantID)
		return nil, nil, nil, errors.NewErrorf(errors.TenantNotExist, "query tenant %s not found", user.DefaultTenantID)
	}

	// create user
	err = arw.DB(ctx).Create(user).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// create user login
	userLogin := &UserLogin{
		LoginName: name,
		UserID:    user.ID,
	}
	err = arw.DB(ctx).Create(userLogin).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// create user tenant relation
	relation := &UserTenantRelation{
		UserID:   user.ID,
		TenantID: user.DefaultTenantID,
	}
	err = arw.DB(ctx).Create(relation).Error
	if err != nil {
		return nil, nil, nil, err
	}
	return user, userLogin, relation, nil
}

func (arw *AccountReadWrite) DeleteUser(ctx context.Context, userID string) error {
	if "" == userID {
		framework.LogWithContext(ctx).Errorf("delete user %s, parameter invalid", userID)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "delete user %s, parameter invalid", userID)
	}
	// delete user
	err := arw.DB(ctx).Where("id = ?", userID).Unscoped().Delete(&User{}).Error
	if err != nil {
		return err
	}

	// delete user login
	err = arw.DB(ctx).Where("user_id = ?", userID).Unscoped().Delete(&UserLogin{}).Error
	if err != nil {
		return err
	}

	// delete user tenant relation
	err = arw.DB(ctx).Where("user_id = ?", userID).Unscoped().Delete(&UserTenantRelation{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (arw *AccountReadWrite) GetUser(ctx context.Context, userID string) (structs.UserInfo, error) {
	if "" == userID {
		framework.LogWithContext(ctx).Errorf("get user %s profile, parameter invalid", userID)
		return structs.UserInfo{}, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"get user %s profile, parameter invalid", userID)
	}

	user := &User{}
	err := arw.DB(ctx).First(user, "id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		return structs.UserInfo{}, errors.NewErrorf(errors.UserNotExist, "query user %s profile not found", userID)
	}

	// get user login name
	var loginNames []string
	err = arw.DB(ctx).Table("user_logins").Select("login_name").Where("user_id = ?", userID).Find(&loginNames).Error
	if err != nil {
		return structs.UserInfo{}, err
	}

	// get user's tenant
	var tenants []string
	err = arw.DB(ctx).Table("user_tenant_relations").Select("tenant_id").Where("user_id = ?", userID).Find(&tenants).Error
	if err != nil {
		return structs.UserInfo{}, err
	}

	return structs.UserInfo{
		ID:              user.ID,
		DefaultTenantID: user.DefaultTenantID,
		Creator:         user.Creator,
		Name:            loginNames,
		TenantID:        tenants,
		Nickname:        user.Name,
		Email:           user.Email,
		Phone:           user.Phone,
		Status:          user.Status,
		CreateAt:        user.CreatedAt,
		UpdateAt:        user.UpdatedAt,
	}, nil
}

func (arw *AccountReadWrite) QueryUsers(ctx context.Context) (map[string]structs.UserInfo, error) {
	userInfos := make(map[string]structs.UserInfo)

	var users []string
	err := arw.DB(ctx).Table("users").Select("id").Find(&users).Error
	if err != nil {
		return userInfos, err
	}

	for _, id := range users {
		info, err := arw.GetUser(ctx, id)
		if err != nil {
			return userInfos, err
		}
		_, ok := userInfos[info.ID]
		if !ok {
			userInfos[info.ID] = info
		}
	}

	return userInfos, nil
}

func (arw *AccountReadWrite) UpdateUserStatus(ctx context.Context, userID string, status string) error {
	if "" == userID {
		framework.LogWithContext(ctx).Errorf("update user %s status %s, parameter invalid", userID, status)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"update user %s status %s, parameter invalid", userID, status)
	}
	return arw.DB(ctx).Model(&User{}).Where("id = ?", userID).Update("status", status).Error
}

func (arw *AccountReadWrite) UpdateUserProfile(ctx context.Context, userID, nickname, email, phone string) error {
	if "" == userID {
		framework.LogWithContext(ctx).Errorf(
			"update user %s profile, nickname: %s, email: %s, phone: %s, parameter invalid", userID, nickname, email, phone)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"update user %s profile, nickname: %s, email: %s, phone: %s, parameter invalid", userID, nickname, email, phone)
	}
	return arw.DB(ctx).Model(&User{}).Where("id = ?", userID).Update("email", email).
		Update("phone", phone).Update("name", nickname).Error
}

func (arw *AccountReadWrite) UpdateUserPassword(ctx context.Context, userID, salt, finalHash string) error {
	if "" == userID {
		framework.LogWithContext(ctx).Errorf(
			"update user %s password, salt: %s, finalHash: %s, parameter invalid", userID, salt, finalHash)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"update user %s password, salt: %s, finalHash: %s, parameter invalid", userID, salt, finalHash)
	}

	//value := dbCommon.Password{Val: finalHash, UpdateTime: time.Now()}
	value := dbCommon.PasswordInExpired{
		Val: finalHash,
		UpdateTime: time.Now(),
	}
	return arw.DB(ctx).Model(&User{}).Where("id = ?",
		userID).Update("salt", salt).Update("final_hash", value).Error
}

func (arw *AccountReadWrite) GetUserByName(ctx context.Context, name string) (*User, error) {
	if "" == name {
		framework.LogWithContext(ctx).Errorf("get user by name %s, parameter invalid", name)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"get user by name %s, parameter invalid", name)
	}
	// find user_id by login name
	userLogin := &UserLogin{}
	err := arw.DB(ctx).First(userLogin, "login_name = ?", name).Error
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = arw.DB(ctx).First(user, "id = ?", userLogin.UserID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewErrorf(errors.UserNotExist, "query user %s profile not found", userLogin.UserID)
	}
	return user, nil
}
func (arw *AccountReadWrite) GetUserByID(ctx context.Context, id string) (*User, error) {
	if "" == id {
		framework.LogWithContext(ctx).Errorf("get user by id %s, parameter invalid", id)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"get user by id %s, parameter invalid", id)
	}

	user := &User{}
	err := arw.DB(ctx).First(user, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewErrorf(errors.UserNotExist, "query user %s profile not found", id)
	}
	return user, nil
}

func (arw *AccountReadWrite) CreateTenant(ctx context.Context, tenant *Tenant) (*structs.TenantInfo, error) {
	if "" == tenant.ID || "" == tenant.Name {
		framework.LogWithContext(ctx).Errorf("create tenant %v, parameter invalid", tenant)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "create tenant %v, parameter invalid", tenant)
	}
	return &structs.TenantInfo{ID: tenant.ID, Name: tenant.Name}, arw.DB(ctx).Create(tenant).Error
}

func (arw *AccountReadWrite) DeleteTenant(ctx context.Context, tenantID string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("delete tenant,tenantID: %s, parameter invalid", tenantID)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "delete tenant, tenantID: %s, parameter invalid", tenantID)
	}
	return arw.DB(ctx).Where("id = ?", tenantID).Unscoped().Delete(&Tenant{}).Error
}

func (arw *AccountReadWrite) GetTenant(ctx context.Context, tenantID string) (structs.TenantInfo, error) {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("get tenant,tenantID: %s, parameter invalid", tenantID)
		return structs.TenantInfo{}, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "get tenant: tenantID: %s, parameter invalid", tenantID)
	}
	info := &Tenant{}
	err := arw.DB(ctx).First(info, "id = ?", tenantID).Error
	if err == gorm.ErrRecordNotFound {
		return structs.TenantInfo{}, errors.NewErrorf(errors.TenantNotExist, "query tenant %s profile not found", tenantID)
	}
	if err != nil {
		return structs.TenantInfo{}, errors.NewErrorf(errors.QueryTenantScanRowError, "get tenant profile,tenantID: %s, query data error: %v,", tenantID, err)
	}
	return structs.TenantInfo{ID: info.ID, Name: info.Name, Creator: info.Creator, MaxCPU: info.MaxCPU, MaxMemory: info.MaxMemory, MaxStorage: info.MaxStorage,
		OnBoardingStatus: info.OnBoardingStatus, MaxCluster: info.MaxCluster, Status: info.Status, CreateAt: info.CreatedAt, UpdateAt: info.UpdatedAt}, err
}

func (arw *AccountReadWrite) QueryTenants(ctx context.Context) (map[string]structs.TenantInfo, error) {
	var info structs.TenantInfo
	tenants := make(map[string]structs.TenantInfo)
	SQL := "SELECT id,name,creator,status,on_boarding_status,max_cluster,created_at,updated_at,creator,max_cpu,max_memory,max_storage FROM tenants;"
	rows, err := arw.DB(ctx).Raw(SQL).Rows()
	if err != nil {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "query all user error: %v, SQL: %s", err, SQL)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&info.ID, &info.Name, &info.Creator, &info.Status, &info.OnBoardingStatus, &info.MaxCluster, &info.CreateAt,
			&info.UpdateAt, &info.Creator, &info.MaxCPU, &info.MaxMemory, &info.MaxStorage)
		if err == nil {
			_, ok := tenants[info.ID]
			if !ok {
				tenants[info.ID] = info
			}
		} else {
			return nil, errors.NewErrorf(errors.QueryTenantScanRowError, "query all tenant, scan data error: %v, SQL: %s", err, SQL)
		}
	}
	return tenants, err
}

func (arw *AccountReadWrite) UpdateTenantStatus(ctx context.Context, tenantID, status string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("update tenant status,tenantID: %s, status: %s, parameter invalid", tenantID, status)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "update tenant status: tenantID: %s, status:%s, parameter invalid", tenantID, status)
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Update("status", status).Error
}

func (arw *AccountReadWrite) UpdateTenantProfile(ctx context.Context, tenantID, name string, maxCluster, maxCPU, maxMemory, maxStorage int32) error {
	if tenantID == "" {
		framework.LogWithContext(ctx).Errorf("update tenant profile,tenantID: %s, name: %s, maxCluster: %d, maxCPU: %d, maxMemory: %d, maxStorage: %d,parameter invalid", tenantID, name, maxCluster, maxCPU, maxMemory, maxStorage)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "update tenant profile: tenantID: %s, name: %s, maxCluster: %d, maxCPU: %d, maxMemory: %d, maxStorage: %d, parameter invalid", tenantID, name, maxCluster, maxCPU, maxMemory, maxStorage)
	}

	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).
		Update("name", name).Update("max_cluster", maxCluster).
		Update("max_cpu", maxCPU).Update("max_memory", maxMemory).Update("max_storage", maxStorage).Error
}

func (arw *AccountReadWrite) UpdateTenantOnBoardingStatus(ctx context.Context, tenantID, status string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Errorf("update tenant on boarding status,tenantID: %s, status: %s, parameter invalid", tenantID, status)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "update tenant on boarding status: tenantID: %s, status:%s, parameter invalid", tenantID, status)
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Update("on_boarding_status", status).Error
}
