package models

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", "root:toor@/tcp?charset=utf8&parseTime=True&loc=Local")

	if err != nil {
		fmt.Printf("mysql connect error %v", err)
		panic(err)
	}

	if db.Error != nil {
		fmt.Printf("database error %v", db.Error)
		panic(db.Error)
	}

	if db.HasTable(&User{}) == false {
		db.CreateTable(&User{})
	}

}
