package tenant

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
)

type Manager struct{}

func NewTenantManager() *Manager {
	return &Manager{}
}
// CreateTenant 创建租户
func (p *Manager) CreateTenant(ctx context.Context, request message.CreateTenantReq) (*message.CreateTenantResp, error) {
	req := message.FindTenantByNameReq{Name: request.Name}
	tenant, err := p.FindTenantByName(ctx, req)

	if err == nil && tenant != nil {

		return &message.CreateTenantResp{Tenant: tenant.Tenant}, fmt.Errorf("tenant already exist")
	}
	ten := structs.Tenant{
		Name: request.Name,
		Type: constants.InstanceWorkspace,
		Status: constants.Valid,
	}
	resp := &message.CreateTenantResp{Tenant: ten}

	return resp, nil
}

func (p *Manager) FindTenantByName(ctx context.Context, req message.FindTenantByNameReq) (resp *message.FindTenantByNameResp, err error) {
	tenant, err := models.GetTenantReaderWriter().FindTenantByName(ctx, req.Name)
	resp.Name = tenant.Name
	resp.ID = tenant.ID
	resp.Type = constants.TenantType(tenant.Type)
	resp.Status = constants.CommonStatus(tenant.Status)
	return
}

func (p *Manager) FindTenantById(ctx context.Context, req message.FindTenantByIdReq) (resp *message.FindTenantByIdResp, err error) {
	tenant, err := models.GetTenantReaderWriter().FindTenantById(ctx, req.ID)
	//if err != nil {
	//	framework.LogWithContext(ctx).Errorf(
	//		"find tenant by ID: %s error: %s", req.ID, err.Error())
	//	return
	//}
	resp.Name = tenant.Name
	resp.ID = tenant.ID
	resp.Type = constants.TenantType(tenant.Type)
	resp.Status = constants.CommonStatus(tenant.Status)

	return
}
