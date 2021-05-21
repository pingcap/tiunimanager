package models

import (
	"context"
	cryrand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/pingcap/tcp/addon/logger"

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
	Name string `gorm:"size:255"`
	Salt string `gorm:"size:512"`
	// hash(salt + pwd)
	FinalHash string `gorm:"size:512"`
}

func (u User) TableName() string {
	return "user"
}

func genSalt(ctx context.Context) (string, error) {
	l := logger.WithContext(ctx).WithField("func", "genSalt")
	b := make([]byte, 128)
	_, err := cryrand.Read(b)
	if err == nil {
		salt := base64.URLEncoding.EncodeToString(b)
		l.Debug("salt:", salt)
		return salt, nil
	} else {
		l.Errorf("met an error:%s", err)
		return "", err
	}
}

func genFinalSalt(ctx context.Context, salt, passwd string) (string, error) {
	s := salt + passwd
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	l := logger.WithContext(ctx).WithField("func", "genFinalSalt")
	l.Debug("salt:", salt, "passwd:", passwd)
	if err == nil {
		l.Debug("finalSalt:", string(finalSalt))
		return string(finalSalt), nil
	} else {
		l.Errorf("met an error:%s", err)
		return "", err
	}
}

func CheckUser(ctx context.Context, name, passwd string) error {
	log := logger.WithContext(ctx).WithField("models", "CheckUser").WithField("name", name)
	u, err := FindUserByName(ctx, name)
	if err == nil {
		log.Info("user:", u)
		if nil == bcrypt.CompareHashAndPassword([]byte(u.FinalHash), []byte(u.Salt+passwd)) {
			log.Debug("check success")
			return nil
		} else {
			return fmt.Errorf("failed")
		}
	} else {
		return err
	}
}

func FindUserByName(ctx context.Context, name string) (user User, err error) {
	log := logger.WithContext(ctx).WithField("models", "FindUserByName").WithField("name", name)
	log.Debug("entry")
	err = db.Where("name = ?", name).First(&user).Error
	if err != nil {
		log.Errorf("err: %s", err)
	} else {
		u := user
		u.Salt = "******"
		u.FinalHash = "******"
		log.Infof("sucess with return: %v", u)
	}
	return
}

func CreateUser(ctx context.Context, name, passwd string) error {
	user := User{
		Name: name,
	}
	u := user
	u.Salt = "******"
	u.FinalHash = "******"
	log := logger.WithContext(ctx).WithField("models", "CreateUser").WithField("user", u)
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
			t0 := time.Now()
			user.Salt, err = genSalt(ctx)
			if err != nil {
				log.Errorf("genSalt err: %s", err)
				tx.Rollback()
				return err
			}
			t1 := time.Now()
			user.FinalHash, err = genFinalSalt(ctx, user.Salt, passwd)
			t2 := time.Now()
			if err != nil {
				log.Errorf("genFinalSalt err: %s", err)
				tx.Rollback()
				return err
			}
			log.Infof("genSalt time cost:%v genFinalSalt time cost:%v", t1.Sub(t0), t2.Sub(t1))
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
