/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

	"runtime"
	"runtime/debug"
	"time"

	jsoniter "github.com/json-iterator/go"
	sqleditorparam "github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
	"github.com/pingcap/tiunimanager/micro-cluster/dataapps/sqleditor"

	platformLog "github.com/pingcap/tiunimanager/micro-cluster/platform/log"

	"github.com/pingcap/tiunimanager/micro-cluster/platform/check"
	"github.com/pingcap/tiunimanager/micro-cluster/platform/system"

	"github.com/pingcap/tiunimanager/common/constants"

	"github.com/pingcap/tiunimanager/metrics"
	"github.com/pingcap/tiunimanager/micro-cluster/user/rbac"
	"github.com/pingcap/tiunimanager/proto/clusterservices"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/micro-cluster/platform/product"
	"github.com/pingcap/tiunimanager/micro-cluster/user/account"
	"github.com/pingcap/tiunimanager/micro-cluster/user/identification"

	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/backuprestore"

	"github.com/pingcap/tiunimanager/micro-cluster/platform/config"

	"github.com/pingcap/tiunimanager/micro-cluster/cluster/changefeed"
	clusterLog "github.com/pingcap/tiunimanager/micro-cluster/cluster/log"
	clusterManager "github.com/pingcap/tiunimanager/micro-cluster/cluster/management"
	clusterParameter "github.com/pingcap/tiunimanager/micro-cluster/cluster/parameter"
	switchoverManager "github.com/pingcap/tiunimanager/micro-cluster/cluster/switchover"
	"github.com/pingcap/tiunimanager/micro-cluster/datatransfer/importexport"
	"github.com/pingcap/tiunimanager/micro-cluster/parametergroup"
	"github.com/pingcap/tiunimanager/micro-cluster/resourcemanager"
	workflow "github.com/pingcap/tiunimanager/workflow2"

	"github.com/pingcap/tiunimanager/library/framework"
)

var TiUniManagerClusterServiceName = "go.micro.tiunimanager.cluster"

type ClusterServiceHandler struct {
	resourceManager         *resourcemanager.ResourceManager
	changeFeedManager       *changefeed.Manager
	switchoverManager       *switchoverManager.Manager
	parameterGroupManager   *parametergroup.Manager
	clusterParameterManager *clusterParameter.Manager
	clusterManager          *clusterManager.Manager
	systemConfigManager     *config.SystemConfigManager
	systemManager           *system.SystemManager
	brManager               backuprestore.BRService
	importexportManager     importexport.ImportExportService
	clusterLogManager       *clusterLog.Manager
	accountManager          *account.Manager
	authManager             *identification.Manager
	productManager          *product.Manager
	rbacManager             rbac.RBACService
	checkManager            check.CheckService
	platformLogManager      *platformLog.Manager
	sqleditorManager        *sqleditor.Manager
}

func handleRequest(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse, requestBody interface{}, permissions []structs.RbacPermission) bool {
	if len(permissions) > 0 {
		result, err := rbac.GetRBACService().CheckPermissionForUser(ctx, message.CheckPermissionForUserReq{UserID: framework.GetUserIDFromContext(ctx), Permissions: permissions})
		if err != nil {
			errMsg := fmt.Sprintf("check permission for user %s error, permission %+v, err = %s", framework.GetUserIDFromContext(ctx), permissions, err.Error())
			handleResponse(ctx, resp, errors.NewError(errors.TIUNIMANAGER_RBAC_PERMISSION_CHECK_FAILED, errMsg), nil, nil)
			return false
		}
		if !result.Result {
			errMsg := fmt.Sprintf("check permission for user %s has no permisson %+v", framework.GetUserIDFromContext(ctx), permissions)
			handleResponse(ctx, resp, errors.NewError(errors.TIUNIMANAGER_RBAC_PERMISSION_CHECK_FAILED, errMsg), nil, nil)
			return false
		}
	}

	err := json.Unmarshal([]byte(req.GetRequest()), requestBody)

	if err != nil {
		errMsg := fmt.Sprintf("unmarshal request failed, err = %s", err.Error())
		handleResponse(ctx, resp, errors.NewError(errors.TIUNIMANAGER_UNMARSHAL_ERROR, errMsg), nil, nil)
		return false
	} else {
		if pc, _, _, ok := runtime.Caller(1); ok {
			desensitizeLog(ctx, runtime.FuncForPC(pc).Name(), "start", requestBody)
		}
		return true
	}
}

func desensitizeLog(ctx context.Context, methodName, event string, data interface{}) string {
	info, _ := jsoniter.MarshalToString(data)
	framework.LogWithContext(ctx).
		WithField("micro-method", methodName).
		WithField("event", event).
		Infof(info)
	return info
}

func handleResponse(ctx context.Context, resp *clusterservices.RpcResponse, err error, responseData interface{}, page *clusterservices.RpcPage) {
	if err == nil {
		data, getDataError := json.Marshal(responseData)
		if getDataError != nil {
			// deal with err uniformly later
			err = errors.WrapError(errors.TIUNIMANAGER_MARSHAL_ERROR, "marshal response data failed", getDataError)
		} else {
			if pc, _, _, ok := runtime.Caller(1); ok {
				desensitizeLog(ctx, runtime.FuncForPC(pc).Name(), "end", responseData)
			}
			// handle data and page
			resp.Code = int32(errors.TIUNIMANAGER_SUCCESS)
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
			resp.Message = finalError.Error()
			return
		} else {
			resp.Code = int32(errors.TIUNIMANAGER_UNRECOGNIZED_ERROR)
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
		framework.LogWithContext(ctx).Errorf("recover from %s, stacktrace %s", funcName, string(debug.Stack()))
		resp.Code = int32(errors.TIUNIMANAGER_PANIC)
		resp.Message = fmt.Sprintf("%v", r)
	}
}

func NewClusterServiceHandler(fw *framework.BaseFramework) *ClusterServiceHandler {
	jsoniter.RegisterTypeEncoder("structs.SensitiveText", structs.SensitiveTextEncoder{})
	handler := new(ClusterServiceHandler)
	handler.resourceManager = resourcemanager.NewResourceManager()
	handler.changeFeedManager = changefeed.GetManager()
	handler.parameterGroupManager = parametergroup.NewManager()
	handler.clusterParameterManager = clusterParameter.NewManager()
	handler.clusterManager = clusterManager.NewClusterManager()
	handler.switchoverManager = switchoverManager.GetManager()
	handler.systemConfigManager = config.NewSystemConfigManager()
	handler.systemManager = system.GetSystemManager()
	handler.brManager = backuprestore.GetBRService()
	handler.importexportManager = importexport.GetImportExportService()
	handler.clusterLogManager = clusterLog.NewManager()
	handler.accountManager = account.NewAccountManager()
	handler.authManager = identification.NewIdentificationManager()
	handler.productManager = product.NewManager()
	handler.rbacManager = rbac.GetRBACService()
	handler.checkManager = check.GetCheckService()
	handler.platformLogManager = platformLog.NewManager()
	handler.sqleditorManager = sqleditor.GetManager()
	return handler
}

func (handler *ClusterServiceHandler) CreateSQLFile(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateSQLFile", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateSQLFile", resp)

	request := &sqleditorparam.SQLFileReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionCreate)}}) {
		result, err := handler.sqleditorManager.CreateSQLFile(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateSQLFile(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateSQLFile", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateSQLFile", resp)

	request := &sqleditorparam.SQLFileUpdateReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.sqleditorManager.UpdateSQLFile(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteSQLFile(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteSQLFile", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteSQLFile", resp)

	request := &sqleditorparam.SQLFileDeleteReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.sqleditorManager.DeleteSQLFile(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ShowSQLFile(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ShowSQLFile", int(resp.GetCode()))
	defer handlePanic(ctx, "ShowSQLFile", resp)

	request := &sqleditorparam.ShowSQLFileReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.sqleditorManager.ShowSQLFile(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ListSQLFile(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ListSQLFile", int(resp.GetCode()))
	defer handlePanic(ctx, "ListSQLFile", resp)

	request := &sqleditorparam.ListSQLFileReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.sqleditorManager.ListSqlFile(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ShowTableMeta(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ShowTableMeta", int(resp.GetCode()))
	defer handlePanic(ctx, "ShowTableMeta", resp)

	request := &sqleditorparam.ShowTableMetaReq{}
	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.sqleditorManager.ShowTableMeta(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) Statements(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "Statements", int(resp.GetCode()))
	defer handlePanic(ctx, "Statements", resp)

	request := &sqleditorparam.StatementParam{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.sqleditorManager.Statements(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ShowClusterMeta(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "ShowClusterMeta", int(resp.GetCode()))
	defer handlePanic(ctx, "ShowClusterMeta", resp)

	request := &sqleditorparam.ShowClusterMetaReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.sqleditorManager.ShowClusterMeta(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}
func (handler *ClusterServiceHandler) CreateSession(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateSession", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateSession", resp)

	request := &sqleditorparam.CreateSessionReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionAll)}}) {
		result, err := handler.sqleditorManager.CreateSession(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CloseSession(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CloseSession", int(resp.GetCode()))
	defer handlePanic(ctx, "CloseSession", resp)

	request := &sqleditorparam.CloseSessionReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSQLEditor), Action: string(constants.RbacActionAll)}}) {
		result, err := handler.sqleditorManager.CloseSession(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) MasterSlaveSwitchover(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "MasterSlaveSwitchover", int(response.GetCode()))
	defer handlePanic(ctx, "MasterSlaveSwitchover", response)

	framework.LogWithContext(ctx).Info("master/slave switchover")
	reqBody := &cluster.MasterSlaveClusterSwitchoverReq{}

	framework.LogWithContext(ctx).Info("MasterSlaveSwitchover: before handleRequest")
	if handleRequest(ctx, request, response, reqBody, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionCreate)}}) {
		framework.LogWithContext(ctx).Info("MasterSlaveSwitchover: after handleRequest")
		framework.LogWithContext(ctx).Info("MasterSlaveSwitchover: before  handler.switchoverManager.Switchover")
		result, err := handler.switchoverManager.Switchover(ctx, reqBody)
		framework.LogWithContext(ctx).Info("MasterSlaveSwitchover: after  handler.switchoverManager.Switchover", result, err)
		handleResponse(ctx, response, err, result, nil)
	}
	framework.LogWithContext(ctx).Info("MasterSlaveSwitchover: ret")
	return nil
}

func (handler *ClusterServiceHandler) CreateChangeFeedTask(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateChangeFeedTask", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateChangeFeedTask", resp)

	request := &cluster.CreateChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionCreate)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionDelete)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCDC), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionCreate)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionDelete)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.clusterParameterManager.ApplyParameterGroup(framework.NewBackgroundMicroCtx(ctx, false), *request, true)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) CopyParameterGroup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CopyParameterGroup", int(resp.GetCode()))
	defer handlePanic(ctx, "CopyParameterGroup", resp)

	request := &message.CopyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionCreate)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.clusterParameterManager.UpdateClusterParameters(framework.NewBackgroundMicroCtx(ctx, false), *request, true)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) InspectClusterParameters(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "InspectClusterParameters", int(resp.GetCode()))
	defer handlePanic(ctx, "InspectClusterParameters", resp)

	request := &cluster.InspectParametersReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceParameter), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, page, err := handler.clusterLogManager.QueryClusterLog(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryPlatformLog(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryPlatformLog", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryPlatformLog", resp)

	request := &message.QueryPlatformLogReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, page, err := handler.platformLogManager.QueryPlatformLog(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateCluster", resp)

	request := cluster.CreateClusterReq{}

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionCreate)}}) {
		result, err := c.clusterManager.CreateCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) PreviewCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "PreviewCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "PreviewCluster", resp)

	request := cluster.CreateClusterReq{}

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, err := c.clusterManager.PreviewCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) PreviewScaleOutCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "PreviewScaleOutCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "PreviewScaleOutCluster", resp)

	request := cluster.ScaleOutClusterReq{}

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, err := c.clusterManager.PreviewScaleOutCluster(framework.NewBackgroundMicroCtx(ctx, false), request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestoreNewCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "RestoreNewCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "RestoreNewCluster", resp)

	request := cluster.RestoreNewClusterReq{}

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionCreate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionCreate)}}) {
		result, err := handler.clusterManager.Takeover(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler ClusterServiceHandler) DeleteMetadataPhysically(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteMetadataPhysically", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteMetadataPhysically", resp)

	request := cluster.DeleteMetadataPhysicallyReq{}

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionDelete)}}) {
		result, err := handler.clusterManager.DeleteMetadataPhysically(framework.NewBackgroundMicroCtx(ctx, false), request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) (err error) {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryCluster", resp)

	request := cluster.QueryClustersReq{}

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, total, err := c.clusterManager.QueryCluster(ctx, request)
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionDelete)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, &exportReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &importReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &queryReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, &deleteReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionDelete)}}) {
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

	if handleRequest(ctx, req, resp, &getReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionRead)}}) {
		result, err := c.systemConfigManager.GetSystemConfig(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) UpdateSystemConfig(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateSystemConfig", int(resp.GetCode()))
	defer handlePanic(ctx, "UpdateSystemConfig", resp)

	updateReq := message.UpdateSystemConfigReq{}

	if handleRequest(ctx, req, resp, &updateReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionUpdate)}}) {
		result, err := c.systemConfigManager.UpdateSystemConfig(framework.NewBackgroundMicroCtx(ctx, false), updateReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) GetSystemInfo(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetSystemInfo", int(resp.GetCode()))
	defer handlePanic(ctx, "GetSystemInfo", resp)

	getReq := message.GetSystemInfoReq{}

	if handleRequest(ctx, req, resp, &getReq, []structs.RbacPermission{}) {
		result, err := c.systemManager.GetSystemInfo(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateBackup", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateBackup", resp)

	backupReq := cluster.BackupClusterDataReq{}

	if handleRequest(ctx, req, resp, &backupReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
		result, err := c.brManager.BackupCluster(framework.NewBackgroundMicroCtx(ctx, false), backupReq, true)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) CancelBackup(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CancelBackup", int(resp.GetCode()))
	defer handlePanic(ctx, "CancelBackup", resp)

	cancelReq := cluster.CancelBackupReq{}

	if handleRequest(ctx, req, resp, &cancelReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
		result, err := c.brManager.CancelBackup(framework.NewBackgroundMicroCtx(ctx, false), cancelReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DeleteBackupRecords(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteBackupRecord", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteBackupRecord", resp)

	deleteReq := cluster.DeleteBackupDataReq{}

	if handleRequest(ctx, req, resp, &deleteReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionDelete)}}) {
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

	if handleRequest(ctx, req, resp, &saveReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
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

	if handleRequest(ctx, req, resp, &getReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, &queryReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, &dashboardReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
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

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
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
	if handleRequest(ctx, req, resp, &listReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceWorkflow), Action: string(constants.RbacActionRead)}}) {
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
	if handleRequest(ctx, request, resp, &detailReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceWorkflow), Action: string(constants.RbacActionRead)}}) {
		manager := workflow.GetWorkFlowService()
		result, err := manager.DetailWorkFlow(framework.NewBackgroundMicroCtx(ctx, false), detailReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) StartFlow(ctx context.Context, request *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "StartFlow", int(resp.GetCode()))
	defer handlePanic(ctx, "StartFlow", resp)

	startReq := message.StartWorkFlowReq{}
	if handleRequest(ctx, request, resp, &startReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceWorkflow), Action: string(constants.RbacActionUpdate)}}) {
		manager := workflow.GetWorkFlowService()
		err := manager.Start(framework.NewBackgroundMicroCtx(ctx, false), startReq.WorkFlowID)
		handleResponse(ctx, resp, err, message.StartWorkFlowResp{}, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) StopFlow(ctx context.Context, request *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "StopFlow", int(resp.GetCode()))
	defer handlePanic(ctx, "StopFlow", resp)

	stopReq := message.StartWorkFlowReq{}
	if handleRequest(ctx, request, resp, &stopReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceWorkflow), Action: string(constants.RbacActionUpdate)}}) {
		manager := workflow.GetWorkFlowService()
		err := manager.Stop(framework.NewBackgroundMicroCtx(ctx, false), stopReq.WorkFlowID)
		handleResponse(ctx, resp, err, message.StopWorkFlowResp{}, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) Login(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "Login", int(resp.GetCode()))
	defer handlePanic(ctx, "Login", resp)

	loginReq := message.LoginReq{}
	if handleRequest(ctx, req, resp, &loginReq, []structs.RbacPermission{}) {
		result, err := c.authManager.Login(ctx, loginReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) Logout(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "Logout", int(resp.GetCode()))
	defer handlePanic(ctx, "Login", resp)

	logoutReq := message.LogoutReq{}
	if handleRequest(ctx, req, resp, &logoutReq, []structs.RbacPermission{}) {
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
	if handleRequest(ctx, req, resp, &verReq, []structs.RbacPermission{}) {
		result, err := c.authManager.Accessible(ctx, verReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) BindRolesForUser(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "BindRoleForUser", int(resp.GetCode()))
	defer handlePanic(ctx, "BindRoleForUser", resp)

	bindReq := message.BindRolesForUserReq{}
	if handleRequest(ctx, req, resp, &bindReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionUpdate)}}) {
		result, err := c.rbacManager.BindRolesForUser(framework.NewBackgroundMicroCtx(ctx, false), bindReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) CreateRbacRole(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateRbacRole", int(resp.GetCode()))
	defer handlePanic(ctx, "CreateRbacRole", resp)

	createReq := message.CreateRoleReq{}
	if handleRequest(ctx, req, resp, &createReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionCreate)}}) {
		result, err := c.rbacManager.CreateRole(framework.NewBackgroundMicroCtx(ctx, false), createReq, false)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) DeleteRbacRole(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteRbacRole", int(resp.GetCode()))
	defer handlePanic(ctx, "DeleteRbacRole", resp)

	deleteReq := message.DeleteRoleReq{}
	if handleRequest(ctx, req, resp, &deleteReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionDelete)}}) {
		result, err := c.rbacManager.DeleteRole(framework.NewBackgroundMicroCtx(ctx, false), deleteReq, false)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) UnbindRoleForUser(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UnbindRoleForUser", int(resp.GetCode()))
	defer handlePanic(ctx, "UnbindRoleForUser", resp)

	unBindReq := message.UnbindRoleForUserReq{}
	if handleRequest(ctx, req, resp, &unBindReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionUpdate)}}) {
		result, err := c.rbacManager.UnbindRoleForUser(framework.NewBackgroundMicroCtx(ctx, false), unBindReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) AddPermissionsForRole(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "AddPermissionsForRole", int(resp.GetCode()))
	defer handlePanic(ctx, "AddPermissionsForRole", resp)

	addReq := message.AddPermissionsForRoleReq{}
	if handleRequest(ctx, req, resp, &addReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionCreate)}}) {
		result, err := c.rbacManager.AddPermissionsForRole(framework.NewBackgroundMicroCtx(ctx, false), addReq, false)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) DeletePermissionsForRole(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeletePermissionsForRole", int(resp.GetCode()))
	defer handlePanic(ctx, "DeletePermissionsForRole", resp)

	deleteReq := message.DeletePermissionsForRoleReq{}
	if handleRequest(ctx, req, resp, &deleteReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionDelete)}}) {
		result, err := c.rbacManager.DeletePermissionsForRole(framework.NewBackgroundMicroCtx(ctx, false), deleteReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) QueryPermissionsForUser(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryPermissionsForUser", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryPermissionsForUser", resp)

	getReq := message.QueryPermissionsForUserReq{}
	if handleRequest(ctx, req, resp, &getReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		result, err := c.rbacManager.QueryPermissionsForUser(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) CheckPermissionForUser(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CheckPermissionsForUser", int(resp.GetCode()))
	defer handlePanic(ctx, "CheckPermissionsForUser", resp)

	checkReq := message.CheckPermissionForUserReq{}
	if handleRequest(ctx, req, resp, &checkReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		result, err := c.rbacManager.CheckPermissionForUser(framework.NewBackgroundMicroCtx(ctx, false), checkReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) QueryRoles(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryRoles", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryRoles", resp)

	getReq := message.QueryRolesReq{}
	if handleRequest(ctx, req, resp, &getReq, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		result, err := c.rbacManager.QueryRoles(framework.NewBackgroundMicroCtx(ctx, false), getReq)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ImportHosts(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "ImportHosts"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.ImportHostsReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
		flowIds, hostIds, err := handler.resourceManager.ImportHosts(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.Hosts, &reqStruct.Condition)
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
	metricsFuncName := "DeleteHosts"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.DeleteHostsReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionDelete)}}) {
		flowIds, err := handler.resourceManager.DeleteHosts(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Force)
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
	metricsFuncName := "QueryHosts"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.QueryHostsReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionRead)}}) {
		filter := reqStruct.GetHostFilter()
		page := reqStruct.GetPage()
		location := reqStruct.GetLocation()

		hosts, total, err := handler.resourceManager.QueryHosts(framework.NewBackgroundMicroCtx(ctx, false), location, filter, page)
		var rsp message.QueryHostsResp
		if err == nil {
			rsp.Hosts = hosts
		}
		handleResponse(ctx, resp, err, rsp, &clusterservices.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(total),
		})
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostReserved(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "UpdateHostReserved"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.UpdateHostReservedReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
		err := handler.resourceManager.UpdateHostReserved(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Reserved)
		var rsp message.UpdateHostReservedResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostStatus(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "UpdateHostStatus"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.UpdateHostStatusReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
		err := handler.resourceManager.UpdateHostStatus(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostIDs, reqStruct.Status)
		var rsp message.UpdateHostStatusResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostInfo(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "UpdateHostInfo"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.UpdateHostInfoReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
		err := handler.resourceManager.UpdateHostInfo(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.NewHostInfo)
		var rsp message.UpdateHostInfoResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CreateDisks(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "CreateDisks"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.CreateDisksReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
		diskIds, err := handler.resourceManager.CreateDisks(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.HostID, reqStruct.Disks)
		var rsp message.CreateDisksResp
		if err == nil {
			rsp.DiskIDs = diskIds
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteDisks(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "DeleteDisks"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.DeleteDisksReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
		err := handler.resourceManager.DeleteDisks(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.DiskIDs)
		var rsp message.DeleteDisksResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateDisk(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "UpdateDisk"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.UpdateDiskReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
		err := handler.resourceManager.UpdateDisk(framework.NewBackgroundMicroCtx(ctx, false), reqStruct.NewDiskInfo)
		var rsp message.UpdateDiskResp
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetHierarchy(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	metricsFuncName := "GetHierarchy"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.GetHierarchyReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionUpdate)}}) {
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
	metricsFuncName := "GetStocks"
	defer metrics.HandleClusterMetrics(start, metricsFuncName, int(resp.GetCode()))
	defer handlePanic(ctx, metricsFuncName, resp)

	reqStruct := message.GetStocksReq{}

	if handleRequest(ctx, req, resp, &reqStruct, []structs.RbacPermission{{Resource: string(constants.RbacResourceResource), Action: string(constants.RbacActionRead)}}) {
		location := reqStruct.GetLocation()
		hostFilter := reqStruct.GetHostFilter()
		diskFilter := reqStruct.GetDiskFilter()

		stocks, err := handler.resourceManager.GetStocks(framework.NewBackgroundMicroCtx(ctx, false), location, hostFilter, diskFilter)
		var rsp message.GetStocksResp
		if err == nil {
			rsp.Stocks = stocks
		}
		handleResponse(ctx, resp, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateVendors(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateVendors", int(response.GetCode()))

	req := message.UpdateVendorInfoReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionUpdate)}}) {
		resp, err := handler.productManager.UpdateVendors(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryVendors(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryVendors", int(response.GetCode()))

	req := message.QueryVendorInfoReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.productManager.QueryVendors(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryAvailableVendors(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryAvailableVendors", int(response.GetCode()))

	req := message.QueryAvailableVendorsReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.productManager.QueryAvailableVendors(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateProducts(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateProducts", int(response.GetCode()))

	req := message.UpdateProductsInfoReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionUpdate)}}) {
		resp, err := handler.productManager.UpdateProducts(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryProducts(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryProducts", int(response.GetCode()))

	req := message.QueryProductsInfoReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.productManager.QueryProducts(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryAvailableProducts(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryAvailableProducts", int(response.GetCode()))

	req := message.QueryAvailableProductsReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.productManager.QueryAvailableProducts(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryProductDetail(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryProductDetail", int(response.GetCode()))

	req := message.QueryProductDetailReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceProduct), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.productManager.QueryProductDetail(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CreateUser(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateUser", int(response.GetCode()))

	req := message.CreateUserReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionCreate)}}) {
		resp, err := handler.accountManager.CreateUser(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteUser(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteUser", int(response.GetCode()))

	req := message.DeleteUserReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionDelete)}}) {
		resp, err := handler.accountManager.DeleteUser(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) GetUser(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetUser", int(response.GetCode()))

	req := message.GetUserReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.accountManager.GetUser(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryUsers(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryUsers", int(response.GetCode()))

	req := message.QueryUserReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.accountManager.QueryUsers(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateUserProfile(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateUserProfile", int(response.GetCode()))

	req := message.UpdateUserProfileReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionUpdate)}}) {
		resp, err := handler.accountManager.UpdateUserProfile(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil

}

func (handler *ClusterServiceHandler) UpdateUserPassword(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateUserPassword", int(response.GetCode()))

	req := message.UpdateUserPasswordReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionUpdate)}}) {
		resp, err := handler.accountManager.UpdateUserPassword(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil

}

func (handler *ClusterServiceHandler) CreateTenant(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CreateTenant", int(response.GetCode()))

	req := message.CreateTenantReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionCreate)}}) {
		resp, err := handler.accountManager.CreateTenant(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) DeleteTenant(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "DeleteTenant", int(response.GetCode()))

	req := message.DeleteTenantReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionDelete)}}) {
		resp, err := handler.accountManager.DeleteTenant(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) GetTenant(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetTenant", int(response.GetCode()))

	req := message.GetTenantReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.accountManager.GetTenant(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryTenants(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryTenants", int(response.GetCode()))

	req := message.QueryTenantReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.accountManager.QueryTenants(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateTenantOnBoardingStatus(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateTenantOnBoardingStatus", int(response.GetCode()))

	req := message.UpdateTenantOnBoardingStatusReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionUpdate)}}) {
		resp, err := handler.accountManager.UpdateTenantOnBoardingStatus(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateTenantProfile(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpdateTenantProfile", int(response.GetCode()))

	req := message.UpdateTenantProfileReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceUser), Action: string(constants.RbacActionUpdate)}}) {
		resp, err := handler.accountManager.UpdateTenantProfile(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) CheckPlatform(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CheckPlatform", int(response.GetCode()))

	req := message.CheckPlatformReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionCreate)}}) {
		resp, err := handler.checkManager.Check(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) CheckCluster(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "CheckCluster", int(response.GetCode()))

	req := message.CheckClusterReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionCreate)}}) {
		resp, err := handler.checkManager.CheckCluster(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryProductUpgradePath(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryProductUpgradePath", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryProductUpgradePath", resp)

	request := &cluster.QueryUpgradePathReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.clusterManager.QueryProductUpdatePath(ctx, request.ClusterID)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryCheckReports(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryCheckReports", int(response.GetCode()))

	req := message.QueryCheckReportsReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.checkManager.QueryCheckReports(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryUpgradeVersionDiffInfo(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "QueryUpgradeVersionDiffInfo", int(resp.GetCode()))
	defer handlePanic(ctx, "QueryUpgradeVersionDiffInfo", resp)

	request := &cluster.QueryUpgradeVersionDiffInfoReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}) {
		result, err := handler.clusterManager.QueryUpgradeVersionDiffInfo(ctx, request.ClusterID, request.TargetVersion)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetCheckReport(ctx context.Context, request *clusterservices.RpcRequest, response *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "GetCheckReport", int(response.GetCode()))

	req := message.GetCheckReportReq{}
	if handleRequest(ctx, request, response, &req, []structs.RbacPermission{{Resource: string(constants.RbacResourceSystem), Action: string(constants.RbacActionRead)}}) {
		resp, err := handler.checkManager.GetCheckReport(ctx, req)
		handleResponse(ctx, response, err, resp, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpgradeCluster(ctx context.Context, req *clusterservices.RpcRequest, resp *clusterservices.RpcResponse) error {
	start := time.Now()
	defer metrics.HandleClusterMetrics(start, "UpgradeCluster", int(resp.GetCode()))
	defer handlePanic(ctx, "UpgradeCluster", resp)

	request := &cluster.UpgradeClusterReq{}

	if handleRequest(ctx, req, resp, request, []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionUpdate)}}) {
		result, err := handler.clusterManager.InPlaceUpgradeCluster(framework.NewBackgroundMicroCtx(ctx, false), *request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}
