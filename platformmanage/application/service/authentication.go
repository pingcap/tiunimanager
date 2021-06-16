package service

// Login 登录
func Login(tenantId int, userName, password string) (string, error) {
	// 校验密码
	// 生成并记录token
	return "", nil
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
