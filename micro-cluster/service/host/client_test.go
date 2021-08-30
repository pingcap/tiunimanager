package host

import (
	"context"

	"github.com/asim/go-micro/v3/client"
	dbclient "github.com/pingcap-inc/tiem/library/firstparty/client"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

type DBFakeService struct {
	// Auth Module
	mockFindTenant            func(ctx context.Context, in *db.DBFindTenantRequest, opts ...client.CallOption) (*db.DBFindTenantResponse, error)
	mockFindAccount           func(ctx context.Context, in *db.DBFindAccountRequest, opts ...client.CallOption) (*db.DBFindAccountResponse, error)
	mockSaveToken             func(ctx context.Context, in *db.DBSaveTokenRequest, opts ...client.CallOption) (*db.DBSaveTokenResponse, error)
	mockFindToken             func(ctx context.Context, in *db.DBFindTokenRequest, opts ...client.CallOption) (*db.DBFindTokenResponse, error)
	mockFindRolesByPermission func(ctx context.Context, in *db.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*db.DBFindRolesByPermissionResponse, error)
	// Host Module
	mockAddHost            func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error)
	mockAddHostsInBatch    func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error)
	mockRemoveHost         func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error)
	mockRemoveHostsInBatch func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error)
	mockListHost           func(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error)
	mockCheckDetails       func(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error)
	mockAllocHosts         func(ctx context.Context, in *db.DBAllocHostsRequest, opts ...client.CallOption) (*db.DBAllocHostsResponse, error)
	mockGetFailureDomain   func(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error)
	// Cluster
	mockCreateCluster           func(ctx context.Context, in *db.DBCreateClusterRequest, opts ...client.CallOption) (*db.DBCreateClusterResponse, error)
	mockDeleteCluster           func(ctx context.Context, in *db.DBDeleteClusterRequest, opts ...client.CallOption) (*db.DBDeleteClusterResponse, error)
	mockUpdateClusterStatus     func(ctx context.Context, in *db.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*db.DBUpdateClusterStatusResponse, error)
	mockUpdateClusterTiupConfig func(ctx context.Context, in *db.DBUpdateTiupConfigRequest, opts ...client.CallOption) (*db.DBUpdateTiupConfigResponse, error)
	mockLoadCluster             func(ctx context.Context, in *db.DBLoadClusterRequest, opts ...client.CallOption) (*db.DBLoadClusterResponse, error)
	mockListCluster             func(ctx context.Context, in *db.DBListClusterRequest, opts ...client.CallOption) (*db.DBListClusterResponse, error)
	// backup & recover & parameters
	mockSaveBackupRecord           func(ctx context.Context, in *db.DBSaveBackupRecordRequest, opts ...client.CallOption) (*db.DBSaveBackupRecordResponse, error)
	mockDeleteBackupRecord         func(ctx context.Context, in *db.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*db.DBDeleteBackupRecordResponse, error)
	mockListBackupRecords          func(ctx context.Context, in *db.DBListBackupRecordsRequest, opts ...client.CallOption) (*db.DBListBackupRecordsResponse, error)
	mockSaveRecoverRecord          func(ctx context.Context, in *db.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*db.DBSaveRecoverRecordResponse, error)
	mockSaveParametersRecord       func(ctx context.Context, in *db.DBSaveParametersRequest, opts ...client.CallOption) (*db.DBSaveParametersResponse, error)
	mockGetCurrentParametersRecord func(ctx context.Context, in *db.DBGetCurrentParametersRequest, opts ...client.CallOption) (*db.DBGetCurrentParametersResponse, error)
	// Tiup Task
	mockCreateTiupTask           func(ctx context.Context, in *db.CreateTiupTaskRequest, opts ...client.CallOption) (*db.CreateTiupTaskResponse, error)
	mockUpdateTiupTask           func(ctx context.Context, in *db.UpdateTiupTaskRequest, opts ...client.CallOption) (*db.UpdateTiupTaskResponse, error)
	mockFindTiupTaskByID         func(ctx context.Context, in *db.FindTiupTaskByIDRequest, opts ...client.CallOption) (*db.FindTiupTaskByIDResponse, error)
	mockGetTiupTaskStatusByBizID func(ctx context.Context, in *db.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*db.GetTiupTaskStatusByBizIDResponse, error)
	// Workflow and Task
	mockCreateFlow func(ctx context.Context, in *db.DBCreateFlowRequest, opts ...client.CallOption) (*db.DBCreateFlowResponse, error)
	mockCreateTask func(ctx context.Context, in *db.DBCreateTaskRequest, opts ...client.CallOption) (*db.DBCreateTaskResponse, error)
	mockUpdateFlow func(ctx context.Context, in *db.DBUpdateFlowRequest, opts ...client.CallOption) (*db.DBUpdateFlowResponse, error)
	mockUpdateTask func(ctx context.Context, in *db.DBUpdateTaskRequest, opts ...client.CallOption) (*db.DBUpdateTaskResponse, error)
	mockLoadFlow   func(ctx context.Context, in *db.DBLoadFlowRequest, opts ...client.CallOption) (*db.DBLoadFlowResponse, error)
	mockLoadTask   func(ctx context.Context, in *db.DBLoadTaskRequest, opts ...client.CallOption) (*db.DBLoadTaskResponse, error)
}

// Mock Auth Module
func (s *DBFakeService) FindTenant(ctx context.Context, in *db.DBFindTenantRequest, opts ...client.CallOption) (*db.DBFindTenantResponse, error) {
	return s.mockFindTenant(ctx, in, opts...)
}
func (s *DBFakeService) FindAccount(ctx context.Context, in *db.DBFindAccountRequest, opts ...client.CallOption) (*db.DBFindAccountResponse, error) {
	return s.mockFindAccount(ctx, in, opts...)
}
func (s *DBFakeService) SaveToken(ctx context.Context, in *db.DBSaveTokenRequest, opts ...client.CallOption) (*db.DBSaveTokenResponse, error) {
	return s.mockSaveToken(ctx, in, opts...)
}
func (s *DBFakeService) FindToken(ctx context.Context, in *db.DBFindTokenRequest, opts ...client.CallOption) (*db.DBFindTokenResponse, error) {
	return s.mockFindToken(ctx, in, opts...)
}
func (s *DBFakeService) FindRolesByPermission(ctx context.Context, in *db.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*db.DBFindRolesByPermissionResponse, error) {
	return s.mockFindRolesByPermission(ctx, in, opts...)
}

// Mock Host Module
func (s *DBFakeService) AddHost(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
	return s.mockAddHost(ctx, in, opts...)
}
func (s *DBFakeService) AddHostsInBatch(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
	return s.mockAddHostsInBatch(ctx, in, opts...)
}
func (s *DBFakeService) RemoveHost(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
	return s.mockRemoveHost(ctx, in, opts...)
}
func (s *DBFakeService) RemoveHostsInBatch(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
	return s.mockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *DBFakeService) ListHost(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error) {
	return s.mockListHost(ctx, in, opts...)
}
func (s *DBFakeService) CheckDetails(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error) {
	return s.mockCheckDetails(ctx, in, opts...)
}
func (s *DBFakeService) AllocHosts(ctx context.Context, in *db.DBAllocHostsRequest, opts ...client.CallOption) (*db.DBAllocHostsResponse, error) {
	return s.mockAllocHosts(ctx, in, opts...)
}
func (s *DBFakeService) GetFailureDomain(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error) {
	return s.mockGetFailureDomain(ctx, in, opts...)
}

// Mock Cluster
func (s *DBFakeService) CreateCluster(ctx context.Context, in *db.DBCreateClusterRequest, opts ...client.CallOption) (*db.DBCreateClusterResponse, error) {
	return s.mockCreateCluster(ctx, in, opts...)
}
func (s *DBFakeService) DeleteCluster(ctx context.Context, in *db.DBDeleteClusterRequest, opts ...client.CallOption) (*db.DBDeleteClusterResponse, error) {
	return s.mockDeleteCluster(ctx, in, opts...)
}
func (s *DBFakeService) UpdateClusterStatus(ctx context.Context, in *db.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*db.DBUpdateClusterStatusResponse, error) {
	return s.mockUpdateClusterStatus(ctx, in, opts...)
}
func (s *DBFakeService) UpdateClusterTiupConfig(ctx context.Context, in *db.DBUpdateTiupConfigRequest, opts ...client.CallOption) (*db.DBUpdateTiupConfigResponse, error) {
	return s.mockUpdateClusterTiupConfig(ctx, in, opts...)
}
func (s *DBFakeService) LoadCluster(ctx context.Context, in *db.DBLoadClusterRequest, opts ...client.CallOption) (*db.DBLoadClusterResponse, error) {
	return s.mockLoadCluster(ctx, in, opts...)
}
func (s *DBFakeService) ListCluster(ctx context.Context, in *db.DBListClusterRequest, opts ...client.CallOption) (*db.DBListClusterResponse, error) {
	return s.mockListCluster(ctx, in, opts...)
}

// Mock backup & recover & parameters
func (s *DBFakeService) SaveBackupRecord(ctx context.Context, in *db.DBSaveBackupRecordRequest, opts ...client.CallOption) (*db.DBSaveBackupRecordResponse, error) {
	return s.mockSaveBackupRecord(ctx, in, opts...)
}
func (s *DBFakeService) DeleteBackupRecord(ctx context.Context, in *db.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*db.DBDeleteBackupRecordResponse, error) {
	return s.mockDeleteBackupRecord(ctx, in, opts...)
}
func (s *DBFakeService) ListBackupRecords(ctx context.Context, in *db.DBListBackupRecordsRequest, opts ...client.CallOption) (*db.DBListBackupRecordsResponse, error) {
	return s.mockListBackupRecords(ctx, in, opts...)
}
func (s *DBFakeService) SaveRecoverRecord(ctx context.Context, in *db.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*db.DBSaveRecoverRecordResponse, error) {
	return s.mockSaveRecoverRecord(ctx, in, opts...)
}
func (s *DBFakeService) SaveParametersRecord(ctx context.Context, in *db.DBSaveParametersRequest, opts ...client.CallOption) (*db.DBSaveParametersResponse, error) {
	return s.mockSaveParametersRecord(ctx, in, opts...)
}
func (s *DBFakeService) GetCurrentParametersRecord(ctx context.Context, in *db.DBGetCurrentParametersRequest, opts ...client.CallOption) (*db.DBGetCurrentParametersResponse, error) {
	return s.mockGetCurrentParametersRecord(ctx, in, opts...)
}

// Mock Tiup Task
func (s *DBFakeService) CreateTiupTask(ctx context.Context, in *db.CreateTiupTaskRequest, opts ...client.CallOption) (*db.CreateTiupTaskResponse, error) {
	return s.mockCreateTiupTask(ctx, in, opts...)
}
func (s *DBFakeService) UpdateTiupTask(ctx context.Context, in *db.UpdateTiupTaskRequest, opts ...client.CallOption) (*db.UpdateTiupTaskResponse, error) {
	return s.mockUpdateTiupTask(ctx, in, opts...)
}
func (s *DBFakeService) FindTiupTaskByID(ctx context.Context, in *db.FindTiupTaskByIDRequest, opts ...client.CallOption) (*db.FindTiupTaskByIDResponse, error) {
	return s.mockFindTiupTaskByID(ctx, in, opts...)
}
func (s *DBFakeService) GetTiupTaskStatusByBizID(ctx context.Context, in *db.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*db.GetTiupTaskStatusByBizIDResponse, error) {
	return s.mockGetTiupTaskStatusByBizID(ctx, in, opts...)
}

// Mock Workflow and Task
func (s *DBFakeService) CreateFlow(ctx context.Context, in *db.DBCreateFlowRequest, opts ...client.CallOption) (*db.DBCreateFlowResponse, error) {
	return s.mockCreateFlow(ctx, in, opts...)
}
func (s *DBFakeService) CreateTask(ctx context.Context, in *db.DBCreateTaskRequest, opts ...client.CallOption) (*db.DBCreateTaskResponse, error) {
	return s.mockCreateTask(ctx, in, opts...)
}
func (s *DBFakeService) UpdateFlow(ctx context.Context, in *db.DBUpdateFlowRequest, opts ...client.CallOption) (*db.DBUpdateFlowResponse, error) {
	return s.mockUpdateFlow(ctx, in, opts...)
}
func (s *DBFakeService) UpdateTask(ctx context.Context, in *db.DBUpdateTaskRequest, opts ...client.CallOption) (*db.DBUpdateTaskResponse, error) {
	return s.mockUpdateTask(ctx, in, opts...)
}
func (s *DBFakeService) LoadFlow(ctx context.Context, in *db.DBLoadFlowRequest, opts ...client.CallOption) (*db.DBLoadFlowResponse, error) {
	return s.mockLoadFlow(ctx, in, opts...)
}
func (s *DBFakeService) LoadTask(ctx context.Context, in *db.DBLoadTaskRequest, opts ...client.CallOption) (*db.DBLoadTaskResponse, error) {
	return s.mockLoadTask(ctx, in, opts...)
}

func InitMockDBClient() *DBFakeService {

	fakeDBClient := &DBFakeService{
		mockFindTenant: func(ctx context.Context, in *db.DBFindTenantRequest, opts ...client.CallOption) (*db.DBFindTenantResponse, error) {
			return nil, nil
		},
		mockFindAccount: func(ctx context.Context, in *db.DBFindAccountRequest, opts ...client.CallOption) (*db.DBFindAccountResponse, error) {
			return nil, nil
		},
		mockSaveToken: func(ctx context.Context, in *db.DBSaveTokenRequest, opts ...client.CallOption) (*db.DBSaveTokenResponse, error) {
			return nil, nil
		},
		mockFindToken: func(ctx context.Context, in *db.DBFindTokenRequest, opts ...client.CallOption) (*db.DBFindTokenResponse, error) {
			return nil, nil
		},
		mockFindRolesByPermission: func(ctx context.Context, in *db.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*db.DBFindRolesByPermissionResponse, error) {
			return nil, nil
		},
		mockAddHost: func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
			return nil, nil
		},
		mockAddHostsInBatch: func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
			return nil, nil
		},
		mockRemoveHost: func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
			return nil, nil
		},
		mockRemoveHostsInBatch: func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		mockListHost: func(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error) {
			return nil, nil
		},
		mockCheckDetails: func(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error) {
			return nil, nil
		},
		mockAllocHosts: func(ctx context.Context, in *db.DBAllocHostsRequest, opts ...client.CallOption) (*db.DBAllocHostsResponse, error) {
			return nil, nil
		},
		mockGetFailureDomain: func(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error) {
			return nil, nil
		},
		mockCreateCluster: func(ctx context.Context, in *db.DBCreateClusterRequest, opts ...client.CallOption) (*db.DBCreateClusterResponse, error) {
			return nil, nil
		},
		mockDeleteCluster: func(ctx context.Context, in *db.DBDeleteClusterRequest, opts ...client.CallOption) (*db.DBDeleteClusterResponse, error) {
			return nil, nil
		},
		mockUpdateClusterStatus: func(ctx context.Context, in *db.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*db.DBUpdateClusterStatusResponse, error) {
			return nil, nil
		},
		mockUpdateClusterTiupConfig: func(ctx context.Context, in *db.DBUpdateTiupConfigRequest, opts ...client.CallOption) (*db.DBUpdateTiupConfigResponse, error) {
			return nil, nil
		},
		mockLoadCluster: func(ctx context.Context, in *db.DBLoadClusterRequest, opts ...client.CallOption) (*db.DBLoadClusterResponse, error) {
			return nil, nil
		},
		mockListCluster: func(ctx context.Context, in *db.DBListClusterRequest, opts ...client.CallOption) (*db.DBListClusterResponse, error) {
			return nil, nil
		},
		mockSaveBackupRecord: func(ctx context.Context, in *db.DBSaveBackupRecordRequest, opts ...client.CallOption) (*db.DBSaveBackupRecordResponse, error) {
			return nil, nil
		},
		mockDeleteBackupRecord: func(ctx context.Context, in *db.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*db.DBDeleteBackupRecordResponse, error) {
			return nil, nil
		},
		mockListBackupRecords: func(ctx context.Context, in *db.DBListBackupRecordsRequest, opts ...client.CallOption) (*db.DBListBackupRecordsResponse, error) {
			return nil, nil
		},
		mockSaveRecoverRecord: func(ctx context.Context, in *db.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*db.DBSaveRecoverRecordResponse, error) {
			return nil, nil
		},
		mockSaveParametersRecord: func(ctx context.Context, in *db.DBSaveParametersRequest, opts ...client.CallOption) (*db.DBSaveParametersResponse, error) {
			return nil, nil
		},
		mockGetCurrentParametersRecord: func(ctx context.Context, in *db.DBGetCurrentParametersRequest, opts ...client.CallOption) (*db.DBGetCurrentParametersResponse, error) {
			return nil, nil
		},
		mockCreateTiupTask: func(ctx context.Context, in *db.CreateTiupTaskRequest, opts ...client.CallOption) (*db.CreateTiupTaskResponse, error) {
			return nil, nil
		},
		mockUpdateTiupTask: func(ctx context.Context, in *db.UpdateTiupTaskRequest, opts ...client.CallOption) (*db.UpdateTiupTaskResponse, error) {
			return nil, nil
		},
		mockFindTiupTaskByID: func(ctx context.Context, in *db.FindTiupTaskByIDRequest, opts ...client.CallOption) (*db.FindTiupTaskByIDResponse, error) {
			return nil, nil
		},
		mockGetTiupTaskStatusByBizID: func(ctx context.Context, in *db.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*db.GetTiupTaskStatusByBizIDResponse, error) {
			return nil, nil
		},
		mockCreateFlow: func(ctx context.Context, in *db.DBCreateFlowRequest, opts ...client.CallOption) (*db.DBCreateFlowResponse, error) {
			return nil, nil
		},
		mockCreateTask: func(ctx context.Context, in *db.DBCreateTaskRequest, opts ...client.CallOption) (*db.DBCreateTaskResponse, error) {
			return nil, nil
		},
		mockUpdateFlow: func(ctx context.Context, in *db.DBUpdateFlowRequest, opts ...client.CallOption) (*db.DBUpdateFlowResponse, error) {
			return nil, nil
		},
		mockUpdateTask: func(ctx context.Context, in *db.DBUpdateTaskRequest, opts ...client.CallOption) (*db.DBUpdateTaskResponse, error) {
			return nil, nil
		},
		mockLoadFlow: func(ctx context.Context, in *db.DBLoadFlowRequest, opts ...client.CallOption) (*db.DBLoadFlowResponse, error) {
			return nil, nil
		},
		mockLoadTask: func(ctx context.Context, in *db.DBLoadTaskRequest, opts ...client.CallOption) (*db.DBLoadTaskResponse, error) {
			return nil, nil
		},
	}

	dbclient.DBClient = fakeDBClient

	return fakeDBClient
}

func (s *DBFakeService) MockAddHost(mock func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error)) {
	s.mockAddHost = mock
}

func (s *DBFakeService) MockAddHostsInBatch(mock func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error)) {
	s.mockAddHostsInBatch = mock
}

func (s *DBFakeService) MockRemoveHost(mock func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error)) {
	s.mockRemoveHost = mock
}

func (s *DBFakeService) MockRemoveHostsInBatch(mock func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error)) {
	s.mockRemoveHostsInBatch = mock
}

func (s *DBFakeService) MockListHost(mock func(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error)) {
	s.mockListHost = mock
}

func (s *DBFakeService) MockCheckDetails(mock func(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error)) {
	s.mockCheckDetails = mock
}

func (s *DBFakeService) MockAllocHosts(mock func(ctx context.Context, in *db.DBAllocHostsRequest, opts ...client.CallOption) (*db.DBAllocHostsResponse, error)) {
	s.mockAllocHosts = mock
}

func (s *DBFakeService) MockGetFailureDomain(mock func(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error)) {
	s.mockGetFailureDomain = mock
}
