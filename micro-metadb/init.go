package main

import (
	"crypto/tls"
	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/ticp/addon/logger"
	mylogger "github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/config"
	"github.com/pingcap/ticp/micro-metadb/models"
	"github.com/pingcap/ticp/micro-metadb/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var MetaDB *gorm.DB

func initConfig() {
	{
		// only use to init the config
		srv := micro.NewService(
			config.GetMicroCliArgsOption(),
		)
		srv.Init()
		config.Init()
		srv = nil
	}
}
func initLogger() {
	// log
	mlog.DefaultLogger = mlogrus.NewLogger(mlogrus.WithLogger(mylogger.WithContext(nil)))
}

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		mlog.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(service.TiCPMetaDBServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
	)
	srv.Init()
}

func initSqliteDB() {
	var err error
	dbFile := config.GetSqliteFilePath()
	log := logger.WithContext(nil).WithField("dbFile", dbFile)
	log.Debug("init: sqlite.open")
	MetaDB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	if err != nil {
		log.Fatalf("sqlite open error %v", err)
	}

	if MetaDB.Error != nil {
		log.Fatalf("database error %v", MetaDB.Error)
	}
	log.Info("sqlite.open success")
	MetaDB.Use(gormopentracing.New())

	err = initTables()

	if err != nil {
		log.Fatalf("sqlite create table failed: %v", err)
	}
}

func initTables() error {
	err := MetaDB.Migrator().CreateTable(
		&models.Tenant{},
		&models.Account{},
		&models.Role{},
		&models.Permission{},
		&models.PermissionBinding{},
		&models.RoleBinding{},
		)
	return err
}