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
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-metadb/service"
	"github.com/pingcap-inc/tiem/models"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ImportExportManager struct {
}

var manager *ImportExportManager

func GetImportExportManager() *ImportExportManager {
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

type ImportInfo struct {
	ClusterId   string
	UserName    string
	Password    string
	FilePath    string
	RecordId    string
	StorageType string
	ConfigPath  string
}

type ExportInfo struct {
	ClusterId    string
	UserName     string
	Password     string
	FileType     string
	RecordId     string
	FilePath     string
	Filter       string
	Sql          string
	StorageType  string
	BucketRegion string
}

var contextDataTransportKey = "dataTransportInfo"

func ExportDataPreCheck(req *clusterpb.DataExportRequest) error {
	if req.GetClusterId() == "" {
		return fmt.Errorf("invalid param clusterId %s", req.GetClusterId())
	}
	if req.GetUserName() == "" {
		return fmt.Errorf("invalid param userName %s", req.GetUserName())
	}
	/*
		if req.GetPassword() == "" {
			return fmt.Errorf("invalid param password %s", req.GetPassword())
		}
	*/

	if FileTypeCSV != req.GetFileType() && FileTypeSQL != req.GetFileType() {
		return fmt.Errorf("invalid param fileType %s", req.GetFileType())
	}
	if req.GetZipName() == "" {
		req.ZipName = common.DefaultZipName
	} else if !strings.HasSuffix(req.GetZipName(), ".zip") {
		req.ZipName = fmt.Sprintf("%s.zip", req.GetZipName())
	}

	switch req.GetStorageType() {
	case common.S3StorageType:
		if req.GetEndpointUrl() == "" {
			return fmt.Errorf("invalid param endpointUrl %s", req.GetEndpointUrl())
		}
		if req.GetBucketUrl() == "" {
			return fmt.Errorf("invalid param bucketUrl %s", req.GetBucketUrl())
		}
		if req.GetAccessKey() == "" {
			return fmt.Errorf("invalid param accessKey %s", req.GetAccessKey())
		}
		if req.GetSecretAccessKey() == "" {
			return fmt.Errorf("invalid param secretAccessKey %s", req.GetSecretAccessKey())
		}
	case common.NfsStorageType:
		absPath, err := filepath.Abs(common.DefaultExportDir)
		if err != nil { //todo: get from config
			return fmt.Errorf("export dir %s is not vaild", common.DefaultExportDir)
		}
		if !checkFilePathExists(absPath) {
			//return fmt.Errorf("export path %s not exist", absPath)
			_ = os.MkdirAll(absPath, os.ModeDir)
		}
	default:
		return fmt.Errorf("invalid param storageType %s", req.GetStorageType())
	}

	return nil
}

func ImportDataPreCheck(ctx context.Context, req *clusterpb.DataImportRequest) error {
	if req.GetClusterId() == "" {
		return fmt.Errorf("invalid param clusterId %s", req.GetClusterId())
	}
	if req.GetUserName() == "" {
		return fmt.Errorf("invalid param userName %s", req.GetUserName())
	}
	/*
		if req.GetPassword() == "" {
			return fmt.Errorf("invalid param password %s", req.GetPassword())
		}
	*/
	absPath, err := filepath.Abs(common.DefaultImportDir)
	if err != nil { //todo: get from config
		return fmt.Errorf("import dir %s is not vaild", common.DefaultImportDir)
	}
	if !checkFilePathExists(absPath) {
		//return fmt.Errorf("import path %s not exist", absPath)
		_ = os.MkdirAll(absPath, os.ModeDir)
	}

	if req.GetRecordId() == 0 {
		switch req.GetStorageType() {
		case common.S3StorageType:
			if req.GetEndpointUrl() == "" {
				return fmt.Errorf("invalid param endpointUrl %s", req.GetEndpointUrl())
			}
			if req.GetBucketUrl() == "" {
				return fmt.Errorf("invalid param bucketUrl %s", req.GetBucketUrl())
			}
			if req.GetAccessKey() == "" {
				return fmt.Errorf("invalid param accessKey %s", req.GetAccessKey())
			}
			if req.GetSecretAccessKey() == "" {
				return fmt.Errorf("invalid param secretAccessKey %s", req.GetSecretAccessKey())
			}
		case common.NfsStorageType:
			break
		default:
			return fmt.Errorf("invalid param storageType %s", req.GetStorageType())
		}
	} else {
		// import from transport record
		req.StorageType = common.NfsStorageType
		dbReq := &dbpb.DBFindTransportRecordByIDRequest{
			RecordId: req.RecordId,
		}
		resp, err := client.DBClient.FindTrasnportRecordByID(ctx, dbReq)
		if err != nil {
			return fmt.Errorf("find transport record %d failed, %s", req.GetRecordId(), err.Error())
		}
		if service.ClusterSuccessResponseStatus.GetCode() != resp.GetStatus().GetCode() {
			return fmt.Errorf("find transport record %d failed, %s", req.GetRecordId(), resp.GetStatus().GetMessage())
		}
		record := resp.GetRecord()
		if !checkFilePathExists(record.GetFilePath()) {
			return fmt.Errorf("data source path %s not exist", record.GetFilePath())
		}
		if record.GetStorageType() != common.NfsStorageType {
			return fmt.Errorf("storage type %s can not support re-import", record.GetStorageType())
		}
		if !record.GetReImportSupport() {
			return fmt.Errorf("transport record not support re-import")
		}
	}

	return nil
}

func (mgr *ImportExportManager) ExportData(ctx context.Context, request *message.DataExportReq) (*message.DataExportResp, error) {
	framework.LogWithContext(ctx).Infof("begin exportdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end exportdata")

	clusterAggregation, err := ClusterRepo.Load(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster %s aggregation from metadb failed", request.ClusterID)
		return nil, err
	}

	flow, err := workflow.GetWorkFlowManager().CreateWorkFlow(ctx, request.ClusterID, constants.WorkFlowExportData)
	if err != nil {
		return nil, err
	}

	exportTime := time.Now()
	exportPrefix, _ := filepath.Abs(common.DefaultExportDir) //todo: get from config
	exportDir := filepath.Join(exportPrefix, request.ClusterID, fmt.Sprintf("%s_%s", exportTime.Format("2006-01-02_15:04:05"), request.StorageType))

	rw := models.GetImportExportReaderWriter()
	record, err := rw.CreateDataTransportRecord(ctx, &importexport.DataTransportRecord{
		Entities: dbCommon.Entities{
			TenantId: "", //todo
			Status:   string(constants.DataImportProcessing),
		},
		ClusterID:       request.ClusterID,
		TransportType:   string(constants.TransportTypeExport),
		FilePath:        getDataExportFilePath(request, exportDir, true),
		ZipName:         request.ZipName,
		StorageType:     request.StorageType,
		Comment:         request.Comment,
		ReImportSupport: checkExportParamSupportReimport(request),
		StartTime:       time.Now(),
		EndTime:         time.Now(),
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create data transport records failed %s", err)
		return nil, err
	}

	info := &ExportInfo{
		ClusterId:    request.ClusterID,
		UserName:     request.UserName,
		Password:     request.Password,
		FileType:     request.FileType,
		RecordId:     record.ID,
		FilePath:     getDataExportFilePath(request, exportDir, false),
		Filter:       request.Filter,
		Sql:          request.Sql,
		StorageType:  request.StorageType,
		BucketRegion: request.BucketRegion,
	}

	// Start the workflow
	workflow.GetWorkFlowManager().AddContext(flow, contextClusterKey, clusterAggregation)
	workflow.GetWorkFlowManager().AddContext(flow, contextDataTransportKey, info)
	workflow.GetWorkFlowManager().AsyncStart(ctx, flow)

	clusterAggregation.CurrentWorkFlow = flow.Flow
	err = ClusterRepo.Persist(ctx, clusterAggregation)
	if err != nil {
		return nil, err
	}
	return &message.DataExportResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
		RecordID: record.ID,
	}, nil
}

func (mgr *ImportExportManager) ImportData(ctx context.Context, request *message.DataImportReq) (*message.DataImportResp, error) {
	framework.LogWithContext(ctx).Infof("begin importdata request %+v", request)
	defer framework.LogWithContext(ctx).Infof("end importdata")

	clusterAggregation, err := ClusterRepo.Load(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster %s aggregation from metadb failed", request.ClusterID)
		return nil, err
	}

	flow, err := workflow.GetWorkFlowManager().CreateWorkFlow(ctx, request.ClusterID, constants.WorkFlowImportData)
	if err != nil {
		return nil, err
	}

	rw := models.GetImportExportReaderWriter()
	var info *ImportInfo
	importTime := time.Now()
	importPrefix, _ := filepath.Abs(common.DefaultImportDir) //todo: get from config
	importDir := filepath.Join(importPrefix, request.ClusterID, fmt.Sprintf("%s_%s", importTime.Format("2006-01-02_15:04:05"), request.StorageType))
	if request.RecordId == "" {
		if common.NfsStorageType == request.StorageType {
			err = os.Rename(filepath.Join(importPrefix, request.ClusterID, "temp"), importDir)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("find import dir failed, %s", err)
				return nil, err
			}
		} else {
			err = os.MkdirAll(importDir, os.ModeDir)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("mkdir import dir failed, %s", err)
				return nil, err
			}
		}

		record, err := rw.CreateDataTransportRecord(ctx, &importexport.DataTransportRecord{
			Entities: dbCommon.Entities{
				TenantId: "", //todo
				Status:   string(constants.DataImportProcessing),
			},
			ClusterID:       request.ClusterID,
			TransportType:   string(constants.TransportTypeImport),
			FilePath:        getDataImportFilePath(request, importDir, true),
			ZipName:         constants.DefaultZipName,
			StorageType:     request.StorageType,
			Comment:         request.Comment,
			ReImportSupport: checkImportParamSupportReimport(request),
			StartTime:       time.Now(),
			EndTime:         time.Now(),
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create data transport records failed %s", err.Error())
			return nil, err
		}

		info = &ImportInfo{
			ClusterId:   request.ClusterID,
			UserName:    request.UserName,
			Password:    request.Password,
			FilePath:    getDataImportFilePath(request, importDir, false),
			RecordId:    record.ID,
			StorageType: request.StorageType,
			ConfigPath:  importDir,
		}
	} else {
		// import from transport record
		srcRecord, err := rw.GetDataTransportRecord(ctx, request.RecordId)
		if err != nil {
			return nil, fmt.Errorf("find transport record %s failed, %s", request.RecordId, err.Error())
		}

		if err := os.MkdirAll(importDir, os.ModeDir); err != nil {
			return nil, fmt.Errorf("make import dir %s failed, %s", importDir, err.Error())
		}

		record, err := rw.CreateDataTransportRecord(ctx, &importexport.DataTransportRecord{
			Entities: dbCommon.Entities{
				TenantId: "", //todo
				Status:   string(constants.DataImportProcessing),
			},
			ClusterID:       request.ClusterID,
			TransportType:   string(constants.TransportTypeImport),
			FilePath:        getDataImportFilePath(request, importDir, true),
			ZipName:         constants.DefaultZipName,
			StorageType:     request.StorageType,
			Comment:         request.Comment,
			ReImportSupport: false,
			StartTime:       time.Now(),
			EndTime:         time.Now(),
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("create data transport records failed %s", err.Error())
			return nil, err
		}

		info = &ImportInfo{
			ClusterId:   request.ClusterID,
			UserName:    request.UserName,
			Password:    request.Password,
			FilePath:    srcRecord.FilePath,
			RecordId:    record.ID,
			StorageType: request.StorageType,
			ConfigPath:  importDir,
		}
	}

	// Start the workflow
	workflow.GetWorkFlowManager().AddContext(flow, contextClusterKey, clusterAggregation)
	workflow.GetWorkFlowManager().AddContext(flow, contextDataTransportKey, info)
	workflow.GetWorkFlowManager().AsyncStart(ctx, flow)

	clusterAggregation.CurrentWorkFlow = flow.Flow
	err = ClusterRepo.Persist(ctx, clusterAggregation)
	if err != nil {
		return nil, err
	}
	return &message.DataImportResp{
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: flow.Flow.ID,
		},
		RecordID: info.RecordId,
	}, nil
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

func convertTomlConfig(clusterAggregation *ClusterAggregation, info *ImportInfo) *DataImportConfig {
	if clusterAggregation == nil || clusterAggregation.CurrentTopologyConfigRecord == nil {
		return nil
	}
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	if configModel == nil || configModel.TiDBServers == nil || configModel.PDServers == nil {
		return nil
	}
	tidbServer := configModel.TiDBServers[0]
	pdServer := configModel.PDServers[0]

	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = common.DefaultTidbPort
	}

	tidbStatusPort := tidbServer.StatusPort
	if tidbStatusPort == 0 {
		tidbStatusPort = common.DefaultTidbStatusPort
	}

	pdClientPort := pdServer.ClientPort
	if pdClientPort == 0 {
		pdClientPort = common.DefaultPDClientPort
	}

	/*
	 * todo: sorted-kv-dir and data-source-dir in the same disk, may slow down import performance,
	 *  and check-requirements = true can not pass lightning pre-check
	 *  in real environment, config data-source-dir = user nfs storage, sorted-kv-dir = other disk, turn on pre-check
	 */
	config := &DataImportConfig{
		Lightning: LightningCfg{
			Level:             "info",
			File:              fmt.Sprintf("%s/tidb-lightning.log", info.ConfigPath),
			CheckRequirements: false, //todo: TBD
		},
		TikvImporter: TikvImporterCfg{
			Backend:     BackendLocal,
			SortedKvDir: info.ConfigPath, //todo: TBD
		},
		MyDumper: MyDumperCfg{
			DataSourceDir: info.FilePath,
		},
		Tidb: TidbCfg{
			Host:       tidbServer.Host,
			Port:       tidbServerPort,
			User:       info.UserName,
			Password:   info.Password,
			StatusPort: tidbStatusPort,
			PdAddr:     fmt.Sprintf("%s:%d", pdServer.Host, pdClientPort),
		},
	}
	return config
}

func checkFilePathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func checkExportParamSupportReimport(request *message.DataExportReq) bool {
	if common.S3StorageType == request.StorageType {
		return false
	}
	if request.Filter == "" && request.Sql != "" && FileTypeCSV == request.FileType {
		return false
	}
	return true
}

func checkImportParamSupportReimport(request *message.DataImportReq) bool {
	if common.NfsStorageType == request.StorageType {
		return true
	}
	return false
}

func getDataExportFilePath(request *message.DataExportReq, exportDir string, persist bool) string {
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

func getDataImportFilePath(request *message.DataImportReq, importDir string, persist bool) string {
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

func buildDataImportConfig(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin buildDataImportConfig")
	defer framework.LogWithContext(ctx).Info("end buildDataImportConfig")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)

	config := convertTomlConfig(clusterAggregation, info)
	if config == nil {
		framework.LogWithContext(ctx).Errorf("convert toml config failed, cluster: %v", clusterAggregation)
		node.Fail(fmt.Errorf("convert toml config failed, cluster: %v", clusterAggregation))
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
	node.Success(nil)
	return true
}

func importDataToCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin importDataToCluster")
	defer framework.LogWithContext(ctx).Info("end importDataToCluster")

	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)

	//tiup tidb-lightning -config tidb-lightning.toml
	importTaskId, err := secondparty.SecondParty.MicroSrvLightning(flowContext.Context, 0,
		[]string{"-config", fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)},
		uint64(node.ID))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup lightning api failed, %s", err.Error())
		node.Fail(fmt.Errorf("call tiup lightning api failed, %s", err.Error()))
		return false
	}
	framework.LogWithContext(ctx).Infof("call tiupmgr tidb-lightning api success, importTaskId %d", importTaskId)
	node.Success(nil)
	return true
}

func updateDataImportRecord(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin updateDataImportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataImportRecord")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	req := &dbpb.DBUpdateTransportRecordRequest{
		Record: &dbpb.TransportRecordDTO{
			RecordId:  info.RecordId,
			ClusterId: cluster.Id,
			EndTime:   time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.UpdateTransportRecord(flowContext, req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		node.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage())
		node.Fail(fmt.Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage()))
		return false
	}
	framework.LogWithContext(ctx).Infof("update data transport record success, %v", resp)
	node.Success(nil)
	return true
}

func exportDataFromCluster(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin exportDataFromCluster")
	defer framework.LogWithContext(ctx).Info("end exportDataFromCluster")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ExportInfo)
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]
	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = common.DefaultTidbPort
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
		"-P", strconv.Itoa(tidbServerPort),
		"--host", tidbServer.Host,
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
	exportTaskId, err := secondparty.SecondParty.MicroSrvDumpling(ctx, 0, cmd, uint64(node.ID))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup dumpling api failed, %s", err.Error())
		node.Fail(fmt.Errorf("call tiup dumpling api failed, %s", err.Error()))
		return false
	}

	framework.LogWithContext(ctx).Infof("call tiupmgr succee, exportTaskId: %d", exportTaskId)
	node.Success(nil)
	return true
}

func updateDataExportRecord(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin updateDataExportRecord")
	defer framework.LogWithContext(ctx).Info("end updateDataExportRecord")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ExportInfo)
	cluster := clusterAggregation.Cluster

	req := &dbpb.DBUpdateTransportRecordRequest{
		Record: &dbpb.TransportRecordDTO{
			RecordId:  info.RecordId,
			ClusterId: cluster.Id,
			EndTime:   time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.UpdateTransportRecord(flowContext, req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		node.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage())
		node.Fail(fmt.Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage()))
		return false
	}
	framework.LogWithContext(ctx).Infof("update data transport record success, %v", resp)
	node.Success(nil)
	return true
}

func importDataFailed(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin importDataFailed")
	defer framework.LogWithContext(ctx).Info("end importDataFailed")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(ctx, info.RecordId, cluster.Id); err != nil {
		node.Fail(err)
		return false
	}

	return clusterFail(node, flowContext)
}

func exportDataFailed(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	ctx := flowContext.Context
	framework.LogWithContext(ctx).Info("begin exportDataFailed")
	defer framework.LogWithContext(ctx).Info("end exportDataFailed")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ExportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(ctx, info.RecordId, cluster.Id); err != nil {
		node.Fail(err)
		return false
	}

	return clusterFail(node, flowContext)
}

func clusterEnd(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Success(nil)
	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true

	return true
}

func clusterFail(node *wfModel.WorkFlowNode, flowContext *workflow.FlowContext) bool {
	node.Status = constants.WorkFlowStatusError
	node.Result = "fail"
	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	clusterAggregation.Cluster.WorkFlowId = 0
	clusterAggregation.FlowModified = true
	return true
}

func updateTransportRecordFailed(ctx context.Context, recordId int64, clusterId string) error {
	req := &dbpb.DBUpdateTransportRecordRequest{
		Record: &dbpb.TransportRecordDTO{
			RecordId:  recordId,
			ClusterId: clusterId,
			EndTime:   time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.UpdateTransportRecord(ctx, req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		framework.LogWithContext(ctx).Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage())
		return errors.New(resp.GetStatus().GetMessage())
	}
	framework.LogWithContext(ctx).Infof("update data transport record success, %v", resp)
	return nil
}
