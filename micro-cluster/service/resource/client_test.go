
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package resource

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

	"github.com/asim/go-micro/v3/client"
	rpc_client "github.com/pingcap-inc/tiem/library/client"
)

type DBFakeService struct {
	// Auth Module
	mockFindTenant            func(ctx context.Context, in *dbpb.DBFindTenantRequest, opts ...client.CallOption) (*dbpb.DBFindTenantResponse, error)
	mockFindAccount           func(ctx context.Context, in *dbpb.DBFindAccountRequest, opts ...client.CallOption) (*dbpb.DBFindAccountResponse, error)
	mockSaveToken             func(ctx context.Context, in *dbpb.DBSaveTokenRequest, opts ...client.CallOption) (*dbpb.DBSaveTokenResponse, error)
	mockFindToken             func(ctx context.Context, in *dbpb.DBFindTokenRequest, opts ...client.CallOption) (*dbpb.DBFindTokenResponse, error)
	mockFindRolesByPermission func(ctx context.Context, in *dbpb.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*dbpb.DBFindRolesByPermissionResponse, error)
	// Host Module
	mockAddHost               func(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error)
	mockAddHostsInBatch       func(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error)
	mockRemoveHost            func(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error)
	mockRemoveHostsInBatch    func(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error)
	mockListHost              func(ctx context.Context, in *dbpb.DBListHostsRequest, opts ...client.CallOption) (*dbpb.DBListHostsResponse, error)
	mockCheckDetails          func(ctx context.Context, in *dbpb.DBCheckDetailsRequest, opts ...client.CallOption) (*dbpb.DBCheckDetailsResponse, error)
	mockAllocHosts            func(ctx context.Context, in *dbpb.DBAllocHostsRequest, opts ...client.CallOption) (*dbpb.DBAllocHostsResponse, error)
	mockGetFailureDomain      func(ctx context.Context, in *dbpb.DBGetFailureDomainRequest, opts ...client.CallOption) (*dbpb.DBGetFailureDomainResponse, error)
	mockAllocResources        func(ctx context.Context, in *dbpb.DBAllocRequest, opts ...client.CallOption) (*dbpb.DBAllocResponse, error)
	mockAllocResourcesInBatch func(ctx context.Context, in *dbpb.DBBatchAllocRequest, opts ...client.CallOption) (*dbpb.DBBatchAllocResponse, error)
	mockRecycleResources      func(ctx context.Context, in *dbpb.DBRecycleRequest, opts ...client.CallOption) (*dbpb.DBRecycleResponse, error)
	// Cluster
	mockCreateCluster               func(ctx context.Context, in *dbpb.DBCreateClusterRequest, opts ...client.CallOption) (*dbpb.DBCreateClusterResponse, error)
	mockDeleteCluster               func(ctx context.Context, in *dbpb.DBDeleteClusterRequest, opts ...client.CallOption) (*dbpb.DBDeleteClusterResponse, error)
	mockUpdateClusterStatus         func(ctx context.Context, in *dbpb.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*dbpb.DBUpdateClusterStatusResponse, error)
	mockUpdateClusterTopologyConfig func(ctx context.Context, in *dbpb.DBUpdateTopologyConfigRequest, opts ...client.CallOption) (*dbpb.DBUpdateTopologyConfigResponse, error)
	mockLoadCluster                 func(ctx context.Context, in *dbpb.DBLoadClusterRequest, opts ...client.CallOption) (*dbpb.DBLoadClusterResponse, error)
	mockListCluster                 func(ctx context.Context, in *dbpb.DBListClusterRequest, opts ...client.CallOption) (*dbpb.DBListClusterResponse, error)
	// backup & recover & parameters
	mockSaveBackupRecord           func(ctx context.Context, in *dbpb.DBSaveBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBSaveBackupRecordResponse, error)
	mockDeleteBackupRecord         func(ctx context.Context, in *dbpb.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBDeleteBackupRecordResponse, error)
	mockListBackupRecords          func(ctx context.Context, in *dbpb.DBListBackupRecordsRequest, opts ...client.CallOption) (*dbpb.DBListBackupRecordsResponse, error)
	mockSaveRecoverRecord          func(ctx context.Context, in *dbpb.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*dbpb.DBSaveRecoverRecordResponse, error)
	mockSaveParametersRecord       func(ctx context.Context, in *dbpb.DBSaveParametersRequest, opts ...client.CallOption) (*dbpb.DBSaveParametersResponse, error)
	mockGetCurrentParametersRecord func(ctx context.Context, in *dbpb.DBGetCurrentParametersRequest, opts ...client.CallOption) (*dbpb.DBGetCurrentParametersResponse, error)
	// Tiup Task
	mockCreateTiupTask           func(ctx context.Context, in *dbpb.CreateTiupTaskRequest, opts ...client.CallOption) (*dbpb.CreateTiupTaskResponse, error)
	mockUpdateTiupTask           func(ctx context.Context, in *dbpb.UpdateTiupTaskRequest, opts ...client.CallOption) (*dbpb.UpdateTiupTaskResponse, error)
	mockFindTiupTaskByID         func(ctx context.Context, in *dbpb.FindTiupTaskByIDRequest, opts ...client.CallOption) (*dbpb.FindTiupTaskByIDResponse, error)
	mockGetTiupTaskStatusByBizID func(ctx context.Context, in *dbpb.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*dbpb.GetTiupTaskStatusByBizIDResponse, error)
	// Workflow and Task
	mockCreateFlow func(ctx context.Context, in *dbpb.DBCreateFlowRequest, opts ...client.CallOption) (*dbpb.DBCreateFlowResponse, error)
	mockCreateTask func(ctx context.Context, in *dbpb.DBCreateTaskRequest, opts ...client.CallOption) (*dbpb.DBCreateTaskResponse, error)
	mockUpdateFlow func(ctx context.Context, in *dbpb.DBUpdateFlowRequest, opts ...client.CallOption) (*dbpb.DBUpdateFlowResponse, error)
	mockUpdateTask func(ctx context.Context, in *dbpb.DBUpdateTaskRequest, opts ...client.CallOption) (*dbpb.DBUpdateTaskResponse, error)
	mockLoadFlow   func(ctx context.Context, in *dbpb.DBLoadFlowRequest, opts ...client.CallOption) (*dbpb.DBLoadFlowResponse, error)
	mockLoadTask   func(ctx context.Context, in *dbpb.DBLoadTaskRequest, opts ...client.CallOption) (*dbpb.DBLoadTaskResponse, error)
}

func (s *DBFakeService) FindAccountById(ctx context.Context, in *dbpb.DBFindAccountByIdRequest, opts ...client.CallOption) (*dbpb.DBFindAccountByIdResponse, error) {
	panic("implement me")
}

func (s *DBFakeService) CreateInstance(ctx context.Context, in *dbpb.DBCreateInstanceRequest, opts ...client.CallOption) (*dbpb.DBCreateInstanceResponse, error) {
	panic("implement me")
}

func (s *DBFakeService) QueryBackupStrategyByTime(ctx context.Context, in *dbpb.DBQueryBackupStrategyByTimeRequest, opts ...client.CallOption) (*dbpb.DBQueryBackupStrategyByTimeResponse, error) {
	panic("implement me")
}

func (s *DBFakeService) ListFlows(ctx context.Context, in *dbpb.DBListFlowsRequest, opts ...client.CallOption) (*dbpb.DBListFlowsResponse, error) {
	panic("implement me")
}

// Mock Auth Module
func (s *DBFakeService) FindTenant(ctx context.Context, in *dbpb.DBFindTenantRequest, opts ...client.CallOption) (*dbpb.DBFindTenantResponse, error) {
	return s.mockFindTenant(ctx, in, opts...)
}
func (s *DBFakeService) FindAccount(ctx context.Context, in *dbpb.DBFindAccountRequest, opts ...client.CallOption) (*dbpb.DBFindAccountResponse, error) {
	return s.mockFindAccount(ctx, in, opts...)
}
func (s *DBFakeService) SaveToken(ctx context.Context, in *dbpb.DBSaveTokenRequest, opts ...client.CallOption) (*dbpb.DBSaveTokenResponse, error) {
	return s.mockSaveToken(ctx, in, opts...)
}
func (s *DBFakeService) FindToken(ctx context.Context, in *dbpb.DBFindTokenRequest, opts ...client.CallOption) (*dbpb.DBFindTokenResponse, error) {
	return s.mockFindToken(ctx, in, opts...)
}
func (s *DBFakeService) FindRolesByPermission(ctx context.Context, in *dbpb.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*dbpb.DBFindRolesByPermissionResponse, error) {
	return s.mockFindRolesByPermission(ctx, in, opts...)
}

// Mock Host Module
func (s *DBFakeService) AddHost(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error) {
	return s.mockAddHost(ctx, in, opts...)
}
func (s *DBFakeService) AddHostsInBatch(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error) {
	return s.mockAddHostsInBatch(ctx, in, opts...)
}
func (s *DBFakeService) RemoveHost(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error) {
	return s.mockRemoveHost(ctx, in, opts...)
}
func (s *DBFakeService) RemoveHostsInBatch(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error) {
	return s.mockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *DBFakeService) ListHost(ctx context.Context, in *dbpb.DBListHostsRequest, opts ...client.CallOption) (*dbpb.DBListHostsResponse, error) {
	return s.mockListHost(ctx, in, opts...)
}
func (s *DBFakeService) CheckDetails(ctx context.Context, in *dbpb.DBCheckDetailsRequest, opts ...client.CallOption) (*dbpb.DBCheckDetailsResponse, error) {
	return s.mockCheckDetails(ctx, in, opts...)
}
func (s *DBFakeService) AllocHosts(ctx context.Context, in *dbpb.DBAllocHostsRequest, opts ...client.CallOption) (*dbpb.DBAllocHostsResponse, error) {
	return s.mockAllocHosts(ctx, in, opts...)
}
func (s *DBFakeService) GetFailureDomain(ctx context.Context, in *dbpb.DBGetFailureDomainRequest, opts ...client.CallOption) (*dbpb.DBGetFailureDomainResponse, error) {
	return s.mockGetFailureDomain(ctx, in, opts...)
}
func (s *DBFakeService) AllocResources(ctx context.Context, in *dbpb.DBAllocRequest, opts ...client.CallOption) (*dbpb.DBAllocResponse, error) {
	return s.mockAllocResources(ctx, in, opts...)
}
func (s *DBFakeService) AllocResourcesInBatch(ctx context.Context, in *dbpb.DBBatchAllocRequest, opts ...client.CallOption) (*dbpb.DBBatchAllocResponse, error) {
	return s.mockAllocResourcesInBatch(ctx, in, opts...)
}
func (s *DBFakeService) RecycleResources(ctx context.Context, in *dbpb.DBRecycleRequest, opts ...client.CallOption) (*dbpb.DBRecycleResponse, error) {
	return s.mockRecycleResources(ctx, in, opts...)
}

// Mock Cluster
func (s *DBFakeService) CreateCluster(ctx context.Context, in *dbpb.DBCreateClusterRequest, opts ...client.CallOption) (*dbpb.DBCreateClusterResponse, error) {
	return s.mockCreateCluster(ctx, in, opts...)
}
func (s *DBFakeService) DeleteCluster(ctx context.Context, in *dbpb.DBDeleteClusterRequest, opts ...client.CallOption) (*dbpb.DBDeleteClusterResponse, error) {
	return s.mockDeleteCluster(ctx, in, opts...)
}
func (s *DBFakeService) UpdateClusterStatus(ctx context.Context, in *dbpb.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*dbpb.DBUpdateClusterStatusResponse, error) {
	return s.mockUpdateClusterStatus(ctx, in, opts...)
}
func (s *DBFakeService) UpdateClusterTopologyConfig(ctx context.Context, in *dbpb.DBUpdateTopologyConfigRequest, opts ...client.CallOption) (*dbpb.DBUpdateTopologyConfigResponse, error) {
	return s.mockUpdateClusterTopologyConfig(ctx, in, opts...)
}
func (s *DBFakeService) LoadCluster(ctx context.Context, in *dbpb.DBLoadClusterRequest, opts ...client.CallOption) (*dbpb.DBLoadClusterResponse, error) {
	return s.mockLoadCluster(ctx, in, opts...)
}
func (s *DBFakeService) ListCluster(ctx context.Context, in *dbpb.DBListClusterRequest, opts ...client.CallOption) (*dbpb.DBListClusterResponse, error) {
	return s.mockListCluster(ctx, in, opts...)
}

// Mock backup & recover & parameters
func (s *DBFakeService) SaveBackupRecord(ctx context.Context, in *dbpb.DBSaveBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBSaveBackupRecordResponse, error) {
	return s.mockSaveBackupRecord(ctx, in, opts...)
}
func (s *DBFakeService) DeleteBackupRecord(ctx context.Context, in *dbpb.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBDeleteBackupRecordResponse, error) {
	return s.mockDeleteBackupRecord(ctx, in, opts...)
}
func (s *DBFakeService) ListBackupRecords(ctx context.Context, in *dbpb.DBListBackupRecordsRequest, opts ...client.CallOption) (*dbpb.DBListBackupRecordsResponse, error) {
	return s.mockListBackupRecords(ctx, in, opts...)
}
func (s *DBFakeService) SaveRecoverRecord(ctx context.Context, in *dbpb.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*dbpb.DBSaveRecoverRecordResponse, error) {
	return s.mockSaveRecoverRecord(ctx, in, opts...)
}
func (s *DBFakeService) SaveParametersRecord(ctx context.Context, in *dbpb.DBSaveParametersRequest, opts ...client.CallOption) (*dbpb.DBSaveParametersResponse, error) {
	return s.mockSaveParametersRecord(ctx, in, opts...)
}
func (s *DBFakeService) GetCurrentParametersRecord(ctx context.Context, in *dbpb.DBGetCurrentParametersRequest, opts ...client.CallOption) (*dbpb.DBGetCurrentParametersResponse, error) {
	return s.mockGetCurrentParametersRecord(ctx, in, opts...)
}

func (s *DBFakeService) CreateTransportRecord(ctx context.Context, in *dbpb.DBCreateTransportRecordRequest, opts ...client.CallOption) (*dbpb.DBCreateTransportRecordResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) UpdateTransportRecord(ctx context.Context, in *dbpb.DBUpdateTransportRecordRequest, opts ...client.CallOption) (*dbpb.DBUpdateTransportRecordResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) FindTrasnportRecordByID(ctx context.Context, in *dbpb.DBFindTransportRecordByIDRequest, opts ...client.CallOption) (*dbpb.DBFindTransportRecordByIDResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) ListTrasnportRecord(ctx context.Context, in *dbpb.DBListTransportRecordRequest, opts ...client.CallOption) (*dbpb.DBListTransportRecordResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) QueryBackupRecords(ctx context.Context, in *dbpb.DBQueryBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBQueryBackupRecordResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) QueryBackupStrategy(ctx context.Context, in *dbpb.DBQueryBackupStrategyRequest, opts ...client.CallOption) (*dbpb.DBQueryBackupStrategyResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) SaveBackupStrategy(ctx context.Context, in *dbpb.DBSaveBackupStrategyRequest, opts ...client.CallOption) (*dbpb.DBSaveBackupStrategyResponse, error) {
	panic("implement me")
}
func (s *DBFakeService) UpdateBackupRecord(ctx context.Context, in *dbpb.DBUpdateBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBUpdateBackupRecordResponse, error) {
	panic("implement me")
}

// Mock Tiup Task
func (s *DBFakeService) CreateTiupTask(ctx context.Context, in *dbpb.CreateTiupTaskRequest, opts ...client.CallOption) (*dbpb.CreateTiupTaskResponse, error) {
	return s.mockCreateTiupTask(ctx, in, opts...)
}
func (s *DBFakeService) UpdateTiupTask(ctx context.Context, in *dbpb.UpdateTiupTaskRequest, opts ...client.CallOption) (*dbpb.UpdateTiupTaskResponse, error) {
	return s.mockUpdateTiupTask(ctx, in, opts...)
}
func (s *DBFakeService) FindTiupTaskByID(ctx context.Context, in *dbpb.FindTiupTaskByIDRequest, opts ...client.CallOption) (*dbpb.FindTiupTaskByIDResponse, error) {
	return s.mockFindTiupTaskByID(ctx, in, opts...)
}
func (s *DBFakeService) GetTiupTaskStatusByBizID(ctx context.Context, in *dbpb.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*dbpb.GetTiupTaskStatusByBizIDResponse, error) {
	return s.mockGetTiupTaskStatusByBizID(ctx, in, opts...)
}

// Mock Workflow and Task
func (s *DBFakeService) CreateFlow(ctx context.Context, in *dbpb.DBCreateFlowRequest, opts ...client.CallOption) (*dbpb.DBCreateFlowResponse, error) {
	return s.mockCreateFlow(ctx, in, opts...)
}
func (s *DBFakeService) CreateTask(ctx context.Context, in *dbpb.DBCreateTaskRequest, opts ...client.CallOption) (*dbpb.DBCreateTaskResponse, error) {
	return s.mockCreateTask(ctx, in, opts...)
}
func (s *DBFakeService) UpdateFlow(ctx context.Context, in *dbpb.DBUpdateFlowRequest, opts ...client.CallOption) (*dbpb.DBUpdateFlowResponse, error) {
	return s.mockUpdateFlow(ctx, in, opts...)
}
func (s *DBFakeService) UpdateTask(ctx context.Context, in *dbpb.DBUpdateTaskRequest, opts ...client.CallOption) (*dbpb.DBUpdateTaskResponse, error) {
	return s.mockUpdateTask(ctx, in, opts...)
}
func (s *DBFakeService) LoadFlow(ctx context.Context, in *dbpb.DBLoadFlowRequest, opts ...client.CallOption) (*dbpb.DBLoadFlowResponse, error) {
	return s.mockLoadFlow(ctx, in, opts...)
}
func (s *DBFakeService) LoadTask(ctx context.Context, in *dbpb.DBLoadTaskRequest, opts ...client.CallOption) (*dbpb.DBLoadTaskResponse, error) {
	return s.mockLoadTask(ctx, in, opts...)
}

func InitMockDBClient() *DBFakeService {

	fakeDBClient := &DBFakeService{
		mockFindTenant: func(ctx context.Context, in *dbpb.DBFindTenantRequest, opts ...client.CallOption) (*dbpb.DBFindTenantResponse, error) {
			return nil, nil
		},
		mockFindAccount: func(ctx context.Context, in *dbpb.DBFindAccountRequest, opts ...client.CallOption) (*dbpb.DBFindAccountResponse, error) {
			return nil, nil
		},
		mockSaveToken: func(ctx context.Context, in *dbpb.DBSaveTokenRequest, opts ...client.CallOption) (*dbpb.DBSaveTokenResponse, error) {
			return nil, nil
		},
		mockFindToken: func(ctx context.Context, in *dbpb.DBFindTokenRequest, opts ...client.CallOption) (*dbpb.DBFindTokenResponse, error) {
			return nil, nil
		},
		mockFindRolesByPermission: func(ctx context.Context, in *dbpb.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*dbpb.DBFindRolesByPermissionResponse, error) {
			return nil, nil
		},
		mockAddHost: func(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error) {
			return nil, nil
		},
		mockAddHostsInBatch: func(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error) {
			return nil, nil
		},
		mockRemoveHost: func(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error) {
			return nil, nil
		},
		mockRemoveHostsInBatch: func(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		mockListHost: func(ctx context.Context, in *dbpb.DBListHostsRequest, opts ...client.CallOption) (*dbpb.DBListHostsResponse, error) {
			return nil, nil
		},
		mockCheckDetails: func(ctx context.Context, in *dbpb.DBCheckDetailsRequest, opts ...client.CallOption) (*dbpb.DBCheckDetailsResponse, error) {
			return nil, nil
		},
		mockAllocHosts: func(ctx context.Context, in *dbpb.DBAllocHostsRequest, opts ...client.CallOption) (*dbpb.DBAllocHostsResponse, error) {
			return nil, nil
		},
		mockGetFailureDomain: func(ctx context.Context, in *dbpb.DBGetFailureDomainRequest, opts ...client.CallOption) (*dbpb.DBGetFailureDomainResponse, error) {
			return nil, nil
		},
		mockCreateCluster: func(ctx context.Context, in *dbpb.DBCreateClusterRequest, opts ...client.CallOption) (*dbpb.DBCreateClusterResponse, error) {
			return nil, nil
		},
		mockDeleteCluster: func(ctx context.Context, in *dbpb.DBDeleteClusterRequest, opts ...client.CallOption) (*dbpb.DBDeleteClusterResponse, error) {
			return nil, nil
		},
		mockUpdateClusterStatus: func(ctx context.Context, in *dbpb.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*dbpb.DBUpdateClusterStatusResponse, error) {
			return nil, nil
		},
		mockUpdateClusterTopologyConfig: func(ctx context.Context, in *dbpb.DBUpdateTopologyConfigRequest, opts ...client.CallOption) (*dbpb.DBUpdateTopologyConfigResponse, error) {
			return nil, nil
		},
		mockLoadCluster: func(ctx context.Context, in *dbpb.DBLoadClusterRequest, opts ...client.CallOption) (*dbpb.DBLoadClusterResponse, error) {
			return nil, nil
		},
		mockListCluster: func(ctx context.Context, in *dbpb.DBListClusterRequest, opts ...client.CallOption) (*dbpb.DBListClusterResponse, error) {
			return nil, nil
		},
		mockSaveBackupRecord: func(ctx context.Context, in *dbpb.DBSaveBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBSaveBackupRecordResponse, error) {
			return nil, nil
		},
		mockDeleteBackupRecord: func(ctx context.Context, in *dbpb.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*dbpb.DBDeleteBackupRecordResponse, error) {
			return nil, nil
		},
		mockListBackupRecords: func(ctx context.Context, in *dbpb.DBListBackupRecordsRequest, opts ...client.CallOption) (*dbpb.DBListBackupRecordsResponse, error) {
			return nil, nil
		},
		mockSaveRecoverRecord: func(ctx context.Context, in *dbpb.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*dbpb.DBSaveRecoverRecordResponse, error) {
			return nil, nil
		},
		mockSaveParametersRecord: func(ctx context.Context, in *dbpb.DBSaveParametersRequest, opts ...client.CallOption) (*dbpb.DBSaveParametersResponse, error) {
			return nil, nil
		},
		mockGetCurrentParametersRecord: func(ctx context.Context, in *dbpb.DBGetCurrentParametersRequest, opts ...client.CallOption) (*dbpb.DBGetCurrentParametersResponse, error) {
			return nil, nil
		},
		mockCreateTiupTask: func(ctx context.Context, in *dbpb.CreateTiupTaskRequest, opts ...client.CallOption) (*dbpb.CreateTiupTaskResponse, error) {
			return nil, nil
		},
		mockUpdateTiupTask: func(ctx context.Context, in *dbpb.UpdateTiupTaskRequest, opts ...client.CallOption) (*dbpb.UpdateTiupTaskResponse, error) {
			return nil, nil
		},
		mockFindTiupTaskByID: func(ctx context.Context, in *dbpb.FindTiupTaskByIDRequest, opts ...client.CallOption) (*dbpb.FindTiupTaskByIDResponse, error) {
			return nil, nil
		},
		mockGetTiupTaskStatusByBizID: func(ctx context.Context, in *dbpb.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*dbpb.GetTiupTaskStatusByBizIDResponse, error) {
			return nil, nil
		},
		mockCreateFlow: func(ctx context.Context, in *dbpb.DBCreateFlowRequest, opts ...client.CallOption) (*dbpb.DBCreateFlowResponse, error) {
			return nil, nil
		},
		mockCreateTask: func(ctx context.Context, in *dbpb.DBCreateTaskRequest, opts ...client.CallOption) (*dbpb.DBCreateTaskResponse, error) {
			return nil, nil
		},
		mockUpdateFlow: func(ctx context.Context, in *dbpb.DBUpdateFlowRequest, opts ...client.CallOption) (*dbpb.DBUpdateFlowResponse, error) {
			return nil, nil
		},
		mockUpdateTask: func(ctx context.Context, in *dbpb.DBUpdateTaskRequest, opts ...client.CallOption) (*dbpb.DBUpdateTaskResponse, error) {
			return nil, nil
		},
		mockLoadFlow: func(ctx context.Context, in *dbpb.DBLoadFlowRequest, opts ...client.CallOption) (*dbpb.DBLoadFlowResponse, error) {
			return nil, nil
		},
		mockLoadTask: func(ctx context.Context, in *dbpb.DBLoadTaskRequest, opts ...client.CallOption) (*dbpb.DBLoadTaskResponse, error) {
			return nil, nil
		},
	}

	rpc_client.DBClient = fakeDBClient

	return fakeDBClient
}

func (s *DBFakeService) MockAddHost(mock func(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error)) {
	s.mockAddHost = mock
}

func (s *DBFakeService) MockAddHostsInBatch(mock func(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error)) {
	s.mockAddHostsInBatch = mock
}

func (s *DBFakeService) MockRemoveHost(mock func(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error)) {
	s.mockRemoveHost = mock
}

func (s *DBFakeService) MockRemoveHostsInBatch(mock func(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error)) {
	s.mockRemoveHostsInBatch = mock
}

func (s *DBFakeService) MockListHost(mock func(ctx context.Context, in *dbpb.DBListHostsRequest, opts ...client.CallOption) (*dbpb.DBListHostsResponse, error)) {
	s.mockListHost = mock
}

func (s *DBFakeService) MockCheckDetails(mock func(ctx context.Context, in *dbpb.DBCheckDetailsRequest, opts ...client.CallOption) (*dbpb.DBCheckDetailsResponse, error)) {
	s.mockCheckDetails = mock
}

func (s *DBFakeService) MockAllocHosts(mock func(ctx context.Context, in *dbpb.DBAllocHostsRequest, opts ...client.CallOption) (*dbpb.DBAllocHostsResponse, error)) {
	s.mockAllocHosts = mock
}

func (s *DBFakeService) MockGetFailureDomain(mock func(ctx context.Context, in *dbpb.DBGetFailureDomainRequest, opts ...client.CallOption) (*dbpb.DBGetFailureDomainResponse, error)) {
	s.mockGetFailureDomain = mock
}

func (s *DBFakeService) MockAllocResources(mock func(ctx context.Context, in *dbpb.DBAllocRequest, opts ...client.CallOption) (*dbpb.DBAllocResponse, error)) {
	s.mockAllocResources = mock
}

func (s *DBFakeService) MockAllocResourcesInBatch(mock func(ctx context.Context, in *dbpb.DBBatchAllocRequest, opts ...client.CallOption) (*dbpb.DBBatchAllocResponse, error)) {
	s.mockAllocResourcesInBatch = mock
}

func (s *DBFakeService) MockRecycleResources(mock func(ctx context.Context, in *dbpb.DBRecycleRequest, opts ...client.CallOption) (*dbpb.DBRecycleResponse, error)) {
	s.mockRecycleResources = mock
}
