package account

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"strconv"
)

type AccountReadWrite struct {
	dbCommon.GormDB
}

func (arw *AccountReadWrite) AddAccount(ctx context.Context, tenantId string, name string, salt string, finalHash string, status int8) (*Account, error) {
	if "" == tenantId || "" == name || "" == salt || "" == finalHash {
		return nil, fmt.Errorf("add account failed, has invalid parameter, tenantID: %s, name: %s, salt: %s, finalHash: %s, status: %d", tenantId, name, salt, finalHash, status)
	}
	result := &Account{
		Entity:    dbCommon.Entity{TenantId: tenantId, Status: strconv.Itoa(int(status))}, //todo: bug
		Name:      name,
		Salt:      salt,
		FinalHash: finalHash,
	}
	return result, arw.DB(ctx).Create(result).Error
}

func (arw *AccountReadWrite) FindAccountByName(ctx context.Context, name string) (*Account, error) {
	if "" == name {
		return nil, fmt.Errorf("find account failed, has invalid parameter, name: %s", name)
	}
	result := &Account{Name: name}
	return result, arw.DB(ctx).Where(&Account{Name: name}).First(result).Error
}

func (arw *AccountReadWrite) FindAccountById(ctx context.Context, id string) (*Account, error) {
	if "" == id {
		return nil, fmt.Errorf("find account failed, has invalid parameter, id: %s", id)
	}
	result := &Account{Entity: dbCommon.Entity{ID: id}}
	return result, arw.DB(ctx).Where(&Account{Entity: dbCommon.Entity{ID: id}}).First(result).Error
}

func NewAccountReadWrite(db *gorm.DB) *AccountReadWrite {
	return &AccountReadWrite{
		dbCommon.WrapDB(db),
	}
}

func (arw *AccountReadWrite) CreateUser(ctx context.Context, user User) (info *structs.UserInfo, err error) {
	if "" == user.ID || "" == user.Name {
		framework.LogWithContext(ctx).Warningf("create user %v, parameter invalid", user)
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "create user %v, parameter invalid", user)
	}
	return &structs.UserInfo{ID: user.ID, Name: user.Name, TenantID: user.TenantID}, arw.DB(ctx).Create(&user).Error
}

func (arw *AccountReadWrite) DeleteUser(ctx context.Context, tenantID, userID string) error {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Warningf("delete user,tenantID: %s, userID: %s, parameter invalid", tenantID, userID)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "create user: tenantID: %s, userID:%s, parameter invalid", tenantID, userID)
	}
	return arw.DB(ctx).Where("tenant_id = ? AND id = ?", tenantID, userID).Unscoped().Delete(&User{}).Error
}

func (arw *AccountReadWrite) GetUser(ctx context.Context, tenantID, userID string) (userInfo structs.UserInfo, err error) {
	if "" == tenantID || "" == userID {
		framework.LogWithContext(ctx).Warningf("get user profile,tenantID: %s, userID: %s, parameter invalid", tenantID, userID)
		return structs.UserInfo{}, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "get user profile: tenantID: %s, userID:%s, parameter invalid", tenantID, userID)
	}
	user := User{ID: userID, TenantID: tenantID}
	err = arw.DB(ctx).Where(&User{ID: userID, TenantID: tenantID}).First(user).Error
	if err == gorm.ErrRecordNotFound {
		return structs.UserInfo{}, errors.NewEMErrorf(errors.UserNotExist, "query user %s profile not found", userID)
	}
	if err != nil {
		return structs.UserInfo{}, errors.NewEMErrorf(errors.QueryUserScanRowError, "query user profile,tenantID: %s, userID:%s, scan data error: %v,", tenantID, userID, err)
	}
	return structs.UserInfo{ID: user.ID, Name: user.Name, TenantID: user.TenantID,
		Email: user.Email, Phone: user.Phone, Status: user.Status, CreateAt: user.CreatedAt, UpdateAt: user.UpdatedAt}, err
}

func (arw *AccountReadWrite) QueryUsers(ctx context.Context) (userInfos map[string]structs.UserInfo, err error) {
	var info structs.UserInfo
	SQL := "SELECT id,name,tenant_id,status,email,phone,create_at,update_at,creator FROM users;"
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
	if "" == tenantID || "" == userID || "" == status {
		framework.LogWithContext(ctx).Warningf("update user status,tenantID: %s, userID: %s, status: %s, parameter invalid", tenantID, userID, status)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update user status: tenantID: %s, userID:%s, status:%s, parameter invalid", tenantID, userID, status)
	}
	return arw.DB(ctx).Model(&User{}).Where("tenant_id = ? AND id = ?", tenantID, userID).Update("status = ?", status).Error
}

func (arw *AccountReadWrite) UpdateUserProfile(ctx context.Context, tenantID, userID, email, phone string) error {
	if "" == tenantID || "" == userID || "" == email || "" == phone {
		framework.LogWithContext(ctx).Warningf("update user profile,tenantID: %s, userID: %s, email: %s, phone: %s, parameter invalid", tenantID, userID, email, phone)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update user profile: tenantID: %s, userID:%s, email: %s, phone: %s, parameter invalid", tenantID, userID, email, phone)
	}
	sql := make(map[string]interface{})
	if "" != email {
		sql["email"] = email
	}
	if "" != phone {
		sql["phone"] = phone
	}
	return arw.DB(ctx).Model(&User{}).Where("tenant_id = ? AND id = ?", tenantID, userID).Updates(sql).Error
}

func (arw *AccountReadWrite) UpdateUserPassword(ctx context.Context, tenantID, userID, salt, finalHash string) error {
	if "" == tenantID || "" == userID || "" == salt || "" == finalHash {
		framework.LogWithContext(ctx).Warningf("update user password,tenantID: %s, userID: %s, salt: %s, finalHash: %s, parameter invalid", tenantID, userID, salt, finalHash)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update user password: tenantID: %s, userID:%s, salt: %s, finalHash: %s, parameter invalid", tenantID, userID, salt, finalHash)
	}
	return arw.DB(ctx).Model(&User{}).Where("tenant_id = ? AND id = ?", tenantID, userID).Updates(map[string]interface{}{"salt": salt, "final_hash": finalHash}).Error
}

func (arw *AccountReadWrite) CreateTenant(ctx context.Context, tenant Tenant) (info *structs.TenantInfo, err error) {
	if "" == tenant.ID || "" == tenant.Name {
		framework.LogWithContext(ctx).Warningf("create tenant %v, parameter invalid", tenant)
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "create tenant %v, parameter invalid", tenant)
	}
	return &structs.TenantInfo{ID: tenant.ID, Name: tenant.Name}, arw.DB(ctx).Create(&tenant).Error
}

func (arw *AccountReadWrite) DeleteTenant(ctx context.Context, tenantID string) error {
	if "" == tenantID {
		framework.LogWithContext(ctx).Warningf("delete tenant,tenantID: %s, parameter invalid", tenantID)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "create tenant: tenantID: %s, parameter invalid", tenantID)
	}
	return arw.DB(ctx).Where("id = ?", tenantID).Unscoped().Delete(&Tenant{}).Error
}

func (arw *AccountReadWrite) GetTenant(ctx context.Context, tenantID string) (tenant structs.TenantInfo, err error) {
	if "" == tenantID {
		framework.LogWithContext(ctx).Warningf("get tenant,tenantID: %s, parameter invalid", tenantID)
		return structs.TenantInfo{}, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "get tenant: tenantID: %s, parameter invalid", tenantID)
	}
	info := &Tenant{ID: tenantID}
	err = arw.DB(ctx).Where(&Tenant{ID: tenantID}).First(info).Error
	if err == gorm.ErrRecordNotFound {
		return structs.TenantInfo{}, errors.NewEMErrorf(errors.TenantNotExist, "query user %s profile not found", tenantID)
	}
	if err != nil {
		return structs.TenantInfo{}, errors.NewEMErrorf(errors.QueryTenantScanRowError, "get tenant profile,tenantID: %s, query data error: %v,", tenantID, err)
	}
	return structs.TenantInfo{ID: info.ID, Name: info.Name, Creator: info.Creator, MaxCPU: info.MaxCPU, MaxMemory: info.MaxMemory, MaxStorage: info.MaxStorage,
		OnBoardingStatus: info.OnBoardingStatus, MaxCluster: info.MaxCluster, Status: info.Status, CreateAt: info.CreatedAt, UpdateAt: info.UpdatedAt}, err
}

func (arw *AccountReadWrite) QueryTenants(ctx context.Context) (tenants map[string]structs.TenantInfo, err error) {
	var info structs.TenantInfo
	SQL := "SELECT id,name,creator,status,on_boarding_status,max_cluster,create_at,update_at,creator,max_cpu,max_memory,max_storage FROM users;"
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
	if "" == tenantID || "" == status {
		framework.LogWithContext(ctx).Warningf("update tenant status,tenantID: %s, status: %s, parameter invalid", tenantID, status)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update tenant status: tenantID: %s, status:%s, parameter invalid", tenantID, status)
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Update("status = ?", status).Error
}

func (arw *AccountReadWrite) UpdateTenantProfile(ctx context.Context, tenantID, name string, maxCluster, maxCPU, maxMemory, maxStorage int32) error {
	if tenantID == "" {
		framework.LogWithContext(ctx).Warningf("update tenant profile,tenantID: %s, name: %s, maxCluster: %d, maxCPU: %d, maxMemory: %d, maxStorage: %d,parameter invalid", tenantID, name, maxCluster, maxCPU, maxMemory, maxStorage)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update tenant profile: tenantID: %s, name: %s, maxCluster: %d, maxCPU: %d, maxMemory: %d, maxStorage: %d, parameter invalid", tenantID, name, maxCluster, maxCPU, maxMemory, maxStorage)
	}
	sql := make(map[string]interface{})
	if "" != name {
		sql["name"] = name
	}
	if maxCluster > 0 {
		sql["max_cluster"] = maxCluster
	}
	if maxCPU > 0 {
		sql["max_cpu"] = maxCPU
	}
	if maxMemory > 0 {
		sql["max_memory"] = maxMemory
	}
	if maxStorage > 0 {
		sql["max_storage"] = maxStorage
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Updates(sql).Error
}

func (arw *AccountReadWrite) UpdateTenantOnBoardingStatus(ctx context.Context, tenantID, status string) error {
	if "" == tenantID || "" == status {
		framework.LogWithContext(ctx).Warningf("update tenant on boarding status,tenantID: %s, status: %s, parameter invalid", tenantID, status)
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "update tenant on boarding status: tenantID: %s, status:%s, parameter invalid", tenantID, status)
	}
	return arw.DB(ctx).Model(&Tenant{}).Where("id = ?", tenantID).Update("on_boarding_status = ?", status).Error
}
