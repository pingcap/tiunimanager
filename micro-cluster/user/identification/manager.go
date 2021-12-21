package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/commons"
	"github.com/pingcap-inc/tiem/models"
	"time"
)

type Manager struct{}

func NewIdentificationManager() *Manager {
	return &Manager{}
}

func (p *Manager) Login(ctx context.Context, request message.LoginReq) (resp message.LoginResp, err error) {
	a, err := models.GetAccountReaderWriter().FindAccountByName(ctx, request.UserName)

	if err != nil {
		return
	}
	account := structs.Account{
		ID: a.ID,
		TenantID: a.TenantId,
		Name: a.Name,
		Salt: a.Salt,
		FinalHash: a.FinalHash,
		//Status: a.Status
	}
	loginSuccess, err := account.CheckPassword(request.Password)
	if err != nil {
		return
	}

	if !loginSuccess {
		err = &structs.UnauthorizedError{}
		return
	}

	req := message.CreateTokenReq{
		AccountID: account.ID,
		AccountName: account.Name,
		TenantID: account.TenantID,
	}
	token, err := p.CreateToken(ctx, req)

	if err != nil {
		return
	} else {
		resp.TokenString = token.TokenString
	}

	return
}

// Logout
func (p *Manager) Logout(ctx context.Context, req message.LogoutReq) (message.LogoutResp, error) {
	token, err := GetToken(ctx, message.GetTokenReq{TokenString: req.TokenString})

	if err != nil {
		return message.LogoutResp{AccountName: ""}, &structs.UnauthorizedError{}
	} else if !token.IsValid() {
		return message.LogoutResp{AccountName: ""}, nil
	} else {
		accountName := token.AccountName
		token.Destroy()

		err := Modify(ctx, message.ModifyTokenReq{Token: &token.Token})
		if err != nil {
			return message.LogoutResp{AccountName: ""}, err
		}

		return message.LogoutResp{AccountName: accountName}, nil
	}
}

var SkipAuth = true

// Accessible
func (p *Manager) Accessible(ctx context.Context, request message.AccessibleReq) (resp message.AccessibleResp, err error) {
	if request.Path == "" {
		err = framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "path empty")
		return
	}

	req := message.GetTokenReq{TokenString: request.TokenString}
	token, err := GetToken(ctx, req)

	if err != nil {
		return
	}

	resp.AccountID = token.AccountID
	resp.AccountName = token.AccountName
	resp.TenantID = token.TenantID


	if SkipAuth {
		// todo checkAuth switch
		return
	}

	if !token.IsValid() {
		err = &structs.UnauthorizedError{}
		return
	}

	return
}

func (p *Manager) CreateToken(ctx context.Context, request message.CreateTokenReq) (message.CreateTokenResp, error) {
	token := structs.Token{
		AccountName: request.AccountName,
		AccountID: request.AccountID,
		TenantID: request.TenantID,
		ExpirationTime: time.Now().Add(commons.DefaultTokenValidPeriod),
	}

	req := message.ProvideTokenReq{Token: &token}
	tokenString, err := Provide(ctx, req)
	token.TokenString = tokenString.TokenString
	return message.CreateTokenResp{Token: token}, err
}