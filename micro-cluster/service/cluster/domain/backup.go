package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty/libbr"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"os"
	"strconv"
	"strings"
	"time"
)

type BackupRecord struct {
	Id         int64
	ClusterId  string
	BackupMethod BackupMethod
	BackupType BackupType
	BackupMode BackupMode
	StorageType StorageType
	OperatorId string
	Size       uint64
	FilePath   string
	StartTime  int64
	EndTime    int64
	BizId      uint64
}

type RecoverRecord struct {
	Id           uint
	ClusterId    string
	OperatorId   string
	BackupRecord BackupRecord
}

//var defaultPathPrefix string = "/tmp/tiem/backup"
var defaultPathPrefix string = "nfs/tiem/backup"

func Backup(ctx context.Context, ope *proto.OperatorDTO, clusterId string, backupMethod string, backupType string, backupMode BackupMode, filePath string) (*ClusterAggregation, error) {
	getLoggerWithContext(ctx).Infof("Begin do Backup, clusterId: %s, backupMethod: %s, backupType: %s, backupMode: %s, filePath: %s", clusterId, backupMethod, backupType, backupMode, filePath)
	defer getLoggerWithContext(ctx).Infof("End do Backup")
	operator := parseOperatorFromDTO(ope)
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil {
		return nil, fmt.Errorf("load cluster %s aggregation failed", clusterId)
	}

	clusterAggregation.CurrentOperator = operator
	cluster := clusterAggregation.Cluster

	flow, _ := CreateFlowWork(clusterId, FlowBackupCluster, operator)

	//todo: only support FULL Physics backup now
	record := &BackupRecord{
		ClusterId:  clusterId,
		StorageType: StorageTypeS3,
		BackupMethod: BackupMethodPhysics,
		BackupType: BackupTypeFull,
		BackupMode: backupMode,
		OperatorId: operator.Id,
		FilePath:   getBackupPath(filePath, clusterId, time.Now(), string(BackupTypeFull)),
		StartTime:  time.Now().Unix(),
	}
	resp, err := client.DBClient.SaveBackupRecord(ctx, &db.DBSaveBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			TenantId:    cluster.TenantId,
			ClusterId:   record.ClusterId,
			BackupType:  string(record.BackupType),
			BackupMethod: string(record.BackupMethod),
			BackupMode:  string(record.BackupMode),
			StorageType: string(record.StorageType),
			OperatorId:  record.OperatorId,
			FilePath:    record.FilePath,
			FlowId:      int64(flow.FlowWork.Id),
			StartTime:   time.Now().Unix(),
			EndTime:     time.Now().Unix(),
		},
	})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("save backup record failed, %s", err.Error())
		return nil, errors.New("save backup record failed")
	}
	record.Id = resp.GetBackupRecord().GetId()
	clusterAggregation.LastBackupRecord = record

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextCtxKey, ctx)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func DeleteBackup(ctx context.Context, ope *proto.OperatorDTO, clusterId string, bakId int64) error {
	getLoggerWithContext(ctx).Infof("Begin do DeleteBackup, clusterId: %s, bakId: %d", clusterId, bakId)
	defer getLoggerWithContext(ctx).Infof("End do DeleteBackup")
	//todo: parma pre check
	resp, err := client.DBClient.QueryBackupRecords(ctx, &db.DBQueryBackupRecordRequest{ClusterId: clusterId, RecordId: bakId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
		return fmt.Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
	}
	getLoggerWithContext(ctx).Infof("query backup record to be deleted, record: %v", resp.GetBackupRecords().GetBackupRecord())
	filePath := resp.GetBackupRecords().GetBackupRecord().GetFilePath() //backup dir

	err = os.RemoveAll(filePath)
	if err != nil {
		getLoggerWithContext(ctx).Errorf("remove backup filePath failed, %s", err.Error())
		return fmt.Errorf("remove backup filePath failed, %s", err.Error())
	}

	_, err = client.DBClient.DeleteBackupRecord(ctx, &db.DBDeleteBackupRecordRequest{Id: bakId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("delete metadb backup record failed, %s", err.Error())
		return fmt.Errorf("delete metadb backup record failed, %s", err.Error())
	}

	return nil
}

func RecoverPreCheck(req *proto.RecoverRequest) error {
	if req.GetCluster() == nil {
		return fmt.Errorf("invalid input cluster info")
	}
	recoverInfo := req.GetCluster().GetRecoverInfo()
	if recoverInfo.GetBackupRecordId() <= 0 || recoverInfo.GetSourceClusterId() == "" {
		return fmt.Errorf("invalid recover info param")
	}

	srcClusterArg, err := ClusterRepo.Load(recoverInfo.SourceClusterId)
	if err != nil || srcClusterArg == nil {
		return fmt.Errorf("load recover src cluster %s aggregation", recoverInfo.SourceClusterId)
	}
	operator := parseOperatorFromDTO(req.GetOperator())
	if srcClusterArg.Cluster.TenantId != operator.TenantId {
		return fmt.Errorf("recover cluster tenant %s not match source cluster tenanat %s", operator.TenantId, srcClusterArg.Cluster.TenantId)
	}

	/*
	 * todo: source cluster contains TiFlash and version < v4.0.0, new cluster for recover must contains TiFlash, otherwise it will cause recover fail
	 * https://docs.pingcap.com/zh/tidb/stable/backup-and-restore-faq
	*/

	//todo: check new cluster storage must > source cluster used storage

	return nil
}

func Recover(ctx context.Context, ope *proto.OperatorDTO, clusterInfo *proto.ClusterBaseInfoDTO, demandDTOs []*proto.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
	getLoggerWithContext(ctx).Infof("Begin do Recover, clusterInfo: %+v, demandDTOs: %+v", clusterInfo, demandDTOs)
	defer getLoggerWithContext(ctx).Infof("End do Recover")
	operator := parseOperatorFromDTO(ope)

	cluster := &Cluster{
		ClusterName:    clusterInfo.ClusterName,
		DbPassword:     clusterInfo.DbPassword,
		ClusterType:    *knowledge.ClusterTypeFromCode(clusterInfo.ClusterType.Code),
		ClusterVersion: *knowledge.ClusterVersionFromCode(clusterInfo.ClusterVersion.Code),
		Tls:            clusterInfo.Tls,
		TenantId:       operator.TenantId,
		OwnerId:        operator.Id,
	}

	demands := make([]*ClusterComponentDemand, len(demandDTOs), len(demandDTOs))

	for i, v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	cluster.Demands = demands

	// persist the cluster into database
	err := ClusterRepo.AddCluster(cluster)

	if err != nil {
		return nil, err
	}
	clusterAggregation := &ClusterAggregation{
		Cluster:          cluster,
		MaintainCronTask: GetDefaultMaintainTask(),
		CurrentOperator:  operator,
	}

	// Start the workflow to create a cluster instance

	flow, err := CreateFlowWork(cluster.Id, FlowRecoverCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextCtxKey, ctx)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func SaveBackupStrategyPreCheck(ope *proto.OperatorDTO, strategy *proto.BackupStrategy) error {
	period := strings.Split(strategy.GetPeriod(), "-")
	if len(period) != 2 {
		return fmt.Errorf("invalid param period, %s", strategy.GetPeriod())
	}

	starts := strings.Split(period[0], ":")
	ends := strings.Split(period[1], ":")
	startHour, err := strconv.Atoi(starts[0])
	if err != nil {
		return fmt.Errorf("invalid param start hour, %s", err.Error())
	}
	endHour, err := strconv.Atoi(ends[0])
	if err != nil {
		return fmt.Errorf("invalid param end hour, %s", err.Error())
	}
	if startHour > 23 || startHour < 0 || endHour > 23 || endHour < 0 || startHour >= endHour {
		return fmt.Errorf("invalid param period, %s", strategy.GetPeriod())
	}

	if strategy.GetBackupDate() != "" {
		backupDates := strings.Split(strategy.GetBackupDate(), ",")
		for _, day := range backupDates {
			if !checkWeekDayValid(day) {
				return fmt.Errorf("backupDate contains invalid weekday, %s", day)
			}
		}
	}

	return nil
}

func SaveBackupStrategy(ctx context.Context, ope *proto.OperatorDTO, strategy *proto.BackupStrategy) error {
	getLoggerWithContext(ctx).Infof("begin save backup strategy: %+v", strategy)
	defer getLoggerWithContext(ctx).Info("end save backup strategy")
	period := strings.Split(strategy.GetPeriod(), "-")
	starts := strings.Split(period[0], ":")
	ends := strings.Split(period[1], ":")
	startHour, _ := strconv.Atoi(starts[0])
	endHour, _ := strconv.Atoi(ends[0])

	_, err := client.DBClient.SaveBackupStrategy(ctx, &db.DBSaveBackupStrategyRequest{
		Strategy: &db.DBBackupStrategyDTO{
			TenantId:    ope.TenantId,
			OperatorId:  ope.GetId(),
			ClusterId:   strategy.ClusterId,
			BackupDate:  strategy.BackupDate,
			StartHour:   uint32(startHour),
			EndHour:     uint32(endHour),
		},
	})
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		return err
	}

	return nil
}

func QueryBackupStrategy(ctx context.Context, ope *proto.OperatorDTO, clusterId string) (*proto.BackupStrategy, error) {
	resp, err := client.DBClient.QueryBackupStrategy(ctx, &db.DBQueryBackupStrategyRequest{
		ClusterId: clusterId,
	})
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		return nil, err
	} else {
		strategy := &proto.BackupStrategy{
			ClusterId:      resp.GetStrategy().GetClusterId(),
			BackupDate:     resp.GetStrategy().GetBackupDate(),
			Period:         fmt.Sprintf("%d:00-%d:00", resp.GetStrategy().GetStartHour(), resp.GetStrategy().GetEndHour()),
		}
		nextBackupTime, err := calculateNextBackupTime(time.Now(), resp.GetStrategy().GetBackupDate(), int(resp.GetStrategy().GetStartHour()))
		if err != nil {
			getLoggerWithContext(ctx).Warnf("calculateNextBackupTime failed, %s", err.Error())
		} else {
			strategy.NextBackupTime = nextBackupTime.Unix()
		}
		return strategy, nil
	}
}

func calculateNextBackupTime(now time.Time, weekdayStr string, hour int) (time.Time, error) {
	days := strings.Split(weekdayStr, ",")
	if len(days) == 0 {
		return time.Time{}, fmt.Errorf("weekday invalid, %s", weekdayStr)
	}
	for _, day := range days {
		if !checkWeekDayValid(day) {
			return time.Time{}, fmt.Errorf("weekday invalid, %s", day)
		}
	}

	var subDays int = 7
	for _, day := range days {
		if WeekDayMap[day] < int(now.Weekday()) {
			if WeekDayMap[day]+7-int(now.Weekday()) < subDays {
				subDays = WeekDayMap[day] + 7 - int(now.Weekday())
			}
		} else if WeekDayMap[day] > int(now.Weekday()) {
			if WeekDayMap[day] - int(now.Weekday()) < subDays {
				subDays = WeekDayMap[day] - int(now.Weekday())
			}
		} else {
			if now.Hour() < hour {
				subDays = 0
			}
		}
	}
	nextAutoBackupTime := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local).AddDate(0, 0, subDays)

	return nextAutoBackupTime, nil
}

func getBackupPath(filePrefix string, clusterId string, time time.Time, backupRange string) string {
	if filePrefix != "" {
		//use user spec filepath
		//local://br_data/[clusterId]/16242354365-FULL/(lock/SST/metadata)
		return fmt.Sprintf("%s/%s/%s_%s", filePrefix, clusterId, time.Format("2021-01-01_01:01:01"), backupRange)
	}
	//return fmt.Sprintf("%s/%s/%d-%s", defaultPathPrefix, clusterId, timeStamp, backupRange)
	//todo: test env s3 config
	return fmt.Sprintf("%s/%s/%s_%s", defaultPathPrefix, clusterId, time.Format("2021-01-01_01:01:01"), backupRange)
}

func backupCluster(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.value(contextCtxKey).(context.Context)
	getLoggerWithContext(ctx).Info("begin backupCluster")
	defer getLoggerWithContext(ctx).Info("end backupCluster")

	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	record := clusterAggregation.LastBackupRecord
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]
	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = DefaultTidbPort
	}

	storageType, err := convertBrStorageType(string(record.StorageType))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("convert storage type failed, %s", err.Error())
		return false
	}

	clusterFacade := libbr.ClusterFacade{
		DbConnParameter: libbr.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			Ip:       tidbServer.Host,
			Port:     strconv.Itoa(tidbServerPort),
		},
		DbName:      "", //todo: support db table backup
		TableName:   "",
		ClusterId:   cluster.Id,
		ClusterName: cluster.ClusterName,
		TaskID:      uint64(task.Id),
	}
	storage := libbr.BrStorage{
		StorageType: storageType,
		Root:        fmt.Sprintf("%s/%s", record.FilePath, "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"), //todo: test env s3 ak sk
	}

	getLoggerWithContext(ctx).Infof("begin call brmgr backup api, clusterFacade[%v], storage[%v]", clusterFacade, storage)
	_, err = libbr.BackUp(clusterFacade, storage, uint64(task.Id))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("call backup api failed, %s", err.Error())
		return false
	}
	record.BizId = uint64(task.Id)

	return true
}

func updateBackupRecord(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.value(contextCtxKey).(context.Context)
	getLoggerWithContext(ctx).Info("begin updateBackupRecord")
	defer getLoggerWithContext(ctx).Info("end updateBackupRecord")

	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)
	record := clusterAggregation.LastBackupRecord

	//todo: update size
	/*
		configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
		cluster := clusterAggregation.Cluster
		tidbServer := configModel.TiDBServers[0]

		clusterFacade := libbr.ClusterFacade{
			DbConnParameter: libbr.DbConnParam{
				Username: "root", //todo: replace admin account
				Password: "",
				Ip:	tidbServer.Host,
				Port: strconv.Itoa(tidbServer.Port),
			},
			ClusterId: cluster.Id,
			ClusterName: cluster.ClusterName,
			TaskID: record.BizId,
		}
		getLogger().Infof("begin call libbr api ShowBackUpInfo, %v", clusterFacade)
		resp := libbr.ShowBackUpInfo(clusterFacade)
		record.Size = resp.Size
		getLogger().Infof("call libbr api ShowBackUpInfo resp, %v", resp)
	*/
	_, err := client.DBClient.UpdateBackupRecord(ctx, &db.DBUpdateBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			Id:      record.Id,
			Size:    record.Size,
			EndTime: time.Now().Unix(),
		},
	})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("update backup record for cluster[%s] failed, %s", clusterAggregation.Cluster.Id, err.Error())
		return false
	}
	return true
}

func recoverFromSrcCluster(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.value(contextCtxKey).(context.Context)
	getLoggerWithContext(ctx).Info("begin recoverFromSrcCluster")
	defer getLoggerWithContext(ctx).Info("end recoverFromSrcCluster")

	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	recoverInfo := cluster.RecoverInfo
	if recoverInfo.SourceClusterId == "" || recoverInfo.BackupRecordId <= 0 {
		getLoggerWithContext(ctx).Infof("cluster %s no need recover", cluster.Id)
		return true
	}

	//todo: wait start task finished, temporary solution
	var req db.FindTiupTaskByIDRequest
	req.Id = flowContext.value("startTaskId").(uint64)

	for i := 0; i < 30; i++ {
		time.Sleep(5 * time.Second)
		rsp, err := client.DBClient.FindTiupTaskByID(ctx, &req)
		if err != nil {
			getLoggerWithContext(ctx).Errorf("get start task err = %s", err.Error())
			task.Fail(err)
			return false
		}
		if rsp.TiupTask.Status == db.TiupTaskStatus_Error{
			getLoggerWithContext(ctx).Errorf("start cluster error, %s", rsp.TiupTask.ErrorStr)
			task.Fail(errors.New(rsp.TiupTask.ErrorStr))
			return false
		}
		if rsp.TiupTask.Status == db.TiupTaskStatus_Finished {
			break
		}
	}

	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	record, err := client.DBClient.QueryBackupRecords(ctx, &db.DBQueryBackupRecordRequest{ClusterId: recoverInfo.SourceClusterId, RecordId: recoverInfo.BackupRecordId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("query backup record failed, %s", err.Error())
		return false
	}

	storageType, err := convertBrStorageType(record.GetBackupRecords().GetBackupRecord().GetStorageType())
	if err != nil {
		getLoggerWithContext(ctx).Errorf("convert br storage type failed, %s", err.Error())
		return false
	}

	clusterFacade := libbr.ClusterFacade{
		DbConnParameter: libbr.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			Ip:	tidbServer.Host,
			Port: strconv.Itoa(tidbServer.Port),
		},
		DbName: "",	//todo: support db table restore
		TableName: "",
		ClusterId: cluster.Id,
		ClusterName: cluster.ClusterName,
	}
	storage := libbr.BrStorage{
		StorageType: storageType,
		Root:        fmt.Sprintf("%s/%s", record.GetBackupRecords().GetBackupRecord().GetFilePath(), "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"), //todo: test env s3 ak sk
	}
	getLoggerWithContext(ctx).Infof("begin call brmgr restore api, clusterFacade %v, storage %v", clusterFacade, storage)
	_, err = libbr.Restore(clusterFacade, storage, uint64(task.Id))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("call restore api failed, %s", err.Error())
		return false
	}
	return true
}

func convertBrStorageType(storageType string) (libbr.StorageType, error) {
	if string(StorageTypeS3) == storageType{
		return libbr.StorageTypeS3, nil
	} else if string(StorageTypeLocal) == storageType{
		return libbr.StorageTypeLocal, nil
	} else {
		return "", fmt.Errorf("invalid storage type, %s", storageType)
	}
}