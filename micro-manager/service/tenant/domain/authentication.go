package domain

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
	} else {
		accountName := token.AccountName
		err := token.destroy()
		
		if err != nil {
			return "", err
		}
		
		return accountName, nil
	}
}

// Accessible 路径鉴权
func Accessible(pathType string, path string, tokenString string) (tenantId uint, accountName string, err error) {
	token, err := TokenMNG.GetToken(tokenString)
	
	if err != nil {
		return
	}

	accountName = token.AccountName
	tenantId = token.TenantId

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
		err = &UnauthorizedError{}
	}

	return
}

type UnauthorizedError struct {}
func (*UnauthorizedError) Error() string{
	return "认证失败"
}

type ForbiddenError struct {}
func (*ForbiddenError) Error() string{
	return "无访问权限"
}

