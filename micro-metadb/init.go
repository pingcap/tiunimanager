package main

import (
	cryrand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/pingcap/ticp/library/firstparty/config"
	"github.com/pingcap/ticp/library/thirdparty/logger"
	"github.com/pingcap/ticp/library/thirdparty/tracer"

	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/ticp/micro-metadb/models"
	db "github.com/pingcap/ticp/micro-metadb/proto"
	"github.com/pingcap/ticp/micro-metadb/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Global LogRecord object
var log *logger.LogRecord

func initConfig() {
	config.InitForMonolith()
}

func initLogger() {
	log = logger.GetLogger()
	service.InitLogger()
	// use log
	log.Debug("init logger completed!")
}

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		log.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(service.TiCPMetaDBServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)
	srv.Init()

	db.RegisterTiCPDBServiceHandler(srv.Server(), new(service.DBServiceHandler))

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

func initSqliteDB() {
	var err error
	dbFile := config.GetSqliteFilePath()
	log.Record("dbFile", dbFile).Info("init: sqlite.open")
	models.MetaDB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	if err != nil {
		log.Fatalf("sqlite open error %v", err)
	}

	if models.MetaDB.Error != nil {
		log.Fatalf("database error %v", models.MetaDB.Error)
	}
	log.Info("sqlite.open success")

	initTables()

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
		&models.TiUPConfigDO{},
		&models.TaskDO{},
		&models.FlowDO{},
		&models.Host{},
		&models.Disk{},
		&models.TiupTask{},
		&models.ParametersRecordDO{},
		&models.BackupRecordDO{},
		&models.BackupRecordDO{},
	)
	return err
}

func initDataForDemo() {
	tenant, _ := models.AddTenant("Ticp系统管理", 1, 0)
	fmt.Println("tenantId = ", tenant.ID)

	role1, _ := models.AddRole(tenant.ID, "管理员", "管理员", 0)
	fmt.Println("role1.Id = ", role1.ID)

	role2, _ := models.AddRole(tenant.ID, "DBA", "DBA", 0)
	fmt.Println("role2.Id = ", role2.ID)

	userId1 := initUser(tenant.ID, "admin")
	fmt.Println("user1.Id = ", userId1)

	userId2 := initUser(tenant.ID, "peijin")
	fmt.Println("user2.Id = ", userId2)

	userId3 := initUser(tenant.ID, "nopermission")
	fmt.Println("user3.Id = ", userId3)

	models.AddRoleBindings([]models.RoleBinding{
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, AccountId: userId1},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, AccountId: userId2, },
	})

	permission1, _ := models.AddPermission(tenant.ID, "/api/v1/host/query", "查询主机", "查询主机", 2, 0)
	permission2, _ := models.AddPermission(tenant.ID, "/api/v1/instance/query", "查询集群", "查询集群", 2, 0)
	permission3, _ := models.AddPermission(tenant.ID, "/api/v1/instance/create", "创建集群", "创建集群", 2, 0)

	models.AddPermissionBindings([]models.PermissionBinding{
		// 管理员可做所有事
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission1.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission2.ID},
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role1.ID, PermissionId: permission3.ID},
		// 用户可做查询主机
		{Entity: models.Entity{TenantId: tenant.ID, Status: 0}, RoleId: role2.ID, PermissionId: permission1.ID},
	})

	// 添加一些demo使用的host和disk数据
	models.CreateHost(&models.Host{
		HostName: "主机1",
		IP:       "192.168.125.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 32,
		Memory:   64,
		Nic:      "1GE",
		AZ:       "East China",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []models.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	})

	return
}

func initUser(tenantId string, name string) string {

	b := make([]byte, 16)
	_, _ = cryrand.Read(b)

	salt := base64.URLEncoding.EncodeToString(b)

	s := salt + name
	finalSalt, _ := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	account, _ := models.AddAccount(tenantId, name, salt, string(finalSalt), 0)

	return account.ID
}
