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
	db             *gorm.DB
	tables         map[string]interface{}
	clusterManager *DAOClusterManager
	accountManager *DAOAccountManager
}

func NewDAOManager(d *gorm.DB) *DAOManager {
	m := new(DAOManager)
	return m
}

func (dao *DAOManager) AccountManager() *DAOAccountManager {
	return dao.accountManager
}

func (dao *DAOManager) SetAccountManager(accountManager *DAOAccountManager) {
	dao.accountManager = accountManager
}

func (dao *DAOManager) ClusterManager() *DAOClusterManager {
	return dao.clusterManager
}

func (dao *DAOManager) SetClusterManager(clusterManager *DAOClusterManager) {
	dao.clusterManager = clusterManager
}

func (dao *DAOManager) Db() *gorm.DB {
	return dao.db
}

func (dao *DAOManager) SetDb(db *gorm.DB) {
	dao.db = db
}

func (dao *DAOManager) Tables() map[string]interface{} {
	return dao.tables
}

func (dao *DAOManager) InitDB(dataDir string) error {
	var err error
	dbFile := dataDir + common2.DBDirPrefix + common2.SqliteFileName
	logins := framework.GetLogger().Record("database file path", dbFile)
	dao.db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil || dao.db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, dao.db.Error)
	} else {
		logins.Infof("open database successful, filepath: %s", dbFile)
	}

	dao.SetAccountManager(NewDAOAccountManager(dao.Db()))
	dao.SetClusterManager(NewDAOClusterManager(dao.Db()))

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
	dao.SetClusterManager(new(DAOClusterManager))
	dao.ClusterManager().SetDb(dao.Db())
	dao.SetAccountManager(new(DAOAccountManager))
	dao.AccountManager().SetDb(dao.Db())
	return err
}

func (dao *DAOManager) initTables() error {
	log := framework.GetLogger()

	log.Info("start create TiEM system tables.")

	dao.tables = make(map[string]interface{})
	dao.tables[TABLE_NAME_CLUSTER] = new(Cluster)
	dao.tables[TABLE_NAME_DEMAND_RECORD] = new(DemandRecord)
	dao.tables[TABLE_NAME_ACCOUNT] = new(Account)
	dao.tables[TABLE_NAME_TENANT] = new(Tenant)
	dao.tables[TABLE_NAME_ROLE] = new(Role)
	dao.tables[TABLE_NAME_ROLE_BINDING] = new(RoleBinding)
	dao.tables[TABLE_NAME_PERMISSION] = new(Permission)
	dao.tables[TABLE_NAME_PERMISSION_BINDING] = new(PermissionBinding)
	dao.tables[TABLE_NAME_TOKEN] = new(Token)
	dao.tables[TABLE_NAME_TASK] = new(TaskDO)
	dao.tables[TABLE_NAME_HOST] = new(Host)
	dao.tables[TABLE_NAME_DISK] = new(Disk)
	dao.tables[TABLE_NAME_TIUP_CONFIG] = new(TiUPConfig)
	dao.tables[TABLE_NAME_TIUP_TASK] = new(TiupTask)
	dao.tables[TABLE_NAME_FLOW] = new(FlowDO)
	dao.tables[TABLE_NAME_PARAMETERS_RECORD] = new(ParametersRecord)
	dao.tables[TABLE_NAME_BACKUP_RECORD] = new(BackupRecord)
	dao.tables[TABLE_NAME_RECOVER_RECORD] = new(RecoverRecord)

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
func (dao *DAOManager) initSystemDefaultData() error {
	accountManager := dao.AccountManager()
	log := framework.GetLogger()
	rt, err := accountManager.AddTenant("TiEM system administration", 1, 0)
	framework.AssertNoErr(err)
	role1, err := accountManager.AddRole(rt.ID, "administrators", "administrators", 0)
	framework.AssertNoErr(err)
	role2, err := accountManager.AddRole(rt.ID, "DBA", "DBA", 0)
	framework.AssertNoErr(err)
	userId1, err := dao.initUser(rt.ID, "admin")
	framework.AssertNoErr(err)
	userId2, err := dao.initUser(rt.ID, "nopermission")
	framework.AssertNoErr(err)

	log.Infof("initialization default tenant: %s, roles: %s, %s, users: %s, %s",
		rt.Name, role1.Name, role2.Name, userId1, userId2)

	err = accountManager.AddRoleBindings([]RoleBinding{
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role1.ID, AccountId: userId1},
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role2.ID, AccountId: userId2},
	})

	permission1, err := accountManager.AddPermission(rt.ID, "/api/v1/host/query", " Query hosts", "Query hosts", 2, 0)
	permission2, err := accountManager.AddPermission(rt.ID, "/api/v1/instance/query", "Query cluster", "Query cluster", 2, 0)
	permission3, err := accountManager.AddPermission(rt.ID, "/api/v1/instance/create", "Create cluster", "Create cluster", 2, 0)

	err = accountManager.AddPermissionBindings([]PermissionBinding{
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

func (dao *DAOManager) initResourceDataForDev() error {
	_, err := CreateHost(dao.Db(), &Host{
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

	_, err = CreateHost(dao.Db(), &Host{
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

	_, err = CreateHost(dao.Db(), &Host{
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

	_, err = CreateHost(dao.Db(), &Host{
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

	_, err = CreateHost(dao.Db(), &Host{
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

	_, err = CreateHost(dao.Db(), &Host{
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

	framework.GetLogger().Error(err)
	return nil
}

func (dao *DAOManager) initUser(tenantId string, name string) (string, error) {
	log := framework.GetLogger()
	accountManager := dao.AccountManager()

	b := make([]byte, 16)
	_, err := cryrand.Read(b)
	framework.AssertNoErr(err)

	salt := base64.URLEncoding.EncodeToString(b)

	s := salt + name
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	framework.AssertNoErr(err)
	rt, err := accountManager.Add(tenantId, name, salt, string(finalSalt), 0)

	if nil != err {
		log.Errorf("add account failed, error: %v, id: %s, name: %s", err, rt.ID, name)
	} else {
		log.Infof("add account successful, id: %s, name: %s", rt.ID, name)
	}
	return rt.ID, err
}
