package models

import (
	"tcp/addon/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var db *gorm.DB

func init() {
	var err error
	dsn := "root:toor@/tcp?charset=utf8&parseTime=True&loc=Local"
	log := logger.WithContext(nil).WithField("dsn", dsn)
	log.Debug("init: mysql.open")
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("mysql connect error %v", err)
	}

	if db.Error != nil {
		log.Fatalf("database error %v", db.Error)
	}
	log.Info("mysql.open success")
	db.Use(gormopentracing.New())
}
