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

	"github.com/pingcap-inc/tiem/models/tiup"
	"github.com/pingcap-inc/tiem/test/mockmodels/mocktiupconfig"

	spec2 "github.com/pingcap/tiup/pkg/cluster/spec"

	"github.com/pingcap-inc/tiem/library/spec"

	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/workflow/secondparty"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/test/mockmodels/mocksecondparty"
)

var secondPartyManager1 *SecondPartyManager

func init() {
	secondPartyManager1 = &SecondPartyManager{
		TiUPBinPath: "mock_tiup",
	}
	models.MockDB()
	initForTestLibtiup()
}

func TestSecondPartyManager_ClusterDeploy_Fail(t *testing.T) {

	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterDeploy, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, TestWorkFlowNodeID, "")
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterDeploy_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterDeploy, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, TestWorkFlowNodeID, "")
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterScaleOut_Fail(t *testing.T) {

	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterDeploy, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterDeploy(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", "", 0, []string{}, TestWorkFlowNodeID, "")
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	_, err := secondPartyManager1.ClusterList(context.TODO(), ClusterComponentTypeStr, 0, []string{})
	if err == nil {
		t.Errorf("case: create secondparty task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterList_WithTimeOut(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	_, err := secondPartyManager1.ClusterList(context.TODO(), ClusterComponentTypeStr, 1, []string{})
	if err == nil {
		t.Errorf("case: create secondparty task successfully. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterStart_Fail(t *testing.T) {

	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterStart, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterStart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterStart_Success(t *testing.T) {
	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterStart, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterStart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterRestart_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterRestart, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterRestart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterRestart_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterRestart, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterRestart(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterStop_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterStop, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterStop(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterStop_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterStop, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterStop(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterDestroy_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterDestroy, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterDestroy(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterDestroy_Success(t *testing.T) {
	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterDestroy, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterDestroy(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_Transfer_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_Transfer, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.Transfer(context.TODO(), TiEMComponentTypeStr, "test-tidb", "test-yaml", "/remote/path", 0, []string{"-N", "test-host"}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterTransfer_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_Transfer, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(TiEMComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.Transfer(context.TODO(), TiEMComponentTypeStr, "test-tidb", "test-yaml", "/remote/path", 0, []string{"-N", "test-host"}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterUpgrade_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterUpgrade, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterUpgrade(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterUpgrade_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterUpgrade, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterUpgrade(context.TODO(), ClusterComponentTypeStr, "test-tidb", "v1", 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterShowConfig(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

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
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

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

	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterEditGlobalConfig, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterEditGlobalConfig(context.TODO(), cmdEditGlobalConfigReq, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
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

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterEditGlobalConfig, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterEditGlobalConfig(context.TODO(), cmdEditGlobalConfigReq, TestWorkFlowNodeID)
	if operationID != TestOperationID || err == nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_startTiupEditGlobalConfigTask(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

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
	secondPartyManager1.startTiUPEditGlobalConfigOperation(context.TODO(), TestOperationID, &req, &topo)
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

	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterEditInstanceConfig, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.ClusterEditInstanceConfig(context.TODO(), cmdEditInstanceConfigReq, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
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

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterEditInstanceConfig, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.ClusterEditInstanceConfig(context.TODO(), cmdEditInstanceConfigReq, TestWorkFlowNodeID)
	if operationID != TestOperationID || err == nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_startTiupEditInstanceConfigTask(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)

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
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiKV
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiFlash
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_PD
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_Pump
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_Drainer
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_CDC
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiSparkMasters
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_TiSparkWorkers
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_Prometheus
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_Grafana
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)

	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)
	req.TiDBClusterComponent = spec.TiDBClusterComponent_Alertmanager
	secondPartyManager1.startTiUPEditInstanceConfigOperation(context.TODO(), TestOperationID, &req, &topo)
}

func TestSecondPartyManager_ClusterReload_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterReload, TestWorkFlowNodeID).Return(nil, expectedErr)

	cmdReloadConfigReq := CmdReloadConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	operationID, err := secondPartyManager1.ClusterReload(context.TODO(), cmdReloadConfigReq, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterReload_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterReload, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	cmdReloadConfigReq := CmdReloadConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	operationID, err := secondPartyManager1.ClusterReload(context.TODO(), cmdReloadConfigReq, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterExec_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterExec, TestWorkFlowNodeID).Return(nil, expectedErr)

	req := CmdClusterExecReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	operationID, err := secondPartyManager1.ClusterExec(context.TODO(), req, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterExec_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_ClusterExec, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	req := CmdClusterExecReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  "test-tidb",
		TimeoutS:      0,
		Flags:         []string{},
	}
	operationID, err := secondPartyManager1.ClusterExec(context.TODO(), req, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterDumpling_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_Dumpling, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.Dumpling(context.TODO(), 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterDumpling_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_Dumpling, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.Dumpling(context.TODO(), 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterLightning_Fail(t *testing.T) {
	expectedErr := errors.New("fail create second party operation")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_Lightning, TestWorkFlowNodeID).Return(nil, expectedErr)

	operationID, err := secondPartyManager1.Lightning(context.TODO(), 0, []string{}, TestWorkFlowNodeID)
	if operationID != "" || err == nil {
		t.Errorf("case: fail create second party operation intentionally. taskid(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", operationID, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterLightning_Success(t *testing.T) {

	secondPartyOperation := secondparty.SecondPartyOperation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), secondparty.OperationType_Lightning, TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	operationID, err := secondPartyManager1.Lightning(context.TODO(), 0, []string{}, TestWorkFlowNodeID)
	if operationID != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, operationID, nil, err)
	}
}

func TestSecondPartyManager_ClusterClusterDisplay(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	_, err := secondPartyManager1.ClusterDisplay(context.TODO(), ClusterComponentTypeStr, "test-tidb", 0, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterClusterDisplay_WithTimeout(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	_, err := secondPartyManager1.ClusterDisplay(context.TODO(), ClusterComponentTypeStr, "test-tidb", 1, []string{})
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_GetOperationStatus_Fail(t *testing.T) {

	expectedErr := errors.New("Fail Find secondparty task by Id")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Get(context.Background(), TestOperationID).Return(nil, expectedErr)

	resp, err := secondPartyManager1.GetOperationStatus(context.TODO(), TestOperationID)
	if resp.Status != "" || err == nil {
		t.Errorf("case: fail find secondparty opeartion by id intentionally. operationStatus(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", resp.Status, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterGetTaskStatus_Success(t *testing.T) {
	secondPartyOperation := secondparty.SecondPartyOperation{
		ID:     TestOperationID,
		Status: secondparty.OperationStatus_Finished,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Get(context.Background(), TestOperationID).Return(&secondPartyOperation, nil)

	resp, err := secondPartyManager1.GetOperationStatus(context.TODO(), TestOperationID)
	if resp.Status != secondparty.OperationStatus_Finished || err != nil {
		t.Errorf("case: find secondparty operation by id successfully. operationStatus(expected: %v, actual: %v), err(expected: %v, actual: %v)", secondparty.OperationStatus_Finished, resp.Status, nil, err)
	}
}

func TestSecondPartyManager_ClusterGetTaskStatusByBizID_Fail(t *testing.T) {
	expectedErr := errors.New("Fail Find secondparty operation by workflownodeid")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().QueryByWorkFlowNodeID(context.Background(), TestWorkFlowNodeID).Return(nil, expectedErr)

	resp, err := secondPartyManager1.GetOperationStatusByWorkFlowNodeID(context.TODO(), TestWorkFlowNodeID)
	if resp.Status != "" || err == nil {
		t.Errorf("case: fail find secondparty operation by workflownodeid intentionally. operationStatus(expected: %s, actual: %s), err(expected: %v, actual: %v)", "", resp.Status, expectedErr, err)
	}
}

func TestSecondPartyManager_ClusterGetTaskStatusByBizID_Success(t *testing.T) {
	secondPartyOperation := secondparty.SecondPartyOperation{
		ID:     TestOperationID,
		Status: secondparty.OperationStatus_Finished,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mocksecondparty.NewMockReaderWriter(mockCtl)
	models.SetSecondPartyOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().QueryByWorkFlowNodeID(context.Background(), TestWorkFlowNodeID).Return(&secondPartyOperation, nil)

	resp, err := secondPartyManager1.GetOperationStatusByWorkFlowNodeID(context.TODO(), TestWorkFlowNodeID)
	if resp.Status != secondparty.OperationStatus_Finished || err != nil {
		t.Errorf("case: find secondparty operation by id successfully. operationStatus(expected: %v, actual: %v), err(expected: %v, actual: %v)", secondparty.OperationStatus_Finished, resp.Status, nil, err)
	}
}

func TestSecondPartyManager_ClusterComponentCtl(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	_, err := secondPartyManager1.ClusterComponentCtl(context.TODO(), CTLComponentTypeStr, "v5.0.0", spec.TiDBClusterComponent_PD, []string{}, 0)
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_ClusterComponentCtl_WithTimeout(t *testing.T) {
	mockCtl := gomock.NewController(t)
	tiUPConfig := &tiup.TiupConfig{
		TiupHome: "",
	}
	mockTiUPConfigReaderWriter := mocktiupconfig.NewMockReaderWriter(mockCtl)
	models.SetTiUPConfigReaderWriter(mockTiUPConfigReaderWriter)
	mockTiUPConfigReaderWriter.EXPECT().QueryByComponentType(context.Background(), string(DefaultComponentTypeStr)).Return(tiUPConfig, nil)

	_, err := secondPartyManager1.ClusterComponentCtl(context.TODO(), CTLComponentTypeStr, "v5.0.0", spec.TiDBClusterComponent_PD, []string{}, 1)
	if err == nil {
		t.Errorf("case: cluster display. err(expected: not nil, actual: nil)")
	}
}

func TestSecondPartyManager_startTiUPTask_Wrong(t *testing.T) {
	secondPartyManager1.startTiUPOperation(context.TODO(), TestOperationID, "ls", []string{"-2"}, 1, "", "")
}

func TestSecondPartyManager_startTiUPTask(t *testing.T) {
	secondPartyManager1.startTiUPOperation(context.TODO(), TestOperationID, "ls", []string{}, 1, "", "")
}

func initForTestLibtiup() {
	secondPartyManager1.syncedOperationStatusMap = make(map[string]OperationStatusMapValue)
	secondPartyManager1.operationStatusCh = make(chan OperationStatusMember, 1024)
	secondPartyManager1.operationStatusMap = make(map[string]OperationStatusMapValue)
}
