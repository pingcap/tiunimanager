package tenant

// Login 登录
func Login(userName, password string) (string, error) {
	//account, err := entity.FindAccountByName(userName)
	//
	//if err != nil {
	//	return "", err
	//}
	//
	//loginSuccess, err := account.CheckPassword(password)
	//if err != nil {
	//	return "", err
	//}
	//tokenString, err := port.TokenMNG.Provide(account.Id, account.TenantId, time.Now().Add(commons.DefaultTokenValidPeriod))
	//
	//return tokenString, err
	return "",nil
}

// Logout 退出登录
func Logout(tenantId int, userName string) (string, error) {
	// 清理token
	return "", nil
}

// Accessible 路径鉴权
func Accessible(tenantId int, token string, path string) (bool, error) {
	// 校验token有效
	// 根据token查用户
	// 校验账号权限
	return true, nil
}

// Operable 操作鉴权
func Operable(tenantId int, token string, path string) (bool, error) {
	// 校验token有效
	// 根据token查用户
	// 校验账号权限
	return true, nil
}
