package client

import (
	"context"
	"fmt"

	"github.com/asim/go-micro/v3/client"
	db "github.com/pingcap/tiem/micro-metadb/proto"
)

type DBFakeService struct {
	// Auth Module
	MockFindTenant            func(ctx context.Context, in *db.DBFindTenantRequest, opts ...client.CallOption) (*db.DBFindTenantResponse, error)
	MockFindAccount           func(ctx context.Context, in *db.DBFindAccountRequest, opts ...client.CallOption) (*db.DBFindAccountResponse, error)
	MockSaveToken             func(ctx context.Context, in *db.DBSaveTokenRequest, opts ...client.CallOption) (*db.DBSaveTokenResponse, error)
	MockFindToken             func(ctx context.Context, in *db.DBFindTokenRequest, opts ...client.CallOption) (*db.DBFindTokenResponse, error)
	MockFindRolesByPermission func(ctx context.Context, in *db.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*db.DBFindRolesByPermissionResponse, error)
	// Host Module
	MockAddHost            func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error)
	MockAddHostsInBatch    func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error)
	MockRemoveHost         func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error)
	MockRemoveHostsInBatch func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error)
	MockListHost           func(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error)
	MockCheckDetails       func(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error)
	MockPreAllocHosts      func(ctx context.Context, in *db.DBPreAllocHostsRequest, opts ...client.CallOption) (*db.DBPreAllocHostsResponse, error)
	MockLockHosts          func(ctx context.Context, in *db.DBLockHostsRequest, opts ...client.CallOption) (*db.DBLockHostsResponse, error)
	MockGetFailureDomain   func(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error)
	// Cluster
	MockCreateCluster           func(ctx context.Context, in *db.DBCreateClusterRequest, opts ...client.CallOption) (*db.DBCreateClusterResponse, error)
	MockDeleteCluster           func(ctx context.Context, in *db.DBDeleteClusterRequest, opts ...client.CallOption) (*db.DBDeleteClusterResponse, error)
	MockUpdateClusterStatus     func(ctx context.Context, in *db.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*db.DBUpdateClusterStatusResponse, error)
	MockUpdateClusterTiupConfig func(ctx context.Context, in *db.DBUpdateTiupConfigRequest, opts ...client.CallOption) (*db.DBUpdateTiupConfigResponse, error)
	MockLoadCluster             func(ctx context.Context, in *db.DBLoadClusterRequest, opts ...client.CallOption) (*db.DBLoadClusterResponse, error)
	MockListCluster             func(ctx context.Context, in *db.DBListClusterRequest, opts ...client.CallOption) (*db.DBListClusterResponse, error)
	// backup & recover & parameters
	MockSaveBackupRecord           func(ctx context.Context, in *db.DBSaveBackupRecordRequest, opts ...client.CallOption) (*db.DBSaveBackupRecordResponse, error)
	MockDeleteBackupRecord         func(ctx context.Context, in *db.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*db.DBDeleteBackupRecordResponse, error)
	MockListBackupRecords          func(ctx context.Context, in *db.DBListBackupRecordsRequest, opts ...client.CallOption) (*db.DBListBackupRecordsResponse, error)
	MockSaveRecoverRecord          func(ctx context.Context, in *db.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*db.DBSaveRecoverRecordResponse, error)
	MockSaveParametersRecord       func(ctx context.Context, in *db.DBSaveParametersRequest, opts ...client.CallOption) (*db.DBSaveParametersResponse, error)
	MockGetCurrentParametersRecord func(ctx context.Context, in *db.DBGetCurrentParametersRequest, opts ...client.CallOption) (*db.DBGetCurrentParametersResponse, error)
	// Tiup Task
	MockCreateTiupTask           func(ctx context.Context, in *db.CreateTiupTaskRequest, opts ...client.CallOption) (*db.CreateTiupTaskResponse, error)
	MockUpdateTiupTask           func(ctx context.Context, in *db.UpdateTiupTaskRequest, opts ...client.CallOption) (*db.UpdateTiupTaskResponse, error)
	MockFindTiupTaskByID         func(ctx context.Context, in *db.FindTiupTaskByIDRequest, opts ...client.CallOption) (*db.FindTiupTaskByIDResponse, error)
	MockGetTiupTaskStatusByBizID func(ctx context.Context, in *db.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*db.GetTiupTaskStatusByBizIDResponse, error)
	// Workflow and Task
	MockCreateFlow func(ctx context.Context, in *db.DBCreateFlowRequest, opts ...client.CallOption) (*db.DBCreateFlowResponse, error)
	MockCreateTask func(ctx context.Context, in *db.DBCreateTaskRequest, opts ...client.CallOption) (*db.DBCreateTaskResponse, error)
	MockUpdateFlow func(ctx context.Context, in *db.DBUpdateFlowRequest, opts ...client.CallOption) (*db.DBUpdateFlowResponse, error)
	MockUpdateTask func(ctx context.Context, in *db.DBUpdateTaskRequest, opts ...client.CallOption) (*db.DBUpdateTaskResponse, error)
	MockLoadFlow   func(ctx context.Context, in *db.DBLoadFlowRequest, opts ...client.CallOption) (*db.DBLoadFlowResponse, error)
	MockLoadTask   func(ctx context.Context, in *db.DBLoadTaskRequest, opts ...client.CallOption) (*db.DBLoadTaskResponse, error)
}

// Mock Auth Module
func (s *DBFakeService) FindTenant(ctx context.Context, in *db.DBFindTenantRequest, opts ...client.CallOption) (*db.DBFindTenantResponse, error) {
	return s.MockFindTenant(ctx, in, opts...)
}
func (s *DBFakeService) FindAccount(ctx context.Context, in *db.DBFindAccountRequest, opts ...client.CallOption) (*db.DBFindAccountResponse, error) {
	return s.MockFindAccount(ctx, in, opts...)
}
func (s *DBFakeService) SaveToken(ctx context.Context, in *db.DBSaveTokenRequest, opts ...client.CallOption) (*db.DBSaveTokenResponse, error) {
	return s.MockSaveToken(ctx, in, opts...)
}
func (s *DBFakeService) FindToken(ctx context.Context, in *db.DBFindTokenRequest, opts ...client.CallOption) (*db.DBFindTokenResponse, error) {
	return s.MockFindToken(ctx, in, opts...)
}
func (s *DBFakeService) FindRolesByPermission(ctx context.Context, in *db.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*db.DBFindRolesByPermissionResponse, error) {
	return s.MockFindRolesByPermission(ctx, in, opts...)
}

// Mock Host Module
func (s *DBFakeService) AddHost(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
	return s.MockAddHost(ctx, in, opts...)
}
func (s *DBFakeService) AddHostsInBatch(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
	return s.MockAddHostsInBatch(ctx, in, opts...)
}
func (s *DBFakeService) RemoveHost(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
	return s.MockRemoveHost(ctx, in, opts...)
}
func (s *DBFakeService) RemoveHostsInBatch(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
	return s.MockRemoveHostsInBatch(ctx, in, opts...)
}
func (s *DBFakeService) ListHost(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error) {
	return s.MockListHost(ctx, in, opts...)
}
func (s *DBFakeService) CheckDetails(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error) {
	return s.MockCheckDetails(ctx, in, opts...)
}
func (s *DBFakeService) PreAllocHosts(ctx context.Context, in *db.DBPreAllocHostsRequest, opts ...client.CallOption) (*db.DBPreAllocHostsResponse, error) {
	return s.MockPreAllocHosts(ctx, in, opts...)
}
func (s *DBFakeService) LockHosts(ctx context.Context, in *db.DBLockHostsRequest, opts ...client.CallOption) (*db.DBLockHostsResponse, error) {
	return s.MockLockHosts(ctx, in, opts...)
}
func (s *DBFakeService) GetFailureDomain(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error) {
	return s.MockGetFailureDomain(ctx, in, opts...)
}

// Mock Cluster
func (s *DBFakeService) CreateCluster(ctx context.Context, in *db.DBCreateClusterRequest, opts ...client.CallOption) (*db.DBCreateClusterResponse, error) {
	return s.MockCreateCluster(ctx, in, opts...)
}
func (s *DBFakeService) DeleteCluster(ctx context.Context, in *db.DBDeleteClusterRequest, opts ...client.CallOption) (*db.DBDeleteClusterResponse, error) {
	return s.MockDeleteCluster(ctx, in, opts...)
}
func (s *DBFakeService) UpdateClusterStatus(ctx context.Context, in *db.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*db.DBUpdateClusterStatusResponse, error) {
	return s.MockUpdateClusterStatus(ctx, in, opts...)
}
func (s *DBFakeService) UpdateClusterTiupConfig(ctx context.Context, in *db.DBUpdateTiupConfigRequest, opts ...client.CallOption) (*db.DBUpdateTiupConfigResponse, error) {
	return s.MockUpdateClusterTiupConfig(ctx, in, opts...)
}
func (s *DBFakeService) LoadCluster(ctx context.Context, in *db.DBLoadClusterRequest, opts ...client.CallOption) (*db.DBLoadClusterResponse, error) {
	return s.MockLoadCluster(ctx, in, opts...)
}
func (s *DBFakeService) ListCluster(ctx context.Context, in *db.DBListClusterRequest, opts ...client.CallOption) (*db.DBListClusterResponse, error) {
	return s.MockListCluster(ctx, in, opts...)
}

// Mock backup & recover & parameters
func (s *DBFakeService) SaveBackupRecord(ctx context.Context, in *db.DBSaveBackupRecordRequest, opts ...client.CallOption) (*db.DBSaveBackupRecordResponse, error) {
	return s.MockSaveBackupRecord(ctx, in, opts...)
}
func (s *DBFakeService) DeleteBackupRecord(ctx context.Context, in *db.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*db.DBDeleteBackupRecordResponse, error) {
	return s.MockDeleteBackupRecord(ctx, in, opts...)
}
func (s *DBFakeService) ListBackupRecords(ctx context.Context, in *db.DBListBackupRecordsRequest, opts ...client.CallOption) (*db.DBListBackupRecordsResponse, error) {
	return s.MockListBackupRecords(ctx, in, opts...)
}
func (s *DBFakeService) SaveRecoverRecord(ctx context.Context, in *db.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*db.DBSaveRecoverRecordResponse, error) {
	return s.MockSaveRecoverRecord(ctx, in, opts...)
}
func (s *DBFakeService) SaveParametersRecord(ctx context.Context, in *db.DBSaveParametersRequest, opts ...client.CallOption) (*db.DBSaveParametersResponse, error) {
	return s.MockSaveParametersRecord(ctx, in, opts...)
}
func (s *DBFakeService) GetCurrentParametersRecord(ctx context.Context, in *db.DBGetCurrentParametersRequest, opts ...client.CallOption) (*db.DBGetCurrentParametersResponse, error) {
	return s.MockGetCurrentParametersRecord(ctx, in, opts...)
}

// Mock Tiup Task
func (s *DBFakeService) CreateTiupTask(ctx context.Context, in *db.CreateTiupTaskRequest, opts ...client.CallOption) (*db.CreateTiupTaskResponse, error) {
	return s.MockCreateTiupTask(ctx, in, opts...)
}
func (s *DBFakeService) UpdateTiupTask(ctx context.Context, in *db.UpdateTiupTaskRequest, opts ...client.CallOption) (*db.UpdateTiupTaskResponse, error) {
	return s.MockUpdateTiupTask(ctx, in, opts...)
}
func (s *DBFakeService) FindTiupTaskByID(ctx context.Context, in *db.FindTiupTaskByIDRequest, opts ...client.CallOption) (*db.FindTiupTaskByIDResponse, error) {
	return s.MockFindTiupTaskByID(ctx, in, opts...)
}
func (s *DBFakeService) GetTiupTaskStatusByBizID(ctx context.Context, in *db.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*db.GetTiupTaskStatusByBizIDResponse, error) {
	return s.MockGetTiupTaskStatusByBizID(ctx, in, opts...)
}

// Mock Workflow and Task
func (s *DBFakeService) CreateFlow(ctx context.Context, in *db.DBCreateFlowRequest, opts ...client.CallOption) (*db.DBCreateFlowResponse, error) {
	return s.MockCreateFlow(ctx, in, opts...)
}
func (s *DBFakeService) CreateTask(ctx context.Context, in *db.DBCreateTaskRequest, opts ...client.CallOption) (*db.DBCreateTaskResponse, error) {
	return s.MockCreateTask(ctx, in, opts...)
}
func (s *DBFakeService) UpdateFlow(ctx context.Context, in *db.DBUpdateFlowRequest, opts ...client.CallOption) (*db.DBUpdateFlowResponse, error) {
	return s.MockUpdateFlow(ctx, in, opts...)
}
func (s *DBFakeService) UpdateTask(ctx context.Context, in *db.DBUpdateTaskRequest, opts ...client.CallOption) (*db.DBUpdateTaskResponse, error) {
	return s.MockUpdateTask(ctx, in, opts...)
}
func (s *DBFakeService) LoadFlow(ctx context.Context, in *db.DBLoadFlowRequest, opts ...client.CallOption) (*db.DBLoadFlowResponse, error) {
	return s.MockLoadFlow(ctx, in, opts...)
}
func (s *DBFakeService) LoadTask(ctx context.Context, in *db.DBLoadTaskRequest, opts ...client.CallOption) (*db.DBLoadTaskResponse, error) {
	return s.MockLoadTask(ctx, in, opts...)
}

func MockDBClient() {

	DBClient = &DBFakeService{
		// Auth Module
		MockFindTenant: func(ctx context.Context, in *db.DBFindTenantRequest, opts ...client.CallOption) (*db.DBFindTenantResponse, error) {
			return nil, nil
		},
		MockFindAccount: func(ctx context.Context, in *db.DBFindAccountRequest, opts ...client.CallOption) (*db.DBFindAccountResponse, error) {
			return nil, nil
		},
		MockSaveToken: func(ctx context.Context, in *db.DBSaveTokenRequest, opts ...client.CallOption) (*db.DBSaveTokenResponse, error) {
			return nil, nil
		},
		MockFindToken: func(ctx context.Context, in *db.DBFindTokenRequest, opts ...client.CallOption) (*db.DBFindTokenResponse, error) {
			return nil, nil
		},
		MockFindRolesByPermission: func(ctx context.Context, in *db.DBFindRolesByPermissionRequest, opts ...client.CallOption) (*db.DBFindRolesByPermissionResponse, error) {
			return nil, nil
		},
		// Host Module
		MockAddHost: func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
			return nil, nil
		},
		MockAddHostsInBatch: func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
			return nil, nil
		},
		MockRemoveHost: func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
			return nil, nil
		},
		MockRemoveHostsInBatch: func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
			return nil, nil
		},
		MockListHost: func(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error) {
			return nil, nil
		},
		MockCheckDetails: func(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error) {
			return nil, nil
		},
		MockPreAllocHosts: func(ctx context.Context, in *db.DBPreAllocHostsRequest, opts ...client.CallOption) (*db.DBPreAllocHostsResponse, error) {
			return nil, nil
		},
		MockLockHosts: func(ctx context.Context, in *db.DBLockHostsRequest, opts ...client.CallOption) (*db.DBLockHostsResponse, error) {
			return nil, nil
		},
		MockGetFailureDomain: func(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error) {
			return nil, nil
		},
		// Cluster
		MockCreateCluster: func(ctx context.Context, in *db.DBCreateClusterRequest, opts ...client.CallOption) (*db.DBCreateClusterResponse, error) {
			return nil, nil
		},
		MockDeleteCluster: func(ctx context.Context, in *db.DBDeleteClusterRequest, opts ...client.CallOption) (*db.DBDeleteClusterResponse, error) {
			return nil, nil
		},
		MockUpdateClusterStatus: func(ctx context.Context, in *db.DBUpdateClusterStatusRequest, opts ...client.CallOption) (*db.DBUpdateClusterStatusResponse, error) {
			return nil, nil
		},
		MockUpdateClusterTiupConfig: func(ctx context.Context, in *db.DBUpdateTiupConfigRequest, opts ...client.CallOption) (*db.DBUpdateTiupConfigResponse, error) {
			return nil, nil
		},
		MockLoadCluster: func(ctx context.Context, in *db.DBLoadClusterRequest, opts ...client.CallOption) (*db.DBLoadClusterResponse, error) {
			return nil, nil
		},
		MockListCluster: func(ctx context.Context, in *db.DBListClusterRequest, opts ...client.CallOption) (*db.DBListClusterResponse, error) {
			return nil, nil
		},
		// backup & recover & parameters
		MockSaveBackupRecord: func(ctx context.Context, in *db.DBSaveBackupRecordRequest, opts ...client.CallOption) (*db.DBSaveBackupRecordResponse, error) {
			return nil, nil
		},
		MockDeleteBackupRecord: func(ctx context.Context, in *db.DBDeleteBackupRecordRequest, opts ...client.CallOption) (*db.DBDeleteBackupRecordResponse, error) {
			return nil, nil
		},
		MockListBackupRecords: func(ctx context.Context, in *db.DBListBackupRecordsRequest, opts ...client.CallOption) (*db.DBListBackupRecordsResponse, error) {
			return nil, nil
		},
		MockSaveRecoverRecord: func(ctx context.Context, in *db.DBSaveRecoverRecordRequest, opts ...client.CallOption) (*db.DBSaveRecoverRecordResponse, error) {
			return nil, nil
		},
		MockSaveParametersRecord: func(ctx context.Context, in *db.DBSaveParametersRequest, opts ...client.CallOption) (*db.DBSaveParametersResponse, error) {
			return nil, nil
		},
		MockGetCurrentParametersRecord: func(ctx context.Context, in *db.DBGetCurrentParametersRequest, opts ...client.CallOption) (*db.DBGetCurrentParametersResponse, error) {
			return nil, nil
		},
		// Tiup Task
		MockCreateTiupTask: func(ctx context.Context, in *db.CreateTiupTaskRequest, opts ...client.CallOption) (*db.CreateTiupTaskResponse, error) {
			return nil, nil
		},
		MockUpdateTiupTask: func(ctx context.Context, in *db.UpdateTiupTaskRequest, opts ...client.CallOption) (*db.UpdateTiupTaskResponse, error) {
			return nil, nil
		},
		MockFindTiupTaskByID: func(ctx context.Context, in *db.FindTiupTaskByIDRequest, opts ...client.CallOption) (*db.FindTiupTaskByIDResponse, error) {
			return nil, nil
		},
		MockGetTiupTaskStatusByBizID: func(ctx context.Context, in *db.GetTiupTaskStatusByBizIDRequest, opts ...client.CallOption) (*db.GetTiupTaskStatusByBizIDResponse, error) {
			return nil, nil
		},
		// Workflow and Task
		MockCreateFlow: func(ctx context.Context, in *db.DBCreateFlowRequest, opts ...client.CallOption) (*db.DBCreateFlowResponse, error) {
			return nil, nil
		},
		MockCreateTask: func(ctx context.Context, in *db.DBCreateTaskRequest, opts ...client.CallOption) (*db.DBCreateTaskResponse, error) {
			return nil, nil
		},
		MockUpdateFlow: func(ctx context.Context, in *db.DBUpdateFlowRequest, opts ...client.CallOption) (*db.DBUpdateFlowResponse, error) {
			return nil, nil
		},
		MockUpdateTask: func(ctx context.Context, in *db.DBUpdateTaskRequest, opts ...client.CallOption) (*db.DBUpdateTaskResponse, error) {
			return nil, nil
		},
		MockLoadFlow: func(ctx context.Context, in *db.DBLoadFlowRequest, opts ...client.CallOption) (*db.DBLoadFlowResponse, error) {
			return nil, nil
		},
		MockLoadTask: func(ctx context.Context, in *db.DBLoadTaskRequest, opts ...client.CallOption) (*db.DBLoadTaskResponse, error) {
			return nil, nil
		},
	}

	fmt.Println("Mock DBClient Completed...")
}
