package models

import (
	"context"
	"errors"
	"fmt"
	"tcp/addon/logger"

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

func FindUserByName(ctx context.Context, name string) (user User, err error) {
	log := logger.WithContext(ctx).WithField("models", "FindUserByName").WithField("name", name)
	log.Debug("entry")
	err = db.Where("name = ?", name).First(&user).Error
	if err != nil {
		log.Errorf("err: %s", err)
	} else {
		log.Infof("sucess with return: %v", user)
	}
	return
}

func CreateUser(ctx context.Context, user User) error {
	log := logger.WithContext(ctx).WithField("models", "CreateUser").WithField("user", user)
	log.Debug("entry")
	tx := db.Begin()
	var findUser User
	err := db.Where("name = ?", user.Name).First(&findUser).Error
	notFound := errors.Is(err, gorm.ErrRecordNotFound)
	if err == nil { //already exist
		tx.Rollback()
		err = fmt.Errorf("record already exist")
		log.Errorf("err: %s", err)
		return err
	} else {
		if notFound {
			err = db.Create(&user).Error
			if err == nil {
				err = tx.Commit().Error
				if err != nil {
					log.Errorf("err: %s", err)
				} else {
					log.Debugf("tx Commit success")
				}
				return err
			} else {
				log.Errorf("err: %s", err)
				tx.Rollback()
				return err
			}
		} else {
			log.Errorf("expect to get not found but got err: %s", err)
			tx.Rollback()
			return err
		}
	}
}
