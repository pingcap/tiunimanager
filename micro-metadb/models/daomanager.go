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
 *                                                                            *
 ******************************************************************************/

package models

import (
	"context"
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"strconv"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"

	"github.com/pingcap/errors"

	"github.com/pingcap-inc/tiem/library/framework"
)

type DAOManager struct {
	db              *gorm.DB
	daoLogger       gormLog.Interface
	tables          map[string]interface{}
	clusterManager  *DAOClusterManager
	accountManager  *DAOAccountManager
	resourceManager *DAOResourceManager
	paramManager    *DAOParamManager
}

func NewDAOManager(fw *framework.BaseFramework) *DAOManager {
	m := new(DAOManager)
	m.daoLogger = &DaoLogger{
		p:             fw,
		SlowThreshold: common.SlowSqlThreshold,
	}
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

func (dao *DAOManager) ResourceManager() *DAOResourceManager {
	return dao.resourceManager
}

func (dao *DAOManager) SetResourceManager(resourceManager *DAOResourceManager) {
	dao.resourceManager = resourceManager
}

func (dao *DAOManager) ParamManager() *DAOParamManager {
	return dao.paramManager
}

func (dao *DAOManager) SetParamManager(paramManager *DAOParamManager) {
	dao.paramManager = paramManager
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

const StartTime = "StartTime"

func (dao *DAOManager) InitMetrics() {
	framework.LogForkFile(common.LogFileSystem).Infof("init sqlite metrics")

	before := func(db *gorm.DB) {
		db.InstanceSet(StartTime, time.Now())
		return
	}

	after := func(method string, db *gorm.DB) {
		code := db.Error == nil

		framework.Current.GetMetrics().SqliteRequestsCounterMetric.With(prometheus.Labels{
			metrics.ServiceLabel: db.Name(),
			metrics.MethodLabel:  method,
			metrics.CodeLabel:    strconv.FormatBool(code)}).
			Inc()

		startTime, ok := db.InstanceGet(StartTime)

		if ok {
			duration := time.Since(startTime.(time.Time)).Milliseconds()
			framework.Current.GetMetrics().SqliteDurationHistogramMetric.With(prometheus.Labels{
				metrics.ServiceLabel: db.Name(),
				metrics.MethodLabel:  method,
				metrics.CodeLabel:    strconv.FormatBool(code)}).
				Observe(float64(duration))

			if duration > common.SlowSqlThreshold {
				framework.Current.GetMetrics().SqliteRequestsCounterMetric.With(prometheus.Labels{
					metrics.ServiceLabel: db.Name(),
					metrics.MethodLabel:  method,
					metrics.CodeLabel:    "slow"}).
					Inc()
			}
		}

	}
	dao.db.Callback().Create().Before("gorm:before_create").Register("metrics_before_create", before)
	dao.db.Callback().Create().After("gorm:after_create").Register("metrics_after_create", func(db *gorm.DB) {
		after("Create", db)
	})

	dao.db.Callback().Query().Before("gorm:before_query").Register("metrics_before_query", before)
	dao.db.Callback().Query().After("gorm:after_query").Register("metrics_after_query", func(db *gorm.DB) {
		after("Query", db)
	})

	dao.db.Callback().Delete().Before("gorm:before_delete").Register("metrics_before_delete", before)
	dao.db.Callback().Delete().After("gorm:after_delete").Register("metrics_after_delete", func(db *gorm.DB) {
		after("Delete", db)
	})

	dao.db.Callback().Update().Before("gorm:before_update").Register("metrics_before_update", before)
	dao.db.Callback().Update().After("gorm:after_update").Register("metrics_after_update", func(db *gorm.DB) {
		after("Update", db)
	})
}

func (dao *DAOManager) InitDB(dataDir string) error {
	var err error
	dbFile := dataDir + common.DBDirPrefix + common.SqliteFileName
	logins := framework.LogForkFile(common.LogFileSystem).WithField("database file path", dbFile)
	dao.db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: dao.daoLogger,
	})

	if err != nil || dao.db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, dao.db.Error)
	} else {
		logins.Infof("open database successful, filepath: %s", dbFile)
	}

	dao.SetAccountManager(NewDAOAccountManager(dao.Db()))
	dao.SetClusterManager(NewDAOClusterManager(dao.Db()))
	dao.SetResourceManager(NewDAOResourceManager(dao.Db()))
	dao.SetParamManager(NewDAOParamManager(dao.Db()))

	return err
}

func (dao *DAOManager) InitTables() error {
	log := framework.LogForkFile(common.LogFileSystem)

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
	dao.AddTable(TABLE_NAME_HOST, new(resource.Host))
	dao.AddTable(TABLE_NAME_DISK, new(resource.Disk))
	dao.AddTable(TABLE_NAME_USED_COMPUTE, new(resource.UsedCompute))
	dao.AddTable(TABLE_NAME_USED_PORT, new(resource.UsedPort))
	dao.AddTable(TABLE_NAME_USED_DISK, new(resource.UsedDisk))
	dao.AddTable(TABLE_NAME_LABEL, new(resource.Label))
	dao.AddTable(TABLE_NAME_TIUP_CONFIG, new(TopologyConfig))
	dao.AddTable(TABLE_NAME_TIUP_TASK, new(TiupTask))
	dao.AddTable(TABLE_NAME_FLOW, new(FlowDO))
	dao.AddTable(TABLE_NAME_PARAMETERS_RECORD, new(ParametersRecord))
	dao.AddTable(TABLE_NAME_BACKUP_RECORD, new(BackupRecord))
	dao.AddTable(TABLE_NAME_BACKUP_STRATEGY, new(BackupStrategy))
	dao.AddTable(TABLE_NAME_TRANSPORT_RECORD, new(TransportRecord))
	dao.AddTable(TABLE_NAME_RECOVER_RECORD, new(RecoverRecord))
	dao.AddTable(TABLE_NAME_COMPONENT_INSTANCE, new(ComponentInstance))
	dao.AddTable(TABLE_NAME_PARAM, new(ParamDO))
	dao.AddTable(TABLE_NAME_PARAM_GROUP, new(ParamGroupDO))
	dao.AddTable(TABLE_NAME_PARAM_GROUP_MAP, new(ParamGroupMapDO))
	dao.AddTable(TABLE_NAME_CLUSTER_PARAM_MAP, new(ClusterParamMapDO))

	log.Info("create TiEM all tables successful.")
	return nil
}

func (dao *DAOManager) InitData() error {
	err := dao.initSystemDefaultData()
	innerLog := framework.LogForkFile(common.LogFileSystem)
	if nil != err {
		innerLog.Errorf("initialize TiEM system data failed, error: %v", err)
	}

	innerLog.Infof(" initialization system default data successful")

	err = dao.initSystemDefaultLabels()
	if nil != err {
		innerLog.Errorf("init system default labels failed, %v\n", err)
	}
	return err
}

func (dao DAOManager) AddTable(tableName string, tableModel interface{}) error {
	log := framework.LogForkFile(common.LogFileSystem)
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
func (dao *DAOManager) initSystemDefaultData() error {
	accountManager := dao.AccountManager()
	log := framework.LogForkFile(common.LogFileSystem)
	rt, err := accountManager.AddTenant(context.TODO(), "TiEM system administration", 1, 0)
	framework.AssertNoErr(err)
	role1, err := accountManager.AddRole(context.TODO(), rt.ID, "administrators", "administrators", 0)
	framework.AssertNoErr(err)
	role2, err := accountManager.AddRole(context.TODO(), rt.ID, "DBA", "DBA", 0)
	framework.AssertNoErr(err)
	userId1, err := dao.initUser(rt.ID, "admin")
	framework.AssertNoErr(err)
	userId2, err := dao.initUser(rt.ID, "nopermission")
	framework.AssertNoErr(err)

	log.Infof("initialization default tenant: %s, roles: %s, %s, users: %s, %s",
		rt.Name, role1.Name, role2.Name, userId1, userId2)

	err = accountManager.AddRoleBindings(context.TODO(), []RoleBinding{
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role1.ID, AccountId: userId1},
		{Entity: Entity{TenantId: rt.ID, Status: 0}, RoleId: role2.ID, AccountId: userId2},
	})
	framework.AssertNoErr(err)

	permission1, err := accountManager.AddPermission(context.TODO(), rt.ID, "/api/v1/host/query", " Query hosts", "Query hosts", 2, 0)
	framework.AssertNoErr(err)
	permission2, err := accountManager.AddPermission(context.TODO(), rt.ID, "/api/v1/instance/query", "Query cluster", "Query cluster", 2, 0)
	framework.AssertNoErr(err)
	permission3, err := accountManager.AddPermission(context.TODO(), rt.ID, "/api/v1/instance/create", "Create cluster", "Create cluster", 2, 0)
	framework.AssertNoErr(err)

	err = accountManager.AddPermissionBindings(context.TODO(), []PermissionBinding{
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

func (dao *DAOManager) initSystemDefaultLabels() (err error) {
	resourceManager := dao.ResourceManager()
	log := framework.LogForkFile(common.LogFileSystem)
	err = resourceManager.InitSystemDefaultLabels(context.TODO())
	if err != nil {
		log.Debugf("init system default labels failed, %v\n", err)
		return err
	}
	log.Debugln("init system default labels succeed")
	return nil
}

func (dao *DAOManager) InitResourceDataForDev(region, zone, rack, hostIp1, hostIp2, hostIp3 string) error {
	log := framework.LogForkFile(common.LogFileSystem)
	zoneCode := resource.GenDomainCodeByName(region, zone)
	rackCode := resource.GenDomainCodeByName(zoneCode, rack)
	id1, err := dao.ResourceManager().CreateHost(context.TODO(), &resource.Host{
		HostName:     "TEST_HOST1",
		IP:           hostIp1,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:       0,
		Arch:         string(resource.X86),
		OS:           "CentOS",
		Kernel:       "5.0.0",
		CpuCores:     16,
		Memory:       64,
		FreeCpuCores: 16,
		FreeMemory:   64,
		Nic:          "1GE",
		Region:       region,
		AZ:           zoneCode,
		Rack:         rackCode,
		Purpose:      string(resource.Compute),
		DiskType:     string(resource.Sata),
		Reserved:     false,
		Disks: []resource.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1, Type: string(resource.Sata)},
			{Name: "sdb", Path: "/mnt1", Capacity: 256, Status: 0, Type: string(resource.Sata)},
			{Name: "sdc", Path: "/mnt2", Capacity: 256, Status: 0, Type: string(resource.Sata)},
			{Name: "sdd", Path: "/mnt3", Capacity: 1024, Status: 0, Type: string(resource.Sata)},
		},
	})
	if err != nil {
		log.Errorf("create TEST_HOST1 failed, %v\n", err)
		return err
	}
	id2, err := dao.ResourceManager().CreateHost(context.TODO(), &resource.Host{
		HostName:     "TEST_HOST2",
		IP:           hostIp2,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:       0,
		Arch:         string(resource.X86),
		OS:           "CentOS",
		Kernel:       "5.0.0",
		CpuCores:     16,
		Memory:       64,
		FreeCpuCores: 16,
		FreeMemory:   64,
		Nic:          "1GE",
		Region:       region,
		AZ:           zoneCode,
		Rack:         rackCode,
		Purpose:      string(resource.Compute),
		DiskType:     string(resource.Sata),
		Reserved:     false,
		Disks: []resource.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1, Type: string(resource.Sata)},
			{Name: "sdb", Path: "/mnt1", Capacity: 256, Status: 0, Type: string(resource.Sata)},
			{Name: "sdc", Path: "/mnt2", Capacity: 256, Status: 0, Type: string(resource.Sata)},
			{Name: "sdd", Path: "/mnt3", Capacity: 1024, Status: 0, Type: string(resource.Sata)},
		},
	})
	if err != nil {
		log.Errorf("create TEST_HOST2 failed, %v\n", err)
		return err
	}
	id3, err := dao.ResourceManager().CreateHost(context.TODO(), &resource.Host{
		HostName:     "TEST_HOST3",
		IP:           hostIp3,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:       0,
		Arch:         string(resource.X86),
		OS:           "CentOS",
		Kernel:       "5.0.0",
		CpuCores:     12,
		Memory:       24,
		FreeCpuCores: 12,
		FreeMemory:   24,
		Nic:          "1GE",
		Region:       region,
		AZ:           zoneCode,
		Rack:         rackCode,
		Purpose:      string(resource.Compute),
		DiskType:     string(resource.Sata),
		Reserved:     false,
		Disks: []resource.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1, Type: string(resource.Sata)},
			{Name: "sdb", Path: "/mnt1", Capacity: 256, Status: 0, Type: string(resource.Sata)},
			{Name: "sdc", Path: "/mnt2", Capacity: 256, Status: 0, Type: string(resource.Sata)},
			{Name: "sdd", Path: "/mnt3", Capacity: 1024, Status: 0, Type: string(resource.Sata)},
		},
	})
	if err != nil {
		log.Errorf("create TEST_HOST3 failed, %v\n", err)
		return err
	}

	log.Infof("insert 3 test hosts[%s, %s, %s] completed.\n", id1, id2, id3)
	return nil
}

func (dao *DAOManager) initUser(tenantId string, name string) (string, error) {
	log := framework.LogForkFile(common.LogFileSystem)
	accountManager := dao.AccountManager()

	b := make([]byte, 16)
	_, err := cryrand.Read(b)
	framework.AssertNoErr(err)

	salt := base64.URLEncoding.EncodeToString(b)

	s := salt + name
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	framework.AssertNoErr(err)
	rt, err := accountManager.Add(context.TODO(), tenantId, name, salt, string(finalSalt), 0)

	if nil != err {
		log.Errorf("add account failed, error: %v, id: %s, name: %s", err, rt.ID, name)
	} else {
		log.Infof("add account successful, id: %s, name: %s", rt.ID, name)
	}
	return rt.ID, err
}
