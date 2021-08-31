package userapi

type UserIdentity struct {
	UserName string `json:"userName"`
	Token    string `json:"token"`
}
