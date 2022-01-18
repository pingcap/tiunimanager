/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: deploymentImpl_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/17
*******************************************************************************/

package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/deployment/operation"
	mockoperation "github.com/pingcap-inc/tiem/deployment/operation/mock"
)

const (
	TestWorkFlowID  = "testworkflownodeid"
	TestOperationID = "testoperationid"
	TestClusterID   = "testclusterid"
	TestVersion     = "v4.0.12"
	TestTiUPHome    = "/root/.tiup"
	TestTiDBTopo    = ""
)

var manager *Manager

func init() {
	manager = &Manager{
		TiUPBinPath: "mock_tiup",
	}
	MockDB()
}

func TestManager_Deploy_Fail(t *testing.T) {
	expectedErr := errors.New("fail create operation record")

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mockoperation.NewMockReaderWriter(mockCtl)
	SetOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), CMDDeploy, TestWorkFlowID).Return(nil, expectedErr)

	_, err := manager.Deploy(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestVersion,
		TestTiDBTopo, TestTiUPHome, TestWorkFlowID, []string{"-u", "root", "-i", "/root/.ssh/tiup_rsa"}, 360)
	if err == nil {
		t.Error("case: fail create second party operation intentionally. got nil err")
	}
}

func TestManager_Deploy_Success(t *testing.T) {
	op := &operation.Operation{
		ID: TestOperationID,
	}

	mockCtl := gomock.NewController(t)
	mockReaderWriter := mockoperation.NewMockReaderWriter(mockCtl)
	SetOperationReaderWriter(mockReaderWriter)
	mockReaderWriter.EXPECT().Create(context.Background(), CMDDeploy, TestWorkFlowID).Return(op, nil)

	id, err := manager.Deploy(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestVersion,
		TestTiDBTopo, TestTiUPHome, TestWorkFlowID, []string{"-u", "root", "-i", "/root/.ssh/tiup_rsa"}, 360)
	if id != TestOperationID || err != nil {
		t.Errorf("case: create secondparty operation successfully and start asyncoperation. operationid(expected: %s, actual: %s), err(expected: %v, actual: %v)", TestOperationID, id, nil, err)
	}
}
