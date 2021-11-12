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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-metadb/service"
	"os"
	"strconv"
	"strings"
	"time"
)

type BackupRecord struct {
	Id           int64
	ClusterId    string
	BackupMethod BackupMethod
	BackupType   BackupType
	BackupMode   BackupMode
	StorageType  StorageType
	OperatorId   string
	Size         uint64
	FilePath     string
	StartTime    int64
	EndTime      int64
}

type RecoverRecord struct {
	Id           uint
	ClusterId    string
	OperatorId   string
	BackupRecord BackupRecord
}

//var defaultPathPrefix string = "/tmp/tiem/backup"
var defaultPathPrefix string = "nfs/tiem/backup"

func Backup(ctx context.Context, ope *clusterpb.OperatorDTO, clusterId string, backupMethod string, backupType string, backupMode BackupMode, filePath string) (*ClusterAggregation, error) {
	getLoggerWithContext(ctx).Infof("Begin do Backup, clusterId: %s, backupMethod: %s, backupType: %s, backupMode: %s, filePath: %s", clusterId, backupMethod, backupType, backupMode, filePath)
	defer getLoggerWithContext(ctx).Infof("End do Backup")
	operator := parseOperatorFromDTO(ope)
	clusterAggregation, err := ClusterRepo.Load(ctx, clusterId)
	if err != nil || clusterAggregation == nil {
		return nil, fmt.Errorf("load cluster %s aggregation failed", clusterId)
	}

	clusterAggregation.CurrentOperator = operator
	cluster := clusterAggregation.Cluster

	flow, _ := CreateFlowWork(ctx, clusterId, FlowBackupCluster, operator)

	//todo: only support FULL Physics backup now
	record := &BackupRecord{
		ClusterId:    clusterId,
		StorageType:  StorageTypeS3,
		BackupMethod: BackupMethodPhysics,
		BackupType:   BackupTypeFull,
		BackupMode:   backupMode,
		OperatorId:   operator.Id,
		FilePath:     getBackupPath(filePath, clusterId, time.Now(), string(BackupTypeFull)),
		StartTime:    time.Now().Unix(),
	}
	resp, err := client.DBClient.SaveBackupRecord(ctx, &dbpb.DBSaveBackupRecordRequest{
		BackupRecord: &dbpb.DBBackupRecordDTO{
			TenantId:     cluster.TenantId,
			ClusterId:    record.ClusterId,
			BackupType:   string(record.BackupType),
			BackupMethod: string(record.BackupMethod),
			BackupMode:   string(record.BackupMode),
			StorageType:  string(record.StorageType),
			OperatorId:   record.OperatorId,
			FilePath:     record.FilePath,
			FlowId:       int64(flow.FlowWork.Id),
			StartTime:    time.Now().Unix(),
			EndTime:      time.Now().Unix(),
		},
	})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("save backup record failed, %s", err.Error())
		return nil, err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("save backup record failed, %s", resp.GetStatus().GetMessage())
		return nil, fmt.Errorf("save backup record failed, %s", resp.GetStatus().GetMessage())
	}
	record.Id = resp.GetBackupRecord().GetId()
	clusterAggregation.LastBackupRecord = record

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextCtxKey, ctx)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	err = ClusterRepo.Persist(ctx, clusterAggregation)
	if err != nil {
		return nil, err
	}
	return clusterAggregation, nil
}

func DeleteBackup(ctx context.Context, ope *clusterpb.OperatorDTO, clusterId string, bakId int64) error {
	getLoggerWithContext(ctx).Infof("Begin do DeleteBackup, clusterId: %s, bakId: %d", clusterId, bakId)
	defer getLoggerWithContext(ctx).Infof("End do DeleteBackup")
	//todo: parma pre check
	resp, err := client.DBClient.QueryBackupRecords(ctx, &dbpb.DBQueryBackupRecordRequest{ClusterId: clusterId, RecordId: bakId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
		return fmt.Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, resp.GetStatus().GetMessage())
		return fmt.Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, resp.GetStatus().GetMessage())
	}
	getLoggerWithContext(ctx).Infof("query backup record to be deleted, record: %v", resp.GetBackupRecords().GetBackupRecord())
	filePath := resp.GetBackupRecords().GetBackupRecord().GetFilePath() //backup dir

	err = os.RemoveAll(filePath)
	if err != nil {
		getLoggerWithContext(ctx).Errorf("remove backup filePath failed, %s", err.Error())
		return fmt.Errorf("remove backup filePath failed, %s", err.Error())
	}

	delResp, err := client.DBClient.DeleteBackupRecord(ctx, &dbpb.DBDeleteBackupRecordRequest{Id: bakId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("delete metadb backup record failed, %s", err.Error())
		return fmt.Errorf("delete metadb backup record failed, %s", err.Error())
	}
	if delResp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("delete metadb backup record failed, %s", delResp.GetStatus().GetMessage())
		return fmt.Errorf("delete metadb backup record failed, %s", delResp.GetStatus().GetMessage())
	}

	return nil
}

func RecoverPreCheck(ctx context.Context, req *clusterpb.RecoverRequest) error {
	if req.GetCluster() == nil {
		return fmt.Errorf("invalid input cluster info")
	}
	recoverInfo := req.GetCluster().GetRecoverInfo()
	if recoverInfo.GetBackupRecordId() <= 0 || recoverInfo.GetSourceClusterId() == "" {
		return fmt.Errorf("invalid recover info param")
	}

	srcClusterArg, err := ClusterRepo.Load(ctx, recoverInfo.SourceClusterId)
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

func Recover(ctx context.Context, ope *clusterpb.OperatorDTO, clusterInfo *clusterpb.ClusterBaseInfoDTO, demandDTOs []*clusterpb.ClusterNodeDemandDTO) (*ClusterAggregation, error) {
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
		RecoverInfo: RecoverInfo{
			SourceClusterId: clusterInfo.GetRecoverInfo().GetSourceClusterId(),
			BackupRecordId:  clusterInfo.GetRecoverInfo().GetBackupRecordId(),
		},
	}

	demands := make([]*ClusterComponentDemand, len(demandDTOs))

	for i, v := range demandDTOs {
		demands[i] = parseNodeDemandFromDTO(v)
	}

	cluster.Demands = demands

	// persist the cluster into database
	err := ClusterRepo.AddCluster(ctx, cluster)

	if err != nil {
		return nil, err
	}
	clusterAggregation := &ClusterAggregation{
		Cluster:          cluster,
		MaintainCronTask: GetDefaultMaintainTask(),
		CurrentOperator:  operator,
	}

	// Start the workflow to create a cluster instance

	flow, err := CreateFlowWork(ctx, cluster.Id, FlowRecoverCluster, operator)
	if err != nil {
		return nil, err
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextCtxKey, ctx)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	err = ClusterRepo.Persist(ctx, clusterAggregation)
	if err != nil {
		return nil, err
	}
	return clusterAggregation, nil
}

func SaveBackupStrategyPreCheck(ope *clusterpb.OperatorDTO, strategy *clusterpb.BackupStrategy) error {
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

func SaveBackupStrategy(ctx context.Context, ope *clusterpb.OperatorDTO, strategy *clusterpb.BackupStrategy) error {
	getLoggerWithContext(ctx).Infof("begin save backup strategy: %+v", strategy)
	defer getLoggerWithContext(ctx).Info("end save backup strategy")
	period := strings.Split(strategy.GetPeriod(), "-")
	starts := strings.Split(period[0], ":")
	ends := strings.Split(period[1], ":")
	startHour, _ := strconv.Atoi(starts[0])
	endHour, _ := strconv.Atoi(ends[0])

	resp, err := client.DBClient.SaveBackupStrategy(ctx, &dbpb.DBSaveBackupStrategyRequest{
		Strategy: &dbpb.DBBackupStrategyDTO{
			TenantId:   ope.TenantId,
			OperatorId: ope.GetId(),
			ClusterId:  strategy.ClusterId,
			BackupDate: strategy.BackupDate,
			StartHour:  uint32(startHour),
			EndHour:    uint32(endHour),
		},
	})
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		return err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Error(resp.GetStatus().GetMessage())
		return fmt.Errorf("save backup strategy failed %s", resp.GetStatus().GetMessage())
	}

	return nil
}

func QueryBackupStrategy(ctx context.Context, ope *clusterpb.OperatorDTO, clusterId string) (*clusterpb.BackupStrategy, error) {
	resp, err := client.DBClient.QueryBackupStrategy(ctx, &dbpb.DBQueryBackupStrategyRequest{
		ClusterId: clusterId,
	})
	if err != nil {
		getLoggerWithContext(ctx).Error(err)
		return nil, err
	}
	if resp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Error(resp.GetStatus().GetMessage())
		return nil, fmt.Errorf("query backup strategy failed %s", resp.GetStatus().GetMessage())
	}

	strategy := &clusterpb.BackupStrategy{
		ClusterId:  resp.GetStrategy().GetClusterId(),
		BackupDate: resp.GetStrategy().GetBackupDate(),
		Period:     fmt.Sprintf("%d:00-%d:00", resp.GetStrategy().GetStartHour(), resp.GetStrategy().GetEndHour()),
	}
	nextBackupTime, err := calculateNextBackupTime(time.Now(), resp.GetStrategy().GetBackupDate(), int(resp.GetStrategy().GetStartHour()))
	if err != nil {
		getLoggerWithContext(ctx).Warnf("calculateNextBackupTime failed, %s", err.Error())
	} else {
		strategy.NextBackupTime = nextBackupTime.Unix()
	}
	return strategy, nil
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
			if WeekDayMap[day]-int(now.Weekday()) < subDays {
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
		return fmt.Sprintf("%s/%s/%s_%s", filePrefix, clusterId, time.Format("2006-01-02_15:04:05"), backupRange)
	}
	//return fmt.Sprintf("%s/%s/%d-%s", defaultPathPrefix, clusterId, timeStamp, backupRange)
	//todo: test env s3 config
	return fmt.Sprintf("%s/%s/%s_%s", defaultPathPrefix, clusterId, time.Format("2006-01-02_15:04:05"), backupRange)
}

func backupCluster(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.GetData(contextCtxKey).(context.Context)
	getLoggerWithContext(ctx).Info("begin backupCluster")
	defer getLoggerWithContext(ctx).Info("end backupCluster")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
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
		task.Fail(err)
		return false
	}

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
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
	storage := secondparty.BrStorage{
		StorageType: storageType,
		Root:        fmt.Sprintf("%s/%s", record.FilePath, "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"), //todo: test env s3 ak sk
	}

	getLoggerWithContext(ctx).Infof("begin call brmgr backup api, clusterFacade[%v], storage[%v]", clusterFacade, storage)
	_, err = secondparty.SecondParty.MicroSrvBackUp(flowContext.Context, clusterFacade, storage, uint64(task.Id))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("call backup api failed, %s", err.Error())
		task.Fail(err)
		return false
	}

	return true
}

func updateBackupRecord(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.GetData(contextCtxKey).(context.Context)
	getLoggerWithContext(ctx).Info("begin updateBackupRecord")
	defer getLoggerWithContext(ctx).Info("end updateBackupRecord")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	record := clusterAggregation.LastBackupRecord

	var req dbpb.FindTiupTaskByIDRequest
	var resp *dbpb.FindTiupTaskByIDResponse
	var err error
	req.Id = flowContext.GetData("backupTaskId").(uint64)

	resp, err = client.DBClient.FindTiupTaskByID(flowContext, &req)
	if err != nil {
		getLoggerWithContext(ctx).Errorf("get backup task err = %s", err.Error())
		task.Fail(err)
		return false
	}
	if resp.TiupTask.Status == dbpb.TiupTaskStatus_Error {
		getLoggerWithContext(ctx).Errorf("backup cluster error, %s", resp.TiupTask.ErrorStr)
		task.Fail(errors.New(resp.TiupTask.ErrorStr))
		return false
	}

	var backupInfo secondparty.CmdBrResp
	err = json.Unmarshal([]byte(resp.GetTiupTask().GetErrorStr()), &backupInfo)
	if err != nil {
		getLoggerWithContext(ctx).Errorf("json unmarshal backup info resp: %+v, failed, %s", resp, err.Error())
	} else {
		record.Size = backupInfo.Size
	}

	updateResp, err := client.DBClient.UpdateBackupRecord(flowContext, &dbpb.DBUpdateBackupRecordRequest{
		BackupRecord: &dbpb.DBBackupRecordDTO{
			Id:      record.Id,
			Size:    record.Size,
			EndTime: time.Now().Unix(),
		},
	})
	if err != nil {
		msg := fmt.Sprintf("update backup record for cluster %s failed", clusterAggregation.Cluster.Id)
		tiemError := framework.WrapError(common.TIEM_METADB_SERVER_CALL_ERROR, msg, err)
		getLoggerWithContext(ctx).Error(tiemError)
		task.Fail(tiemError)
		return false
	}
	if updateResp.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		msg := fmt.Sprintf("update backup record for cluster %s failed, %s", clusterAggregation.Cluster.Id, updateResp.Status.Message)
		tiemError := framework.NewTiEMError(common.TIEM_ERROR_CODE(updateResp.GetStatus().GetCode()), msg)

		getLoggerWithContext(ctx).Error(tiemError)
		task.Fail(tiemError)
		return false
	}
	return true
}

func recoverFromSrcCluster(task *TaskEntity, flowContext *FlowContext) bool {
	ctx := flowContext.GetData(contextCtxKey).(context.Context)
	getLoggerWithContext(ctx).Info("begin recoverFromSrcCluster")
	defer getLoggerWithContext(ctx).Info("end recoverFromSrcCluster")

	clusterAggregation := flowContext.GetData(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	recoverInfo := cluster.RecoverInfo
	if recoverInfo.SourceClusterId == "" || recoverInfo.BackupRecordId <= 0 {
		getLoggerWithContext(ctx).Infof("cluster %s no need recover", cluster.Id)
		task.Success(nil)
		return true
	}

	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	record, err := client.DBClient.QueryBackupRecords(flowContext, &dbpb.DBQueryBackupRecordRequest{ClusterId: recoverInfo.SourceClusterId, RecordId: recoverInfo.BackupRecordId})
	if err != nil {
		getLoggerWithContext(ctx).Errorf("query backup record failed, %s", err.Error())
		task.Fail(fmt.Errorf("query backup record failed, %s", err.Error()))
		return false
	}
	if record.GetStatus().GetCode() != service.ClusterSuccessResponseStatus.GetCode() {
		getLoggerWithContext(ctx).Errorf("query backup record failed, %s", record.GetStatus().GetMessage())
		task.Fail(fmt.Errorf("query backup record failed, %s", record.GetStatus().GetMessage()))
		return false
	}

	storageType, err := convertBrStorageType(record.GetBackupRecords().GetBackupRecord().GetStorageType())
	if err != nil {
		getLoggerWithContext(ctx).Errorf("convert br storage type failed, %s", err.Error())
		task.Fail(fmt.Errorf("convert br storage type failed, %s", err.Error()))
		return false
	}

	clusterFacade := secondparty.ClusterFacade{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			Ip:       tidbServer.Host,
			Port:     strconv.Itoa(tidbServer.Port),
		},
		DbName:      "", //todo: support db table restore
		TableName:   "",
		ClusterId:   cluster.Id,
		ClusterName: cluster.ClusterName,
	}
	storage := secondparty.BrStorage{
		StorageType: storageType,
		Root:        fmt.Sprintf("%s/%s", record.GetBackupRecords().GetBackupRecord().GetFilePath(), "?access-key=minioadmin\\&secret-access-key=minioadmin\\&endpoint=http://minio.pingcap.net:9000\\&force-path-style=true"), //todo: test env s3 ak sk
	}
	getLoggerWithContext(ctx).Infof("begin call brmgr restore api, clusterFacade %v, storage %v", clusterFacade, storage)
	_, err = secondparty.SecondParty.MicroSrvRestore(flowContext.Context, clusterFacade, storage, uint64(task.Id))
	if err != nil {
		getLoggerWithContext(ctx).Errorf("call restore api failed, %s", err.Error())
		task.Fail(err)
		return false
	}
	return true
}

func convertBrStorageType(storageType string) (secondparty.StorageType, error) {
	if string(StorageTypeS3) == storageType {
		return secondparty.StorageTypeS3, nil
	} else if string(StorageTypeLocal) == storageType {
		return secondparty.StorageTypeLocal, nil
	} else {
		return "", fmt.Errorf("invalid storage type, %s", storageType)
	}
}
