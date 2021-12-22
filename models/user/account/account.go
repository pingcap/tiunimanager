package account

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap/errors"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	common.Entity

	Name      string `gorm:"default:null;not null"`
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