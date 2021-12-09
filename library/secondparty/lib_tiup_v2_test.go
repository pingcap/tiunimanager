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

/*******************************************************************************
 * @File: lib_tiup_v2_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"context"
	"errors"
	"strings"

	"github.com/pingcap-inc/tiem/library/spec"
	spec2 "github.com/pingcap/tiup/pkg/cluster/spec"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/test/mockdb"
)

var secondPartyManager1 *SecondPartyManager

func init() {
	secondPartyManager1 = &SecondPartyManager{
		TiupBinPath: "mock_tiup",
	}
	initForTestLibtiup()
}

func TestSecondPartyManager_ClusterDeploy_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterDeploy_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterList(t *testing.T) {
	_, err := secondPartyManager1.ClusterList(context.TODO(), ClusterComponentTypeStr, 0, []string{})
	if err == nil {
		t.Errorf("case: create tiup task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterList_WithTimeOut(t *testing.T) {
	_, err := secondPartyManager1.ClusterList(context.TODO(), ClusterComponentTypeStr, 1, []string{})
	if err == nil {
		t.Errorf("case: create tiup task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterStart_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterStart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterStart_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterStart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterRestart_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterRestart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterRestart_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterRestart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterStop_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterStop(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterStop_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterStop(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterDestroy_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterDestroy(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterDestroy_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterDestroy(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterTransfer_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Transfer
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.Transfer(context.TODO(), TiEMComponentTypeStr, "test-tidb", "test-yaml", "/remote/path", 0, []string{"-N", "test-host"}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterTransfer_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Transfer
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.Transfer(context.TODO(), TiEMComponentTypeStr, "test-tidb", "test-yaml", "/remote/path", 0, []string{"-N", "test-host"}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterUpgrade_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Upgrade
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterUpgrade(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterUpgrade_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Upgrade
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterUpgrade(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterShowConfig(t *testing.T) {
	req := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	_, err := secondPartyManager1.ClusterShowConfig(context.TODO(), &req)
	if err == nil {
		t.Errorf("case: cluster show-config. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterShowConfig_WithTimeout(t *testing.T) {
	req := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      1,
		Flags:         []string{},
	}
	_, err := secondPartyManager1.ClusterShowConfig(context.TODO(), &req)
	if err == nil {
		t.Errorf("case: cluster show-config. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterEditGlobalConfig_Fail(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["foo"] = "bar"
	globalComponentConfigTiDB := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
		ConfigMap:            configMap,
	}
	cmdEditGlobalConfigReq := CmdEditGlobalConfigReq{
		TiUPComponent:          ClusterComponentTypeStr,
		InstanceName:           "test-tidb",
		GlobalComponentConfigs: []GlobalComponentConfig{globalComponentConfigTiDB},
		TimeoutS:               0,
		Flags:                  []string{},
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_EditGlobalConfig
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")
	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterEditGlobalConfig(context.TODO(), cmdEditGlobalConfigReq, TESTBIZID)
	if taskID != 0 || err == nil || !strings.Contains(err.Error(), "rsp") {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterEditGlobalConfig_Success(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["foo"] = "bar"
	globalComponentConfigTiDB := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
		ConfigMap:            configMap,
	}
	cmdEditGlobalConfigReq := CmdEditGlobalConfigReq{
		TiUPComponent:          ClusterComponentTypeStr,
		InstanceName:           "test-tidb",
		GlobalComponentConfigs: []GlobalComponentConfig{globalComponentConfigTiDB},
		TimeoutS:               0,
		Flags:                  []string{},
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_EditGlobalConfig
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterEditGlobalConfig(context.TODO(), cmdEditGlobalConfigReq, TESTBIZID)
	if taskID != 1 || err == nil || !strings.Contains(err.Error(), "cmd start err") {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_startTiupEditGlobalConfigTask(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["foo"] = "bar"
	topo := spec2.Specification{}
	globalComponentConfigTiDB := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
		ConfigMap:            configMap,
	}
	globalComponentConfigTiKV := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_TiKV,
		ConfigMap:            configMap,
	}
	globalComponentConfigPD := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_PD,
		ConfigMap:            configMap,
	}
	globalComponentConfigTiFlash := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_TiFlash,
		ConfigMap:            configMap,
	}
	globalComponentConfigTiFlashLearner := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_TiFlashLearner,
		ConfigMap:            configMap,
	}
	globalComponentConfigPump := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_Pump,
		ConfigMap:            configMap,
	}
	globalComponentConfigDrainer := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_Drainer,
		ConfigMap:            configMap,
	}
	globalComponentConfigCDC := GlobalComponentConfig{
		TiDBClusterComponent: spec.TiDBClusterComponent_CDC,
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
			globalComponentConfigTiFlashLearner,
			globalComponentConfigPump,
			globalComponentConfigDrainer,
			globalComponentConfigCDC,
		},
	}
	secondPartyManager1.startTiupEditGlobalConfigTask(context.TODO(), 1, &req, &topo)
}

func TestSecondPartyManager_ClusterEditInstanceConfig_Fail(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["host"] = "1.2.3.4"
	cmdEditInstanceConfigReq := CmdEditInstanceConfigReq{
		TiUPComponent:        ClusterComponentTypeStr,
		InstanceName:         "test-tidb",
		TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
		Host:                 "1.2.3.4",
		Port:                 1234,
		ConfigMap:            configMap,
		TimeoutS:             0,
		Flags:                []string{},
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_EditInstanceConfig
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")
	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.ClusterEditInstanceConfig(context.TODO(), cmdEditInstanceConfigReq, TESTBIZID)
	if taskID != 0 || err == nil || !strings.Contains(err.Error(), "rsp") {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterEditInstanceConfig_Success(t *testing.T) {
	configMap := make(map[string]interface{})
	configMap["host"] = "1.2.3.4"
	cmdEditInstanceConfigReq := CmdEditInstanceConfigReq{
		TiUPComponent:        ClusterComponentTypeStr,
		InstanceName:         "test-tidb",
		TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
		Host:                 "1.2.3.4",
		Port:                 1234,
		ConfigMap:            configMap,
		TimeoutS:             0,
		Flags:                []string{},
	}

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_EditInstanceConfig
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.ClusterEditInstanceConfig(context.TODO(), cmdEditInstanceConfigReq, TESTBIZID)
	if taskID != 1 || err == nil || !strings.Contains(err.Error(), "cmd start err") {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_startTiupEditInstanceConfigTask(t *testing.T) {
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
		TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
		Host:                 "1.2.3.4",
		Port:                 1234,
		ConfigMap:            configMap,
		TimeoutS:             0,
		Flags:                []string{},
	}
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiKV
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiFlash
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_PD
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_Pump
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_Drainer
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_CDC
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiSparkMasters
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiSparkWorkers
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_Prometheus
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_Grafana
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)

	req.TiDBClusterComponent = spec.TiDBClusterComponent_Alertmanager
	secondPartyManager1.startTiupEditInstanceConfigTask(context.TODO(), 1, &req, &topo)
}

func TestSecondPartyManager_ClusterReload_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Reload
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	cmdReloadConfigReq := CmdReloadConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	taskID, err := secondPartyManager1.ClusterReload(context.TODO(), cmdReloadConfigReq, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterReload_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Reload
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	cmdReloadConfigReq := CmdReloadConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	taskID, err := secondPartyManager1.ClusterReload(context.TODO(), cmdReloadConfigReq, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterDumpling_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.Dumpling(context.TODO(), 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterDumpling_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.Dumpling(context.TODO(), 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterLightning_Fail(t *testing.T) {
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = TESTBIZID

	expectedErr := errors.New("Fail Create tiup task")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(nil, expectedErr)

	taskID, err := secondPartyManager1.Lightning(context.TODO(), 0, []string{}, TESTBIZID)
	if taskID != 0 || err == nil {
		t.Errorf("case: fail create tiup task intentionally. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 0, taskID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterLightning_Success(t *testing.T) {

	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = TESTBIZID

	var resp dbPb.CreateTiupOperatorRecordResponse
	resp.ErrCode = 0
	resp.Id = 1

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().CreateTiupOperatorRecord(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	taskID, err := secondPartyManager1.Lightning(context.TODO(), 0, []string{}, TESTBIZID)
	if taskID != 1 || err != nil {
		t.Errorf("case: create tiup task successfully. taskid(expected: %d, actual: %d), err(expected: %v, actual: %v)", 1, taskID, nil, err)
	}
}

func TestSecondPartyManager_ClusterClusterDisplay(t *testing.T) {

	_, err := secondPartyManager1.ClusterDisplay(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterClusterDisplay_WithTimeout(t *testing.T) {

	_, err := secondPartyManager1.ClusterDisplay(context.TODO(), ClusterComponentTypeStr, "test-tidb", 1, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterGetTaskStatus_Fail(t *testing.T) {
	var req dbPb.FindTiupOperatorRecordByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupOperatorRecordByIDResponse
	resp.ErrCode = 1
	resp.ErrStr = "Fail Find tiup task by Id"

	expectedErr := errors.New("Fail Find tiup task by Id")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupOperatorRecordByID(context.Background(), gomock.Eq(&req)).Return(&resp, expectedErr)

	_, errStr, err := secondPartyManager1.GetTaskStatus(context.TODO(), 1)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by id intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterGetTaskStatus_Success(t *testing.T) {
	var req dbPb.FindTiupOperatorRecordByIDRequest
	req.Id = 1

	var resp dbPb.FindTiupOperatorRecordByIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.TiupTask = &dbPb.TiupOperatorRecord{
		ID:     1,
		Status: dbPb.TiupTaskStatus_Init,
	}

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().FindTiupOperatorRecordByID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, errStr, err := secondPartyManager1.GetTaskStatus(context.TODO(), 1)
	if stat != dbPb.TiupTaskStatus_Init || errStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", errStr, nil, err)
	}
}

func TestSecondPartyManager_ClusterGetTaskStatusByBizID_Fail(t *testing.T) {
	var req dbPb.GetTiupOperatorRecordStatusByBizIDRequest
	req.BizID = TESTBIZID

	var resp dbPb.GetTiupOperatorRecordStatusByBizIDResponse
	resp.ErrCode = 1
	resp.ErrStr = "Fail Find tiup task by BizId"

	expectedErr := errors.New("Fail Find tiup task by Id")

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().GetTiupOperatorRecordStatusByBizID(context.Background(), gomock.Eq(&req)).Return(&resp, expectedErr)

	_, errStr, err := secondPartyManager1.GetTaskStatusByBizID(context.TODO(), TESTBIZID)
	if errStr != "" || err == nil {
		t.Errorf("case: fail find tiup task by BizId intentionally. errStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", errStr, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterGetTaskStatusByBizID_Success(t *testing.T) {
	var req dbPb.GetTiupOperatorRecordStatusByBizIDRequest
	req.BizID = TESTBIZID

	var resp dbPb.GetTiupOperatorRecordStatusByBizIDResponse
	resp.ErrCode = 0
	resp.ErrStr = ""
	resp.Stat = dbPb.TiupTaskStatus_Init
	resp.StatErrStr = ""

	mockCtl := gomock.NewController(t)
	mockDBClient := mockdb.NewMockTiEMDBService(mockCtl)
	client.DBClient = mockDBClient
	mockDBClient.EXPECT().GetTiupOperatorRecordStatusByBizID(context.Background(), gomock.Eq(&req)).Return(&resp, nil)

	stat, statErrStr, err := secondPartyManager1.GetTaskStatusByBizID(context.TODO(), TESTBIZID)
	if stat != dbPb.TiupTaskStatus_Init || statErrStr != "" || err != nil {
		t.Errorf("case: find tiup task by id successfully. stat(expected: %d, actual: %d), statErrStr(expected: %s, actual: %s), err(expected: %v, actual: %v)", dbPb.TiupTaskStatus_Init, stat, "", statErrStr, nil, err)
	}
}

func TestSecondPartyManager_startNewTiupTask_Wrong(t *testing.T) {
	secondPartyManager1.startNewTiupTask(context.TODO(), 1, "ls", []string{"-2"}, 1)
}

func TestSecondPartyManager_startNewTiupTask(t *testing.T) {
	secondPartyManager1.startNewTiupTask(context.TODO(), 1, "ls", []string{}, 1)
}

func initForTestLibtiup() {
	secondPartyManager1.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondPartyManager1.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondPartyManager1.taskStatusMap = make(map[uint64]TaskStatusMapValue)
}
