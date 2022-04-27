# TiEM backup and restore design doc

- Author(s): [cchenkey](http://github.com/cchenkey)

## table of Contents

- [TiEM backup and restore design doc](#TiEM backup and restore design doc)
  - [table of Contents](#Table of Contents)
  - [backup and restore API](#backup and restore API)
  - [design](#design)
    - [interface](#interface)
    - [workflow](#workflow)
    - [backup process](#backup process)
    - [restore process](#restore process)
    - [errors](#errors)

## backup and restore API
micro-api/route/route.go，"/clusters" and "/backups" group define backup and restore API,
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
The `HandleFunc` of each route can be tracked as the entry of each API，
- `backuprestore.GetBackupStrategy` get backup strategy of one cluster；
- `backuprestore.SaveBackupStrategy` save backup strategy of one cluster；
- `backuprestore.Backup` do backup immediately；
- `backuprestore.CancelBackup` cancel running backup；
- `backuprestore.QueryBackupRecords` query backup records of one cluster；
- `backuprestore.DeleteBackup` delete backup record and backup files；


## design

backupRestoreManager is manager of backup and restore（micro-cluster/cluster/backuprestore），each file function：

- `service.go` define backup and restore interface；
- `manger.go` implement of interface；
- `excutor.go` implement of backup and restore workflow nodes；
- `autobackup.go` implement of auto backup；
- `common.go` constant definition of backup and restore；

### interface
interface of backup and restore：
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
workflow definition of backup and restore：
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
### backup process

backup data include manual backup and auto backup, manual backup will do backup to user config default backup address(S3/nfs),auto backup will do backup at backup strategy by user setting

- get backup db account of cluster，then do backup sql to save backup files to user config default backup address
- update backup result, tso and backup file size to backup record

### restore process

- restore new cluster will reuse the process of create cluster
- at last workflow node of restore process, it will call backuprestore manager api to restore one backup record to the cluster
- the restoreExistCluster will do restore sql to cluster

### errors

backup and restore API errors define in file common/errors/errorcode.go
``` go
	// backup && restore
	TIEM_BACKUP_SYSTEM_CONFIG_NOT_FOUND: {"backup system config not found", 404},
	TIEM_BACKUP_SYSTEM_CONFIG_INVAILD:   {"backup system config invalid", 400},
	TIEM_BACKUP_RECORD_CREATE_FAILED:    {"create backup record failed", 500},
	TIEM_BACKUP_RECORD_DELETE_FAILED:    {"delete backup record failed", 500},
	TIEM_BACKUP_RECORD_QUERY_FAILED:     {"query backup record failed", 500},
	TIEM_BACKUP_STRATEGY_SAVE_FAILED:    {"save backup strategy failed", 500},
	TIEM_BACKUP_STRATEGY_QUERY_FAILED:   {"query backup strategy failed", 500},
	TIEM_BACKUP_STRATEGY_DELETE_FAILED:  {"delete backup strategy failed", 500},
	TIEM_BACKUP_FILE_DELETE_FAILED:      {"remove backup file failed", 500},
	TIEM_BACKUP_PATH_CREATE_FAILED:      {"backup filepath create failed", 500},
	TIEM_BACKUP_RECORD_INVALID:          {"backup record invalid", 400},
	TIEM_BACKUP_RECORD_CANCEL_FAILED:    {"cancel backup record failed", 500},
```
