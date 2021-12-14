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
 ******************************************************************************/

/*******************************************************************************
 * @File: manager_test.go
 * @Description: cluster parameter cluster service unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/14 09:52
*******************************************************************************/

package parameter

import (
	"context"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/models/parametergroup"

	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclusterparameter"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
)

var manager = NewManager()

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)

			return models.Open(d, false)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

func TestManager_QueryClusterParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterId string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
			return "1", []*parameter.ClusterParamDetail{
				&parameter.ClusterParamDetail{
					Parameter: parametergroup.Parameter{
						ID:             "1",
						Category:       "basic",
						Name:           "param1",
						InstanceType:   "TiKV",
						SystemVariable: "",
						Type:           0,
					},
					DefaultValue: "10",
					RealValue:    "{\"clusterValue\":\"1\"}",
					Note:         "test parameter",
				},
			}, 1, nil
		})

	resp, page, err := manager.QueryClusterParameters(context.TODO(), cluster.QueryClusterParametersReq{
		ClusterID:   "1",
		PageRequest: structs.PageRequest{Page: 1, PageSize: 10},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.EqualValues(t, 1, page.Total)
}

func TestManager_UpdateClusterParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	clusterParameterRW.EXPECT().UpdateClusterParameter(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterId string, params []*parameter.ClusterParameterMapping) (err error) {
			return nil
		})

	resp, err := manager.UpdateClusterParameters(context.TODO(), cluster.UpdateClusterParametersReq{
		ClusterID: "1",
		Params: []structs.ClusterParameterSampleInfo{
			{
				ParamId:       "1",
				Name:          "param1",
				ComponentType: "TiKV",
				HasReboot:     0,
				HasApply:      1,
				UpdateSource:  1,
				Type:          1,
				RealValue:     structs.ParameterRealValue{},
			},
		},
		Reboot: false,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

func TestManager_InspectClusterParameters(t *testing.T) {
	parameters, err := manager.InspectClusterParameters(context.TODO(), cluster.InspectClusterParametersReq{ClusterID: "1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, parameters)
}
