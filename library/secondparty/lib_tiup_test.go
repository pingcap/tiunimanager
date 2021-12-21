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

package secondparty

import (
	"context"
	"errors"
	"strings"

	"github.com/pingcap-inc/tiem/common/constants"

	spec2 "github.com/pingcap/tiup/pkg/cluster/spec"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/test/mockdb"
)

var secondMicro1 *SecondMicro

func init() {
	secondMicro1 = &SecondMicro{
		TiupBinPath: "mock_tiup",
	}
	microInitForTestLibtiup()
}

func TestSecondMicro_MicroSrvTiupDeploy_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupDeploy_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupList(t *testing.T) {
	_, err := secondMicro1.MicroSrvTiupList(context.TODO(), ClusterComponentTypeStr, 0, []string{})
	if err == nil {
		t.Errorf("case: create secondparty task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupList_WithTimeOut(t *testing.T) {
	_, err := secondMicro1.MicroSrvTiupList(context.TODO(), ClusterComponentTypeStr, 1, []string{})
	if err == nil {
		t.Errorf("case: create secondparty task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupStart_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupStart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupStart_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupStart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupRestart_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupRestart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupRestart_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupRestart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupStop_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupStop(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupStop_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupStop(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupDestroy_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupDestroy(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupDestroy_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupDestroy(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupTransfer_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Transfer
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupTransfer(context.TODO(), TiEMComponentTypeStr, "test-tidb", "test-yaml", "/remote/path", 0, []string{"-N", "test-host"}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupTransfer_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Transfer
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupTransfer(context.TODO(), TiEMComponentTypeStr, "test-tidb", "test-yaml", "/remote/path", 0, []string{"-N", "test-host"}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupUpgrade_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Upgrade
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupUpgrade(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupUpgrade_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Upgrade
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupUpgrade(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupShowConfig(t *testing.T) {
	req := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	_, err := secondMicro1.MicroSrvTiupShowConfig(context.TODO(), &req)
	if err == nil {
		t.Errorf("case: cluster show-config. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupShowConfig_WithTimeout(t *testing.T) {
	req := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      1,
		Flags:         []string{},
	}
	_, err := secondMicro1.MicroSrvTiupShowConfig(context.TODO(), &req)
	if err == nil {
		t.Errorf("case: cluster show-config. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupEditGlobalConfig_Fail(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["foo"] = "bar"
	globalComponentConfigTiDB := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDTiDB,
		ConfigMap:            configMap,
	}
	cmdEditGlobalConfigReq := CmdEditGlobalConfigReq{
		TiUPComponent:          ClusterComponentTypeStr,
		InstanceName:           "test-tidb",
		GlobalComponentConfigs: []GlobalComponentConfig{globalComponentConfigTiDB},
		TimeoutS:               0,
		Flags:                  []string{},
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_EditGlobalConfig
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")
	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupEditGlobalConfig(context.TODO(), cmdEditGlobalConfigReq, 0)
	if taskID != 0 || err == nil || !strings.Contains(err.Error(), "rsp") {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupEditGlobalConfig_Success(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["foo"] = "bar"
	globalComponentConfigTiDB := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDTiDB,
		ConfigMap:            configMap,
	}
	cmdEditGlobalConfigReq := CmdEditGlobalConfigReq{
		TiUPComponent:          ClusterComponentTypeStr,
		InstanceName:           "test-tidb",
		GlobalComponentConfigs: []GlobalComponentConfig{globalComponentConfigTiDB},
		TimeoutS:               0,
		Flags:                  []string{},
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_EditGlobalConfig
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupEditGlobalConfig(context.TODO(), cmdEditGlobalConfigReq, 0)
	if taskID != 1 || err == nil || !strings.Contains(err.Error(), "cmd start err") {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_startTiupEditGlobalConfigTask(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["foo"] = "bar"
	topo := spec2.Specification{}
	globalComponentConfigTiDB := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDTiDB,
		ConfigMap:            configMap,
	}
	globalComponentConfigTiKV := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDTiKV,
		ConfigMap:            configMap,
	}
	globalComponentConfigPD := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDPD,
		ConfigMap:            configMap,
	}
	globalComponentConfigTiFlash := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDTiFlash,
		ConfigMap:            configMap,
	}
	globalComponentConfigCDC := GlobalComponentConfig{
		TiDBClusterComponent: constants.ComponentIDTiCDC,
		ConfigMap:            configMap,
	}
	req := CmdEditGlobalConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		GlobalComponentConfigs: []GlobalComponentConfig{
			globalComponentConfigTiDB,
			globalComponentConfigTiKV,
			globalComponentConfigPD,
			globalComponentConfigTiFlash,
			globalComponentConfigCDC,
		},
	}
	secondMicro1.startTiupEditGlobalConfigTask(context.TODO(), 1, &req, &topo)
}

func TestSecondMicro_MicroSrvTiupEditInstanceConfig_Fail(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["host"] = "1.2.3.4"
	cmdEditInstanceConfigReq := CmdEditInstanceConfigReq{
		TiUPComponent:        ClusterComponentTypeStr,
		InstanceName:         "test-tidb",
		TiDBClusterComponent: constants.ComponentIDTiDB,
		Host:                 "1.2.3.4",
		Port:                 1234,
		ConfigMap:            configMap,
		TimeoutS:             0,
		Flags:                []string{},
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_EditInstanceConfig
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")
	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvTiupEditInstanceConfig(context.TODO(), cmdEditInstanceConfigReq, 0)
	if taskID != 0 || err == nil || !strings.Contains(err.Error(), "rsp") {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupEditInstanceConfig_Success(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["host"] = "1.2.3.4"
	cmdEditInstanceConfigReq := CmdEditInstanceConfigReq{
		TiUPComponent:        ClusterComponentTypeStr,
		InstanceName:         "test-tidb",
		TiDBClusterComponent: constants.ComponentIDTiDB,
		Host:                 "1.2.3.4",
		Port:                 1234,
		ConfigMap:            configMap,
		TimeoutS:             0,
		Flags:                []string{},
	}

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_EditInstanceConfig
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvTiupEditInstanceConfig(context.TODO(), cmdEditInstanceConfigReq, 0)
	if taskID != 1 || err == nil || !strings.Contains(err.Error(), "cmd start err") {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_startTiupEditInstanceConfigTask(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["host"] = "1.2.3.5"
	tiDBSpec := spec2.TiDBSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	tiKVSpec := spec2.TiKVSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	tiFlashSpec := spec2.TiFlashSpec{
		Host:             "1.2.3.4",
		FlashServicePort: 1234,
	}
	pdSpec := spec2.PDSpec{
		Host:       "1.2.3.4",
		ClientPort: 1234,
	}
	pumpSpec := spec2.PumpSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	drainerSpec := spec2.DrainerSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	cdcSpec := spec2.CDCSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	tiSparkMasterSpec := spec2.TiSparkMasterSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	tiSparkWorkerSpec := spec2.TiSparkWorkerSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	prometheusSpec := spec2.PrometheusSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	grafanaSpec := spec2.GrafanaSpec{
		Host: "1.2.3.4",
		Port: 1234,
	}
	alertmanagerSpec := spec2.AlertmanagerSpec{
		Host:    "1.2.3.4",
		WebPort: 1234,
	}
	topo := spec2.Specification{
		TiDBServers:    []*spec2.TiDBSpec{&tiDBSpec},
		TiKVServers:    []*spec2.TiKVSpec{&tiKVSpec},
		TiFlashServers: []*spec2.TiFlashSpec{&tiFlashSpec},
		PDServers:      []*spec2.PDSpec{&pdSpec},
		PumpServers:    []*spec2.PumpSpec{&pumpSpec},
		Drainers:       []*spec2.DrainerSpec{&drainerSpec},
		CDCServers:     []*spec2.CDCSpec{&cdcSpec},
		TiSparkMasters: []*spec2.TiSparkMasterSpec{&tiSparkMasterSpec},
		TiSparkWorkers: []*spec2.TiSparkWorkerSpec{&tiSparkWorkerSpec},
		Monitors:       []*spec2.PrometheusSpec{&prometheusSpec},
		Grafanas:       []*spec2.GrafanaSpec{&grafanaSpec},
		Alertmanagers:  []*spec2.AlertmanagerSpec{&alertmanagerSpec},
	}
	req := CmdEditInstanceConfigReq{
		TiUPComponent:        ClusterComponentTypeStr,
		InstanceName:         "test-tidb",
		TiDBClusterComponent: constants.ComponentIDTiDB,
		Host:                 "1.2.3.4",
		Port:                 1234,
		ConfigMap:            configMap,
		TimeoutS:             0,
		Flags:                []string{},
	}
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDTiKV
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDTiFlash
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDPD
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDTiCDC
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDPrometheus
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDGrafana
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = constants.ComponentIDAlertManager
	secondMicro1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)
}

func TestSecondMicro_MicroSrvTiupReload_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Reload
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	cmdReloadConfigReq := CmdReloadConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	taskID, err := secondMicro1.MicroSrvTiupReload(context.TODO(), cmdReloadConfigReq, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupReload_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Reload
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	cmdReloadConfigReq := CmdReloadConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	taskID, err := secondMicro1.MicroSrvTiupReload(context.TODO(), cmdReloadConfigReq, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupDumpling_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvDumpling(context.TODO(), 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupDumpling_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvDumpling(context.TODO(), 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupLightning_Fail(t *testing.T) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = 0

	expectedErr := errors.New("Fail Create secondparty task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondMicro1.MicroSrvLightning(context.TODO(), 0, []string{}, 0)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create secondparty task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupLightning_Success(t *testing.T) {

	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = 0

	var resp dbPb.CreateTiupTaskResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupTask(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondMicro1.MicroSrvLightning(context.TODO(), 0, []string{}, 0)
	if taskID != 1 || err != nil {
		t.Errorf("case: create secondparty task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupClusterDisplay(t *testing.T) {

	_, err := secondMicro1.MicroSrvTiupDisplay(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupClusterDisplay_WithTimeout(t *testing.T) {

	_, err := secondMicro1.MicroSrvTiupDisplay(context.TODO(), ClusterComponentTypeStr, "test-tidb", 1, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatus_Fail(t *testing.T) {
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupTaskByIDResponse
	resp.ErrCode = 1
	resp.ErrStr = "Fail Find secondparty task by Id"

	expectedErr := errors.New("Fail Find secondparty task by Id")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupTaskByID(context.Background(), gomock.Eq(&req)).Return(&resp, expectedErr)

	_, errStr, err := secondMicro1.MicroSrvGetTaskStatus(context.TODO(), 1)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find secondparty task by id intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatus_Success(t *testing.T) {
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupTaskByIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.TiupTask = &dbPb.TiupTask{
		ID:     1,
		Status: dbPb.TiupTaskStatus_Init,
	}

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupTaskByID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, errStr, err := secondMicro1.MicroSrvGetTaskStatus(context.TODO(), 1)
	if stat != dbPb.TiupTaskStatus_Init || errStr != "" || err != nil {
		t.Errorf("case: find secondparty task by id successfully. stat(expected: %d, actual: %d), errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", errStr, nil, err)
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatusByBizID_Fail(t *testing.T) {
	var req dbPb.GetTiupTaskStatusByBizIDRequest
	req.BizID = 0

	var resp dbPb.GetTiupTaskStatusByBizIDResponse
	resp.ErrCode = 1
	resp.ErrStr = "Fail Find secondparty task by BizId"

	expectedErr := errors.New("Fail Find secondparty task by Id")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().GetTiupTaskStatusByBizID(context.Background(), gomock.Eq(&req)).Return(&resp, expectedErr)

	_, errStr, err := secondMicro1.MicroSrvGetTaskStatusByBizID(context.TODO(), 0)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find secondparty task by BizId intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestSecondMicro_MicroSrvTiupGetTaskStatusByBizID_Success(t *testing.T) {
	var req dbPb.GetTiupTaskStatusByBizIDRequest
	req.BizID = 0

	var resp dbPb.GetTiupTaskStatusByBizIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.Stat = dbPb.TiupTaskStatus_Init
	resp.StatErrStr = ""

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().GetTiupTaskStatusByBizID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, statErrStr, err := secondMicro1.MicroSrvGetTaskStatusByBizID(context.TODO(), 0)
	if stat != dbPb.TiupTaskStatus_Init || statErrStr != "" || err != nil {
		t.Errorf("case: find secondparty task by id successfully. stat(expected: %d, actual: %d), statErrStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", statErrStr, nil, err)
	}
}

func TestSecondMicro_startNewTiupTask_Wrong(t *testing.T) {
	secondMicro1.startNewTiupTask(context.TODO(), 1, "ls", []string{"-2"}, 1)
}

func TestSecondMicro_startNewTiupTask(t *testing.T) {
	secondMicro1.startNewTiupTask(context.TODO(), 1, "ls", []string{}, 1)
}

func microInitForTestLibtiup() {
	secondMicro1.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondMicro1.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondMicro1.taskStatusMap = make(map[uint64]TaskStatusMapValue)
}
