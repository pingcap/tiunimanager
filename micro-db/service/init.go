package service

import (
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var DB *gorm.DB

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
		&Tenant{},
		&Account{},
		&Role{},
		&Permission{},
		&PermissionBinding{},
		&RoleBinding{},
		)
	return err
}