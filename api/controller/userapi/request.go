package userapi

type LoginInfo struct {
	UserName 		string 	`json:"userName"`
	UserPassword 	string 	`json:"userPassword"`
}

type LogoutInfo struct {
	UserName 		string 	`json:"userName"`
}
