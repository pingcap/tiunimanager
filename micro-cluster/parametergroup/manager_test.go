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
	"errors"
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
	parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, offset, size int) (params []*parametergroup.Parameter, total int64, err error) {
			resp := []*parametergroup.Parameter{
				{
					ID:    "1",
					Type:  0,
					Unit:  "",
					Range: "[\"0\", \"10\"]",
				},
			}
			return resp, 1, nil
		})
	parameterGroupRW.EXPECT().ExistsParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	parameterGroupRW.EXPECT().CreateParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, pg *parametergroup.ParameterGroup, pgm []*parametergroup.ParameterGroupMapping, addParams []message.ParameterInfo) (*parametergroup.ParameterGroup, error) {
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
		AddParams: []message.ParameterInfo{
			{
				Category:       "log",
				Name:           "binlog_cache",
				InstanceType:   "TiDB",
				SystemVariable: "log.binlog_cache",
				Type:           0,
				Unit:           "mb",
				Range:          []string{"0", "1024"},
				HasReboot:      0,
				HasApply:       1,
				UpdateSource:   0,
				ReadOnly:       0,
				Description:    "binlog cache",
				DefaultValue:   "512",
				Note:           "binlog cache",
			},
		},
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_UpdateParameterGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)

	parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, offset, size int) (params []*parametergroup.Parameter, total int64, err error) {
			resp := []*parametergroup.Parameter{
				{
					ID:    "1",
					Type:  0,
					Unit:  "",
					Range: "[\"0\", \"10\"]",
				},
			}
			return resp, 1, nil
		})
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, nil, nil
		})
	parameterGroupRW.EXPECT().ExistsParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	parameterGroupRW.EXPECT().UpdateParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, pg *parametergroup.ParameterGroup, pgm []*parametergroup.ParameterGroupMapping, addParams []message.ParameterInfo, delParams []string) error {
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
		AddParams: []message.ParameterInfo{
			{
				Category:       "log",
				Name:           "binlog_cache",
				InstanceType:   "TiDB",
				SystemVariable: "log.binlog_cache",
				Type:           0,
				Unit:           "mb",
				Range:          []string{"0", "1024"},
				HasReboot:      0,
				HasApply:       1,
				UpdateSource:   0,
				ReadOnly:       0,
				Description:    "binlog cache",
				DefaultValue:   "512",
				Note:           "binlog cache",
			},
		},
		DelParams: []string{},
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_UpdateParameterGroup_Error1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1", HasDefault: 1}, nil, nil
		})
	_, err := manager.UpdateParameterGroup(context.TODO(), message.UpdateParameterGroupReq{
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
	assert.Error(t, err)
}

func TestManager_UpdateParameterGroup_Error2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "", HasDefault: 1}, nil, nil
		})
	_, err := manager.UpdateParameterGroup(context.TODO(), message.UpdateParameterGroupReq{
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
	assert.Error(t, err)
}

func TestManager_DeleteParameterGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
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

func TestManager_DeleteParameterGroup_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
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
	parameterGroupRW.EXPECT().QueryParametersByGroupId(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (params []*parametergroup.ParamDetail, err error) {
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
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, []*parametergroup.ParamDetail{
				{
					Parameter:    parametergroup.Parameter{ID: "1"},
					DefaultValue: "10",
					Note:         "param1",
				},
			}, nil
		})
	resp, err := manager.DetailParameterGroup(context.TODO(), message.DetailParameterGroupReq{ParamGroupID: "1"})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_CopyParameterGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
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

	parameterGroupRW.EXPECT().CreateParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, pg *parametergroup.ParameterGroup, pgm []*parametergroup.ParameterGroupMapping, addParams []message.ParameterInfo) (*parametergroup.ParameterGroup, error) {
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

func TestManager_CopyParameterGroup_Error1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, []*parametergroup.ParamDetail{
				{},
			}, errors.New("failed to create parameter group")
		})

	_, err := manager.CopyParameterGroup(context.TODO(), message.CopyParameterGroupReq{
		ParamGroupID: "1",
		Name:         "copy_parameter_group",
		Note:         "copy parameter group",
	})
	assert.Error(t, err)
}

func TestManager_CopyParameterGroup_Error2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId, paramName string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1", Name: "default_parameter_group"}, []*parametergroup.ParamDetail{
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

	_, err := manager.CopyParameterGroup(context.TODO(), message.CopyParameterGroupReq{
		ParamGroupID: "1",
		Name:         "default_parameter_group",
		Note:         "default parameter group",
	})
	assert.Error(t, err)
}

func TestManager_validateParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	t.Run("normal", func(t *testing.T) {
		parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, offset, size int) (params []*parametergroup.Parameter, total int64, err error) {
				resp := []*parametergroup.Parameter{
					{
						ID:    "1",
						Type:  0,
						Unit:  "",
						Range: "[\"0\", \"10\"]",
					},
				}
				return resp, 1, nil
			})
		err := validateParameter(context.TODO(), []structs.ParameterGroupParameterSampleInfo{
			{
				ID:           "1",
				DefaultValue: "5",
				Note:         "test",
			},
		})
		assert.NoError(t, err)
	})
	t.Run("query parameters error", func(t *testing.T) {
		parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, offset, size int) (params []*parametergroup.Parameter, total int64, err error) {
				return nil, 1, errors.New("query parameters error")
			})
		err := validateParameter(context.TODO(), []structs.ParameterGroupParameterSampleInfo{
			{
				ID:           "1",
				DefaultValue: "5",
				Note:         "test",
			},
		})
		assert.Error(t, err)
	})
	t.Run("range validate error", func(t *testing.T) {
		parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, offset, size int) (params []*parametergroup.Parameter, total int64, err error) {
				resp := []*parametergroup.Parameter{
					{
						ID:    "1",
						Type:  0,
						Unit:  "",
						Range: "[\"0\", \"10\"]",
					},
				}
				return resp, 1, nil
			})
		err := validateParameter(context.TODO(), []structs.ParameterGroupParameterSampleInfo{
			{
				ID:           "1",
				DefaultValue: "20",
				Note:         "test",
			},
		})
		assert.Error(t, err)
	})
	t.Run("range validate error 2", func(t *testing.T) {
		parameterGroupRW.EXPECT().QueryParameters(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, offset, size int) (params []*parametergroup.Parameter, total int64, err error) {
				resp := []*parametergroup.Parameter{
					{
						ID:    "1",
						Type:  0,
						Unit:  "",
						Range: "[\"10\"]",
					},
				}
				return resp, 1, nil
			})
		err := validateParameter(context.TODO(), []structs.ParameterGroupParameterSampleInfo{
			{
				ID:           "1",
				DefaultValue: "20",
				Note:         "test",
			},
		})
		assert.Error(t, err)
	})
}
