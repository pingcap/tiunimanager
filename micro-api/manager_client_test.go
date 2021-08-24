package main

import (
	"context"

	"github.com/asim/go-micro/v3/client"
	clusterClient "github.com/pingcap/tiem/micro-cluster/client"
	pb "github.com/pingcap/tiem/micro-cluster/proto"
)

type ClusterFakeService struct {
	// Auth manager module
	mockLogin          func(ctx context.Context, in *pb.LoginRequest, opts ...client.CallOption) (*pb.LoginResponse, error)
	mockLogout         func(ctx context.Context, in *pb.LogoutRequest, opts ...client.CallOption) (*pb.LogoutResponse, error)
	mockVerifyIdentity func(ctx context.Context, in *pb.VerifyIdentityRequest, opts ...client.CallOption) (*pb.VerifyIdentityResponse, error)
	// host manager module
	mockImportHost         func(ctx context.Context, in *pb.ImportHostRequest, opts ...client.CallOption) (*pb.ImportHostResponse, error)
	mockImportHostsInBatch func(ctx context.Context, in *pb.ImportHostsInBatchRequest, opts ...client.CallOption) (*pb.ImportHostsInBatchResponse, error)
	mockRemoveHost         func(ctx context.Context, in *pb.RemoveHostRequest, opts ...client.CallOption) (*pb.RemoveHostResponse, error)
	mockRemoveHostsInBatch func(ctx context.Context, in *pb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*pb.RemoveHostsInBatchResponse, error)
	mockListHost           func(ctx context.Context, in *pb.ListHostsRequest, opts ...client.CallOption) (*pb.ListHostsResponse, error)
	mockCheckDetails       func(ctx context.Context, in *pb.CheckDetailsRequest, opts ...client.CallOption) (*pb.CheckDetailsResponse, error)
	mockAllocHosts         func(ctx context.Context, in *pb.AllocHostsRequest, opts ...client.CallOption) (*pb.AllocHostResponse, error)
	mockGetFailureDomain   func(ctx context.Context, in *pb.GetFailureDomainRequest, opts ...client.CallOption) (*pb.GetFailureDomainResponse, error)
}

func (s *ClusterFakeService) Login(ctx context.Context, in *pb.LoginRequest, opts ...client.CallOption) (*pb.LoginResponse, error) {
	return s.mockLogin(ctx, in, opts...)
}

func (s *ClusterFakeService) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...client.CallOption) (*pb.LogoutResponse, error) {
	return s.mockLogout(ctx, in, opts...)
}

func (s *ClusterFakeService) VerifyIdentity(ctx context.Context, in *pb.VerifyIdentityRequest, opts ...client.CallOption) (*pb.VerifyIdentityResponse, error) {
	return s.mockVerifyIdentity(ctx, in, opts...)
}

func (s *ClusterFakeService) ImportHost(ctx context.Context, in *pb.ImportHostRequest, opts ...client.CallOption) (*pb.ImportHostResponse, error) {
	return s.mockImportHost(ctx, in, opts...)
}
func (s *ClusterFakeService) ImportHostsInBatch(ctx context.Context, in *pb.ImportHostsInBatchRequest, opts ...client.CallOption) (*pb.ImportHostsInBatchResponse, error) {
	return s.mockImportHostsInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) RemoveHost(ctx context.Context, in *pb.RemoveHostRequest, opts ...client.CallOption) (*pb.RemoveHostResponse, error) {
	return s.mockRemoveHost(ctx, in, opts...)
}
func (s *ClusterFakeService) RemoveHostsInBatch(ctx context.Context, in *pb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*pb.RemoveHostsInBatchResponse, error) {
	return s.mockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *ClusterFakeService) ListHost(ctx context.Context, in *pb.ListHostsRequest, opts ...client.CallOption) (*pb.ListHostsResponse, error) {
	return s.mockListHost(ctx, in, opts...)
}
func (s *ClusterFakeService) CheckDetails(ctx context.Context, in *pb.CheckDetailsRequest, opts ...client.CallOption) (*pb.CheckDetailsResponse, error) {
	return s.mockCheckDetails(ctx, in, opts...)
}
func (s *ClusterFakeService) AllocHosts(ctx context.Context, in *pb.AllocHostsRequest, opts ...client.CallOption) (*pb.AllocHostResponse, error) {
	return s.mockAllocHosts(ctx, in, opts...)
}
func (s *ClusterFakeService) GetFailureDomain(ctx context.Context, in *pb.GetFailureDomainRequest, opts ...client.CallOption) (*pb.GetFailureDomainResponse, error) {
	return s.mockGetFailureDomain(ctx, in, opts...)
}

func InitFakeManagerClient() *ClusterFakeService {

	fakeClient := &ClusterFakeService{
		mockLogin: func(ctx context.Context, in *pb.LoginRequest, opts ...client.CallOption) (*pb.LoginResponse, error) {
			return nil, nil
		},
		mockLogout: func(ctx context.Context, in *pb.LogoutRequest, opts ...client.CallOption) (*pb.LogoutResponse, error) {
			return nil, nil
		},
		mockVerifyIdentity: func(ctx context.Context, in *pb.VerifyIdentityRequest, opts ...client.CallOption) (*pb.VerifyIdentityResponse, error) {
			rsp := new(pb.VerifyIdentityResponse)
			rsp.Status = new(pb.ManagerResponseStatus)
			rsp.AccountId = "fake-accountID"
			rsp.TenantId = "fake-tenantID"
			rsp.AccountName = "fake-accountName"
			rsp.Status.Code = 0
			return rsp, nil
		},

		mockImportHost: func(ctx context.Context, in *pb.ImportHostRequest, opts ...client.CallOption) (*pb.ImportHostResponse, error) {
			return nil, nil
		},
		mockImportHostsInBatch: func(ctx context.Context, in *pb.ImportHostsInBatchRequest, opts ...client.CallOption) (*pb.ImportHostsInBatchResponse, error) {
			return nil, nil
		},
		mockRemoveHost: func(ctx context.Context, in *pb.RemoveHostRequest, opts ...client.CallOption) (*pb.RemoveHostResponse, error) {
			return nil, nil
		},
		mockRemoveHostsInBatch: func(ctx context.Context, in *pb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*pb.RemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		mockListHost: func(ctx context.Context, in *pb.ListHostsRequest, opts ...client.CallOption) (*pb.ListHostsResponse, error) {
			return nil, nil
		},
		mockCheckDetails: func(ctx context.Context, in *pb.CheckDetailsRequest, opts ...client.CallOption) (*pb.CheckDetailsResponse, error) {
			return nil, nil
		},
		mockAllocHosts: func(ctx context.Context, in *pb.AllocHostsRequest, opts ...client.CallOption) (*pb.AllocHostResponse, error) {
			return nil, nil
		},
		mockGetFailureDomain: func(ctx context.Context, in *pb.GetFailureDomainRequest, opts ...client.CallOption) (*pb.GetFailureDomainResponse, error) {
			return nil, nil
		},
	}
	clusterClient.ClusterClient = fakeClient
	return fakeClient
}

func (s *ClusterFakeService) MockImportHost(mock func(ctx context.Context, in *pb.ImportHostRequest, opts ...client.CallOption) (*pb.ImportHostResponse, error)) {
	s.mockImportHost = mock
}
func (s *ClusterFakeService) MockImportHostsInBatch(mock func(ctx context.Context, in *pb.ImportHostsInBatchRequest, opts ...client.CallOption) (*pb.ImportHostsInBatchResponse, error)) {
	s.mockImportHostsInBatch = mock
}
func (s *ClusterFakeService) MockRemoveHost(mock func(ctx context.Context, in *pb.RemoveHostRequest, opts ...client.CallOption) (*pb.RemoveHostResponse, error)) {
	s.mockRemoveHost = mock
}
func (s *ClusterFakeService) MockRemoveHostsInBatch(mock func(ctx context.Context, in *pb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*pb.RemoveHostsInBatchResponse, error)) {
	s.mockRemoveHostsInBatch = mock
}
func (s *ClusterFakeService) MockListHost(mock func(ctx context.Context, in *pb.ListHostsRequest, opts ...client.CallOption) (*pb.ListHostsResponse, error)) {
	s.mockListHost = mock
}
func (s *ClusterFakeService) MockCheckDetails(mock func(ctx context.Context, in *pb.CheckDetailsRequest, opts ...client.CallOption) (*pb.CheckDetailsResponse, error)) {
	s.mockCheckDetails = mock
}
func (s *ClusterFakeService) MockAllocHosts(mock func(ctx context.Context, in *pb.AllocHostsRequest, opts ...client.CallOption) (*pb.AllocHostResponse, error)) {
	s.mockAllocHosts = mock
}
func (s *ClusterFakeService) MockGetFailureDomain(mock func(ctx context.Context, in *pb.GetFailureDomainRequest, opts ...client.CallOption) (*pb.GetFailureDomainResponse, error)) {
	s.mockGetFailureDomain = mock
}

func (s *ClusterFakeService) CreateCluster(ctx context.Context, in *pb.ClusterCreateReqDTO, opts ...client.CallOption) (*pb.ClusterCreateRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryCluster(ctx context.Context, in *pb.ClusterQueryReqDTO, opts ...client.CallOption) (*pb.ClusterQueryRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DeleteCluster(ctx context.Context, in *pb.ClusterDeleteReqDTO, opts ...client.CallOption) (*pb.ClusterDeleteRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DetailCluster(ctx context.Context, in *pb.ClusterDetailReqDTO, opts ...client.CallOption) (*pb.ClusterDetailRespDTO, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryBackupRecord(ctx context.Context, in *pb.QueryBackupRequest, opts ...client.CallOption) (*pb.QueryBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) CreateBackup(ctx context.Context, in *pb.CreateBackupRequest, opts ...client.CallOption) (*pb.CreateBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) RecoverBackupRecord(ctx context.Context, in *pb.RecoverBackupRequest, opts ...client.CallOption) (*pb.RecoverBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DeleteBackupRecord(ctx context.Context, in *pb.DeleteBackupRequest, opts ...client.CallOption) (*pb.DeleteBackupResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) SaveBackupStrategy(ctx context.Context, in *pb.SaveBackupStrategyRequest, opts ...client.CallOption) (*pb.SaveBackupStrategyResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) GetBackupStrategy(ctx context.Context, in *pb.GetBackupStrategyRequest, opts ...client.CallOption) (*pb.GetBackupStrategyResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) QueryParameters(ctx context.Context, in *pb.QueryClusterParametersRequest, opts ...client.CallOption) (*pb.QueryClusterParametersResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) SaveParameters(ctx context.Context, in *pb.SaveClusterParametersRequest, opts ...client.CallOption) (*pb.SaveClusterParametersResponse, error) {
	panic("implement me")
}

func (s *ClusterFakeService) DescribeDashboard(ctx context.Context, in *pb.DescribeDashboardRequest, opts ...client.CallOption) (*pb.DescribeDashboardResponse, error) {
	panic("implement me")
}

