package main

import (
	"context"

	"github.com/pingcap-inc/tiem/library/client"

	micro "github.com/asim/go-micro/v3/client"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
)

type ClusterFakeService struct {
	// Auth manager module
	mockLogin          func(ctx context.Context, in *cluster.LoginRequest, opts ...micro.CallOption) (*cluster.LoginResponse, error)
	mockLogout         func(ctx context.Context, in *cluster.LogoutRequest, opts ...micro.CallOption) (*cluster.LogoutResponse, error)
	mockVerifyIdentity func(ctx context.Context, in *cluster.VerifyIdentityRequest, opts ...micro.CallOption) (*cluster.VerifyIdentityResponse, error)
	// host manager module
	mockImportHost            func(ctx context.Context, in *cluster.ImportHostRequest, opts ...micro.CallOption) (*cluster.ImportHostResponse, error)
	mockImportHostsInBatch    func(ctx context.Context, in *cluster.ImportHostsInBatchRequest, opts ...micro.CallOption) (*cluster.ImportHostsInBatchResponse, error)
	mockRemoveHost            func(ctx context.Context, in *cluster.RemoveHostRequest, opts ...micro.CallOption) (*cluster.RemoveHostResponse, error)
	mockRemoveHostsInBatch    func(ctx context.Context, in *cluster.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*cluster.RemoveHostsInBatchResponse, error)
	mockListHost              func(ctx context.Context, in *cluster.ListHostsRequest, opts ...micro.CallOption) (*cluster.ListHostsResponse, error)
	mockCheckDetails          func(ctx context.Context, in *cluster.CheckDetailsRequest, opts ...micro.CallOption) (*cluster.CheckDetailsResponse, error)
	mockAllocHosts            func(ctx context.Context, in *cluster.AllocHostsRequest, opts ...micro.CallOption) (*cluster.AllocHostResponse, error)
	mockGetFailureDomain      func(ctx context.Context, in *cluster.GetFailureDomainRequest, opts ...micro.CallOption) (*cluster.GetFailureDomainResponse, error)
	mockAllocResourcesInBatch func(ctx context.Context, in *cluster.BatchAllocRequest, opts ...micro.CallOption) (*cluster.BatchAllocResponse, error)
	mockRecycleResources      func(ctx context.Context, in *cluster.RecycleRequest, opts ...micro.CallOption) (*cluster.RecycleResponse, error)
}

func (s *ClusterFakeService) ListFlows(ctx context.Context, in *cluster.ListFlowsRequest, opts ...micro.CallOption) (*cluster.ListFlowsResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) Login(ctx context.Context, in *cluster.LoginRequest, opts ...micro.CallOption) (*cluster.LoginResponse, error) {
	return s.mockLogin(ctx, in, opts...)
}

func (s *ClusterFakeService) Logout(ctx context.Context, in *cluster.LogoutRequest, opts ...micro.CallOption) (*cluster.LogoutResponse, error) {
	return s.mockLogout(ctx, in, opts...)
}

func (s *ClusterFakeService) VerifyIdentity(ctx context.Context, in *cluster.VerifyIdentityRequest, opts ...micro.CallOption) (*cluster.VerifyIdentityResponse, error) {
	return s.mockVerifyIdentity(ctx, in, opts...)
}

func (s *ClusterFakeService) ImportHost(ctx context.Context, in *cluster.ImportHostRequest, opts ...micro.CallOption) (*cluster.ImportHostResponse, error) {
	return s.mockImportHost(ctx, in, opts...)
}
func (s *ClusterFakeService) ImportHostsInBatch(ctx context.Context, in *cluster.ImportHostsInBatchRequest, opts ...micro.CallOption) (*cluster.ImportHostsInBatchResponse, error) {
	return s.mockImportHostsInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) RemoveHost(ctx context.Context, in *cluster.RemoveHostRequest, opts ...micro.CallOption) (*cluster.RemoveHostResponse, error) {
	return s.mockRemoveHost(ctx, in, opts...)
}
func (s *ClusterFakeService) RemoveHostsInBatch(ctx context.Context, in *cluster.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*cluster.RemoveHostsInBatchResponse, error) {
	return s.mockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) ListHost(ctx context.Context, in *cluster.ListHostsRequest, opts ...micro.CallOption) (*cluster.ListHostsResponse, error) {
	return s.mockListHost(ctx, in, opts...)
}
func (s *ClusterFakeService) CheckDetails(ctx context.Context, in *cluster.CheckDetailsRequest, opts ...micro.CallOption) (*cluster.CheckDetailsResponse, error) {
	return s.mockCheckDetails(ctx, in, opts...)
}
func (s *ClusterFakeService) AllocHosts(ctx context.Context, in *cluster.AllocHostsRequest, opts ...micro.CallOption) (*cluster.AllocHostResponse, error) {
	return s.mockAllocHosts(ctx, in, opts...)
}
func (s *ClusterFakeService) GetFailureDomain(ctx context.Context, in *cluster.GetFailureDomainRequest, opts ...micro.CallOption) (*cluster.GetFailureDomainResponse, error) {
	return s.mockGetFailureDomain(ctx, in, opts...)
}
func (s *ClusterFakeService) AllocResourcesInBatch(ctx context.Context, in *cluster.BatchAllocRequest, opts ...micro.CallOption) (*cluster.BatchAllocResponse, error) {
	return s.mockAllocResourcesInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) RecycleResources(ctx context.Context, in *cluster.RecycleRequest, opts ...micro.CallOption) (*cluster.RecycleResponse, error) {
	return s.mockRecycleResources(ctx, in, opts...)
}

func InitFakeClusterClient() *ClusterFakeService {

	fakeClient := &ClusterFakeService{
		mockLogin: func(ctx context.Context, in *cluster.LoginRequest, opts ...micro.CallOption) (*cluster.LoginResponse, error) {
			return nil, nil
		},
		mockLogout: func(ctx context.Context, in *cluster.LogoutRequest, opts ...micro.CallOption) (*cluster.LogoutResponse, error) {
			return nil, nil
		},
		mockVerifyIdentity: func(ctx context.Context, in *cluster.VerifyIdentityRequest, opts ...micro.CallOption) (*cluster.VerifyIdentityResponse, error) {
			rsp := new(cluster.VerifyIdentityResponse)
			rsp.Status = new(cluster.ManagerResponseStatus)
			rsp.AccountId = "fake-accountID"
			rsp.TenantId = "fake-tenantID"
			rsp.AccountName = "fake-accountName"
			rsp.Status.Code = 0
			return rsp, nil
		},

		mockImportHost: func(ctx context.Context, in *cluster.ImportHostRequest, opts ...micro.CallOption) (*cluster.ImportHostResponse, error) {
			return nil, nil
		},
		mockImportHostsInBatch: func(ctx context.Context, in *cluster.ImportHostsInBatchRequest, opts ...micro.CallOption) (*cluster.ImportHostsInBatchResponse, error) {
			return nil, nil
		},
		mockRemoveHost: func(ctx context.Context, in *cluster.RemoveHostRequest, opts ...micro.CallOption) (*cluster.RemoveHostResponse, error) {
			return nil, nil
		},
		mockRemoveHostsInBatch: func(ctx context.Context, in *cluster.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*cluster.RemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		mockListHost: func(ctx context.Context, in *cluster.ListHostsRequest, opts ...micro.CallOption) (*cluster.ListHostsResponse, error) {
			return nil, nil
		},
		mockCheckDetails: func(ctx context.Context, in *cluster.CheckDetailsRequest, opts ...micro.CallOption) (*cluster.CheckDetailsResponse, error) {
			return nil, nil
		},
		mockAllocHosts: func(ctx context.Context, in *cluster.AllocHostsRequest, opts ...micro.CallOption) (*cluster.AllocHostResponse, error) {
			return nil, nil
		},
		mockGetFailureDomain: func(ctx context.Context, in *cluster.GetFailureDomainRequest, opts ...micro.CallOption) (*cluster.GetFailureDomainResponse, error) {
			return nil, nil
		},
	}
	client.ClusterClient = fakeClient
	return fakeClient
}

func (s *ClusterFakeService) MockImportHost(mock func(ctx context.Context, in *cluster.ImportHostRequest, opts ...micro.CallOption) (*cluster.ImportHostResponse, error)) {
	s.mockImportHost = mock
}
func (s *ClusterFakeService) MockImportHostsInBatch(mock func(ctx context.Context, in *cluster.ImportHostsInBatchRequest, opts ...micro.CallOption) (*cluster.ImportHostsInBatchResponse, error)) {
	s.mockImportHostsInBatch = mock
}
func (s *ClusterFakeService) MockRemoveHost(mock func(ctx context.Context, in *cluster.RemoveHostRequest, opts ...micro.CallOption) (*cluster.RemoveHostResponse, error)) {
	s.mockRemoveHost = mock
}
func (s *ClusterFakeService) MockRemoveHostsInBatch(mock func(ctx context.Context, in *cluster.RemoveHostsInBatchRequest, opts ...micro.CallOption) (*cluster.RemoveHostsInBatchResponse, error)) {
	s.mockRemoveHostsInBatch = mock
}
func (s *ClusterFakeService) MockListHost(mock func(ctx context.Context, in *cluster.ListHostsRequest, opts ...micro.CallOption) (*cluster.ListHostsResponse, error)) {
	s.mockListHost = mock
}
func (s *ClusterFakeService) MockCheckDetails(mock func(ctx context.Context, in *cluster.CheckDetailsRequest, opts ...micro.CallOption) (*cluster.CheckDetailsResponse, error)) {
	s.mockCheckDetails = mock
}
func (s *ClusterFakeService) MockAllocHosts(mock func(ctx context.Context, in *cluster.AllocHostsRequest, opts ...micro.CallOption) (*cluster.AllocHostResponse, error)) {
	s.mockAllocHosts = mock
}
func (s *ClusterFakeService) MockGetFailureDomain(mock func(ctx context.Context, in *cluster.GetFailureDomainRequest, opts ...micro.CallOption) (*cluster.GetFailureDomainResponse, error)) {
	s.mockGetFailureDomain = mock
}

func (s *ClusterFakeService) CreateCluster(ctx context.Context, in *cluster.ClusterCreateReqDTO, opts ...micro.CallOption) (*cluster.ClusterCreateRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryCluster(ctx context.Context, in *cluster.ClusterQueryReqDTO, opts ...micro.CallOption) (*cluster.ClusterQueryRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DeleteCluster(ctx context.Context, in *cluster.ClusterDeleteReqDTO, opts ...micro.CallOption) (*cluster.ClusterDeleteRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DetailCluster(ctx context.Context, in *cluster.ClusterDetailReqDTO, opts ...micro.CallOption) (*cluster.ClusterDetailRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) RecoverCluster(ctx context.Context, in *cluster.RecoverRequest, opts ...micro.CallOption) (*cluster.RecoverResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryBackupRecord(ctx context.Context, in *cluster.QueryBackupRequest, opts ...micro.CallOption) (*cluster.QueryBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) CreateBackup(ctx context.Context, in *cluster.CreateBackupRequest, opts ...micro.CallOption) (*cluster.CreateBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) RecoverBackupRecord(ctx context.Context, in *cluster.RecoverRequest, opts ...micro.CallOption) (*cluster.RecoverResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DeleteBackupRecord(ctx context.Context, in *cluster.DeleteBackupRequest, opts ...micro.CallOption) (*cluster.DeleteBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) SaveBackupStrategy(ctx context.Context, in *cluster.SaveBackupStrategyRequest, opts ...micro.CallOption) (*cluster.SaveBackupStrategyResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) GetBackupStrategy(ctx context.Context, in *cluster.GetBackupStrategyRequest, opts ...micro.CallOption) (*cluster.GetBackupStrategyResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryParameters(ctx context.Context, in *cluster.QueryClusterParametersRequest, opts ...micro.CallOption) (*cluster.QueryClusterParametersResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) SaveParameters(ctx context.Context, in *cluster.SaveClusterParametersRequest, opts ...micro.CallOption) (*cluster.SaveClusterParametersResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DescribeDashboard(ctx context.Context, in *cluster.DescribeDashboardRequest, opts ...micro.CallOption) (*cluster.DescribeDashboardResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) ImportData(ctx context.Context, in *cluster.DataImportRequest, opts ...micro.CallOption) (*cluster.DataImportResponse, error) {
	panic("implement me")
}
func (s *ClusterFakeService) ExportData(ctx context.Context, in *cluster.DataExportRequest, opts ...micro.CallOption) (*cluster.DataExportResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DescribeDataTransport(ctx context.Context, in *cluster.DataTransportQueryRequest, opts ...micro.CallOption) (*cluster.DataTransportQueryResponse, error) {
	panic("implement me")
}
