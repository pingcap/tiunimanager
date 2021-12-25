package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
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
		err = errors.WrapError(errors.TIEM_UNAUTHORIZED_USER, "unauthorized", err)
		return
	}

	if !loginSuccess {
		err = errors.NewError(errors.TIEM_UNAUTHORIZED_USER, "unauthorized")
		return
	}

	req := message.CreateTokenReq{
		AccountID: a.ID,
		AccountName: a.Name,
		TenantID: a.TenantId,
	}
	token, err := p.CreateToken(ctx, req)

	if err != nil {
		err = errors.WrapError(errors.TIEM_UNAUTHORIZED_USER, "unauthorized", err)
		return
	} else {
		resp.TokenString = token.TokenString
		resp.UserName = token.AccountName
		resp.TenantId = token.TenantId
	}

	return
}

// Logout
func (p *Manager) Logout(ctx context.Context, req message.LogoutReq) (message.LogoutResp, error) {
	r := message.GetTokenReq{TokenString: req.TokenString}
	token, err := GetToken(ctx, r)

	if err != nil {

		return message.LogoutResp{AccountName: ""}, errors.NewError(errors.TIEM_UNAUTHORIZED_USER, "unauthorized")
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

	req := message.GetTokenReq{TokenString: request.TokenString}
	token, err := GetToken(ctx, req)

	if err != nil {
		return
	}

	resp.AccountID = token.AccountId
	resp.AccountName = token.AccountName
	resp.TenantID = token.TenantId

	if !token.IsValid() {
		err = errors.NewError(errors.TIEM_UNAUTHORIZED_USER, "invalid token")
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