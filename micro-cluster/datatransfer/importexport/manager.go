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

package importexport

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"path/filepath"
	"time"
)

var manager ImportExportService

type ImportExportManager struct {
}

func GetImportExportService() ImportExportService {
	if manager == nil {
		manager = NewImportExportManager()
	}
	return manager
}

func NewImportExportManager() *ImportExportManager {
	mgr := ImportExportManager{}
	flowManager := workflow.GetWorkFlowManager()
	flowManager.RegisterWorkFlow(context.TODO(), constants.WorkFlowExportData, &workflow.WorkFlowDefine{
		FlowName: constants.WorkFlowExportData,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"exportDataFromCluster", "exportDataDone", "fail", workflow.PollingNode, exportDataFromCluster},
			"exportDataDone":   {"updateDataExportRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateDataExportRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, clusterEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, exportDataFailed},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), constants.WorkFlowImportData, &workflow.WorkFlowDefine{
		FlowName: constants.WorkFlowImportData,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"buildDataImportConfig", "buildConfigDone", "fail", workflow.SyncFuncNode, buildDataImportConfig},
			"buildConfigDone":  {"importDataToCluster", "importDataDone", "fail", workflow.PollingNode, importDataToCluster},
			"importDataDone":   {"updateDataImportRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateDataImportRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, clusterEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, importDataFailed},
		},
	})

	return &mgr
}

func (mgr *ImportExportManager) ExportData(ctx context.Context, request *message.DataExportReq) (*message.DataExportResp, error) {
	framework.LogWithContext(ctx).Infof("begin exportdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end exportdata")

	//todo
	return nil, nil
}

func (mgr *ImportExportManager) ImportData(ctx context.Context, request *message.DataImportReq) (*message.DataImportResp, error) {
	framework.LogWithContext(ctx).Infof("begin importdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end importdata")

	//todo
	return nil, nil
}

func (mgr *ImportExportManager) QueryDataTransportRecords(ctx context.Context, request *message.QueryDataImportExportRecordsReq) (*message.QueryDataImportExportRecordsResp, error) {
	framework.LogWithContext(ctx).Infof("begin QueryDataTransportRecords request: %+v", request)
	defer framework.LogWithContext(ctx).Info("end QueryDataTransportRecords")

	rw := models.GetImportExportReaderWriter()
	records, _, err := rw.QueryDataTransportRecords(ctx, request.RecordID, request.ClusterID, request.ReImport, time.Unix(request.StartTime, 0), time.Unix(request.EndTime, 0), request.Page, request.PageSize)
	if err != nil {
		return nil, err
	}

	respRecords := make([]*structs.DataImportExportRecordInfo, len(records))
	for index, record := range records {
		respRecords[index] = &structs.DataImportExportRecordInfo{
			RecordID:      record.ID,
			ClusterID:     record.ClusterID,
			TransportType: record.TransportType,
			FilePath:      record.FilePath,
			ZipName:       record.ZipName,
			StorageType:   record.StorageType,
			Comment:       record.Comment,
			Status:        record.Status,
			StartTime:     record.StartTime,
			EndTime:       record.EndTime,
			CreateTime:    record.CreatedAt,
			UpdateTime:    record.UpdatedAt,
			DeleteTime:    record.DeletedAt.Time,
		}
	}

	return &message.QueryDataImportExportRecordsResp{
		Records: respRecords,
	}, nil
}

func (mgr *ImportExportManager) DeleteDataTransportRecord(ctx context.Context, request *message.DeleteImportExportRecordReq) (*message.DeleteImportExportRecordResp, error) {
	framework.LogWithContext(ctx).Infof("begin DeleteDataTransportRecord request: %+v", request)
	defer framework.LogWithContext(ctx).Info("end DeleteDataTransportRecord")

	rw := models.GetImportExportReaderWriter()
	record, err := rw.GetDataTransportRecord(ctx, request.RecordID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get data transport record %s failed %s", request.RecordID, err)
		return nil, err
	}

	if common.S3StorageType != record.StorageType {
		filePath := filepath.Dir(record.FilePath)
		_ = os.RemoveAll(filePath)
		framework.LogWithContext(ctx).Infof("remove file path %s, record: %+v", filePath, record.ID)
	}

	err = rw.DeleteDataTransportRecord(ctx, request.RecordID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("delete data transport record %s failed %s", request.RecordID, err)
		return nil, err
	}
	framework.LogWithContext(ctx).Infof("delete transport record %+v success", record.ID)

	return nil, nil
}

func buildDataImportConfig(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func importDataToCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func updateDataImportRecord(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func exportDataFromCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func updateDataExportRecord(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func importDataFailed(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	return clusterFail(node, flowContext)
}

func exportDataFailed(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	return clusterFail(node, flowContext)
}

func clusterEnd(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}

func clusterFail(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	return true
}
