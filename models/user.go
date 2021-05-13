package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
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
	err := db.Where("name = ?", user.Name).First(&findUser).Error
	notFound := errors.Is(err, gorm.ErrRecordNotFound)
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
