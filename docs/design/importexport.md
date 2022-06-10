# TiUniManager improt and export design

- Author(s): [cchenkey](http://github.com/cchenkey)

## table of Contents

- [TiUniManager improt and export design](#TiUniManager-improt-and-export-design)
  - [table of Contents](#table-of-Contents)
  - [import export API](#import-export-API)
  - [design](#design)
    - [interface](#interface)
    - [workflow](#workflow)
    - [import process](#import-process)
    - [export process](#export-process)
    - [errors](#errors)

## import export API
micro-api/route/route.go，"/clusters" group define import export API,
``` go
		cluster := apiV1.Group("/clusters")
		{
		    // ...
			//Import and Export
			cluster.POST("/import", metrics.HandleMetrics(constants.MetricsDataImport), importexport.ImportData)
			cluster.POST("/export", metrics.HandleMetrics(constants.MetricsDataExport), importexport.ExportData)
			cluster.GET("/transport", metrics.HandleMetrics(constants.MetricsDataExportImportQuery), importexport.QueryDataTransport)
			cluster.DELETE("/transport/:recordId", metrics.HandleMetrics(constants.MetricsDataExportImportDelete), importexport.DeleteDataTransportRecord)
			// ...
		}
```
The `HandleFunc` of each route can be tracked as the entry of each API，
- `importexport.ImportData` import data to cluster；
- `importexport.ExportData` export data from cluster；
- `importexport.QueryDataTransport` query import and export records；
- `importexport.DeleteDataTransportRecord` delete import and export records；

## design

importexportManager is manager of import and export（micro-cluster/datatransfer/importexport），each file function：

- `service.go` definition of import export interface；
- `manger.go` implement of import export interface；
- `excutor.go` implement of improt export workflow nodes；
- `lighting.go` import tool lighting config；
- `common.go` constant definition of import and export；

### interface
interface of improt and export：
``` go
type ImportExportService interface {
	// ExportData
	// @Description: export data
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.DataExportResp
	// @Return error
	ExportData(ctx context.Context, request message.DataExportReq) (resp message.DataExportResp, exportErr error)

	// ImportData
	// @Description: import data
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.DataImportResp
	// @Return error
	ImportData(ctx context.Context, request message.DataImportReq) (resp message.DataImportResp, importErr error)

	// QueryDataTransportRecords
	// @Description: query data import & export records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryDataImportExportRecordsResp
	// @Return structs.Page
	// @Return error
	QueryDataTransportRecords(ctx context.Context, request message.QueryDataImportExportRecordsReq) (resp message.QueryDataImportExportRecordsResp, page structs.Page, err error)

	// DeleteDataTransportRecord
	// @Description: delete data import & export records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.DeleteImportExportRecordResp
	// @Return error
	DeleteDataTransportRecord(ctx context.Context, request message.DeleteImportExportRecordReq) (reps message.DeleteImportExportRecordResp, err error)
}
```

### workflow
import and export workflow definition：
``` go
flowManager.RegisterWorkFlow(context.TODO(), constants.FlowExportData, &workflow.WorkFlowDefine{
		FlowName: constants.FlowExportData,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"exportDataFromCluster", "exportDataDone", "fail", workflow.PollingNode, exportDataFromCluster},
			"exportDataDone":   {"updateDataExportRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateDataExportRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, exportDataFailed},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowImportData, &workflow.WorkFlowDefine{
		FlowName: constants.FlowImportData,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"buildDataImportConfig", "buildConfigDone", "fail", workflow.SyncFuncNode, buildDataImportConfig},
			"buildConfigDone":  {"importDataToCluster", "importDataDone", "fail", workflow.PollingNode, importDataToCluster},
			"importDataDone":   {"updateDataImportRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateDataImportRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, importDataFailed},
		},
	})
```
### import process

import data to cluster include 3 method: import from S3, upload file to nfs shared storage then import, import from history record. user need config improt and export path.

- import export manager will pre-check params and config is available
- build lighting.toml config by user params
- call tiup lightning to import data
- update import result to import-export record

### export process

export data from cluster support 2 methods: export data to S3, export data to nfs shared storage. user need config export path.

- importexport manager will pre-check params and config is available
- call tiup dumpling to export data
- update export result to import-export record

### errors

import export API errors define in file common/errors/errorcode.go
``` go
	// import && export
	TIEM_TRANSPORT_SYSTEM_CONFIG_NOT_FOUND: {"data transport system config not found", 404},
	TIEM_TRANSPORT_SYSTEM_CONFIG_INVALID:   {"data transport system config invalid", 400},
	TIEM_TRANSPORT_RECORD_NOT_FOUND:        {"data transport record is not found", 404},
	TIEM_TRANSPORT_RECORD_CREATE_FAILED:    {"create data transport record failed", 500},
	TIEM_TRANSPORT_RECORD_DELETE_FAILED:    {"delete data transport record failed", 500},
	TIEM_TRANSPORT_RECORD_QUERY_FAILED:     {"query data transport record failed", 500},
	TIEM_TRANSPORT_FILE_DELETE_FAILED:      {"remove transport file failed", 500},
	TIEM_TRANSPORT_PATH_CREATE_FAILED:      {"data transport filepath create failed", 500},
	TIEM_TRANSPORT_FILE_SIZE_INVALID:       {"data transport file size invalid", 400},
	TIEM_TRANSPORT_FILE_UPLOAD_FAILED:      {"data transport file upload failed", 500},
	TIEM_TRANSPORT_FILE_DOWNLOAD_FAILED:    {"data transport file download failed", 500},
	TIEM_TRANSPORT_FILE_TRANSFER_LIMITED:   {"exceed limit file transfer num", 400},
```
