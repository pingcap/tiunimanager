package domain

import "errors"

type AccountAggregation struct {
	Account
	Roles []Role
}

type PermissionAggregation struct {
	Permission
	Roles []Role
}

// Login 登录
func Login(userName, password string) (tokenString string, err error) {
	account, err := findAccountByName(userName)

	if err != nil {
		return
	}

	loginSuccess, err := account.checkPassword(password)
	if err != nil {
		return
	}

	if !loginSuccess {
		err = &UnauthorizedError{}
		return
	}

	token, err := createToken(account.Id, account.Name, account.TenantId)

	if err != nil {
		return
	} else {
		tokenString = token.TokenString
	}

	return
}

// Logout 退出登录
func Logout(tokenString string) (string, error) {
	token,err := TokenMNG.GetToken(tokenString)

	if err != nil {
		return "", &UnauthorizedError{}
	} else if !token.isValid() {
		return "", nil
	} else {
		accountName := token.AccountName
		err := token.destroy()
		
		if err != nil {
			return "", err
		}
		
		return accountName, nil
	}
}

var SkipAuth = true

// Accessible 路径鉴权
func Accessible(pathType string, path string, tokenString string) (tenantId string, accountId, accountName string, err error) {
	if path == "" {
		err = errors.New("path cannot be blank")
		return
	}

	token, err := TokenMNG.GetToken(tokenString)
	
	if err != nil {
		return
	}

	accountId = token.AccountId
	accountName = token.AccountName
	tenantId = token.TenantId

	if (SkipAuth) {
		// todo checkAuth switch
		return
	}

	// 校验token有效
	if !token.isValid() {
		err =  &UnauthorizedError{}
		return
	}

	// 根据token查用户
	account, err := findAccountAggregation(accountName)
	if err != nil {
		return
	}

	// 查权限
	permission, err := findPermissionAggregationByCode(tenantId, path)
	if err != nil {
		return
	}

	ok, err := checkAuth(account, permission)

	if err != nil {
		return
	}

	if !ok {
		err = &ForbiddenError{}
	}

	return
}

// findAccountExtendInfo 根据名称获取账号及扩展信息
func findAccountAggregation(name string) (*AccountAggregation, error) {
	a,err := RbacRepo.LoadAccountAggregation(name)
	if err != nil {
		return nil, err
	}

	return &a, err
}

func findPermissionAggregationByCode(tenantId string, code string) (*PermissionAggregation, error) {
	a,e := RbacRepo.LoadPermissionAggregation(tenantId, code)
	return &a, e
}

// checkAuth 校验权限
func checkAuth(account *AccountAggregation, permission *PermissionAggregation) (bool, error){

	accountRoles := account.Roles

	if accountRoles == nil || len(accountRoles) == 0 {
		return false, nil
	}

	accountRoleMap := make(map[string]bool)

	for _,r := range accountRoles {
		accountRoleMap[r.Id] = true
	}

	allowedRoles := permission.Roles

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

type UnauthorizedError struct {}
func (*UnauthorizedError) Error() string{
	return "Unauthorized"
}

type ForbiddenError struct {}
func (*ForbiddenError) Error() string{
	return "Access Forbidden"
}

