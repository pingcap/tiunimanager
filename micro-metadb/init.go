package main

import (
	cryrand "crypto/rand"
	"crypto/tls"
	"encoding/base64"

	"github.com/asim/go-micro/v3"

	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	"github.com/pingcap/tiem/library/thirdparty/tracer"

	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/tiem/micro-metadb/models"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	"github.com/pingcap/tiem/micro-metadb/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Global LogRecord object
var log *logger.LogRecord

func initConfig() {
	srv := micro.NewService(
		config.GetMicroMetaDBCliArgsOption(),
	)
	srv.Init()
	srv = nil

	config.InitForMonolith(config.MicroMetaDBMod)
}

func initLogger(key config.Key) {
	log = logger.GetLogger(key)
	service.InitLogger(key)
	// use log
	log.Debug("init logger completed!")
}

func initTracer() {
	tracer.InitTracer()
}

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		log.Fatalf(" load certificate file %s failed, error %v", config.GetCertificateCrtFilePath(), err)
		return
	}
	log.Infof(" load certificate file %s successful", config.GetCertificateCrtFilePath())
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(service.TiEMMetaDBServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Address(config.GetMetaDBServiceAddress()),
		micro.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)
	srv.Init()

	db.RegisterTiEMDBServiceHandler(srv.Server(), new(service.DBServiceHandler))

	if err := srv.Run(); err != nil {
		log.Fatalf(" Initialization micro service failed, error %v, listening address %s, etcd registry address %s", err, config.GetMetaDBServiceAddress(), config.GetRegistryAddress())
	}
	log.Info(" Initialization micro service successful, listening address %s, etcd registry address %s", config.GetMetaDBServiceAddress(), config.GetRegistryAddress())
}

func initSqliteDB() {
	var err error
	dbFile := config.GetSqliteFilePath()
	logins := log.Record("database file path", dbFile)
	models.MetaDB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil || models.MetaDB.Error != nil {
		logins.Fatalf("open database failed, database error: %v, metadb error: %v", err, models.MetaDB.Error)
	} else {
		logins.Infof("open database successful")
	}
	err = initTables()
	if err != nil {
		logins.Fatalf(" load system tables from database failed, error: %v", err)
	} else {
		logins.Infof(" load system tables from database successful")
	}

	initDataForSystemDefault()

	logins.Infof(" initialization system default data successful")

	initDataForDemo()
}

func initTables() error {
	err := models.MetaDB.Migrator().CreateTable(
		&models.Tenant{},
		&models.Account{},
		&models.Role{},
		&models.Permission{},
		&models.PermissionBinding{},
		&models.RoleBinding{},
		&models.Token{},
		&models.ClusterDO{},
		&models.DemandRecordDO{},
		&models.TiUPConfigDO{},
		&models.TaskDO{},
		&models.FlowDO{},
		&models.Host{},
		&models.Disk{},
		&models.TiupTask{},
		&models.ParametersRecordDO{},
		&models.BackupRecordDO{},
		&models.RecoverRecordDO{},
	)
	return err
}

func initDataForSystemDefault() {
	var err error
	tenant, err := models.AddTenant("TiEM system administration", 1, 0)
	//TODO assert(err == nil)
	if err != nil {
		log.Fatal(" TODO ")
	}
	role1, err := models.AddRole(tenant.ID, "administrators", "administrators", 0)
	//TODO assert(err == nil)
	role2, err := models.AddRole(tenant.ID, "DBA", "DBA", 0)
	//TODO assert(err == nil)
	userId1, err := initUser(tenant.ID, "admin")
	//TODO assert(err == nil)
	userId2, err := initUser(tenant.ID, "nopermission")
	//TODO assert(err == nil)

	log.Infof("initialization default tencent: %s, roles: %s, %s, users:%s, %s", tenant, role1, role2, userId1, userId2)

	err = models.AddRoleBindings([]models.RoleBinding{
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, AccountId: userId1},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, AccountId: userId2},
	})
	//TODO assert(err == nil)

	permission1, err := models.AddPermission(tenant.ID, "/api/v1/host/query", " Query hosts", "Query hosts", 2, 0)
	//TODO assert(err == nil)
	permission2, err := models.AddPermission(tenant.ID, "/api/v1/instance/query", "Query cluster", "Query cluster", 2, 0)
	//TODO assert(err == nil)
	permission3, err := models.AddPermission(tenant.ID, "/api/v1/instance/create", "Create cluster", "Create cluster", 2, 0)
	//TODO assert(err == nil)

	err = models.AddPermissionBindings([]models.PermissionBinding{
		// Administrators can do everything
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission1.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission2.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission3.ID},

		// User can do query host and cluster
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission1.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission2.ID},
	})
	//TODO assert(err == nil)
}

func initDataForDemo() {
	// 添加一些demo使用的host和disk数据
	models.CreateHost(&models.Host{
		HostName: "主机1",
		IP:       "192.168.125.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 4,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	})
	// 添加一些demo使用的host和disk数据
	models.CreateHost(&models.Host{
		HostName: "主机2",
		IP:       "192.168.125.133",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 4,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/tikv", Capacity: 256, Status: 1},
		},
	})
	// 添加一些demo使用的host和disk数据
	models.CreateHost(&models.Host{
		HostName: "主机3",
		IP:       "192.168.125.134",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 4,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/pd", Capacity: 256, Status: 1},
		},
	})

	models.CreateHost(&models.Host{
		HostName: "主机4",
		IP:       "192.168.125.135",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 4,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone2",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})

	models.CreateHost(&models.Host{
		HostName: "主机4",
		IP:       "192.168.125.136",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 4,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})

	models.CreateHost(&models.Host{
		HostName: "主机4",
		IP:       "192.168.125.137",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 4,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/www", Capacity: 256, Status: 1},
		},
	})
	return
}

func initUser(tenantId string, name string) (string, error) {

	var err error
	b := make([]byte, 16)
	_, err = cryrand.Read(b)
	//TODO assert(err == nil)

	salt := base64.URLEncoding.EncodeToString(b)

	s := salt + name
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	//TODO assert(err == nil)
	account, err := models.AddAccount(tenantId, name, salt, string(finalSalt), 0)

	return account.ID, err
}
