package domain

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"fmt"
	port2 "github.com/pingcap/ticp/micro-manager/service/tenant/port"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Tenant    *Tenant
	Id        uint
	TenantId  uint
	Name      string
	Salt      string
	FinalHash string
	Status    CommonStatus
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

func  (account *Account) checkPassword(passwd string) (bool, error) {
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
	port2.RbacRepo.AddAccount(account)
	return nil
}

// createAccount 创建账号
func createAccount(tenant *Tenant, name, passwd string) (*Account, error) {
	if tenant == nil || !tenant.Status.IsValid(){
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := findAccountByName(name)

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

// findAccountByName 根据名称获取账号
func findAccountByName(name string) (*Account, error) {
	a,err := port2.RbacRepo.FetchAccountByName(name)
	if err != nil {
		return nil, err
	}

	FindTenantById(a.Id)
	return &a, err
}

// assignRoles 给账号赋予角色
func (account *Account) assignRoles(roles []Role) error {
	bindings := make([]RoleBinding, len(roles), len(roles))

	for index,r := range roles {
		bindings[index] = RoleBinding{Account: account, Role: &r, Status: Valid}
	}
	return port2.RbacRepo.AddRoleBindings(bindings)
}

// listAllRoles 获取账号的所有角色
func (account *Account) listAllRoles() ([]Role, error){
	return port2.RbacRepo.FetchAllRolesByAccount(account)
}

// checkAuth 校验权限
func checkAuth(account *Account, permission *Permission) (bool, error){
	accountRoles,err := account.listAllRoles()
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
	
	allowedRoles,err := permission.listAllRoles()
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