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
 *                                                                            *
 ******************************************************************************/

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/metrics"
	"github.com/pingcap-inc/tiem/proto/clusterservices"
	"time"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/micro-cluster/platform/product"
	"github.com/pingcap-inc/tiem/micro-cluster/user/identification"
	"github.com/pingcap-inc/tiem/micro-cluster/user/tenant"
	"github.com/pingcap-inc/tiem/micro-cluster/user/userinfo"

	"github.com/pingcap-inc/tiem/micro-cluster/platform/config"

	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/changefeed"
	clusterLog "github.com/pingcap-inc/tiem/micro-cluster/cluster/log"
	clusterManager "github.com/pingcap-inc/tiem/micro-cluster/cluster/management"
	clusterParameter "github.com/pingcap-inc/tiem/micro-cluster/cluster/parameter"
	"github.com/pingcap-inc/tiem/micro-cluster/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/micro-cluster/parametergroup"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager"
	"github.com/pingcap-inc/tiem/workflow"

	"github.com/pingcap-inc/tiem/library/framework"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

type ClusterServiceHandler struct {
	resourceManager         *resourcemanager.ResourceManager
	changeFeedManager       *changefeed.Manager
	parameterGroupManager   *parametergroup.Manager
	clusterParameterManager *clusterParameter.Manager
	clusterManager          *clusterManager.Manager
	systemConfigManager     *config.SystemConfigManager
	brManager               backuprestore.BRService
	importexportManager     importexport.ImportExportService
	clusterLogManager       *clusterLog.Manager
	tenantManager           *tenant.Manager
	accountManager          *userinfo.Manager
	authManager             *identification.Manager
	productManager          *product.ProductManager
}

func handleRequest(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse, requestBody interface{}) bool {
	err := json.Unmarshal([]byte(req.GetRequest()), requestBody)
	if err != nil {
		errMsg := fmt.Sprintf("unmarshal request error, request = %s, err = %s", req.GetRequest(), err.Error())
		handleResponse(ctx, resp, errors.NewError(errors.TIEM_UNMARSHAL_ERROR, errMsg), nil, nil)
		return false
	} else {
		return true
	}
}

func handleResponse(ctx context.Context, resp *clusterservices.RpcResponse, err error, responseData interface{}, page *clusterservices.RpcPage) {
	if err == nil {
		data, getDataError := json.Marshal(responseData)
		if getDataError != nil {
			// deal with err uniformly later
			err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, fmt.Sprintf("marshal request data error, data = %v", responseData), getDataError)
		} else {
			// handle data and page
			resp.Code = int32(errors.TIEM_SUCCESS)
			resp.Response = string(data)
			if page != nil {
				resp.Page = page
			}
			return
		}
	}

	if err != nil {
		if finalError, ok := err.(errors.EMError); ok {
			framework.LogWithContext(ctx).Errorf("rpc method failed with error, %s", err.Error())
			resp.Code = int32(finalError.GetCode())
			resp.Message = finalError.GetMsg()
			return
		} else {
			resp.Code = int32(errors.TIEM_UNRECOGNIZED_ERROR)
			resp.Message = err.Error()
		}

		return
	}
}

// handlePanic
// @Description: recover from any panic from user request
// @Parameter ctx
// @Parameter funcName
// @Parameter resp
func handlePanic(ctx context.Context, funcName string, resp *clusterservices.RpcResponse) {
	if r := recover(); r != nil {
		framework.LogWithContext(ctx).Errorf("recover from %s", funcName)
		resp.Code = int32(errors.TIEM_PANIC)
		resp.Message = fmt.Sprintf("%v", r)
	}
}

func NewClusterServiceHandler(fw *framework.BaseFramework) *ClusterServiceHandler {
	handler := new(ClusterServiceHandler)
	handler.resourceManager = resourcemanager.NewResourceManager()
	handler.changeFeedManager = changefeed.GetManager()
	handler.parameterGroupManager = parametergroup.NewManager()
	handler.clusterParameterManager = clusterParameter.NewManager()
	handler.clusterManager = clusterManager.NewClusterManager()
	handler.systemConfigManager = config.NewSystemConfigManager()
	handler.brManager = backuprestore.GetBRService()
	handler.importexportManager = importexport.GetImportExportService()
	handler.clusterLogManager = clusterLog.NewManager()
	handler.tenantManager = tenant.NewTenantManager()
	handler.accountManager = userinfo.NewAccountManager()
	handler.authManager = identification.NewIdentificationManager()
	handler.productManager = product.NewProductManager()

	return handler
}

func (handler *ClusterServiceHandler) CreateChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateChangeFeedTask", resp)

	request := &cluster.CreateChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Create(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}
func (handler *ClusterServiceHandler) DetailChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DetailChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "DetailChangeFeedTask", resp)

	request := &cluster.DetailChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Detail(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}
func (handler *ClusterServiceHandler) PauseChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "PauseChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "PauseChangeFeedTask", resp)

	request := &cluster.PauseChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Pause(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ResumeChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ResumeChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "ResumeChangeFeedTask", resp)

	request := &cluster.ResumeChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Resume(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) DeleteChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteChangeFeedTask", resp)

	request := &cluster.DeleteChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Delete(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateChangeFeedTask", resp)

	request := &cluster.UpdateChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Update(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryChangeFeedTasks(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryChangeFeedTasks", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryChangeFeedTasks", resp)

	request := cluster.QueryChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, total, err := handler.changeFeedManager.Query(ctx, request)
		handleResponse(ctx, resp, err, result, &clusterservices.RpcPage{
			Page:     int32(request.Page),
			PageSize: int32(request.PageSize),
			Total:    int32(total),
		})
	}

	return nil
}

func (handler *ClusterServiceHandler) CreateParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()

	defer metrics.HandleClusterMetrics(start, "CreateParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateParameterGroup", resp)

	request := &message.CreateParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.CreateParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateParameterGroup", resp)

	request := &message.UpdateParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.UpdateParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) DeleteParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteParameterGroup", resp)

	request := &message.DeleteParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.DeleteParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryParameterGroup", resp)

	request := &message.QueryParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.parameterGroupManager.QueryParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) DetailParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DetailParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "DetailParameterGroup", resp)

	request := &message.DetailParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.DetailParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) ApplyParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ApplyParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "ApplyParameterGroup", resp)

	request := &message.ApplyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.ApplyParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) CopyParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CopyParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "CopyParameterGroup", resp)

	request := &message.CopyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.CopyParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryClusterParameters(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryClusterParameters", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryClusterParameters", resp)

	request := &cluster.QueryClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.clusterParameterManager.QueryClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateClusterParameters(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateClusterParameters", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateClusterParameters", resp)

	request := &cluster.UpdateClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.UpdateClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request, true)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) InspectClusterParameters(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "InspectClusterParameters", int(resp.GetCode()))
	defer handlePanic(ctx, "InspectClusterParameters", resp)

	request := &cluster.InspectClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.InspectClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryClusterLog(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryClusterLog", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryClusterLog", resp)

	request := &cluster.QueryClusterLogReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.clusterLogManager.QueryClusterLog(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateCluster", resp)

	request := cluster.CreateClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.CreateCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestoreNewCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "RestoreNewCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "RestoreNewCluster", resp)

	request := cluster.RestoreNewClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.RestoreNewCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ScaleOutCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ScaleOutCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "ScaleOutCluster", resp)

	request := cluster.ScaleOutClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.ScaleOut(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ScaleInCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ScaleInCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "ScaleInCluster", resp)

	request := cluster.ScaleInClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.ScaleIn(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CloneCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CloneCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "CloneCluster", resp)

	request := cluster.CloneClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.Clone(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler ClusterServiceHandler) TakeoverClusters(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "TakeoverClusters", int(resp.GetCode()))
	defer handlePanic(ctx, "TakeoverClusters", resp)

	request := cluster.TakeoverClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.Takeover(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryCluster", resp)

	request := cluster.QueryClustersReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, total, err := c.clusterManager.QueryCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, &clusterservices.RpcPage{
			Page:     int32(request.Page),
			PageSize: int32(request.PageSize),
			Total:    int32(total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteCluster", resp)

	request := cluster.DeleteClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.DeleteCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestartCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "RestartCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "RestartCluster", resp)

	request := cluster.RestartClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.RestartCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) StopCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "StopCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "StopCluster", resp)

	request := cluster.StopClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.StopCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DetailCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "DetailCluster", resp)

	request := cluster.QueryClusterDetailReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.DetailCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ExportData", int(resp.GetCode()))
	defer handlePanic(ctx, "ExportData", resp)

	exportReq := message.DataExportReq{}

	if handleRequest(ctx, req, resp, &exportReq) {
		result, err := c.importexportManager.ExportData(framework.NewBackgroundMicroCtx(ctx, false), exportReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ImportData", int(resp.GetCode()))
	defer handlePanic(ctx, "ImportData", resp)

	importReq := message.DataImportReq{}

	if handleRequest(ctx, req, resp, &importReq) {
		result, err := c.importexportManager.ImportData(framework.NewBackgroundMicroCtx(ctx, false), importReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryDataTransport(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryDataTransport", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryDataTransport", resp)

	queryReq := message.QueryDataImportExportRecordsReq{}

	if handleRequest(ctx, req, resp, &queryReq) {
		result, page, err := c.importexportManager.QueryDataTransportRecords(framework.NewBackgroundMicroCtx(ctx, false), queryReq)
		handleResponse(ctx, resp, err, result, &clusterservices.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) DeleteDataTransportRecord(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteDataTransportRecord", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteDataTransportRecord", resp)

	deleteReq := message.DeleteImportExportRecordReq{}

	if handleRequest(ctx, req, resp, &deleteReq) {
		result, err := c.importexportManager.DeleteDataTransportRecord(framework.NewBackgroundMicroCtx(ctx, false), deleteReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) GetSystemConfig(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetSystemConfig", int(resp.GetCode()))
	defer handlePanic(ctx, "GetSystemConfig", resp)

	getReq := message.GetSystemConfigReq{}

	if handleRequest(ctx, req, resp, &getReq) {
		result, err := c.systemConfigManager.GetSystemConfig(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateBackup", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateBackup", resp)

	backupReq := cluster.BackupClusterDataReq{}

	if handleRequest(ctx, req, resp, &backupReq) {
		result, err := c.brManager.BackupCluster(framework.NewBackgroundMicroCtx(ctx, false), backupReq, true)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DeleteBackupRecords(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteBackupRecord", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteBackupRecord", resp)

	deleteReq := cluster.DeleteBackupDataReq{}

	if handleRequest(ctx, req, resp, &deleteReq) {
		result, err := c.brManager.DeleteBackupRecords(framework.NewBackgroundMicroCtx(ctx, false), deleteReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "SaveBackupStrategy", int(resp.GetCode()))
	defer handlePanic(ctx, "SaveBackupStrategy", resp)

	saveReq := cluster.SaveBackupStrategyReq{}

	if handleRequest(ctx, req, resp, &saveReq) {
		result, err := c.brManager.SaveBackupStrategy(framework.NewBackgroundMicroCtx(ctx, false), saveReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetBackupStrategy", int(resp.GetCode()))
	defer handlePanic(ctx, "GetBackupStrategy", resp)

	getReq := cluster.GetBackupStrategyReq{}

	if handleRequest(ctx, req, resp, &getReq) {
		result, err := c.brManager.GetBackupStrategy(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryBackupRecords(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryBackupRecords", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryBackupRecords", resp)

	queryReq := cluster.QueryBackupRecordsReq{}

	if handleRequest(ctx, req, resp, &queryReq) {
		result, page, err := c.brManager.QueryClusterBackupRecords(framework.NewBackgroundMicroCtx(ctx, false), queryReq)
		handleResponse(ctx, resp, err, result, &clusterservices.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) GetDashboardInfo(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DescribeDashboard", int(resp.GetCode()))
	defer handlePanic(ctx, "DescribeDashboard", resp)

	dashboardReq := cluster.GetDashboardInfoReq{}

	if handleRequest(ctx, req, resp, &dashboardReq) {
		result, err := c.clusterManager.GetClusterDashboardInfo(framework.NewBackgroundMicroCtx(ctx, false), dashboardReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) GetMonitorInfo(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetMonitorInfo", int(resp.GetCode()))
	defer handlePanic(ctx, "GetMonitorInfo", resp)

	request := &cluster.QueryMonitorInfoReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := c.clusterManager.GetMonitorInfo(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (c ClusterServiceHandler) ListFlows(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ListFlows", int(resp.GetCode()))
	defer handlePanic(ctx, "ListFlows", resp)

	listReq := message.QueryWorkFlowsReq{}
	if handleRequest(ctx, req, resp, &listReq) {
		manager := workflow.GetWorkFlowService()
		result, page, err := manager.ListWorkFlows(framework.NewBackgroundMicroCtx(ctx, false), listReq)
		handleResponse(ctx, resp, err, result, &clusterservices.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c *ClusterServiceHandler) DetailFlow(ctx context.Context, request *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DetailFlow", int(resp.GetCode()))
	defer handlePanic(ctx, "DetailFlow", resp)

	detailReq := message.QueryWorkFlowDetailReq{}
	if handleRequest(ctx, request, resp, &detailReq) {
		manager := workflow.GetWorkFlowService()
		result, err := manager.DetailWorkFlow(framework.NewBackgroundMicroCtx(ctx, false), detailReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) Login(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "Login", int(resp.GetCode()))
	defer handlePanic(ctx, "Login", resp)

	loginReq := message.LoginReq{}
	if handleRequest(ctx, req, resp, &loginReq) {
		result, err := c.authManager.Login(framework.NewBackgroundMicroCtx(ctx, false), loginReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) Logout(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "Logout", int(resp.GetCode()))
	defer handlePanic(ctx, "Login", resp)

	logoutReq := message.LogoutReq{}
	if handleRequest(ctx, req, resp, &logoutReq) {
		result, err := c.authManager.Logout(framework.NewBackgroundMicroCtx(ctx, false), logoutReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) VerifyIdentity(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "VerifyIdentity", int(resp.GetCode()))
	defer handlePanic(ctx, "VerifyIdentity", resp)

	verReq := message.AccessibleReq{}
	if handleRequest(ctx, req, resp, &verReq) {
		result, err := c.authManager.Accessible(framework.NewBackgroundMicroCtx(ctx, false), verReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ImportHosts(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ImportHosts", int(resp.GetCode()))
	defer handlePanic(ctx, "ImportHosts", resp)

	reqStruct := message.ImportHostsReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		flowIds, hostIds, err := handler.resourceManager.ImportHosts(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.Hosts)
		var rsp message.ImportHostsResp
		if err == nil {
			rsp.HostIDS = hostIds
			for _, flowId := range flowIds {
				rsp.FlowInfo = append(rsp.FlowInfo, structs.AsyncTaskWorkFlowInfo{WorkFlowID: flowId})
			}
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteHosts(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteHosts", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteHosts", resp)

	reqStruct := message.DeleteHostsReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		flowIds, err := handler.resourceManager.DeleteHosts(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs)
		var rsp message.DeleteHostsResp
		if err == nil {
			for _, flowId := range flowIds {
				rsp.FlowInfo = append(rsp.FlowInfo, structs.AsyncTaskWorkFlowInfo{WorkFlowID: flowId})
			}
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryHosts(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryHosts", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryHosts", resp)

	reqStruct := message.QueryHostsReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		filter := reqStruct.GetHostFilter()
		page := reqStruct.GetPage()

		hosts, err := handler.resourceManager.QueryHosts(framework.NewBackgroundMicroCtx(ctx, false), filter, page)
		var rsp message.QueryHostsResp
		if err == nil {
			rsp.Hosts = hosts
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostReserved(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateHostReserved", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateHostReserved", resp)

	reqStruct := message.UpdateHostReservedReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		err := handler.resourceManager.UpdateHostReserved(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Reserved)
		var rsp message.UpdateHostReservedResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostStatus(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateHostStatus", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateHostReserved", resp)

	reqStruct := message.UpdateHostStatusReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		err := handler.resourceManager.UpdateHostStatus(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Status)
		var rsp message.UpdateHostStatusResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetHierarchy(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetHierarchy", int(resp.GetCode()))
	defer handlePanic(ctx, "GetHierarchy", resp)

	reqStruct := message.GetHierarchyReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		filter := reqStruct.GetHostFilter()

		root, err := handler.resourceManager.GetHierarchy(framework.NewBackgroundMicroCtx(ctx, false), filter, reqStruct.Level, reqStruct.Depth)
		var rsp message.GetHierarchyResp
		if err == nil {
			rsp.Root = *root
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetStocks(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetStocks", int(resp.GetCode()))
	defer handlePanic(ctx, "GetStocks", resp)

	reqStruct := message.GetStocksReq{}

	if handleRequest(ctx, req, resp, &reqStruct) {
		location := reqStruct.GetLocation()
		hostFilter := reqStruct.GetHostFilter()
		diskFilter := reqStruct.GetDiskFilter()

		stocks, err := handler.resourceManager.GetStocks(framework.NewBackgroundMicroCtx(ctx, false), location, hostFilter, diskFilter)
		var rsp message.GetStocksResp
		if err == nil {
			rsp.Stocks = *stocks
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CreateZones(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateZones", int(response.GetCode()))

	req := message.CreateZonesReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.CreateZones(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteZone(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteZone", int(response.GetCode()))

	req := message.DeleteZoneReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.DeleteZones(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryZones(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryZones", int(response.GetCode()))

	req := message.QueryZonesReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.QueryZones(ctx)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CreateProduct(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateProduct", int(response.GetCode()))

	req := message.CreateProductReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.CreateProduct(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteProduct(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteProduct", int(response.GetCode()))

	req := message.DeleteProductReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.DeleteProduct(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryProducts(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryProducts", int(response.GetCode()))

	req := message.QueryProductsReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.QueryProducts(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryProductDetail(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryProductDetail", int(response.GetCode()))

	req := message.QueryProductDetailReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.QueryProductDetail(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CreateSpecs(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateSpecs", int(response.GetCode()))

	req := message.CreateSpecsReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.CreateSpecs(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteSpecs(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteSpecs", int(response.GetCode()))

	req := message.DeleteSpecsReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.DeleteSpecs(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QuerySpecs(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QuerySpecs", int(response.GetCode()))

	req := message.QuerySpecsReq{}
	if handleRequest(ctx, request, response, &req) {
		resp, err := handler.productManager.QuerySpecs(ctx)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}
