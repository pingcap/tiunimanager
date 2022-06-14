# TiUniManager 备份恢复设计文档

- Author(s): [cchenkey](http://github.com/cchenkey)

## 目录

- [TiUniManager 备份恢复设计文档](#tiunimanager-备份恢复设计文档)
  - [目录](#目录)
  - [备份恢复模块API](#备份恢复模块API)
  - [设计](#设计)
    - [接口](#接口)
    - [workflow](#workflow)
    - [备份数据流程](#备份数据流程)
    - [恢复数据流程](#恢复数据流程)
    - [异常错误](#异常错误)

## 备份恢复模块API
在micro-api/route/route.go中，"/clusters"和"/backups" group定义了备份恢复模块提供的API,
``` go
		cluster := apiV1.Group("/clusters")
		{
		    // ...
			// Backup Strategy
			cluster.GET("/:clusterId/strategy", metrics.HandleMetrics(constants.MetricsBackupQueryStrategy), backuprestore.GetBackupStrategy)
			cluster.PUT("/:clusterId/strategy", metrics.HandleMetrics(constants.MetricsBackupModifyStrategy), backuprestore.SaveBackupStrategy)
			// ...
		}
		backup := apiV1.Group("/backups")
		{
			backup.POST("/", metrics.HandleMetrics(constants.MetricsBackupCreate), backuprestore.Backup)
			backup.POST("/cancel", metrics.HandleMetrics(constants.MetricsBackupCancel), backuprestore.CancelBackup)
			backup.GET("/", metrics.HandleMetrics(constants.MetricsBackupQuery), backuprestore.QueryBackupRecords)
			backup.DELETE("/:backupId", metrics.HandleMetrics(constants.MetricsBackupDelete), backuprestore.DeleteBackup)
		}
```
可以跟踪每条route后的HandleFunc进入每个API的入口，其中，
- `backuprestore.GetBackupStrategy` 查询集群自动备份策略；
- `backuprestore.SaveBackupStrategy` 保存集群自动备份策略；
- `backuprestore.Backup` 执行立即备份接口；
- `backuprestore.CancelBackup` 取消进行中的备份；
- `backuprestore.QueryBackupRecords` 查询备份记录；
- `backuprestore.DeleteBackup` 删除备份记录及备份文件；


## 设计

backupRestoreManager是备份恢复具体流程的入口（micro-cluster/cluster/backuprestore），文件功能如下：

- `service.go` 备份恢复服务的接口定义；
- `manger.go` 备份恢复 manager定义和接口实现；
- `excutor.go` 备份恢复workflow步骤实现；
- `autobackup.go` 自动备份逻辑实现；
- `common.go` 备份恢复相关常量定义；

### 接口
备份恢复的接口如下：
``` go
type BRService interface {
	// BackupCluster
	// @Description: backup cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter maintenanceStatusChange
	// @Return cluster.BackupClusterDataResp
	// @Return error
	BackupCluster(ctx context.Context, request cluster.BackupClusterDataReq, maintenanceStatusChange bool) (resp cluster.BackupClusterDataResp, backupErr error)

	// RestoreExistCluster
	// @Description: restore exist cluster by backup record
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Parameter maintenanceStatusChange
	// @Return cluster.RestoreExistClusterResp
	// @Return error
	RestoreExistCluster(ctx context.Context, request cluster.RestoreExistClusterReq, maintenanceStatusChange bool) (resp cluster.RestoreExistClusterResp, restoreErr error)

	// CancelBackup
	// @Description: cancel backup cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.CancelBackupResp
	// @Return error
	CancelBackup(ctx context.Context, request cluster.CancelBackupReq) (resp cluster.CancelBackupResp, cancelErr error)

	// QueryClusterBackupRecords
	// @Description: query backup records of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.QueryBackupRecordsResp
	// @Return structs.Page
	// @Return error
	QueryClusterBackupRecords(ctx context.Context, request cluster.QueryBackupRecordsReq) (resp cluster.QueryBackupRecordsResp, page structs.Page, err error)

	// DeleteBackupRecords
	// @Description: delete backup records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.QueryBackupRecordsResp
	// @Return error
	DeleteBackupRecords(ctx context.Context, request cluster.DeleteBackupDataReq) (resp cluster.DeleteBackupDataResp, err error)

	// SaveBackupStrategy
	// @Description: save backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.SaveBackupStrategyResp
	// @Return error
	SaveBackupStrategy(ctx context.Context, request cluster.SaveBackupStrategyReq) (resp cluster.SaveBackupStrategyResp, err error)

	// GetBackupStrategy
	// @Description: get backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.GetBackupStrategyResp
	// @Return error
	GetBackupStrategy(ctx context.Context, request cluster.GetBackupStrategyReq) (resp cluster.GetBackupStrategyResp, err error)

	// DeleteBackupStrategy
	// @Description: save backup strategy of cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return cluster.UpdateBackupStrategyResp
	// @Return error
	DeleteBackupStrategy(ctx context.Context, request cluster.DeleteBackupStrategyReq) (resp cluster.DeleteBackupStrategyResp, err error)
}

```

### workflow
备份和恢复任务流步骤如下：
``` go
flowManager.RegisterWorkFlow(context.TODO(), constants.FlowBackupCluster, &workflow.WorkFlowDefine{
		FlowName: constants.FlowBackupCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":            {"backup", "backupDone", "fail", workflow.SyncFuncNode, backupCluster},
			"backupDone":       {"updateBackupRecord", "updateRecordDone", "fail", workflow.SyncFuncNode, updateBackupRecord},
			"updateRecordDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":             {"fail", "", "", workflow.SyncFuncNode, backupFail},
		},
	})
	flowManager.RegisterWorkFlow(context.TODO(), constants.FlowRestoreExistCluster, &workflow.WorkFlowDefine{
		FlowName: constants.FlowRestoreExistCluster,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":       {"restoreFromSrcCluster", "restoreDone", "fail", workflow.SyncFuncNode, restoreFromSrcCluster},
			"restoreDone": {"end", "", "", workflow.SyncFuncNode, defaultEnd},
			"fail":        {"fail", "", "", workflow.SyncFuncNode, restoreFail},
		},
	})
```
### 备份数据流程

备份数据包括立即备份和自动备份，立即备份会根据用户设置的默认备份地址（S3/nfs）执行备份，自动备份会按用户设置的自动备份策略，在相应时间点触发自动备份

- 获取集群的备份账号，执行backup sql到用户配置的默认备份地址
- 更新全量备份结果，备份tso，备份大小等信息到备份的历史记录

### 恢复数据流程

- 恢复集群流程会复用创建集群流程
- 在最后一步调用backuprestore manager的恢复数据子流程
- 子流程执行restore sql

### 异常错误

备份恢复的API报错定义在common/errors/errorcode.go
``` go
	// backup && restore
	TIUNIMANAGER_BACKUP_SYSTEM_CONFIG_NOT_FOUND: {"backup system config not found", 404},
	TIUNIMANAGER_BACKUP_SYSTEM_CONFIG_INVAILD:   {"backup system config invalid", 400},
	TIUNIMANAGER_BACKUP_RECORD_CREATE_FAILED:    {"create backup record failed", 500},
	TIUNIMANAGER_BACKUP_RECORD_DELETE_FAILED:    {"delete backup record failed", 500},
	TIUNIMANAGER_BACKUP_RECORD_QUERY_FAILED:     {"query backup record failed", 500},
	TIUNIMANAGER_BACKUP_STRATEGY_SAVE_FAILED:    {"save backup strategy failed", 500},
	TIUNIMANAGER_BACKUP_STRATEGY_QUERY_FAILED:   {"query backup strategy failed", 500},
	TIUNIMANAGER_BACKUP_STRATEGY_DELETE_FAILED:  {"delete backup strategy failed", 500},
	TIUNIMANAGER_BACKUP_FILE_DELETE_FAILED:      {"remove backup file failed", 500},
	TIUNIMANAGER_BACKUP_PATH_CREATE_FAILED:      {"backup filepath create failed", 500},
	TIUNIMANAGER_BACKUP_RECORD_INVALID:          {"backup record invalid", 400},
	TIUNIMANAGER_BACKUP_RECORD_CANCEL_FAILED:    {"cancel backup record failed", 500},
```
