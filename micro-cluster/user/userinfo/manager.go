package userinfo

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/user/account"
	"github.com/pingcap-inc/tiem/models/user/tenant"
	"strconv"
)

type Manager struct {}

func NewAccountManager() *Manager {
	return &Manager{}
}

func (p *Manager) CreateAccount(ctx context.Context, tenant *tenant.Tenant, name, passwd string) (*account.Account, error) {
	if tenant == nil || !tenant.Status.IsValid() {
		return nil, fmt.Errorf("tenant not valid")
	}

	existed, e := p.FindAccountByName(ctx, name)

	if e == nil && existed != nil {
		return existed, fmt.Errorf("account already exist")
	}
	account := account.Account{Name: name, Entity: common.Entity{Status: strconv.Itoa(int(constants.Valid))}}
	account.GenSaltAndHash(passwd)
	a, _ := models.GetAccountReaderWriter().AddAccount(ctx, tenant.ID, name, account.Salt, account.FinalHash, int8(constants.Valid))
	return a, nil
}

// CreateAccount CreateAccount
//func (p *Manager) CreateAccount(ctx context.Context, req message.CreateAccountReq) (*message.CreateAccountResp, error) {
//	if req.Tenant == nil || req.Tenant.Status != 0 {
//		return nil, fmt.Errorf("tenant not valid")
//	}
//	//findReq := message.FindAccountByNameReq{Name: req.Name}
//	//findResp, e := p.FindAccountByName(ctx, findReq)
//	a, e := models.GetAccountReaderWriter().FindAccountByName(ctx, req.Name)
//	if e == nil && a != nil {
//		return &message.CreateAccountResp{Account: *a}, fmt.Errorf("account already exist")
//	}
//
//	//account := domain.Account{Name: name, Status: domain.Valid}
//	account := account.Account{Name: req.Name, Entity: common.Entity{Status: strconv.Itoa(int(constants.Valid))}}
//	account.GenSaltAndHash(req.Password)
//	a, _ = models.GetAccountReaderWriter().AddAccount(ctx, req.Tenant.ID, req.Name, account.Salt, account.FinalHash, int8(constants.Valid))
//
//	return &message.CreateAccountResp{*a}, nil
//}

func (p *Manager) FindAccountByName(ctx context.Context, name string) (*account.Account, error) {
	a, err := models.GetAccountReaderWriter().FindAccountByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return a, err
}

//FindAccountByName FindAccountByName
//func (p *Manager) FindAccountByName(ctx context.Context, name string) (*message.FindAccountByNameResp, error) {
//	a, err := models.GetAccountReaderWriter().FindAccountByName(ctx, name)
//	if err != nil {
//		return nil, err
//	}
//	account := structs.Account{
//		ID: a.ID,
//		TenantID: a.TenantId,
//		Name: a.Name,
//		Salt: a.Salt,
//		FinalHash: a.FinalHash,
//		//Status: a.Status,
//	}
//	resp := &message.FindAccountByNameResp{Account: account}
//	return resp, err
//}

