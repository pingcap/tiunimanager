
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package main

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/pingcap-inc/tiem/library/client"

	micro "github.com/asim/go-micro/v3/client"
)

type ClusterFakeService struct {
	// Auth manager module
	mockLogin          func(ctx context.Context, in *clusterpb.LoginRequest, opts ...micro.CallOption) (*clusterpb.LoginResponse, error)
	mockLogout         func(ctx context.Context, in *clusterpb.LogoutRequest, opts ...micro.CallOption) (*clusterpb.LogoutResponse, error)
	mockVerifyIdentity func(ctx context.Context, in *clusterpb.VerifyIdentityRequest, opts ...micro.CallOption) (*clusterpb.VerifyIdentityResponse, error)
	// host manager module
	mockImportHost            func(ctx context.Context, in *clusterpb.ImportHostRequest, opts ...micro.CallOption) (*clusterpb.ImportHostResponse, error)
	mockImportHostsInBatch    func(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.ImportHostsInBatchResponse, error)
	mockRemoveHost            func(ctx context.Context, in *clusterpb.RemoveHostRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostResponse, error)
	mockRemoveHostsInBatch    func(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostsInBatchResponse, error)
	mockListHost              func(ctx context.Context, in *clusterpb.ListHostsRequest, opts ...micro.CallOption) (*clusterpb.ListHostsResponse, error)
	mockCheckDetails          func(ctx context.Context, in *clusterpb.CheckDetailsRequest, opts ...micro.CallOption) (*clusterpb.CheckDetailsResponse, error)
	mockAllocHosts            func(ctx context.Context, in *clusterpb.AllocHostsRequest, opts ...micro.CallOption) (*clusterpb.AllocHostResponse, error)
	mockGetFailureDomain      func(ctx context.Context, in *clusterpb.GetFailureDomainRequest, opts ...micro.CallOption) (*clusterpb.GetFailureDomainResponse, error)
	mockAllocResourcesInBatch func(ctx context.Context, in *clusterpb.BatchAllocRequest, opts ...micro.CallOption) (*clusterpb.BatchAllocResponse, error)
	mockRecycleResources      func(ctx context.Context, in *clusterpb.RecycleRequest, opts ...micro.CallOption) (*clusterpb.RecycleResponse, error)
}

func (s *ClusterFakeService) ListFlows(ctx context.Context, in *clusterpb.ListFlowsRequest, opts ...micro.CallOption) (*clusterpb.ListFlowsResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) Login(ctx context.Context, in *clusterpb.LoginRequest, opts ...micro.CallOption) (*clusterpb.LoginResponse, error) {
	return s.mockLogin(ctx, in, opts...)
}

func (s *ClusterFakeService) Logout(ctx context.Context, in *clusterpb.LogoutRequest, opts ...micro.CallOption) (*clusterpb.LogoutResponse, error) {
	return s.mockLogout(ctx, in, opts...)
}

func (s *ClusterFakeService) VerifyIdentity(ctx context.Context, in *clusterpb.VerifyIdentityRequest, opts ...micro.CallOption) (*clusterpb.VerifyIdentityResponse, error) {
	return s.mockVerifyIdentity(ctx, in, opts...)
}

func (s *ClusterFakeService) ImportHost(ctx context.Context, in *clusterpb.ImportHostRequest, opts ...micro.CallOption) (*clusterpb.ImportHostResponse, error) {
	return s.mockImportHost(ctx, in, opts...)
}
func (s *ClusterFakeService) ImportHostsInBatch(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.ImportHostsInBatchResponse, error) {
	return s.mockImportHostsInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) RemoveHost(ctx context.Context, in *clusterpb.RemoveHostRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostResponse, error) {
	return s.mockRemoveHost(ctx, in, opts...)
}
func (s *ClusterFakeService) RemoveHostsInBatch(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostsInBatchResponse, error) {
	return s.mockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) ListHost(ctx context.Context, in *clusterpb.ListHostsRequest, opts ...micro.CallOption) (*clusterpb.ListHostsResponse, error) {
	return s.mockListHost(ctx, in, opts...)
}
func (s *ClusterFakeService) CheckDetails(ctx context.Context, in *clusterpb.CheckDetailsRequest, opts ...micro.CallOption) (*clusterpb.CheckDetailsResponse, error) {
	return s.mockCheckDetails(ctx, in, opts...)
}
func (s *ClusterFakeService) AllocHosts(ctx context.Context, in *clusterpb.AllocHostsRequest, opts ...micro.CallOption) (*clusterpb.AllocHostResponse, error) {
	return s.mockAllocHosts(ctx, in, opts...)
}
func (s *ClusterFakeService) GetFailureDomain(ctx context.Context, in *clusterpb.GetFailureDomainRequest, opts ...micro.CallOption) (*clusterpb.GetFailureDomainResponse, error) {
	return s.mockGetFailureDomain(ctx, in, opts...)
}
func (s *ClusterFakeService) AllocResourcesInBatch(ctx context.Context, in *clusterpb.BatchAllocRequest, opts ...micro.CallOption) (*clusterpb.BatchAllocResponse, error) {
	return s.mockAllocResourcesInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) RecycleResources(ctx context.Context, in *clusterpb.RecycleRequest, opts ...micro.CallOption) (*clusterpb.RecycleResponse, error) {
	return s.mockRecycleResources(ctx, in, opts...)
}

func InitFakeClusterClient() *ClusterFakeService {

	fakeClient := &ClusterFakeService{
		mockLogin: func(ctx context.Context, in *clusterpb.LoginRequest, opts ...micro.CallOption) (*clusterpb.LoginResponse, error) {
			return nil, nil
		},
		mockLogout: func(ctx context.Context, in *clusterpb.LogoutRequest, opts ...micro.CallOption) (*clusterpb.LogoutResponse, error) {
			return nil, nil
		},
		mockVerifyIdentity: func(ctx context.Context, in *clusterpb.VerifyIdentityRequest, opts ...micro.CallOption) (*clusterpb.VerifyIdentityResponse, error) {
			rsp := new(clusterpb.VerifyIdentityResponse)
			rsp.Status = new(clusterpb.ManagerResponseStatus)
			rsp.AccountId = "fake-accountID"
			rsp.TenantId = "fake-tenantID"
			rsp.AccountName = "fake-accountName"
			rsp.Status.Code = 0
			return rsp, nil
		},

		mockImportHost: func(ctx context.Context, in *clusterpb.ImportHostRequest, opts ...micro.CallOption) (*clusterpb.ImportHostResponse, error) {
			return nil, nil
		},
		mockImportHostsInBatch: func(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.ImportHostsInBatchResponse, error) {
			return nil, nil
		},
		mockRemoveHost: func(ctx context.Context, in *clusterpb.RemoveHostRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostResponse, error) {
			return nil, nil
		},
		mockRemoveHostsInBatch: func(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		mockListHost: func(ctx context.Context, in *clusterpb.ListHostsRequest, opts ...micro.CallOption) (*clusterpb.ListHostsResponse, error) {
			return nil, nil
		},
		mockCheckDetails: func(ctx context.Context, in *clusterpb.CheckDetailsRequest, opts ...micro.CallOption) (*clusterpb.CheckDetailsResponse, error) {
			return nil, nil
		},
		mockAllocHosts: func(ctx context.Context, in *clusterpb.AllocHostsRequest, opts ...micro.CallOption) (*clusterpb.AllocHostResponse, error) {
			return nil, nil
		},
		mockGetFailureDomain: func(ctx context.Context, in *clusterpb.GetFailureDomainRequest, opts ...micro.CallOption) (*clusterpb.GetFailureDomainResponse, error) {
			return nil, nil
		},
	}
	client.ClusterClient = fakeClient
	return fakeClient
}

func (s *ClusterFakeService) MockImportHost(mock func(ctx context.Context, in *clusterpb.ImportHostRequest, opts ...micro.CallOption) (*clusterpb.ImportHostResponse, error)) {
	s.mockImportHost = mock
}
func (s *ClusterFakeService) MockImportHostsInBatch(mock func(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.ImportHostsInBatchResponse, error)) {
	s.mockImportHostsInBatch = mock
}
func (s *ClusterFakeService) MockRemoveHost(mock func(ctx context.Context, in *clusterpb.RemoveHostRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostResponse, error)) {
	s.mockRemoveHost = mock
}
func (s *ClusterFakeService) MockRemoveHostsInBatch(mock func(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*clusterpb.RemoveHostsInBatchResponse, error)) {
	s.mockRemoveHostsInBatch = mock
}
func (s *ClusterFakeService) MockListHost(mock func(ctx context.Context, in *clusterpb.ListHostsRequest, opts ...micro.CallOption) (*clusterpb.ListHostsResponse, error)) {
	s.mockListHost = mock
}
func (s *ClusterFakeService) MockCheckDetails(mock func(ctx context.Context, in *clusterpb.CheckDetailsRequest, opts ...micro.CallOption) (*clusterpb.CheckDetailsResponse, error)) {
	s.mockCheckDetails = mock
}
func (s *ClusterFakeService) MockAllocHosts(mock func(ctx context.Context, in *clusterpb.AllocHostsRequest, opts ...micro.CallOption) (*clusterpb.AllocHostResponse, error)) {
	s.mockAllocHosts = mock
}
func (s *ClusterFakeService) MockGetFailureDomain(mock func(ctx context.Context, in *clusterpb.GetFailureDomainRequest, opts ...micro.CallOption) (*clusterpb.GetFailureDomainResponse, error)) {
	s.mockGetFailureDomain = mock
}

func (s *ClusterFakeService) CreateCluster(ctx context.Context, in *clusterpb.ClusterCreateReqDTO, opts ...micro.CallOption) (*clusterpb.ClusterCreateRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryCluster(ctx context.Context, in *clusterpb.ClusterQueryReqDTO, opts ...micro.CallOption) (*clusterpb.ClusterQueryRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DeleteCluster(ctx context.Context, in *clusterpb.ClusterDeleteReqDTO, opts ...micro.CallOption) (*clusterpb.ClusterDeleteRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DetailCluster(ctx context.Context, in *clusterpb.ClusterDetailReqDTO, opts ...micro.CallOption) (*clusterpb.ClusterDetailRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) RecoverCluster(ctx context.Context, in *clusterpb.RecoverRequest, opts ...micro.CallOption) (*clusterpb.RecoverResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryBackupRecord(ctx context.Context, in *clusterpb.QueryBackupRequest, opts ...micro.CallOption) (*clusterpb.QueryBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) CreateBackup(ctx context.Context, in *clusterpb.CreateBackupRequest, opts ...micro.CallOption) (*clusterpb.CreateBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) RecoverBackupRecord(ctx context.Context, in *clusterpb.RecoverRequest, opts ...micro.CallOption) (*clusterpb.RecoverResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DeleteBackupRecord(ctx context.Context, in *clusterpb.DeleteBackupRequest, opts ...micro.CallOption) (*clusterpb.DeleteBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) SaveBackupStrategy(ctx context.Context, in *clusterpb.SaveBackupStrategyRequest, opts ...micro.CallOption) (*clusterpb.SaveBackupStrategyResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) GetBackupStrategy(ctx context.Context, in *clusterpb.GetBackupStrategyRequest, opts ...micro.CallOption) (*clusterpb.GetBackupStrategyResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryParameters(ctx context.Context, in *clusterpb.QueryClusterParametersRequest, opts ...micro.CallOption) (*clusterpb.QueryClusterParametersResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) SaveParameters(ctx context.Context, in *clusterpb.SaveClusterParametersRequest, opts ...micro.CallOption) (*clusterpb.SaveClusterParametersResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DescribeDashboard(ctx context.Context, in *clusterpb.DescribeDashboardRequest, opts ...micro.CallOption) (*clusterpb.DescribeDashboardResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) ImportData(ctx context.Context, in *clusterpb.DataImportRequest, opts ...micro.CallOption) (*clusterpb.DataImportResponse, error) {
	panic("implement me")
}
func (s *ClusterFakeService) ExportData(ctx context.Context, in *clusterpb.DataExportRequest, opts ...micro.CallOption) (*clusterpb.DataExportResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DescribeDataTransport(ctx context.Context, in *clusterpb.DataTransportQueryRequest, opts ...micro.CallOption) (*clusterpb.DataTransportQueryResponse, error) {
	panic("implement me")
}
