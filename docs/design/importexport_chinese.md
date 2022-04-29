# TiEM 导入导出设计文档

- Author(s): [cchenkey](http://github.com/cchenkey)

## 目录

- [TiEM 导入导出设计文档](#tiem-导入导出设计文档)
  - [目录](#目录)
  - [导入导出模块API](#导入导出模块API)
  - [设计](#设计)
    - [接口](#接口)
    - [workflow](#workflow)
    - [导入数据流程](#导入数据流程)
    - [导出数据流程](#导出数据流程)
    - [异常错误](#异常错误)

## 导入导出模块API
在micro-api/route/route.go中，"/clusters" group定义了导入导出模块提供的API,
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
可以跟踪每条route后的HandleFunc进入每个API的入口，其中，
- `importexport.ImportData` 导入数据到集群；
- `importexport.ExportData` 从集群中导出数据；
- `importexport.QueryDataTransport` 查询导入导出记录；
- `importexport.DeleteDataTransportRecord` 删除导入导出记录；

## 设计

ImportExportManager是导入导出具体流程的入口（micro-cluster/datatransfer/importexport），文件功能如下：

- `service.go` 导入导出服务的接口定义；
- `manger.go` importexport manager定义和接口实现；
- `excutor.go` 导入导出workflow步骤实现；
- `lighting.go` 导入工具lighting的配置管理；
- `common.go` 导入导出相关常量定义；

### 接口
导入导出的接口如下：
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
导入和导出任务流步骤如下：
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
### 导入数据流程

导入数据支持三种模式导入：从S3存储导入数据，上传文件从nfs共享设备导入数据，从历史记录中导入数据。用户需要配置好中控机的文件存储位置

- importexport manager进行前置参数校验，检查导入配置项是否配置
- 根据用户提供的导入数据源，构建lighting.toml配置文件
- 调用tiup执行lighting导入过程
- 更新导入结果到导入导出历史记录

### 导出数据流程

导出数据支持两种模式导出：导出数据到S3存储，导出数据到用户的nfs共享设备上。用户需要配置好中控机的文件存储位置

- importexport manager进行前置参数校验，检查导出配置项是否配置
- 调用tiup执行dunpling导出过程
- 更新导出结果到导入导出历史记录

### 异常错误

导入导出的API报错定义在common/errors/errorcode.go
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
