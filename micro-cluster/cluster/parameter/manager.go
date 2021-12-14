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
 * @File: manager
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 10:01
*******************************************************************************/

package parameter

import (
	"context"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/convert"
	"github.com/pingcap-inc/tiem/message/cluster"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) QueryClusterParameters(ctx context.Context, req cluster.QueryClusterParametersReq) (resp []cluster.QueryClusterParametersResp, err error) {
	var dbReq *dbpb.DBFindParamsByClusterIdRequest
	err = convert.ConvertObj(req, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("list cluster params req: %v, err: %v", req, err)
		return
	}

	_, err = client.DBClient.FindParamsByClusterId(ctx, dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("list cluster param invoke metadb err: %v", err)
		return
	}

	//resp.Page = convertPage(dbRsp.Page)
	//resp.RespStatus = convertRespStatus(dbRsp.Status)
	//resp.ID = dbRsp.ID
	//if dbRsp.Params != nil {
	//	ps := make([]*clusterpb.ClusterParamDTO, len(dbRsp.Params))
	//	err = convert.ConvertObj(dbRsp.Params, &ps)
	//	if err != nil {
	//		framework.LogWithContext(ctx).Errorf("list cluster params convert resp err: %v", err)
	//		return err
	//	}
	//	resp.Params = ps
	//}
	return resp, nil
}

func (m *Manager) UpdateClusterParameters(ctx context.Context, req cluster.UpdateClusterParametersReq) (resp cluster.UpdateClusterParametersResp, err error) {
	var dbReq *dbpb.DBUpdateClusterParamsRequest
	err = convert.ConvertObj(req, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster params req: %v, err: %v", req, err)
		return
	}

	dbRsp, err := client.DBClient.UpdateClusterParams(ctx, dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster param invoke metadb err: %v", err)
		return
	}
	resp.ClusterID = dbRsp.ClusterId
	return resp, nil
}

func (m *Manager) InspectClusterParams(ctx context.Context, req cluster.InspectClusterParametersReq) (resp cluster.InspectClusterParametersResp, err error) {
	// todo: Reliance on parameter source query implementation
	return
}
