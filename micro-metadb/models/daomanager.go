package models

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	common2 "github.com/pingcap-inc/tiem/library/common"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pingcap/errors"

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

	dao.SetClusterManager(new(DAOClusterManager))
	dao.ClusterManager().SetDb(dao.Db())
	dao.SetAccountManager(new(DAOAccountManager))
	dao.AccountManager().SetDb(dao.Db())
	return err
}

func (dao *DAOManager) InitTables() error {
	log := framework.GetLogger()

	log.Info("start create TiEM system tables.")

	dao.tables = make(map[string]interface{})
	dao.AddTable(TABLE_NAME_CLUSTER, new(Cluster))
	dao.AddTable(TABLE_NAME_DEMAND_RECORD, new(DemandRecord))
	dao.AddTable(TABLE_NAME_ACCOUNT, new(Account))
	dao.AddTable(TABLE_NAME_TENANT, new(Tenant))
	dao.AddTable(TABLE_NAME_ROLE, new(Role))
	dao.AddTable(TABLE_NAME_ROLE_BINDING, new(RoleBinding))
	dao.AddTable(TABLE_NAME_PERMISSION, new(Permission))
	dao.AddTable(TABLE_NAME_PERMISSION_BINDING, new(PermissionBinding))
	dao.AddTable(TABLE_NAME_TOKEN, new(Token))
	dao.AddTable(TABLE_NAME_TASK, new(TaskDO))
	dao.AddTable(TABLE_NAME_HOST, new(Host))
	dao.AddTable(TABLE_NAME_DISK, new(Disk))
	dao.AddTable(TABLE_NAME_TIUP_CONFIG, new(TiUPConfig))
	dao.AddTable(TABLE_NAME_TIUP_TASK, new(TiupTask))
	dao.AddTable(TABLE_NAME_FLOW, new(FlowDO))
	dao.AddTable(TABLE_NAME_PARAMETERS_RECORD, new(ParametersRecord))
	dao.AddTable(TABLE_NAME_BACKUP_RECORD, new(BackupRecord))
	dao.AddTable(TABLE_NAME_RECOVER_RECORD, new(RecoverRecord))
	dao.AddTable(TABLE_NAME_RECOVER_RECORD, new(RecoverRecord))

	log.Info("create TiEM all tables successful.")
	return nil
}

func (dao *DAOManager) InitData() error {

	err := dao.InitSystemDefaultData()
	if nil != err {
		framework.GetLogger().Errorf("initialize TiEM system data failed, error: %v", err)
	}

	framework.GetLogger().Infof(" initialization system default data successful")

	err = dao.InitResourceDataForDev()
	if nil != err {
		framework.GetLogger().Errorf("initialize TiEM system test resource failed, error: %v", err)
	}
	return err
}

func (dao DAOManager) AddTable(tableName string, tableModel interface{}) error {
	log := framework.GetLogger()
	dao.tables[tableName] = tableModel
	if !dao.db.Migrator().HasTable(dao.tables[tableName]) {
		er := dao.db.Migrator().CreateTable(dao.tables[tableName])
		if nil != er {
			log.Errorf("create table %s failed, error : %v.", tableName, er)
			return errors.New(fmt.Sprintf("crete table %s failed, error: %v.", tableName, er))
		}
	}
	return nil
}

/*
Initial TiEM system default system account and tenant information
*/
func (dao *DAOManager) InitSystemDefaultData() error {
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

func (dao *DAOManager) InitResourceDataForDev() error {
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
