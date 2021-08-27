package models

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pingcap/errors"

	common2 "github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
)

type DAOManager struct {
	db *gorm.DB
	framework *framework.BaseFramework
	tables map[string]interface {}
}

func (dao *DAOManager) Db() *gorm.DB {
	return dao.db
}

func (dao *DAOManager) Framework() *framework.BaseFramework {
	return dao.framework
}

func (dao *DAOManager) SetFramework(framework *framework.BaseFramework) {
	dao.framework = framework
}

func (dao *DAOManager) SetDb(db *gorm.DB) {
	dao.db = db
}

func (dao * DAOManager) Tables() map[string]interface{} {
	return dao.tables
}

func (dao * DAOManager) InitDB () error {
	var err error
	dbFile := dao.Framework().GetDataDir() + common2.DBDirPrefix + common2.SqliteFileName
	logins := dao.Framework().GetLogger().Record("database file path", dbFile)

	dao.db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil || dao.db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, dao.db.Error)
	} else {
		logins.Infof("open database successful, filepath: %s", dbFile)
	}

	err = dao.initTables()
	if nil != err {
		logins.Errorf("initialize TiEM system tables failed, error: %v", err)
		return err
	}

	err = dao.initSystemDefaultData()
	if nil != err {
		logins.Errorf("initialize TiEM system data failed, error: %v", err)
	}

	logins.Infof(" initialization system default data successful")

	err = dao.initResourceDataForDev()
	if nil != err {
		logins.Errorf("initialize TiEM system test resource failed, error: %v", err)
	}
	return err
}

func (dao *DAOManager) initTables() error {
	log := dao.Framework().GetLogger()

	log.Info("start create TiEM system tables.")

	dao.tables[TALBE_NAME_CLUSTER] = new(Cluster)
	dao.tables[TABLE_NAME_DEMAND_RECORD]  = new(DemandRecord)
	dao.tables[TABLE_NAME_ACCOUNT] = new(Account)
	dao.tables[TABLE_NAME_TENANT]  = new(Tenant)
	dao.tables[TABLE_NAME_ROLE]    = new(Role)
	dao.tables[TABLE_NAME_ROLE_BINDING] = new(RoleBinding)
	dao.tables[TABLE_NAME_PERMISSION] = new(Permission)
	dao.tables[TABLE_NAME_PERMISSION_BINDING] = new(PermissionBinding)
	dao.tables[TABLE_NAME_TOKEN]   = new(Token)
	dao.tables[TABLE_NAME_TASK]   = new(TaskDO)
	dao.tables[TABLE_NAME_HOST]	   = new(Host)
	dao.tables[TABLE_NAME_DISK]    = new(Disk)
	dao.tables[TABLE_NAME_TIUP_CONFIG]    = new(TiUPConfig)
	dao.tables[TABLE_NAME_TIUP_TASK] = new(TiupTask)
	dao.tables[TABLE_NAME_FLOW]    = new (FlowDO)
	dao.tables[TABLE_NAME_PARAMETERS_RECORD] = new (ParametersRecord)
	dao.tables[TABLE_NAME_BACKUP_RECORD] = new (BackupRecord)
	dao.tables[TABLE_NAME_RECOVER_RECORD] = new (RecoverRecord)

	var er error

	for name := range dao.tables {
		if !dao.db.Migrator().HasTable(dao.tables[name]) {
			er = dao.db.Migrator().CreateTable(dao.tables[name])
			if nil != er {
				log.Errorf("create table %s failed, error : %v.", name, er)
				return errors.New(fmt.Sprintf("crete table %s failed, error: %v.", name, er))
			}
		}
	}
	log.Info("create TiEM all tables successful.")
	return er
}
/*
Initial TiEM system default system account and tenant information
 */
func (dao * DAOManager) initSystemDefaultData() error {
	tenant := dao.tables[TABLE_NAME_TENANT].(*Tenant)
	role:= dao.tables[TABLE_NAME_ROLE].(*Role)
	log := dao.Framework().GetLogger()
	var rt, err = tenant.AddTenant(dao.Db(),"TiEM system administration", 1, 0)
	framework.AssertNoErr(err)
	role1, err := role.AddRole(dao.Db(),rt.ID, "administrators", "administrators", 0)
	framework.AssertNoErr(err)
	role2, err := role.AddRole(dao.Db(),rt.ID, "DBA", "DBA", 0)
	framework.AssertNoErr(err)
	userId1,err := dao.initUser(rt.ID, "admin")
	framework.AssertNoErr(err)
	userId2,err := dao.initUser(rt.ID, "nopermission")
	framework.AssertNoErr(err)

	log.Infof("initialization default tencent: %s, roles: %s, %s, users:%s, %s", tenant, role1, role2, userId1, userId2)

	roleBinding := dao.tables[TABLE_NAME_ROLE_BINDING].(*RoleBinding)
	err = roleBinding.AddRoleBindings(dao.Db(),[]RoleBinding{
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role1.ID, AccountId: userId1},
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role2.ID, AccountId: userId2},
	})

	permission := dao.tables[TABLE_NAME_PERMISSION].(*Permission)
	permission1, err := permission.AddPermission(dao.Db(),rt.ID, "/api/v1/host/query", " Query hosts", "Query hosts", 2, 0)
	permission2, err := permission.AddPermission(dao.Db(),rt.ID, "/api/v1/instance/query", "Query cluster", "Query cluster", 2, 0)
	permission3, err := permission.AddPermission(dao.Db(),rt.ID, "/api/v1/instance/create", "Create cluster", "Create cluster", 2, 0)

	pb := dao.tables[TABLE_NAME_PERMISSION_BINDING].(*PermissionBinding)
	err = pb.AddPermissionBindings(dao.Db(),[]PermissionBinding{
		// Administrators can do everything
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission1.ID},
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission2.ID},
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission3.ID},

		// User can do query host and cluster
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission1.ID},
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission2.ID},
	})
	framework.AssertNoErr(err)
	return err
}

func (dao * DAOManager) initResourceDataForDev() error {
	_, err := CreateHost(dao.Db(),&Host{
		HostName: "主机1",
		IP:       "192.168.125.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	})

	_, err = CreateHost(dao.Db(),&Host{
		HostName: "主机2",
		IP:       "192.168.125.133",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tikv", Capacity: 256, Status: 1},
		},
	})

	_, err = CreateHost(dao.Db(),&Host{
		HostName: "主机3",
		IP:       "192.168.125.134",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/pd", Capacity: 256, Status: 1},
		},
	})

	_, err = CreateHost(dao.Db(),&Host{
		HostName: "主机4",
		IP:       "192.168.125.135",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone2",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})

	_, err = CreateHost(dao.Db(),&Host{
		HostName: "主机4",
		IP:       "192.168.125.136",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})

	_, err = CreateHost(dao.Db(),&Host{
		HostName: "主机4",
		IP:       "192.168.125.137",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})

	dao.Framework().GetLogger().Error(err)
	return nil
}

func (dao * DAOManager) initUser(tenantId string, name string) (string, error) {
	log := dao.Framework().GetLogger()

	b := make([]byte, 16)
	_, err := cryrand.Read(b)
	framework.AssertNoErr(err)

	salt := base64.URLEncoding.EncodeToString(b)

	s := salt + name
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	framework.AssertNoErr(err)
	account := dao.tables[TABLE_NAME_ACCOUNT].(*Account)
	rt, err := account.Add(dao.Db(),tenantId, name, salt, string(finalSalt), 0)

	if nil != err {
		log.Errorf("add account failed, error: %v, id: %s, name: %s", err,rt.ID, name)
	} else {
		log.Info("add account successful, id: %s, name: %s", rt.ID, name)
	}
	return rt.ID, err
}
