package models

import (
	"context"

	"github.com/pingcap/tcp/addon/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var db *gorm.DB

func init() {
	var err error
	dbFile := "tcp.sqlite.db"
	log := logger.WithContext(nil).WithField("dbFile", dbFile)
	log.Debug("init: sqlite.open")
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	if err != nil {
		log.Fatalf("sqlite open error %v", err)
	}

	if db.Error != nil {
		log.Fatalf("database error %v", db.Error)
	}
	log.Info("sqlite.open success")
	db.Use(gormopentracing.New())
	if db.Migrator().HasTable(&User{}) {
	} else {
		err = db.Migrator().CreateTable(&User{})
		if err != nil {
			log.Fatalf("sqlite create table failed: %v", err)
		}
		err = CreateUser(context.Background(), "admin", "admin")
		if err != nil {
			log.Fatalf("sqlite create admin user failed: %v", err)
		}
	}
}
