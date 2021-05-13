package models

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var db *gorm.DB

func init() {
	var err error
	dsn := "root:toor@/tcp?charset=utf8&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Printf("mysql connect error %v", err)
		panic(err)
	}

	if db.Error != nil {
		fmt.Printf("database error %v", db.Error)
		panic(db.Error)
	}
	db.Use(gormopentracing.New())
}
