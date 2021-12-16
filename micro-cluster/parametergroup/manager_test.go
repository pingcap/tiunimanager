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
 * @Description: parameter group cluster service unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/14 16:54
*******************************************************************************/

package parametergroup

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/models/parametergroup"

	"github.com/pingcap-inc/tiem/test/mockmodels/mockparametergroup"

	"github.com/golang/mock/gomock"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"

	"github.com/alecthomas/assert"
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

func TestManager_CreateParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().CreateParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, pg *parametergroup.ParameterGroup, pgm []*parametergroup.ParameterGroupMapping) (*parametergroup.ParameterGroup, error) {
			resp := &parametergroup.ParameterGroup{
				ID:             "1",
				Name:           "test_parameter_group",
				ParentID:       "",
				ClusterSpec:    "8C16G",
				HasDefault:     1,
				DBType:         1,
				GroupType:      1,
				ClusterVersion: "5.0",
				Note:           "test parameter group",
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
			}
			return resp, nil
		})
	resp, err := manager.CreateParameterGroup(context.TODO(), message.CreateParameterGroupReq{
		Name:           "test_parameter_group",
		DBType:         1,
		HasDefault:     1,
		ClusterVersion: "5.0",
		ClusterSpec:    "8C16G",
		GroupType:      1,
		Note:           "test parameter group",
		Params: []structs.ParameterGroupParameterSampleInfo{
			{
				ID:           "1",
				DefaultValue: "10",
				Note:         "test param",
			},
		},
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_UpdateParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().UpdateParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, pg *parametergroup.ParameterGroup, pgm []*parametergroup.ParameterGroupMapping) error {
			return nil
		})
	resp, err := manager.UpdateParameterGroup(context.TODO(), message.UpdateParameterGroupReq{
		ParamGroupID:   "1",
		Name:           "test_parameter_group",
		ClusterVersion: "5.0",
		ClusterSpec:    "8C16G",
		Note:           "test parameter group",
		Params: []structs.ParameterGroupParameterSampleInfo{
			{
				ID:           "1",
				DefaultValue: "10",
				Note:         "test param",
			},
		},
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_DeleteParameterGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, nil, nil
		})
	parameterGroupRW.EXPECT().DeleteParameterGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) error {
			return nil
		})
	resp, err := manager.DeleteParameterGroup(context.TODO(), message.DeleteParameterGroupReq{ParamGroupID: "1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_DeleteParameterGroup_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: ""}, nil, nil
		})
	_, err := manager.DeleteParameterGroup(context.TODO(), message.DeleteParameterGroupReq{ParamGroupID: ""})
	assert.Error(t, err)
}

func TestManager_QueryParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().QueryParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name, clusterSpec, clusterVersion string, dbType, hasDefault int, offset, size int) (groups []*parametergroup.ParameterGroup, total int64, err error) {
			return []*parametergroup.ParameterGroup{
				{
					ID:             "1",
					Name:           "test_parameter_group",
					ParentID:       "",
					ClusterSpec:    "8C16G",
					HasDefault:     1,
					DBType:         1,
					GroupType:      1,
					ClusterVersion: "5.0",
					Note:           "test parameter group",
					CreatedAt:      time.Time{},
					UpdatedAt:      time.Time{},
				},
			}, 1, nil
		})
	parameterGroupRW.EXPECT().QueryParametersByGroupId(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) (params []*parametergroup.ParamDetail, err error) {
			return []*parametergroup.ParamDetail{
				{
					Parameter: parametergroup.Parameter{
						ID:             "1",
						Category:       "basic",
						Name:           "param1",
						InstanceType:   "TiKV",
						SystemVariable: "",
						Type:           0,
					},
					DefaultValue: "10",
					Note:         "test parameter",
				},
			}, nil
		})

	resp, page, err := manager.QueryParameterGroup(context.TODO(), message.QueryParameterGroupReq{
		PageRequest:    structs.PageRequest{Page: 1, PageSize: 1},
		Name:           "",
		DBType:         0,
		HasDefault:     0,
		ClusterVersion: "",
		ClusterSpec:    "",
		HasDetail:      true,
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, page.Total)
}

func TestManager_DetailParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, nil, nil
		})
	resp, err := manager.DetailParameterGroup(context.TODO(), message.DetailParameterGroupReq{ParamGroupID: "1"})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_CopyParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, []*parametergroup.ParamDetail{
				{
					Parameter: parametergroup.Parameter{
						ID:             "1",
						Category:       "basic",
						Name:           "param1",
						InstanceType:   "TiKV",
						SystemVariable: "",
						Type:           0,
					},
					DefaultValue: "10",
					Note:         "test parameter",
				},
			}, nil
		})

	parameterGroupRW.EXPECT().CreateParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, pg *parametergroup.ParameterGroup, pgm []*parametergroup.ParameterGroupMapping) (*parametergroup.ParameterGroup, error) {
			resp := &parametergroup.ParameterGroup{
				ID:             "1",
				Name:           "test_parameter_group",
				ParentID:       "",
				ClusterSpec:    "8C16G",
				HasDefault:     1,
				DBType:         1,
				GroupType:      1,
				ClusterVersion: "5.0",
				Note:           "test parameter group",
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
			}
			return resp, nil
		})
	resp, err := manager.CopyParameterGroup(context.TODO(), message.CopyParameterGroupReq{
		ParamGroupID: "1",
		Name:         "copy_parameter_group",
		Note:         "copy parameter group",
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}
