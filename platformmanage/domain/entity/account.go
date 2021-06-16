package entity

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"

	"github.com/pingcap/ticp/platformmanage/domain/port"
)

type Account struct {
	Tenant 		*Tenant
	Id          int
	Name 		string
	Salt 		string
	FinalHash 	string
	Status 		CommonStatus
}

func (account *Account) genSaltAndHash(passwd string) error {
	b := make([]byte, 128)
	_, err := cryrand.Read(b)

	if err != nil {
		return err
	}

	salt := base64.URLEncoding.EncodeToString(b)

	finalSalt, err := finalHash(salt, passwd)

	if err == nil {
		account.Salt = salt
		account.FinalHash = string(finalSalt)
	}

	return nil
}

func  (account *Account) CheckPassword(passwd string) (bool, error) {
	finalSalt, err := finalHash(account.Salt, passwd)
	if err != nil {
		return false, err
	}

	return string(finalSalt) == account.FinalHash, err
}

func finalHash(salt string, passwd string) ([]byte, error) {
	s := salt + passwd
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)

	return finalSalt, err
}

func (account *Account) persist() error{
	port.RbacRepo.AddAccount(account)
	return nil
}

// CreateAccount 创建账号
func CreateAccount(tenant *Tenant, name, passwd string) (*Account, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := FindAccountByName(tenant, name)

	if e != nil {
		return nil, e
	} else if !(nil == existed) {
		return nil, fmt.Errorf("account already exist")
	}

	account := Account{Tenant: tenant, Name: name, Status: Valid}

	account.genSaltAndHash(passwd)
	account.persist()

	return &account, nil
}

// FindAccountByName 根据名称获取账号
func FindAccountByName(tenant *Tenant, name string) (*Account, error) {
	a,e := port.RbacRepo.FetchAccount(tenant.Id, name)
	return &a, e
}

// AssignRoles 给账号赋予角色
func (account *Account) AssignRoles(roles []Role) error {
	bindings := make([]RoleBinding, len(roles), len(roles))

	for index,r := range roles {
		bindings[index] = RoleBinding{Account: account, Role: &r, Status: Valid}
	}
	return port.RbacRepo.AddRoleBindings(bindings)
}

// ListAllRoles 获取账号的所有角色
func (account *Account) ListAllRoles() ([]Role, error){
	return port.RbacRepo.FetchAllRolesByAccount(account)
}

// CheckAuth 校验权限
func CheckAuth(account *Account, permission *Permission) (bool, error){
	accountRoles,err := account.ListAllRoles()
	if err != nil {
		return false, err
	}
	if accountRoles == nil || len(accountRoles) == 0 {
		return false, nil
	}

	accountRoleMap := make(map[int]bool)
	
	for _,r := range accountRoles {
		accountRoleMap[r.Id] = true
	}
	
	allowedRoles,err := permission.ListAllRoles()
	if err != nil {
		return false, err
	}
	if allowedRoles == nil || len(allowedRoles) == 0 {
		return false, nil
	}

	for _,r := range allowedRoles {
		if _,exist := accountRoleMap[r.Id]; exist  {
			return true, nil
		}
	}

	return false, nil
}