/*
 * Copyright (c)  2022 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*******************************************************************************
 * @File: manager.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/01/06
*******************************************************************************/

package tenant

import (
	"context"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/user/account"
)

type Manager struct{}

func NewTenantManager() *Manager {
	return &Manager{}
}
func (p *Manager) CreateTenant(ctx context.Context, request message.CreateTenantReqV1) (resp message.CreateTenantRespV1, er error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()

	tenant, err := rw.GetTenant(ctx, request.ID)
	if err == nil && tenant.ID == request.ID {
		return resp, errors.NewEMErrorf(errors.TenantAlreadyExist, "tenantID: %s", request.ID)
	} else if err.(errors.EMError).GetCode() == errors.TenantNotExist {
		_, err = rw.CreateTenant(ctx, account.Tenant{ID: request.ID, Name: request.Name, Status: request.Status,
			OnBoardingStatus: request.OnBoardingStatus, MaxCluster: request.MaxCluster, MaxCPU: request.MaxCPU,
			MaxMemory: request.MaxMemory, MaxStorage: request.MaxStorage, /* TODO Creator: request.Creator,*/
		})
		if nil != err {
			log.Warningf("create tenant %s error: %v", request.ID, err)
		}
	}

	return resp, err
}

func (p *Manager) DeleteTenant(ctx context.Context, request message.DeleteTenantReq) (resp message.DeleteTenantResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	_, er := rw.GetTenant(ctx, request.ID)
	if er != nil {
		log.Warningf("delete tenant error: %v,tenantID: %s", er, request.ID)
		return resp, errors.NewEMErrorf(errors.TenantNotExist, "tenantID: %s", request.ID)
	}

	er = rw.DeleteTenant(ctx, request.ID)
	if er != nil {
		log.Warningf("delete tenant error: %v,tenantID: %s", er, request.ID)
		return resp, errors.NewEMErrorf(errors.DeleteTenantFailed, "tenantID: %s", request.ID)
	}
	return resp, err
}

func (p *Manager) GetTenant(ctx context.Context, request message.GetTenantReq) (resp message.GetTenantResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.Info, err = rw.GetTenant(ctx, request.ID)
	if err != nil {
		log.Warningf("get  tenant error: %v,tenantID: %s", err, request.ID)
		return resp, errors.NewEMErrorf(errors.TenantNotExist, "tenantID: %s", request.ID)
	}
	return resp, err
}

func (p *Manager) QueryTenants(ctx context.Context, request message.QueryTenantReq) (resp message.QueryTenantResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	resp.Tenants, err = rw.QueryTenants(ctx)
	if err != nil {
		log.Warningf("query all tenant error: %v", err)
		return resp, errors.NewEMErrorf(errors.QueryTenantScanRowError, "error: %v", err)
	}
	return resp, err
}

func (p *Manager) UpdateTenantOnBoardingStatus(ctx context.Context, request message.UpdateTenantOnBoardingStatusReq) (resp message.UpdateTenantOnBoardingStatusResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	_, err = rw.GetTenant(ctx, request.ID)
	if err != nil {
		log.Warningf("update tenant on boarding status error,get tenant error: %v,tenantID: %s", err, request.ID)
		return resp, errors.NewEMErrorf(errors.TenantNotExist, "update tenant on boarding status error,tenantID: %s", request.ID)
	}
	err = rw.UpdateTenantOnBoardingStatus(ctx, request.ID, request.OnBoardingStatus)
	if err != nil {
		log.Warningf("update tenant on boarding status error: %v,tenantID: %s", err, request.ID)
		return resp, errors.NewEMErrorf(errors.UpdateTenantOnBoardingStatusFailed, "error: %v", err)
	}
	return resp, err
}

func (p *Manager) UpdateTenantProfile(ctx context.Context, request message.UpdateTenantProfileReq) (resp message.UpdateTenantProfileResp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetAccountReaderWriter()
	err = rw.UpdateTenantProfile(ctx, request.ID, request.Name, request.MaxCluster, request.MaxCPU, request.MaxMemory, request.MaxStorage)
	if err != nil {
		log.Warningf("update tenant profile error: %v,tenantID: %s,Name:%s, MaxCluster:%d,MaxCPU:%d,MaxMemory:%d,MaxStorage:%d", err, request.ID, request.Name, request.MaxCluster, request.MaxCPU, request.MaxMemory, request.MaxStorage)
		return resp, errors.NewEMErrorf(errors.UpdateTenantOnBoardingStatusFailed, "error: %v,tenantID: %s,Name:%s, MaxCluster:%d,MaxCPU:%d,MaxMemory:%d,MaxStorage:%d", err, request.Name, request.MaxCluster, request.MaxCPU, request.MaxMemory, request.MaxStorage)
	}
	return resp, err
}
