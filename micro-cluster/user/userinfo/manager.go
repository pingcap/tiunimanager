package userinfo

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
)

type Manager struct {}

func NewAccountManager() *Manager {
	return &Manager{}
}

// CreateAccount CreateAccount
func (p *Manager) CreateAccount(ctx context.Context, req message.CreateAccountReq) (*message.CreateAccountResp, error) {
	if req.Tenant == nil || !req.Tenant.Status.IsValid() {
		return nil, fmt.Errorf("tenant not valid")
	}
	findReq := message.FindAccountByNameReq{Name: req.Name}
	findResp, e := p.FindAccountByName(ctx, findReq)

	if e == nil && findResp != nil {
		resp := &message.CreateAccountResp{Account: findResp.Account}
		return resp, fmt.Errorf("account already exist")
	}

	//account := domain.Account{Name: name, Status: domain.Valid}
	account := structs.Account{Name: req.Name, Status: constants.Valid}
	account.GenSaltAndHash(req.Password)
	a, _ := models.GetAccountReaderWriter().AddAccount(ctx, req.Tenant.ID, req.Name, account.Salt, account.FinalHash, int8(account.Status))
	resp := &message.CreateAccountResp{
		Account: structs.Account{
			TenantID: a.ID,
			Name: a.Name,
			Salt: a.Salt,
			FinalHash: a.FinalHash,
			//Status: a.Status,
		},
	}
	return resp, nil
}

// FindAccountByName FindAccountByName
func (p *Manager) FindAccountByName(ctx context.Context, req message.FindAccountByNameReq) (*message.FindAccountByNameResp, error) {
	a, err := models.GetAccountReaderWriter().FindAccountByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	account := structs.Account{
		ID: a.ID,
		TenantID: a.TenantId,
		Name: a.Name,
		Salt: a.Salt,
		FinalHash: a.FinalHash,
		//Status: a.Status,
	}
	resp := &message.FindAccountByNameResp{Account: account}
	return resp, err
}

