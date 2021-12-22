package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/user/identification"
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

	loginSuccess, err := a.CheckPassword(request.Password)
	if err != nil {
		return
	}

	if !loginSuccess {
		err = framework.NewTiEMError(common.TIEM_UNAUTHORIZED, "unauthorized")
		return
	}

	req := message.CreateTokenReq{
		AccountID: a.ID,
		AccountName: a.Name,
		TenantID: a.TenantId,
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
		return message.LogoutResp{AccountName: ""}, framework.NewTiEMError(common.TIEM_UNAUTHORIZED, "unauthorized")
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

	resp.AccountID = token.AccountId
	resp.AccountName = token.AccountName
	resp.TenantID = token.TenantId


	if SkipAuth {
		// todo checkAuth switch
		return
	}

	if !token.IsValid() {
		err = framework.NewTiEMError(common.TIEM_UNAUTHORIZED, "unauthorized")
		return
	}

	return
}

func (p *Manager) CreateToken(ctx context.Context, request message.CreateTokenReq) (message.CreateTokenResp, error) {
	token := identification.Token{
		AccountName: request.AccountName,
		AccountId: request.AccountID,
		TenantId: request.TenantID,
		ExpirationTime: time.Now().Add(constants.DefaultTokenValidPeriod),
	}

	req := message.ProvideTokenReq{Token: &token}
	tokenString, err := Provide(ctx, req)
	token.TokenString = tokenString.TokenString
	return message.CreateTokenResp{Token: token}, err
}