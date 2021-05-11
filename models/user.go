package models

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

/*
create table user (
   	id int auto_increment,
   	name varchar(255),
   	password varchar(64),
   	primary key (id)
);
*/

type User struct {
	gorm.Model
	Name     string `gorm:"size:255"`
	Password string `gorm:"size:64"`
}

func (u User) TableName() string {
	return "user"
}

func FindUserByName(name string) (user User, err error) {
	err = db.Where("name = ?", name).First(&user).Error
	return
}

func CreateUser(user User) error {
	tx := db.Begin()
	var findUser User
	tmp := db.Where("name = ?", user.Name).First(&findUser)
	err := tmp.Error
	notFound := tmp.RecordNotFound()
	tmp = nil
	if err == nil { //already exist
		tx.Rollback()
		return fmt.Errorf("record already exist")
	} else {
		if notFound {
			err = db.Create(&user).Error
			if err == nil {
				return tx.Commit().Error
			} else {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return err
		}
	}
}
