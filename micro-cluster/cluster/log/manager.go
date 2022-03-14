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
 * @File: manager.go
 * @Description: cluster log manager
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/17 13:43
*******************************************************************************/

package log

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/pingcap-inc/tiem/proto/clusterservices"
	"github.com/pingcap-inc/tiem/util/convert"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/workflow"

	"github.com/pingcap-inc/tiem/common/errors"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/message/cluster"
)

type Manager struct{}

var manager *Manager
var once sync.Once

func NewManager() *Manager {
	once.Do(func() {
		if manager == nil {
			workflowManager := workflow.GetWorkFlowService()
			workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowBuildLogConfig, &buildLogConfigDefine)

			manager = &Manager{}
		}
	})
	return manager
}

var buildLogConfigDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowBuildLogConfig,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":   {"collect", "success", "fail", workflow.SyncFuncNode, collectorClusterLogConfig},
		"success": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
		"fail":    {"end", "", "", workflow.SyncFuncNode, defaultEnd},
	},
}

type Service interface {
	BuildClusterLogConfig(ctx context.Context, clusterId string) (flowID string, err error)
}

func GetService() Service {
	return NewManager()
}

func (m Manager) BuildClusterLogConfig(ctx context.Context, clusterId string) (flowID string, err error) {
	framework.LogWithContext(ctx).Infof("begin build cluster log, req clusterId: %+v", clusterId)
	defer framework.LogWithContext(ctx).Infof("end build cluster log")

	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := meta.Get(ctx, clusterId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser [%s] meta from db error: %s", clusterId, err.Error())
		return
	}

	if flow, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, clusterMeta.Cluster.ID, workflow.BizTypeCluster, buildLogConfigDefine.FlowName); err != nil {
		framework.LogWithContext(ctx).Errorf("create flow failed, clusterID = %s, error = %s", clusterMeta.Cluster.ID, err.Error())
		return "", err
	} else {
		flowID = flow.Flow.ID
		flow.Context.SetData(contextClusterMeta, clusterMeta)
		if err = workflow.GetWorkFlowService().AsyncStart(ctx, flow); err != nil {
			framework.LogWithContext(ctx).Errorf("start flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, clusterMeta.Cluster.ID, err.Error())
			return flowID, err
		}
		framework.LogWithContext(ctx).Infof("create flow %s succeed, clusterID = %s", flow.Flow.Name, clusterMeta.Cluster.ID)
	}
	return flowID, nil
}

// QueryClusterLog
// @Description: query cluster log
// @Receiver m
// @Parameter ctx
// @Parameter req
// @return resp
// @return page
// @return err
func (m Manager) QueryClusterLog(ctx context.Context, req cluster.QueryClusterLogReq) (resp cluster.QueryClusterLogResp, page *clusterservices.RpcPage, err error) {
	buf, err := prepareSearchParams(ctx, req)
	if err != nil {
		return resp, page, err
	}
	esResp, err := framework.Current.GetElasticsearchClient().Search(logIndexPrefix, &buf, (req.Page-1)*req.PageSize, req.PageSize)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s query log, search es err: %v", req.ClusterID, err)
		return resp, page, errors.NewErrorf(errors.TIEM_LOG_QUERY_FAILED, errors.TIEM_LOG_QUERY_FAILED.Explain(), err)
	}
	return handleResult(ctx, req, esResp)
}

func prepareSearchParams(ctx context.Context, req cluster.QueryClusterLogReq) (buf bytes.Buffer, err error) {
	// Get cluster
	_, err = models.GetClusterReaderWriter().Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s query log, get cluster err: %v", req.ClusterID, err)
		return buf, errors.NewErrorf(errors.TIEM_CLUSTER_NOT_FOUND, errors.TIEM_CLUSTER_NOT_FOUND.Explain())
	}

	buf = bytes.Buffer{}
	query, err := buildSearchClusterReqParams(req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s query log, build search params err: %v", req.ClusterID, err)
		return buf, errors.NewErrorf(errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR, errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR.Explain(), err)
	}
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s query log, encode err: %v", req.ClusterID, err)
		return buf, errors.NewErrorf(errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR, errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR.Explain(), err)
	}
	return buf, nil
}

func handleResult(ctx context.Context, req cluster.QueryClusterLogReq, esResp *esapi.Response) (resp cluster.QueryClusterLogResp, page *clusterservices.RpcPage, err error) {
	if esResp.IsError() || esResp.StatusCode != 200 {
		framework.LogWithContext(ctx).Errorf("cluster %s query log, search es err: %v", req.ClusterID, err)
		return resp, page, errors.NewErrorf(errors.TIEM_LOG_QUERY_FAILED, errors.TIEM_LOG_QUERY_FAILED.Explain(), esResp.String())
	}
	var esResult ElasticSearchResult
	if err = json.NewDecoder(esResp.Body).Decode(&esResult); err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s query log, decoder err: %v", req.ClusterID, err)
		return resp, page, errors.NewErrorf(errors.TIEM_LOG_QUERY_FAILED, errors.TIEM_LOG_QUERY_FAILED.Explain(), err)
	}

	resp = cluster.QueryClusterLogResp{
		Took: esResult.Took,
	}
	resp.Results = make([]structs.ClusterLogItem, 0)
	for _, hit := range esResult.Hits.Hits {
		var hitItem SearchClusterLogSourceItem
		err := convert.ConvertObj(hit.Source, &hitItem)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("cluster %s query log, convert obj err: %v", req.ClusterID, err)
			return resp, page, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
		}

		resp.Results = append(resp.Results, structs.ClusterLogItem{
			Index:      hit.Index,
			Id:         hit.Id,
			Level:      hitItem.Log.Level,
			SourceLine: hitItem.Log.Logger,
			Message:    hitItem.Message,
			Ip:         hitItem.Ip,
			ClusterId:  hitItem.ClusterId,
			Module:     hitItem.Fileset.Name,
			Ext:        hitItem.Tidb,
			Timestamp:  time.Unix(hitItem.Timestamp.Unix(), 0).Format(DateFormat),
		})
	}

	page = &clusterservices.RpcPage{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Total:    int32(esResult.Hits.Total.Value),
	}
	return resp, page, nil
}

// buildSearchClusterReqParams build request elasticsearch parameters
//response e.g.:
//{
//  "query": {
//    "bool": {
//      "filter": [
//        {
//          "term": {
//            "fileset.name": "pd"
//          }
//        },
//        {
//          "match": {
//            "message": "gc worker"
//          }
//        },
//        {
//          "range": {
//            "@timestamp": {
//              "gte": 1631840244000,
//              "lte": 1631876244000
//            }
//          }
//        }
//      ]
//    }
//  }
//}
func buildSearchClusterReqParams(req cluster.QueryClusterLogReq) (map[string]interface{}, error) {
	filters := make([]interface{}, 0)
	// required parameters: clusterId
	filters = append(filters, map[string]interface{}{
		"term": map[string]interface{}{
			"clusterId": req.ClusterID,
		},
	})
	// option parameters: module
	if req.Module != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"fileset.name": strings.ToLower(req.Module),
			},
		})
	}
	// option parameters: level
	if req.Level != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"log.level": strings.ToUpper(req.Level),
			},
		})
	}
	// option parameters: ip
	if req.Ip != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"ip": req.Ip,
			},
		})
	}
	// option parameters: message
	if req.Message != "" {
		filters = append(filters, map[string]interface{}{
			"match": map[string]interface{}{
				"message": req.Message,
			},
		})
	}
	// option parameters: startTime, endTime
	tsFilter, err := FilterTimestamp(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	if len(tsFilter) > 0 {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": tsFilter,
			},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		},
	}
	return query, nil
}
