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

package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/micro-metadb/service"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/secondparty"
)

type ImportInfo struct {
	ClusterId   string
	UserName    string
	Password    string
	FilePath    string
	RecordId    int64
	StorageType string
	ConfigPath  string
}

type ExportInfo struct {
	ClusterId    string
	UserName     string
	Password     string
	FileType     string
	RecordId     int64
	FilePath     string
	Filter       string
	Sql          string
	StorageType  string
	BucketRegion string
}

/*
	data import toml config for lightning
	https://docs.pingcap.com/zh/tidb/dev/tidb-lightning-configuration
*/
type DataImportConfig struct {
	Lightning    LightningCfg    `toml:"lightning"`
	TikvImporter TikvImporterCfg `toml:"tikv-importer"`
	MyDumper     MyDumperCfg     `toml:"mydumper"`
	Tidb         TidbCfg         `toml:"tidb"`
}

type LightningCfg struct {
	Level             string `toml:"level"`              //lightning log level
	File              string `toml:"file"`               //lightning log path
	CheckRequirements bool   `toml:"check-requirements"` //lightning pre check
}

/*
	tidb-lightning backend
	https://docs.pingcap.com/zh/tidb/stable/tidb-lightning-backends#tidb-lightning-backend
*/
const (
	BackendLocal string = "local"
	//BackendImport string = "importer"
	//BackendTidb   string = "tidb"
)

const (
	FileTypeCSV string = "csv"
	FileTypeSQL string = "sql"
)

const (
	DefaultTidbPort       int = 4000
	DefaultTidbStatusPort int = 10080
	DefaultPDClientPort   int = 2379
	DefaultAlertPort      int = 9093
	DefaultGrafanaPort    int = 3000
)

type TikvImporterCfg struct {
	Backend     string `toml:"backend"`       //backend mode: local/normal
	SortedKvDir string `toml:"sorted-kv-dir"` //temp store path
}

type MyDumperCfg struct {
	DataSourceDir string `toml:"data-source-dir"` //import data filepath
}

type TidbCfg struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	StatusPort int    `toml:"status-port"` //table infomation from tidb status port
	PdAddr     string `toml:"pd-addr"`
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

func ExportData(ctx context.Context, request *clusterpb.DataExportRequest) (int64, error) {
	getLoggerWithContext(ctx).Infof("begin exportdata request %+v", request)
	defer getLoggerWithContext(ctx).Infof("end exportdata")

	operator := parseOperatorFromDTO(request.GetOperator())
	getLoggerWithContext(ctx).Info(operator)
	clusterAggregation, err := ClusterRepo.Load(ctx, request.GetClusterId())
	if err != nil {
		getLoggerWithContext(ctx).Errorf("load cluster %s aggregation from metadb failed", request.GetClusterId())
		return 0, err
	}

	flow, err := CreateFlowWork(ctx, request.GetClusterId(), FlowExportData, operator)
	if err != nil {
		return 0, err
	}

	exportTime := time.Now()
	exportPrefix, _ := filepath.Abs(common.DefaultExportDir) //todo: get from config
	exportDir := filepath.Join(exportPrefix, request.GetClusterId(), fmt.Sprintf("%s_%s", exportTime.Format("2006-01-02_15:04:05"), request.GetStorageType()))

	req := &dbpb.DBCreateTransportRecordRequest{
		Record: &dbpb.TransportRecordDTO{
			ClusterId:       request.GetClusterId(),
			TenantId:        operator.TenantId,
			TransportType:   string(common.TransportTypeExport),
			FilePath:        getDataExportFilePath(request, exportDir, true),
			ZipName:         request.GetZipName(),
			StorageType:     request.GetStorageType(),
			FlowId:          int64(flow.FlowWork.Id),
			ReImportSupport: checkExportParamSupportReimport(request),
			Comment:         request.GetComment(),
			StartTime:       time.Now().Unix(),
			EndTime:         time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.CreateTransportRecord(ctx, req)
	if err != nil {
		return 0, err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		return 0, errors.New(resp.GetStatus().GetMessage())
	}

	info := &ExportInfo{
		ClusterId:    request.GetClusterId(),
		UserName:     request.GetUserName(),
		Password:     request.GetPassword(),
		FileType:     request.GetFileType(),
		RecordId:     resp.GetRecordId(),
		FilePath:     getDataExportFilePath(request, exportDir, false),
		Filter:       request.GetFilter(),
		Sql:          request.GetSql(),
		StorageType:  request.GetStorageType(),
		BucketRegion: request.GetBucketRegion(),
	}

	// Start the workflow
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDataTransportKey, info)
	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	err = ClusterRepo.Persist(ctx, clusterAggregation)
	if err != nil {
		return 0, err
	}
	return info.RecordId, nil
}

func ImportData(ctx context.Context, request *clusterpb.DataImportRequest) (int64, error) {
	getLoggerWithContext(ctx).Infof("begin importdata request %+v", request)
	defer getLoggerWithContext(ctx).Infof("end importdata")

	operator := parseOperatorFromDTO(request.GetOperator())
	getLoggerWithContext(ctx).Info(operator)
	clusterAggregation, err := ClusterRepo.Load(ctx, request.GetClusterId())
	if err != nil {
		getLoggerWithContext(ctx).Errorf("load cluster %s aggregation from metadb failed", request.GetClusterId())
		return 0, err
	}

	flow, err := CreateFlowWork(ctx, request.GetClusterId(), FlowImportData, operator)
	if err != nil {
		return 0, err
	}

	var info *ImportInfo
	importTime := time.Now()
	importPrefix, _ := filepath.Abs(common.DefaultImportDir) //todo: get from config
	importDir := filepath.Join(importPrefix, request.GetClusterId(), fmt.Sprintf("%s_%s", importTime.Format("2006-01-02_15:04:05"), request.GetStorageType()))
	if request.GetRecordId() == 0 {
		if common.NfsStorageType == request.GetStorageType() {
			err = os.Rename(filepath.Join(importPrefix, request.GetClusterId(), "temp"), importDir)
			if err != nil {
				getLoggerWithContext(ctx).Errorf("find import dir failed, %s", err.Error())
				return 0, err
			}
		} else {
			err = os.MkdirAll(importDir, os.ModeDir)
			if err != nil {
				getLoggerWithContext(ctx).Errorf("mkdir import dir failed, %s", err.Error())
				return 0, err
			}
		}

		req := &dbpb.DBCreateTransportRecordRequest{
			Record: &dbpb.TransportRecordDTO{
				ClusterId:       request.GetClusterId(),
				TenantId:        operator.TenantId,
				TransportType:   string(common.TransportTypeImport),
				StorageType:     request.GetStorageType(),
				FilePath:        getDataImportFilePath(request, importDir, true),
				ZipName:         common.DefaultZipName,
				FlowId:          int64(flow.FlowWork.Id),
				ReImportSupport: checkImportParamSupportReimport(request),
				Comment:         request.GetComment(),
				StartTime:       time.Now().Unix(),
				EndTime:         time.Now().Unix(),
			},
		}
		resp, err := client.DBClient.CreateTransportRecord(ctx, req)
		if err != nil {
			return 0, err
		}
		if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
			return 0, errors.New(resp.GetStatus().GetMessage())
		}
		info = &ImportInfo{
			ClusterId:   request.GetClusterId(),
			UserName:    request.GetUserName(),
			Password:    request.GetPassword(),
			FilePath:    getDataImportFilePath(request, importDir, false),
			RecordId:    resp.GetRecordId(),
			StorageType: request.GetStorageType(),
			ConfigPath:  importDir,
		}
	} else {
		// import from transport record
		dbReq := &dbpb.DBFindTransportRecordByIDRequest{
			RecordId: request.GetRecordId(),
		}
		queryResp, err := client.DBClient.FindTrasnportRecordByID(ctx, dbReq)
		if err != nil {
			return 0, fmt.Errorf("find transport record %d failed, %s", request.GetRecordId(), err.Error())
		}
		if service.ClusterSuccessResponseStatus.GetCode() != queryResp.GetStatus().GetCode() {
			return 0, fmt.Errorf("find transport record %d failed, %s", request.GetRecordId(), queryResp.GetStatus().GetMessage())
		}

		record := queryResp.GetRecord()
		if err := os.MkdirAll(importDir, os.ModeDir); err != nil {
			return 0, fmt.Errorf("make import dir %s failed, %s", importDir, err.Error())
		}

		req := &dbpb.DBCreateTransportRecordRequest{
			Record: &dbpb.TransportRecordDTO{
				ClusterId:       request.GetClusterId(),
				TenantId:        operator.TenantId,
				TransportType:   string(common.TransportTypeImport),
				StorageType:     request.GetStorageType(),
				FilePath:        record.GetFilePath(),
				ZipName:         common.DefaultZipName,
				FlowId:          int64(flow.FlowWork.Id),
				ReImportSupport: false, // import from other data source, can not re-import cause it has own data file
				Comment:         request.GetComment(),
				StartTime:       time.Now().Unix(),
				EndTime:         time.Now().Unix(),
			},
		}
		resp, createErr := client.DBClient.CreateTransportRecord(ctx, req)
		if createErr != nil {
			return 0, createErr
		}
		if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
			return 0, errors.New(resp.GetStatus().GetMessage())
		}
		info = &ImportInfo{
			ClusterId:   request.GetClusterId(),
			UserName:    request.GetUserName(),
			Password:    request.GetPassword(),
			FilePath:    record.GetFilePath(),
			RecordId:    resp.GetRecordId(),
			StorageType: request.GetStorageType(),
			ConfigPath:  importDir,
		}
	}

	// Start the workflow
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDataTransportKey, info)
	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	err = ClusterRepo.Persist(ctx, clusterAggregation)
	if err != nil {
		return 0, err
	}
	return info.RecordId, nil
}

func DescribeDataTransportRecord(ctx context.Context, ope *clusterpb.OperatorDTO, recordId int64, clusterId string, reImport bool, startTime, endTime int64, page, pageSize int32) ([]*dbpb.DBTransportRecordDisplayDTO, *dbpb.DBPageDTO, error) {
	getLoggerWithContext(ctx).Infof("begin DescribeDataTransportRecord clusterId: %s, recordId: %d, reImport: %v, page: %d, pageSize: %d", clusterId, recordId, reImport, page, pageSize)
	defer getLoggerWithContext(ctx).Info("end DescribeDataTransportRecord")

	operator := parseOperatorFromDTO(ope)
	getLoggerWithContext(ctx).Info(operator)

	req := &dbpb.DBListTransportRecordRequest{
		Page: &dbpb.DBPageDTO{
			Page:     page,
			PageSize: pageSize,
		},
		ClusterId: clusterId,
		RecordId:  recordId,
		ReImport:  reImport,
		StartTime: startTime,
		EndTime:   endTime,
	}
	resp, err := client.DBClient.ListTrasnportRecord(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		return nil, nil, errors.New(resp.GetStatus().GetMessage())
	}

	return resp.GetRecords(), resp.GetPage(), nil
}

func DeleteDataTransportRecord(ctx context.Context, ope *clusterpb.OperatorDTO, clusterId string, recordId int64) error {
	getLoggerWithContext(ctx).Infof("begin DeleteDataTransportRecord clusterId: %s, recordId: %d", clusterId, recordId)
	defer getLoggerWithContext(ctx).Info("end DeleteDataTransportRecord")

	operator := parseOperatorFromDTO(ope)
	getLoggerWithContext(ctx).Info(operator)

	resp, err := client.DBClient.FindTrasnportRecordByID(ctx, &dbpb.DBFindTransportRecordByIDRequest{RecordId: recordId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("query transport record %d of cluster %s failed, %s", recordId, clusterId, err.Error())
		return fmt.Errorf("query transport record %d of cluster %s failed, %s", recordId, clusterId, err.Error())
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("query transport record %d of cluster %s failed, %s", recordId, clusterId, resp.GetStatus().GetMessage())
		return fmt.Errorf("query backup transport %d of cluster %s failed, %s", recordId, clusterId, resp.GetStatus().GetMessage())
	}
	getLoggerWithContext(ctx).Infof("query transport record to be deleted, record: %+v", resp.GetRecord())
	if resp.GetRecord().GetRecordId() <= 0 {
		getLoggerWithContext(ctx).Infof("record %d not exist", recordId)
		return nil
	}

	if common.S3StorageType != resp.GetRecord().GetStorageType() {
		filePath := filepath.Dir(resp.GetRecord().GetFilePath())
		_ = os.RemoveAll(filePath)
		getLoggerWithContext(ctx).Infof("remove file path %s, record: %+v", filePath, resp.GetRecord())
	}

	delResp, err := client.DBClient.DeleteTransportRecord(ctx, &dbpb.DBDeleteTransportRequest{RecordId: recordId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("delete metadb transport record failed, %s", err.Error())
		return fmt.Errorf("delete metadb transport record failed, %s", err.Error())
	}
	if delResp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("delete metadb transport record failed, %s", delResp.GetStatus().GetMessage())
		return fmt.Errorf("delete metadb transport record failed, %s", delResp.GetStatus().GetMessage())
	}
	getLoggerWithContext(ctx).Infof("delete transport record to be deleted, record: %+v", delResp.GetRecord())

	return nil
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
		tidbServerPort = DefaultTidbPort
	}

	tidbStatusPort := tidbServer.StatusPort
	if tidbStatusPort == 0 {
		tidbStatusPort = DefaultTidbStatusPort
	}

	pdClientPort := pdServer.ClientPort
	if pdClientPort == 0 {
		pdClientPort = DefaultPDClientPort
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

func checkExportParamSupportReimport(request *clusterpb.DataExportRequest) bool {
	if common.S3StorageType == request.GetStorageType() {
		return false
	}
	if request.GetFilter() == "" && request.GetSql() != "" && FileTypeCSV == request.GetFileType() {
		return false
	}
	return true
}

func checkImportParamSupportReimport(request *clusterpb.DataImportRequest) bool {
	if common.NfsStorageType == request.GetStorageType() {
		return true
	}
	return false
}

func getDataExportFilePath(request *clusterpb.DataExportRequest, exportDir string, persist bool) string {
	var filePath string
	if common.S3StorageType == request.GetStorageType() {
		if persist {
			filePath = fmt.Sprintf("%s?&endpoint=%s", request.GetBucketUrl(), request.GetEndpointUrl())
		} else {
			filePath = fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.GetBucketUrl(), request.GetAccessKey(), request.GetSecretAccessKey(), request.GetEndpointUrl())
		}
	} else {
		filePath = filepath.Join(exportDir, "data")
	}
	return filePath
}

func getDataImportFilePath(request *clusterpb.DataImportRequest, importDir string, persist bool) string {
	var filePath string
	if common.S3StorageType == request.GetStorageType() {
		if persist {
			filePath = fmt.Sprintf("%s?&endpoint=%s", request.GetBucketUrl(), request.GetEndpointUrl())
		} else {
			filePath = fmt.Sprintf("%s?access-key=%s&secret-access-key=%s&endpoint=%s&force-path-style=true", request.GetBucketUrl(), request.GetAccessKey(), request.GetSecretAccessKey(), request.GetEndpointUrl())
		}
	} else {
		filePath = filepath.Join(importDir, "data")
	}
	return filePath
}

func cleanDataTransportDir(ctx context.Context, filepath string) error {
	getLoggerWithContext(ctx).Infof("clean and re-mkdir data dir: %s", filepath)
	if err := os.RemoveAll(filepath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath, os.ModeDir); err != nil {
		return err
	}
	return nil
}

func buildDataImportConfig(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin buildDataImportConfig")
	defer getLoggerWithContext(ctx).Info("end buildDataImportConfig")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)

	config := convertTomlConfig(clusterAggregation, info)
	if config == nil {
		getLoggerWithContext(ctx).Errorf("convert toml config failed, cluster: %v", clusterAggregation)
		task.Fail(fmt.Errorf("convert toml config failed, cluster: %v", clusterAggregation))
		return false
	}

	filePath := fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		getLoggerWithContext(ctx).Errorf("create import toml config failed, %s", err.Error())
		task.Fail(fmt.Errorf("create import toml config failed, %s", err.Error()))
		return false
	}

	if err = toml.NewEncoder(file).Encode(config); err != nil {
		getLoggerWithContext(ctx).Errorf("encode data import toml config failed, %s", err.Error())
		task.Fail(fmt.Errorf("encode data import toml config failed, %s", err.Error()))
		return false
	}
	getLoggerWithContext(ctx).Infof("build lightning toml config sucess, %v", config)
	task.Success(nil)
	return true
}

func importDataToCluster(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin importDataToCluster")
	defer getLoggerWithContext(ctx).Info("end importDataToCluster")

	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)

	//tiup tidb-lightning -config tidb-lightning.toml
	importTaskId, err := secondparty.SecondParty.MicroSrvLightning(flowContext.Context, 0,
		[]string{"-config", fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)},
		uint64(task.Id))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("call tiup lightning api failed, %s", err.Error())
		task.Fail(fmt.Errorf("call tiup lightning api failed, %s", err.Error()))
		return false
	}
	getLoggerWithContext(ctx).Infof("call tiupmgr tidb-lightning api success, importTaskId %d", importTaskId)
	task.Success(nil)
	return true
}

func updateDataImportRecord(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin updateDataImportRecord")
	defer getLoggerWithContext(ctx).Info("end updateDataImportRecord")

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
		getLoggerWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		task.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage())
		task.Fail(fmt.Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage()))
		return false
	}
	getLoggerWithContext(ctx).Infof("update data transport record success, %v", resp)
	task.Success(nil)
	return true
}

func exportDataFromCluster(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin exportDataFromCluster")
	defer getLoggerWithContext(ctx).Info("end exportDataFromCluster")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ExportInfo)
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]
	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = DefaultTidbPort
	}

	if common.NfsStorageType == info.StorageType {
		if err := cleanDataTransportDir(ctx, info.FilePath); err != nil {
			getLoggerWithContext(ctx).Errorf("clean export directory failed, %s", err.Error())
			task.Fail(fmt.Errorf("clean export directory failed, %s", err.Error()))
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
	getLoggerWithContext(ctx).Infof("call tiupmgr dumpling api, cmd: %v", cmd)
	exportTaskId, err := secondparty.SecondParty.MicroSrvDumpling(ctx, 0, cmd, uint64(task.Id))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("call tiup dumpling api failed, %s", err.Error())
		task.Fail(fmt.Errorf("call tiup dumpling api failed, %s", err.Error()))
		return false
	}

	getLoggerWithContext(ctx).Infof("call tiupmgr succee, exportTaskId: %d", exportTaskId)
	task.Success(nil)
	return true
}

func updateDataExportRecord(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin updateDataExportRecord")
	defer getLoggerWithContext(ctx).Info("end updateDataExportRecord")

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
		getLoggerWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		task.Fail(fmt.Errorf("update data transport record failed, %s", err.Error()))
		return false
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage())
		task.Fail(fmt.Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage()))
		return false
	}
	getLoggerWithContext(ctx).Infof("update data transport record success, %v", resp)
	task.Success(nil)
	return true
}

func importDataFailed(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin importDataFailed")
	defer getLoggerWithContext(ctx).Info("end importDataFailed")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(ctx, info.RecordId, cluster.Id); err != nil {
		task.Fail(err)
		return false
	}

	return ClusterFail(task, flowContext)
}

func exportDataFailed(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.Context
	getLoggerWithContext(ctx).Info("begin exportDataFailed")
	defer getLoggerWithContext(ctx).Info("end exportDataFailed")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	info := flowContext.GetData(contextDataTransportKey).(*ExportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(ctx, info.RecordId, cluster.Id); err != nil {
		task.Fail(err)
		return false
	}

	return ClusterFail(task, flowContext)
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
		getLoggerWithContext(ctx).Errorf("update data transport record failed, %s", err.Error())
		return err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("update data transport record failed, %s", resp.GetStatus().GetMessage())
		return errors.New(resp.GetStatus().GetMessage())
	}
	getLoggerWithContext(ctx).Infof("update data transport record success, %v", resp)
	return nil
}
