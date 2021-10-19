package identification

type UserIdentity struct {
	UserName string `json:"userName"`
	TenantId string `json:"tenantId"`
	Token    string `json:"token"`
}
