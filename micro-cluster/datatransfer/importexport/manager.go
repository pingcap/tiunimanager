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
	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/models"
	dbModel "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var manager ImportExportService
var once sync.Once

type ImportExportManager struct {
	defaultExportPath string
	defaultImportPath string
}

func GetImportExportService() ImportExportService {
	once.Do(func() {
		manager = NewImportExportManager()
	})
	return manager
}

func NewImportExportManager() *ImportExportManager {
	mgr := ImportExportManager{
		defaultExportPath: constants.DefaultExportPath, //todo: get from config
		defaultImportPath: constants.DefaultImportPath, //todo: get from config
	}
	flowManager := workflow.GetWorkFlowService()
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

	if err := mgr.exportDataPreCheck(ctx, request); err != nil {
		framework.LogWithContext(ctx).Errorf("export data precheck failed, %s", err.Error())
		return nil, err
	}

	meta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return nil, fmt.Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
	}

	exportTime := time.Now()
	exportPrefix, _ := filepath.Abs(mgr.defaultExportPath)
	exportDir := filepath.Join(exportPrefix, request.ClusterID, fmt.Sprintf("%s_%s", exportTime.Format("2006-01-02_15:04:05"), request.StorageType))

	record := &importexport.DataTransportRecord{
		Entity: dbModel.Entity{
			TenantId: meta.GetCluster().TenantId,
			Status:   string(constants.DataImportExportProcessing),
		},
		ClusterID:       request.ClusterID,
		TransportType:   string(constants.TransportTypeExport),
		FilePath:        mgr.getDataExportFilePath(request, exportDir, true),
		ZipName:         request.ZipName,
		StorageType:     request.StorageType,
		Comment:         request.Comment,
		ReImportSupport: mgr.checkExportParamSupportReimport(request),
		StartTime:       time.Now(),
		EndTime:         time.Now(),
	}
	rw := models.GetImportExportReaderWriter()
	recordCreate, err := rw.CreateDataTransportRecord(ctx, record)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create data transport record failed, %s", err.Error())
		return nil, fmt.Errorf("create data transport record failed, %s", err.Error())
	}

	info := &ExportInfo{
		ClusterId:    request.ClusterID,
		UserName:     request.UserName,
		Password:     request.Password,
		FileType:     request.FileType,
		RecordId:     recordCreate.ID,
		FilePath:     mgr.getDataExportFilePath(request, exportDir, false),
		Filter:       request.Filter,
		Sql:          request.Sql,
		StorageType:  request.StorageType,
		BucketRegion: request.BucketRegion,
	}

	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, constants.WorkFlowExportData)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.WorkFlowExportData, err.Error())
		return nil, fmt.Errorf("create %s workflow failed, %s", constants.WorkFlowExportData, err.Error())
	}
	// Start the workflow
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	flowManager.AddContext(flow, contextDataTransportRecordKey, info)
	flowManager.AsyncStart(ctx, flow)

	return &message.DataExportResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
		RecordID: recordCreate.ID,
	}, nil
}

func (mgr *ImportExportManager) ImportData(ctx context.Context, request *message.DataImportReq) (*message.DataImportResp, error) {
	framework.LogWithContext(ctx).Infof("begin importdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end importdata")

	if err := mgr.importDataPreCheck(ctx, request); err != nil {
		framework.LogWithContext(ctx).Errorf("export data precheck failed, %s", err.Error())
		return nil, err
	}

	meta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
		return nil, fmt.Errorf("load cluster meta %s failed, %s", request.ClusterID, err.Error())
	}

	rw := models.GetImportExportReaderWriter()
	var info *ImportInfo
	importTime := time.Now()
	importPrefix, _ := filepath.Abs(mgr.defaultImportPath)
	importDir := filepath.Join(importPrefix, request.ClusterID, fmt.Sprintf("%s_%s", importTime.Format("2006-01-02_15:04:05"), request.StorageType))
	if request.RecordId == "" {
		if common.NfsStorageType == request.StorageType {
			err = os.Rename(filepath.Join(importPrefix, request.ClusterID, "temp"), importDir)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("find import dir failed, %s", err.Error())
				return nil, err
			}
		} else {
			err = os.MkdirAll(importDir, os.ModeDir)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("mkdir import dir failed, %s", err.Error())
				return nil, err
			}
		}

		record := &importexport.DataTransportRecord{
			Entity: dbModel.Entity{
				TenantId: meta.GetCluster().TenantId,
				Status:   string(constants.DataImportExportProcessing),
			},
			ClusterID:       request.ClusterID,
			TransportType:   string(constants.TransportTypeImport),
			FilePath:        mgr.getDataImportFilePath(request, importDir, true),
			ZipName:         constants.DefaultZipName,
			StorageType:     request.StorageType,
			Comment:         request.Comment,
			ReImportSupport: mgr.checkImportParamSupportReimport(request),
			StartTime:       time.Now(),
			EndTime:         time.Now(),
		}
		recordCreate, err := rw.CreateDataTransportRecord(ctx, record)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create data transport record failed, %s", err.Error())
			return nil, fmt.Errorf("create data transport record failed, %s", err.Error())
		}

		info = &ImportInfo{
			ClusterId:   request.ClusterID,
			UserName:    request.UserName,
			Password:    request.Password,
			FilePath:    mgr.getDataImportFilePath(request, importDir, false),
			RecordId:    recordCreate.ID,
			StorageType: request.StorageType,
			ConfigPath:  importDir,
		}
	} else {
		// import from transport record
		recordGet, err := rw.GetDataTransportRecord(ctx, request.RecordId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get data transport record %s failed, %s", request.RecordId, err.Error())
			return nil, fmt.Errorf("get data transport record %s failed, %s", request.RecordId, err.Error())
		}

		if err := os.MkdirAll(importDir, os.ModeDir); err != nil {
			return nil, fmt.Errorf("make import dir %s failed, %s", importDir, err.Error())
		}

		record := &importexport.DataTransportRecord{
			Entity: dbModel.Entity{
				TenantId: meta.GetCluster().TenantId,
				Status:   string(constants.DataImportExportProcessing),
			},
			ClusterID:       request.ClusterID,
			TransportType:   string(constants.TransportTypeImport),
			FilePath:        mgr.getDataImportFilePath(request, importDir, true),
			ZipName:         constants.DefaultZipName,
			StorageType:     request.StorageType,
			Comment:         request.Comment,
			ReImportSupport: false,
			StartTime:       time.Now(),
			EndTime:         time.Now(),
		}
		recordCreate, err := rw.CreateDataTransportRecord(ctx, record)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create data transport record failed, %s", err.Error())
			return nil, fmt.Errorf("create data transport record failed, %s", err.Error())
		}
		info = &ImportInfo{
			ClusterId:   request.ClusterID,
			UserName:    request.UserName,
			Password:    request.Password,
			FilePath:    recordGet.FilePath,
			RecordId:    recordCreate.ID,
			StorageType: request.StorageType,
			ConfigPath:  importDir,
		}
	}
	flowManager := workflow.GetWorkFlowService()
	flow, err := flowManager.CreateWorkFlow(ctx, request.ClusterID, constants.WorkFlowExportData)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create %s workflow failed, %s", constants.WorkFlowExportData, err.Error())
		return nil, fmt.Errorf("create %s workflow failed, %s", constants.WorkFlowExportData, err.Error())
	}
	// Start the workflow
	flowManager.AddContext(flow, contextClusterMetaKey, meta)
	flowManager.AddContext(flow, contextDataTransportRecordKey, info)
	if err = flowManager.AsyncStart(ctx, flow); err != nil {
		framework.LogWithContext(ctx).Errorf("async start %s workflow failed, %s", constants.WorkFlowBackupCluster, err.Error())
		return nil, fmt.Errorf("async start %s workflow failed, %s", constants.WorkFlowBackupCluster, err.Error())
	}

	return &message.DataImportResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
		RecordID: info.RecordId,
	}, nil
}

func (mgr *ImportExportManager) QueryDataTransportRecords(ctx context.Context, request *message.QueryDataImportExportRecordsReq) (*message.QueryDataImportExportRecordsResp, *structs.Page, error) {
	framework.LogWithContext(ctx).Infof("begin QueryDataTransportRecords request: %+v", request)
	defer framework.LogWithContext(ctx).Info("end QueryDataTransportRecords")

	rw := models.GetImportExportReaderWriter()
	records, total, err := rw.QueryDataTransportRecords(ctx, request.RecordID, request.ClusterID, request.ReImport, time.Unix(request.StartTime, 0), time.Unix(request.EndTime, 0), request.Page, request.PageSize)
	if err != nil {
		return nil, nil, err
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

	response := &message.QueryDataImportExportRecordsResp{
		Records: respRecords,
	}
	page := &structs.Page{
		Page:     request.Page,
		PageSize: request.PageSize,
		Total:    int(total),
	}

	return response, page, nil
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

func (mgr *ImportExportManager) exportDataPreCheck(ctx context.Context, request *message.DataExportReq) error {
	if request.ClusterID == "" {
		return fmt.Errorf("invalid param clusterId %s", request.ClusterID)
	}
	if request.UserName == "" {
		return fmt.Errorf("invalid param userName %s", request.UserName)
	}
	/*
		if request.Password == "" {
			return fmt.Errorf("invalid param password %s", request.Password)
		}
	*/

	if FileTypeCSV != request.FileType && FileTypeSQL != request.FileType {
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
		absPath, err := filepath.Abs(mgr.defaultExportPath)
		if err != nil {
			return fmt.Errorf("export dir %s is not vaild", mgr.defaultExportPath)
		}
		if !mgr.checkFilePathExists(absPath) {
			//return fmt.Errorf("export path %s not exist", absPath)
			_ = os.MkdirAll(absPath, os.ModeDir)
		}
	default:
		return fmt.Errorf("invalid param storageType %s", request.StorageType)
	}

	return nil
}

func (mgr *ImportExportManager) importDataPreCheck(ctx context.Context, request *message.DataImportReq) error {
	if request.ClusterID == "" {
		return fmt.Errorf("invalid param clusterId %s", request.ClusterID)
	}
	if request.UserName == "" {
		return fmt.Errorf("invalid param userName %s", request.UserName)
	}
	/*
		if request.Password == "" {
			return fmt.Errorf("invalid param password %s", request.Password)
		}
	*/
	absPath, err := filepath.Abs(mgr.defaultImportPath)
	if err != nil {
		return fmt.Errorf("import dir %s is not vaild", mgr.defaultImportPath)
	}
	if !mgr.checkFilePathExists(absPath) {
		//return fmt.Errorf("import path %s not exist", absPath)
		_ = os.MkdirAll(absPath, os.ModeDir)
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
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (mgr *ImportExportManager) checkExportParamSupportReimport(request *message.DataExportReq) bool {
	if common.S3StorageType == request.StorageType {
		return false
	}
	if request.Filter == "" && request.Sql != "" && FileTypeCSV == request.FileType {
		return false
	}
	return true
}

func (mgr *ImportExportManager) checkImportParamSupportReimport(request *message.DataImportReq) bool {
	if common.NfsStorageType == request.StorageType {
		return true
	}
	return false
}

func (mgr *ImportExportManager) getDataExportFilePath(request *message.DataExportReq, exportDir string, persist bool) string {
	var filePath string
	if common.S3StorageType == request.StorageType {
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
	if common.S3StorageType == request.StorageType {
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

func cleanDataTransportDir(ctx context.Context, filepath string) error {
	framework.LogWithContext(ctx).Infof("clean and re-mkdir data dir: %s", filepath)
	if err := os.RemoveAll(filepath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath, os.ModeDir); err != nil {
		return err
	}
	return nil
}

func buildDataImportConfig(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin buildDataImportConfig")
	defer framework.LogWithContext(ctx).Info("end buildDataImportConfig")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*ImportInfo)

	config := NewDataImportConfig(meta, info)
	if config == nil {
		framework.LogWithContext(ctx).Errorf("convert toml config failed, cluster: %s", meta.GetCluster().ID)
		node.Fail(fmt.Errorf("convert toml config failed, cluster: %s", meta.GetCluster().ID))
		return false
	}

	filePath := fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create import toml config failed, %s", err.Error())
		node.Fail(fmt.Errorf("create import toml config failed, %s", err.Error()))
		return false
	}

	if err = toml.NewEncoder(file).Encode(config); err != nil {
		framework.LogWithContext(ctx).Errorf("encode data import toml config failed, %s", err.Error())
		node.Fail(fmt.Errorf("encode data import toml config failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Infof("build lightning toml config sucess, %v", config)
	node.Success()
	return true
}

func importDataToCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin importDataToCluster")
	defer framework.LogWithContext(ctx).Info("end importDataToCluster")

	info := ctx.GetData(contextDataTransportRecordKey).(*ImportInfo)

	//tiup tidb-lightning -config tidb-lightning.toml
	importTaskId, err := secondparty.Manager.Lightning(ctx, 0,
		[]string{"-config", fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)},
		node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup lightning api failed, %s", err.Error())
		node.Fail(fmt.Errorf("call tiup lightning api failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Infof("call tiupmgr tidb-lightning api success, importTaskId %d", importTaskId)
	node.Success()
	return true
}

func updateDataImportRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin updateDataImportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataImportRecord")

	info := ctx.GetData(contextDataTransportRecordKey).(*ImportInfo)

	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, info.RecordId, string(constants.DataImportExportFinished), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		node.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Info("update data transport record success")
	node.Success()
	return true
}

func exportDataFromCluster(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin exportDataFromCluster")
	defer framework.LogWithContext(ctx).Info("end exportDataFromCluster")

	//meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*ExportInfo)

	//todo: get from meta
	tidbHost := ""
	tidbPort := 4000
	if tidbPort == 0 {
		tidbPort = constants.DefaultTiDBPort
	}

	if common.NfsStorageType == info.StorageType {
		if err := cleanDataTransportDir(ctx, info.FilePath); err != nil {
			framework.LogWithContext(ctx).Errorf("clean export directory failed, %s", err.Error())
			node.Fail(fmt.Errorf("clean export directory failed, %s", err.Error()))
			return false
		}
	}

	//tiup dumpling -u root -P 4000 --host 127.0.0.1 --filetype sql -t 8 -o /tmp/test -r 200000 -F 256MiB --filter "user*"
	//todo: replace root password
	cmd := []string{"-u", info.UserName,
		"-p", info.Password,
		"-P", strconv.Itoa(tidbPort),
		"--host", tidbHost,
		"--filetype", info.FileType,
		"-t", "8",
		"-o", info.FilePath,
		"-r", "200000",
		"-F", "256MiB"}
	if info.Filter != "" {
		cmd = append(cmd, "--filter", info.Filter)
	}
	if FileTypeCSV == info.FileType && info.Filter == "" && info.Sql != "" {
		cmd = append(cmd, "--sql", info.Sql)
	}
	if common.S3StorageType == info.StorageType && info.BucketRegion != "" {
		cmd = append(cmd, "--s3.region", fmt.Sprintf("\"%s\"", info.BucketRegion))
	}
	framework.LogWithContext(ctx).Infof("call tiupmgr dumpling api, cmd: %v", cmd)
	exportTaskId, err := secondparty.Manager.Dumpling(ctx, 0, cmd, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup dumpling api failed, %s", err.Error())
		node.Fail(fmt.Errorf("call tiup dumpling api failed, %s", err.Error()))
		return false
	}

	framework.LogWithContext(ctx).Infof("call tiupmgr succee, exportTaskId: %d", exportTaskId)
	node.Success()
	return true
}

func updateDataExportRecord(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin updateDataExportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataExportRecord")

	info := ctx.GetData(contextDataTransportRecordKey).(*ExportInfo)

	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, info.RecordId, string(constants.DataImportExportFinished), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		node.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}

	framework.LogWithContext(ctx).Info("update data transport record success")
	node.Success()
	return true
}

func importDataFailed(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	framework.LogWithContext(ctx).Info("begin importDataFailed")
	defer framework.LogWithContext(ctx).Info("end importDataFailed")

	meta := ctx.GetData(contextClusterMetaKey).(*handler.ClusterMeta)
	info := ctx.GetData(contextDataTransportRecordKey).(*ExportInfo)
	if err := updateTransportRecordFailed(ctx, info.RecordId, meta.GetCluster().ID); err != nil {
		node.Fail(err)
		return false
	}

	return clusterFail(node, ctx)
}

func exportDataFailed(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	return clusterFail(node, ctx)
}

func clusterEnd(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	node.Success()
	return true
}

func clusterFail(node *wfModel.WorkFlowNode, ctx *workflow.FlowContext) bool {
	node.Success()
	return true
}

func updateTransportRecordFailed(ctx context.Context, recordId string, clusterId string) error {
	rw := models.GetImportExportReaderWriter()
	err := rw.UpdateDataTransportRecord(ctx, recordId, string(constants.DataImportExportFailed), time.Now())
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return err
	}

	framework.LogWithContext(ctx).Info("update data transport record success")
	return nil
}
