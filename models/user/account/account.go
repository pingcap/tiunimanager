package account

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        string    `gorm:"primarykey"`
	TenantID  string    `gorm:"primarykey"`
	Creator   string    `gorm:"default:null;not null;;<-:create"`
	Name      string    `gorm:"default:null;not null;uniqueIndex"`
	Salt      string    `gorm:"default:null;not null;<-:create"` //password
	FinalHash string    `gorm:"default:null;not null"`
	Email     string    `gorm:"default:null"`
	Phone     string    `gorm:"default:null"`
	Status    string    `gorm:"not null;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type Account struct {
	common.Entity

	Name      string `gorm:"default:null;not null;uniqueIndex"`
	Salt      string `gorm:"default:null;not null;<-:create"`
	FinalHash string `gorm:"default:null;not null"`
}

func (account *Account) GenSaltAndHash(passwd string) error {
	b := make([]byte, 16)
	_, err := cryrand.Read(b)

	if err != nil {
		return err
	}

	salt := base64.URLEncoding.EncodeToString(b)

	finalSalt, err := common.FinalHash(salt, passwd)

	if err != nil {
		return err
	}

	account.Salt = salt
	account.FinalHash = string(finalSalt)

	return nil
}

func (account *Account) CheckPassword(passwd string) (bool, error) {
	if passwd == "" {
		return false, errors.New("password cannot be empty")
	}
	if len(passwd) > 20 {
		return false, errors.New("password is too long")
	}
	s := account.Salt + passwd

	err := bcrypt.CompareHashAndPassword([]byte(account.FinalHash), []byte(s))

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
