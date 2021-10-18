package domain

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Id        string
	TenantId  string
	Name      string
	Salt      string
	FinalHash string
	Status    CommonStatus
}

type AccountAggregation struct {
	Account
	Roles []Role
}

func (account *Account) GenSaltAndHash(passwd string) error {
	b := make([]byte, 16)
	_, err := cryrand.Read(b)

	if err != nil {
		return err
	}

	salt := base64.URLEncoding.EncodeToString(b)

	finalSalt, err := finalHash(salt, passwd)

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

func finalHash(salt string, passwd string) ([]byte, error) {
	if passwd == "" {
		return nil, errors.New("password cannot be empty")
	}
	s := salt + passwd
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)

	return finalSalt, err
}
