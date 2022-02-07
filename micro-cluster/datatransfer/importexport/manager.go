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
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/models"
	dbModel "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var dataTransportService ImportExportService
var once sync.Once

type ImportExportManager struct{}

func GetImportExportService() ImportExportService {
	once.Do(func() {
		if dataTransportService == nil {
			dataTransportService = NewImportExportManager()
		}
	})
	return dataTransportService
}

func MockImportExportService(service ImportExportService) {
	dataTransportService = service
}

func NewImportExportManager() *ImportExportManager {
	mgr := ImportExportManager{}
	flowManager := workflow.GetWorkFlowService()
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowExportData, &workflow.WorkFlowDefine{
		FlowName: constants.FlowExportData,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"exportDataFromCluster", "exportDataDone", "fail", workflow.PollingNode, exportDataFromCluster},
			"exportDataDone":   {"updateDataExportRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateDataExportRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"end", "", "", workflow.SyncFuncNode, exportDataFailed},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowImportData, &workflow.WorkFlowDefine{
		FlowName: constants.FlowImportData,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"buildDataImportConfig", "buildConfigDone", "fail", workflow.SyncFuncNode, buildDataImportConfig},
			"buildConfigDone":  {"importDataToCluster", "importDataDone", "fail", workflow.PollingNode, importDataToCluster},
			"importDataDone":   {"updateDataImportRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateDataImportRecord},
			"updateRecordDone": {"cleanImportTempFile", "cleanTempDone", "fail", workflow.SyncFuncNode, cleanImportTempFile},
			"cleanTempDone":    {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"end", "", "", workflow.SyncFuncNode, importDataFailed},
		},
	})

	return &mgr
}

func (mgr *ImportExportManager) ExportData(ctx context.Context, request message.DataExportReq) (resp message.DataExportResp, exportErr error) {
	framework.LogWithContext(ctx).Infof("begin exportdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end exportdata")

	if err := mgr.exportDataPreCheck(ctx, &request); err != nil {
		framework.LogWithContext(ctx).Errorf("export data precheck failed, %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("export data precheck failed, %s", err.Error()), err)
	}

	configRW := models.GetConfigReaderWriter()
	exportPathConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyExportShareStoragePath)
	if err != nil || exportPathConfig.ConfigValue == "" {
		framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyExportShareStoragePath, err.Error())
		return resp, errors.WrapError(errors.TIEM_TRANSPORT_SYSTEM_CONFIG_INVALID, fmt.Sprintf("get conifg %s failed: %s", constants.ConfigKeyExportShareStoragePath, err.Error()), err)
	}

	meta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_NOT_FOUND, fmt.Sprintf("load cluster meta %s failed, %s", request.ClusterID, err.Error()), err)
	}

	exportTime := time.Now()
	exportPrefix, _ := filepath.Abs(exportPathConfig.ConfigValue)
	exportDir := filepath.Join(exportPrefix, request.ClusterID, fmt.Sprintf("%s_%s", exportTime.Format("2006-01-02_15:04:05"), request.StorageType))

	record := &importexport.DataTransportRecord{
		Entity: dbModel.Entity{
			TenantId: meta.Cluster.TenantId,
			Status:   string(constants.DataImportExportProcessing),
		},
		ClusterID:       request.ClusterID,
		TransportType:   string(constants.TransportTypeExport),
		FilePath:        mgr.getDataExportFilePath(&request, exportDir, true),
		ZipName:         request.ZipName,
		StorageType:     request.StorageType,
		Comment:         request.Comment,
		ReImportSupport: mgr.checkExportParamSupportReimport(&request),
		StartTime:       time.Now(),
		EndTime:         time.Now(),
	}
	rw := models.GetImportExportReaderWriter()
	recordCreate, err := rw.CreateDataTransportRecord(ctx, record)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create data transport record failed, %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_CREATE_FAILED, fmt.Sprintf("create data transport record failed, %s", err.Error()), err)
	}
	defer func() {
		if exportErr != nil {
			if delErr := rw.DeleteDataTransportRecord(ctx, recordCreate.ID); delErr != nil {
				framework.LogWithContext(ctx).Warnf("delete transport record %+v failed, %s", recordCreate, err.Error())
			}
		}
	}()

	info := &exportInfo{
		ClusterId:   request.ClusterID,
		UserName:    request.UserName,
		Password:    request.Password,
		FileType:    request.FileType,
		RecordId:    recordCreate.ID,
		FilePath:    mgr.getDataExportFilePath(&request, exportDir, false),
		Filter:      request.Filter,
		Sql:         request.Sql,
		StorageType: request.StorageType,
	}

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, workflow.BizTypeCluster, constants.FlowExportData)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.FlowExportData, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, fmt.Sprintf("create %s workflow failed, %s", constants.FlowExportData, err.Error()), err)
	}
	// Start the workflow
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	flowManager.AddContext(flow, contextDataTransportRecordKey, info)
	if err := flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("start %s workflow failed, %s", constants.FlowExportData, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_START_FAILED, fmt.Sprintf("start %s workflow failed, %s", constants.FlowExportData, err.Error()), err)
	}

	resp.WorkFlowID = flow.Flow.ID
	resp.RecordID = recordCreate.ID
	return resp, nil
}

func (mgr *ImportExportManager) ImportData(ctx context.Context, request message.DataImportReq) (resp message.DataImportResp, importErr error) {
	framework.LogWithContext(ctx).Infof("begin importdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end importdata")

	if err := mgr.importDataPreCheck(ctx, &request); err != nil {
		framework.LogWithContext(ctx).Errorf("import data precheck failed, %s", err.Error())
		return resp, errors.WrapError(errors.TIEM_PARAMETER_INVALID, fmt.Sprintf("import data precheck failed, %s", err.Error()), err)
	}

	configRW := models.GetConfigReaderWriter()
	importPathConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyImportShareStoragePath)
	if err != nil || importPathConfig.ConfigValue == "" {
		framework.LogWithContext(ctx).Errorf("get conifg %s failed: %s", constants.ConfigKeyImportShareStoragePath, err.Error())
		return resp, errors.WrapError(errors.TIEM_TRANSPORT_SYSTEM_CONFIG_INVALID, fmt.Sprintf("get conifg %s failed: %s", constants.ConfigKeyImportShareStoragePath, err.Error()), err)
	}

	meta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return resp, errors.WrapError(errors.TIEM_CLUSTER_NOT_FOUND, fmt.Sprintf("load cluster meta %s failed, %s", request.ClusterID, err.Error()), err)
	}

	rw := models.GetImportExportReaderWriter()
	var info *importInfo
	var recordCreate *importexport.DataTransportRecord
	importTime := time.Now()
	importPrefix, _ := filepath.Abs(importPathConfig.ConfigValue)
	importDir := filepath.Join(importPrefix, request.ClusterID, fmt.Sprintf("%s_%s", importTime.Format("2006-01-02_15:04:05"), request.StorageType))
	if request.RecordId == "" {
		if string(constants.StorageTypeNFS) == request.StorageType {
			err = os.Rename(filepath.Join(importPrefix, request.ClusterID, "temp"), importDir)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("find import dir failed, %s", err.Error())
				return resp, errors.WrapError(errors.TIEM_TRANSPORT_PATH_CREATE_FAILED, fmt.Sprintf("find import dir failed, %s", err.Error()), err)
			}
		} else {
			err = os.MkdirAll(importDir, os.ModePerm)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("mkdir import dir failed, %s", err.Error())
				return resp, errors.WrapError(errors.TIEM_TRANSPORT_PATH_CREATE_FAILED, fmt.Sprintf("mkdir import dir failed, %s", err.Error()), err)
			}
		}

		record := &importexport.DataTransportRecord{
			Entity: dbModel.Entity{
				TenantId: meta.Cluster.TenantId,
				Status:   string(constants.DataImportExportProcessing),
			},
			ClusterID:       request.ClusterID,
			TransportType:   string(constants.TransportTypeImport),
			FilePath:        mgr.getDataImportFilePath(&request, importDir, true),
			ZipName:         constants.DefaultZipName,
			StorageType:     request.StorageType,
			Comment:         request.Comment,
			ReImportSupport: mgr.checkImportParamSupportReimport(&request),
			StartTime:       time.Now(),
			EndTime:         time.Now(),
		}
		recordCreate, err = rw.CreateDataTransportRecord(ctx, record)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create data transport record failed, %s", err.Error())
			return resp, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_CREATE_FAILED, fmt.Sprintf("create data transport record failed, %s", err.Error()), err)
		}

		info = &importInfo{
			ClusterId:   request.ClusterID,
			UserName:    request.UserName,
			Password:    request.Password,
			FilePath:    mgr.getDataImportFilePath(&request, importDir, false),
			RecordId:    recordCreate.ID,
			StorageType: request.StorageType,
			ConfigPath:  importDir,
		}
	} else {
		// import from transport record
		recordGet, err := rw.GetDataTransportRecord(ctx, request.RecordId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get data transport record %s failed, %s", request.RecordId, err.Error())
			return resp, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_QUERY_FAILED, fmt.Sprintf("get data transport record %s failed, %s", request.RecordId, err.Error()), err)
		}

		if err := os.MkdirAll(importDir, os.ModePerm); err != nil {
			return resp, errors.WrapError(errors.TIEM_TRANSPORT_PATH_CREATE_FAILED, fmt.Sprintf("make import dir %s failed, %s", importDir, err.Error()), err)
		}

		record := &importexport.DataTransportRecord{
			Entity: dbModel.Entity{
				TenantId: meta.Cluster.TenantId,
				Status:   string(constants.DataImportExportProcessing),
			},
			ClusterID:       request.ClusterID,
			TransportType:   string(constants.TransportTypeImport),
			FilePath:        mgr.getDataImportFilePath(&request, importDir, true),
			ZipName:         constants.DefaultZipName,
			StorageType:     request.StorageType,
			Comment:         request.Comment,
			ReImportSupport: false,
			StartTime:       time.Now(),
			EndTime:         time.Now(),
		}
		recordCreate, err = rw.CreateDataTransportRecord(ctx, record)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create data transport record failed, %s", err.Error())
			return resp, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_CREATE_FAILED, fmt.Sprintf("create data transport record failed, %s", err.Error()), err)
		}
		info = &importInfo{
			ClusterId:   request.ClusterID,
			UserName:    request.UserName,
			Password:    request.Password,
			FilePath:    recordGet.FilePath,
			RecordId:    recordCreate.ID,
			StorageType: request.StorageType,
			ConfigPath:  importDir,
		}
	}
	defer func() {
		if importErr != nil {
			if delErr := rw.DeleteDataTransportRecord(ctx, recordCreate.ID); delErr != nil {
				framework.LogWithContext(ctx).Warnf("delete transport record %+v failed, %s", recordCreate, err.Error())
			}
		}
	}()

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, workflow.BizTypeCluster, constants.FlowImportData)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.FlowImportData, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, fmt.Sprintf("create %s workflow failed, %s", constants.FlowImportData, err.Error()), err)
	}
	// Start the workflow
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	flowManager.AddContext(flow, contextDataTransportRecordKey, info)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", constants.FlowImportData, err.Error())
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_START_FAILED, fmt.Sprintf("async start %s workflow failed, %s", constants.FlowImportData, err.Error()), err)
	}

	resp.WorkFlowID = flow.Flow.ID
	resp.RecordID = info.RecordId
	return resp, nil
}

func (mgr *ImportExportManager) QueryDataTransportRecords(ctx context.Context, request message.QueryDataImportExportRecordsReq) (resp message.QueryDataImportExportRecordsResp, page structs.Page, err error) {
	framework.LogWithContext(ctx).Infof("begin QueryDataTransportRecords request: %+v", request)
	defer framework.LogWithContext(ctx).Info("end QueryDataTransportRecords")

	rw := models.GetImportExportReaderWriter()
	records, total, err := rw.QueryDataTransportRecords(ctx, request.RecordID, request.ClusterID, request.ReImport, request.StartTime, request.EndTime, request.Page, request.PageSize)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query data transport records request %+v failed, %s", request, err.Error())
		return resp, page, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_QUERY_FAILED, fmt.Sprintf("query data transport records failed, %s", err.Error()), err)
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

	resp.Records = respRecords
	page = structs.Page{
		Page:     request.Page,
		PageSize: request.PageSize,
		Total:    int(total),
	}

	return resp, page, nil
}

func (mgr *ImportExportManager) DeleteDataTransportRecord(ctx context.Context, request message.DeleteImportExportRecordReq) (resp message.DeleteImportExportRecordResp, err error) {
	framework.LogWithContext(ctx).Infof("begin DeleteDataTransportRecord request: %+v", request)
	defer framework.LogWithContext(ctx).Info("end DeleteDataTransportRecord")

	rw := models.GetImportExportReaderWriter()
	record, err := rw.GetDataTransportRecord(ctx, request.RecordID)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get data transport record %s failed %s", request.RecordID, err)
		return resp, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_QUERY_FAILED, fmt.Sprintf("get data transport record %s failed, %s", request.RecordID, err.Error()), err)
	}

	if string(constants.StorageTypeS3) != record.StorageType {
		filePath := filepath.Dir(record.FilePath)
		go func() {
			removeErr := os.RemoveAll(filePath)
			framework.LogWithContext(ctx).Infof("remove file path %s, record: %+v, error: %v", filePath, record, removeErr)
		}()
	}

	err = rw.DeleteDataTransportRecord(ctx, request.RecordID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("delete data transport record %s failed %s", request.RecordID, err.Error())
		return resp, errors.WrapError(errors.TIEM_TRANSPORT_RECORD_DELETE_FAILED, fmt.Sprintf("delete data transport record %s failed %s", request.RecordID, err.Error()), err)
	}
	framework.LogWithContext(ctx).Infof("delete transport record %+v success", record.ID)

	resp.RecordID = request.RecordID
	return resp, nil
}

func (mgr *ImportExportManager) exportDataPreCheck(ctx context.Context, request *message.DataExportReq) error {
	configRW := models.GetConfigReaderWriter()
	exportPathConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyExportShareStoragePath)
	if err != nil || exportPathConfig.ConfigValue == "" {
		return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyExportShareStoragePath, err.Error())
	}

	if request.ClusterID == "" {
		return fmt.Errorf("invalid param clusterId %s", request.ClusterID)
	}
	if request.UserName == "" {
		return fmt.Errorf("invalid param userName %s", request.UserName)
	}
	if request.Password == "" {
		return fmt.Errorf("invalid param password %s", request.Password)
	}

	if fileTypeCSV != request.FileType && fileTypeSQL != request.FileType {
		return fmt.Errorf("invalid param fileType %s", request.FileType)
	}
	if request.ZipName == "" {
		request.ZipName = constants.DefaultZipName
	} else if !strings.HasSuffix(request.ZipName, ".zip") {
		request.ZipName = fmt.Sprintf("%s.zip", request.ZipName)
	}

	switch request.StorageType {
	case string(constants.StorageTypeS3):
		if request.EndpointUrl == "" {
			return fmt.Errorf("invalid param endpointUrl %s", request.EndpointUrl)
		}
		if request.BucketUrl == "" {
			return fmt.Errorf("invalid param bucketUrl %s", request.BucketUrl)
		}
		if request.AccessKey == "" {
			return fmt.Errorf("invalid param accessKey %s", request.AccessKey)
		}
		if request.SecretAccessKey == "" {
			return fmt.Errorf("invalid param secretAccessKey %s", request.SecretAccessKey)
		}
	case string(constants.StorageTypeNFS):
		absPath, err := filepath.Abs(exportPathConfig.ConfigValue)
		if err != nil {
			return fmt.Errorf("export dir %s is not vaild", exportPathConfig.ConfigValue)
		}
		if !mgr.checkFilePathExists(absPath) {
			if err = os.MkdirAll(absPath, os.ModePerm); err != nil {
				return fmt.Errorf("make export path %s failed, %s", absPath, err.Error())
			}
		}
	default:
		return fmt.Errorf("invalid param storageType %s", request.StorageType)
	}

	return nil
}

func (mgr *ImportExportManager) importDataPreCheck(ctx context.Context, request *message.DataImportReq) error {
	configRW := models.GetConfigReaderWriter()
	importPathConfig, err := configRW.GetConfig(ctx, constants.ConfigKeyImportShareStoragePath)
	if err != nil || importPathConfig.ConfigValue == "" {
		return fmt.Errorf("get conifg %s failed: %s", constants.ConfigKeyImportShareStoragePath, err.Error())
	}

	if request.ClusterID == "" {
		return fmt.Errorf("invalid param clusterId %s", request.ClusterID)
	}
	if request.UserName == "" {
		return fmt.Errorf("invalid param userName %s", request.UserName)
	}
	if request.Password == "" {
		return fmt.Errorf("invalid param password %s", request.Password)
	}
	absPath, err := filepath.Abs(importPathConfig.ConfigValue)
	if err != nil {
		return fmt.Errorf("import dir %s is not vaild", importPathConfig.ConfigValue)
	}
	if !mgr.checkFilePathExists(absPath) {
		if err = os.MkdirAll(absPath, os.ModePerm); err != nil {
			return fmt.Errorf("make import path %s failed, %s", absPath, err.Error())
		}
	}

	if request.RecordId == "" {
		switch request.StorageType {
		case string(constants.StorageTypeS3):
			if request.EndpointUrl == "" {
				return fmt.Errorf("invalid param endpointUrl %s", request.EndpointUrl)
			}
			if request.BucketUrl == "" {
				return fmt.Errorf("invalid param bucketUrl %s", request.BucketUrl)
			}
			if request.AccessKey == "" {
				return fmt.Errorf("invalid param accessKey %s", request.AccessKey)
			}
			if request.SecretAccessKey == "" {
				return fmt.Errorf("invalid param secretAccessKey %s", request.SecretAccessKey)
			}
		case string(constants.StorageTypeNFS):
			break
		default:
			return fmt.Errorf("invalid param storageType %s", request.StorageType)
		}
	} else {
		// import from transport record
		request.StorageType = string(constants.StorageTypeNFS)
		rw := models.GetImportExportReaderWriter()
		recordGet, err := rw.GetDataTransportRecord(ctx, request.RecordId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get data transport record %s failed, %s", request.RecordId, err.Error())
			return fmt.Errorf("get data transport record %s failed, %s", request.RecordId, err.Error())
		}

		if !mgr.checkFilePathExists(recordGet.FilePath) {
			return fmt.Errorf("data source path %s not exist", recordGet.FilePath)
		}
		if recordGet.StorageType != string(constants.StorageTypeNFS) {
			return fmt.Errorf("storage type %s can not support re-import", recordGet.StorageType)
		}
		if !recordGet.ReImportSupport {
			return fmt.Errorf("transport record %s not support re-import", recordGet.ID)
		}
	}

	return nil
}

func (mgr *ImportExportManager) checkFilePathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (mgr *ImportExportManager) checkExportParamSupportReimport(request *message.DataExportReq) bool {
	if string(constants.StorageTypeS3) == request.StorageType {
		return false
	}
	if request.Filter == "" && request.Sql != "" && fileTypeCSV == request.FileType {
		return false
	}
	return true
}

func (mgr *ImportExportManager) checkImportParamSupportReimport(request *message.DataImportReq) bool {
	return string(constants.StorageTypeNFS) == request.StorageType
}

func (mgr *ImportExportManager) getDataExportFilePath(request *message.DataExportReq, exportDir string, persist bool) string {
	var filePath string
	if string(constants.StorageTypeS3) == request.StorageType {
		if persist {
			filePath = fmt.Sprintf("%s?&endpoint=%s", request.BucketUrl, request.EndpointUrl)
		} else {
			filePath = fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.BucketUrl, request.AccessKey, request.SecretAccessKey, request.EndpointUrl)
		}
	} else {
		filePath = filepath.Join(exportDir, "data")
	}
	return filePath
}

func (mgr *ImportExportManager) getDataImportFilePath(request *message.DataImportReq, importDir string, persist bool) string {
	var filePath string
	if string(constants.StorageTypeS3) == request.StorageType {
		if persist {
			filePath = fmt.Sprintf("%s?&endpoint=%s", request.BucketUrl, request.EndpointUrl)
		} else {
			filePath = fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.BucketUrl, request.AccessKey, request.SecretAccessKey, request.EndpointUrl)
		}
	} else {
		filePath = filepath.Join(importDir, "data")
	}
	return filePath
}
