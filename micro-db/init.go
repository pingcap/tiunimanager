package main

import (
	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/pingcap/ticp/addon/logger"
	mylogger "github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/config"
	"github.com/pingcap/ticp/micro-db/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var DB *gorm.DB

func initLogger() {
	// log
	mlog.DefaultLogger = mlogrus.NewLogger(mlogrus.WithLogger(mylogger.WithContext(nil)))
}

func InitSqliteDB() {
	var err error
	dbFile := config.GetSqliteFilePath()
	log := logger.WithContext(nil).WithField("dbFile", dbFile)
	log.Debug("init: sqlite.open")
	DB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	if err != nil {
		log.Fatalf("sqlite open error %v", err)
	}

	if DB.Error != nil {
		log.Fatalf("database error %v", DB.Error)
	}
	log.Info("sqlite.open success")
	DB.Use(gormopentracing.New())

	err = initTables()

	if err != nil {
		log.Fatalf("sqlite create table failed: %v", err)
	}
}

func initTables() error {
	err := DB.Migrator().CreateTable(
		&models.Tenant{},
		&models.Account{},
		&models.Role{},
		&models.Permission{},
		&models.PermissionBinding{},
		&models.RoleBinding{},
		)
	return err
}