package main

import (
	cryrand "crypto/rand"
	"encoding/base64"
	common2 "github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initSqliteDB(f *framework.BaseFramework) error {
	var err error
	dbFile := f.GetDataDir() + common2.DBDirPrefix + common2.SqliteFileName
	logins := f.GetLogger().Record("database file path", dbFile)

	models.MetaDB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil || models.MetaDB.Error != nil {
		logins.Fatalf("open database failed, database error: %v, metadb error: %v", err, models.MetaDB.Error)
	} else {
		logins.Infof("open database successful")
	}

	framework.AssertNoErr(initTables(f))

	initTenantDataForDev(f)

	logins.Infof(" initialization system default data successful")

	initResourceDataForDev(f)
	return err
}

func initTables(f *framework.BaseFramework) error {
	models.MetaDB.Migrator().CreateTable(
		&models.Tenant{},
		&models.AccountDO{},
		&models.RoleDO{},
		&models.PermissionDO{},
		&models.PermissionBindingDO{},
		&models.RoleBindingDO{},
		&models.Token{},
		&models.ClusterDO{},
		&models.DemandRecordDO{},
		&models.TiUPConfigDO{},
		&models.TaskDO{},
		&models.FlowDO{},
		&models.Host{},
		&models.Disk{},
		&models.TiupTask{},
		&models.TransportRecord{},
		&models.ParametersRecordDO{},
		&models.BackupRecordDO{},
		&models.RecoverRecordDO{},
	)
	return nil
}

func initTenantDataForDev(f *framework.BaseFramework) error {
	var err error
	tenant, err := models.AddTenant("TiEM system administration", 1, 0)
	if err != nil {
		f.GetLogger().Fatal(" TODO ")
	}
	role1, err := models.AddRole(tenant.ID, "administrators", "administrators", 0)
	framework.AssertNoErr(err)
	role2, err := models.AddRole(tenant.ID, "DBA", "DBA", 0)
	framework.AssertNoErr(err)
	userId1 := initUser(tenant.ID, "admin")
	userId2 := initUser(tenant.ID, "nopermission")

	f.GetLogger().Infof("initialization default tencent: %s, roles: %s, %s, users:%s, %s", tenant, role1, role2, userId1, userId2)

	err = models.AddRoleBindings([]models.RoleBindingDO{
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, AccountId: userId1},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, AccountId: userId2},
	})

	permission1, err := models.AddPermission(tenant.ID, "/api/v1/host/query", " Query hosts", "Query hosts", 2, 0)
	permission2, err := models.AddPermission(tenant.ID, "/api/v1/instance/query", "Query cluster", "Query cluster", 2, 0)
	permission3, err := models.AddPermission(tenant.ID, "/api/v1/instance/create", "Create cluster", "Create cluster", 2, 0)

	err = models.AddPermissionBindings([]models.PermissionBindingDO{
		// Administrators can do everything
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission1.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission2.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission3.ID},

		// User can do query host and cluster
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission1.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission2.ID},
	})
	framework.AssertNoErr(err)
	return nil
}

func initResourceDataForDev(f *framework.BaseFramework) error {
	// 添加一些demo使用的host和disk数据
	_, err := models.CreateHost(&models.Host{
		HostName: "主机1",
		IP:       "192.168.125.132",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		DC:       "DataCenter1",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	})

	// 添加一些demo使用的host和disk数据
	_, err = models.CreateHost(&models.Host{
		HostName: "主机2",
		IP:       "192.168.125.133",
		UserName: "root",
		Passwd:   "e446c4324d4c240b4b07a0a738f2820952ebeaa11ad9",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		DC:       "DataCenter1",
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/tikv", Capacity: 256, Status: 0},
		},
	})

	// 添加一些demo使用的host和disk数据
	_, err = models.CreateHost(&models.Host{
		HostName: "主机3",
		IP:       "192.168.125.134",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		DC:       "DataCenter1",
		AZ:       "Zone2",
		Rack:     "3-2",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/pd", Capacity: 256, Status: 1},
		},
	})

	_, err = models.CreateHost(&models.Host{
		HostName: "主机4",
		IP:       "192.168.125.135",
		UserName: "root",
		Passwd:   "431156422e1746bc108e7f334ebfdb080219be9b6f2b",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		DC:       "DataCenter1",
		AZ:       "Zone2",
		Rack:     "3-2",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 0},
		},
	})

	_, err = models.CreateHost(&models.Host{
		HostName: "主机5",
		IP:       "192.168.125.136",
		UserName: "root",
		Passwd:   "431156422e1746bc108e7f334ebfdb080219be9b6f2b",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		DC:       "DataCenter1",
		AZ:       "Zone3",
		Rack:     "3-3",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})

	_, err = models.CreateHost(&models.Host{
		HostName: "主机6",
		IP:       "192.168.125.137",
		UserName: "root",
		Passwd:   "e446c4324d4c240b4b07a0a738f2820952ebeaa11ad9",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		DC:       "DataCenter1",
		AZ:       "Zone3",
		Rack:     "3-3",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 0},
		},
	})

	f.GetLogger().Error(err)
	return nil
}

func initUser(tenantId string, name string) string {
	b := make([]byte, 16)
	_, err := cryrand.Read(b)
	framework.AssertNoErr(err)

	salt := base64.URLEncoding.EncodeToString(b)

	s := salt + name
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	framework.AssertNoErr(err)
	account, err := models.AddAccount(tenantId, name, salt, string(finalSalt), 0)

	return account.ID
}
