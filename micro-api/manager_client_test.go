package main

import (
	"context"

	"github.com/asim/go-micro/v3/client"
	managerClient "github.com/pingcap/tiem/micro-cluster/client"
	managerPb "github.com/pingcap/tiem/micro-cluster/proto"
)

type ManagerFakeService struct {
	// Auth manager module
	mockLogin          func(ctx context.Context, in *managerPb.LoginRequest, opts ...client.CallOption) (*managerPb.LoginResponse, error)
	mockLogout         func(ctx context.Context, in *managerPb.LogoutRequest, opts ...client.CallOption) (*managerPb.LogoutResponse, error)
	mockVerifyIdentity func(ctx context.Context, in *managerPb.VerifyIdentityRequest, opts ...client.CallOption) (*managerPb.VerifyIdentityResponse, error)
	// host manager module
	mockImportHost         func(ctx context.Context, in *managerPb.ImportHostRequest, opts ...client.CallOption) (*managerPb.ImportHostResponse, error)
	mockImportHostsInBatch func(ctx context.Context, in *managerPb.ImportHostsInBatchRequest, opts ...client.CallOption) (*managerPb.ImportHostsInBatchResponse, error)
	mockRemoveHost         func(ctx context.Context, in *managerPb.RemoveHostRequest, opts ...client.CallOption) (*managerPb.RemoveHostResponse, error)
	mockRemoveHostsInBatch func(ctx context.Context, in *managerPb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*managerPb.RemoveHostsInBatchResponse, error)
	mockListHost           func(ctx context.Context, in *managerPb.ListHostsRequest, opts ...client.CallOption) (*managerPb.ListHostsResponse, error)
	mockCheckDetails       func(ctx context.Context, in *managerPb.CheckDetailsRequest, opts ...client.CallOption) (*managerPb.CheckDetailsResponse, error)
	mockAllocHosts         func(ctx context.Context, in *managerPb.AllocHostsRequest, opts ...client.CallOption) (*managerPb.AllocHostResponse, error)
	mockGetFailureDomain   func(ctx context.Context, in *managerPb.GetFailureDomainRequest, opts ...client.CallOption) (*managerPb.GetFailureDomainResponse, error)
}

func (s *ManagerFakeService) Login(ctx context.Context, in *managerPb.LoginRequest, opts ...client.CallOption) (*managerPb.LoginResponse, error) {
	return s.mockLogin(ctx, in, opts...)
}

func (s *ManagerFakeService) Logout(ctx context.Context, in *managerPb.LogoutRequest, opts ...client.CallOption) (*managerPb.LogoutResponse, error) {
	return s.mockLogout(ctx, in, opts...)
}

func (s *ManagerFakeService) VerifyIdentity(ctx context.Context, in *managerPb.VerifyIdentityRequest, opts ...client.CallOption) (*managerPb.VerifyIdentityResponse, error) {
	return s.mockVerifyIdentity(ctx, in, opts...)
}

func (s *ManagerFakeService) ImportHost(ctx context.Context, in *managerPb.ImportHostRequest, opts ...client.CallOption) (*managerPb.ImportHostResponse, error) {
	return s.mockImportHost(ctx, in, opts...)
}
func (s *ManagerFakeService) ImportHostsInBatch(ctx context.Context, in *managerPb.ImportHostsInBatchRequest, opts ...client.CallOption) (*managerPb.ImportHostsInBatchResponse, error) {
	return s.mockImportHostsInBatch(ctx, in, opts...)
}
func (s *ManagerFakeService) RemoveHost(ctx context.Context, in *managerPb.RemoveHostRequest, opts ...client.CallOption) (*managerPb.RemoveHostResponse, error) {
	return s.mockRemoveHost(ctx, in, opts...)
}
func (s *ManagerFakeService) RemoveHostsInBatch(ctx context.Context, in *managerPb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*managerPb.RemoveHostsInBatchResponse, error) {
	return s.mockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *ManagerFakeService) ListHost(ctx context.Context, in *managerPb.ListHostsRequest, opts ...client.CallOption) (*managerPb.ListHostsResponse, error) {
	return s.mockListHost(ctx, in, opts...)
}
func (s *ManagerFakeService) CheckDetails(ctx context.Context, in *managerPb.CheckDetailsRequest, opts ...client.CallOption) (*managerPb.CheckDetailsResponse, error) {
	return s.mockCheckDetails(ctx, in, opts...)
}
func (s *ManagerFakeService) AllocHosts(ctx context.Context, in *managerPb.AllocHostsRequest, opts ...client.CallOption) (*managerPb.AllocHostResponse, error) {
	return s.mockAllocHosts(ctx, in, opts...)
}
func (s *ManagerFakeService) GetFailureDomain(ctx context.Context, in *managerPb.GetFailureDomainRequest, opts ...client.CallOption) (*managerPb.GetFailureDomainResponse, error) {
	return s.mockGetFailureDomain(ctx, in, opts...)
}

func InitFakeManagerClient() *ManagerFakeService {

	fakeManagerClient := &ManagerFakeService{
		mockLogin: func(ctx context.Context, in *managerPb.LoginRequest, opts ...client.CallOption) (*managerPb.LoginResponse, error) {
			return nil, nil
		},
		mockLogout: func(ctx context.Context, in *managerPb.LogoutRequest, opts ...client.CallOption) (*managerPb.LogoutResponse, error) {
			return nil, nil
		},
		mockVerifyIdentity: func(ctx context.Context, in *managerPb.VerifyIdentityRequest, opts ...client.CallOption) (*managerPb.VerifyIdentityResponse, error) {
			rsp := new(managerPb.VerifyIdentityResponse)
			rsp.Status = new(managerPb.ManagerResponseStatus)
			rsp.AccountId = "fake-accountID"
			rsp.TenantId = "fake-tenantID"
			rsp.AccountName = "fake-accountName"
			rsp.Status.Code = 0
			return rsp, nil
		},

		mockImportHost: func(ctx context.Context, in *managerPb.ImportHostRequest, opts ...client.CallOption) (*managerPb.ImportHostResponse, error) {
			return nil, nil
		},
		mockImportHostsInBatch: func(ctx context.Context, in *managerPb.ImportHostsInBatchRequest, opts ...client.CallOption) (*managerPb.ImportHostsInBatchResponse, error) {
			return nil, nil
		},
		mockRemoveHost: func(ctx context.Context, in *managerPb.RemoveHostRequest, opts ...client.CallOption) (*managerPb.RemoveHostResponse, error) {
			return nil, nil
		},
		mockRemoveHostsInBatch: func(ctx context.Context, in *managerPb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*managerPb.RemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		mockListHost: func(ctx context.Context, in *managerPb.ListHostsRequest, opts ...client.CallOption) (*managerPb.ListHostsResponse, error) {
			return nil, nil
		},
		mockCheckDetails: func(ctx context.Context, in *managerPb.CheckDetailsRequest, opts ...client.CallOption) (*managerPb.CheckDetailsResponse, error) {
			return nil, nil
		},
		mockAllocHosts: func(ctx context.Context, in *managerPb.AllocHostsRequest, opts ...client.CallOption) (*managerPb.AllocHostResponse, error) {
			return nil, nil
		},
		mockGetFailureDomain: func(ctx context.Context, in *managerPb.GetFailureDomainRequest, opts ...client.CallOption) (*managerPb.GetFailureDomainResponse, error) {
			return nil, nil
		},
	}
	managerClient.ManagerClient = fakeManagerClient
	return fakeManagerClient
}

func (s *ManagerFakeService) MockImportHost(mock func(ctx context.Context, in *managerPb.ImportHostRequest, opts ...client.CallOption) (*managerPb.ImportHostResponse, error)) {
	s.mockImportHost = mock
}
func (s *ManagerFakeService) MockImportHostsInBatch(mock func(ctx context.Context, in *managerPb.ImportHostsInBatchRequest, opts ...client.CallOption) (*managerPb.ImportHostsInBatchResponse, error)) {
	s.mockImportHostsInBatch = mock
}
func (s *ManagerFakeService) MockRemoveHost(mock func(ctx context.Context, in *managerPb.RemoveHostRequest, opts ...client.CallOption) (*managerPb.RemoveHostResponse, error)) {
	s.mockRemoveHost = mock
}
func (s *ManagerFakeService) MockRemoveHostsInBatch(mock func(ctx context.Context, in *managerPb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*managerPb.RemoveHostsInBatchResponse, error)) {
	s.mockRemoveHostsInBatch = mock
}
func (s *ManagerFakeService) MockListHost(mock func(ctx context.Context, in *managerPb.ListHostsRequest, opts ...client.CallOption) (*managerPb.ListHostsResponse, error)) {
	s.mockListHost = mock
}
func (s *ManagerFakeService) MockCheckDetails(mock func(ctx context.Context, in *managerPb.CheckDetailsRequest, opts ...client.CallOption) (*managerPb.CheckDetailsResponse, error)) {
	s.mockCheckDetails = mock
}
func (s *ManagerFakeService) MockAllocHosts(mock func(ctx context.Context, in *managerPb.AllocHostsRequest, opts ...client.CallOption) (*managerPb.AllocHostResponse, error)) {
	s.mockAllocHosts = mock
}
func (s *ManagerFakeService) MockGetFailureDomain(mock func(ctx context.Context, in *managerPb.GetFailureDomainRequest, opts ...client.CallOption) (*managerPb.GetFailureDomainResponse, error)) {
	s.mockGetFailureDomain = mock
}
