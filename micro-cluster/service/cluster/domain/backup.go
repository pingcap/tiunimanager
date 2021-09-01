package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/secondparty/libbr"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"os"
	"strconv"
	"time"
)

type BackupRecord struct {
	Id         int64
	ClusterId  string
	Range      BackupRange
	BackupType BackupType
	OperatorId string
	Size       uint64
	FilePath   string
	StartTime  int64
	EndTime    int64
	BizId 	   uint64
}

type RecoverRecord struct {
	Id         uint
	ClusterId  string
	OperatorId string
	BackupRecord   BackupRecord
}

type BackupStrategy struct {
	ValidityPeriod  	int64
	CronString 			string
}

var defaultPathPrefix string = "/tmp/tiem/backup"

func BackupPreCheck(request *proto.CreateBackupRequest) error {
	if !checkBackupRangeValid(request.GetBackupRange()) {
		return errors.New("backupRange invalid")
	}
	if !checkBackupTypeValid(request.GetBackupType()) {
		return errors.New("backupType invalid")
	}
	if request.GetClusterId() == "" {
		return errors.New("clusterId invalid")
	}
	if request.GetFilePath() == "" {
		return errors.New("filePath invalid")
	}
	//todo: check operator is valid, maybe some has default config

	return nil
}

func Backup(ope *proto.OperatorDTO, clusterId string, backupRange string, backupType string, filePath string) (*ClusterAggregation, error){
	getLogger().Infof("Begin do Backup, clusterId: %s, backupRange: %s, backupType: %s, filePath: %s", clusterId, backupRange, backupType, filePath)
	defer getLogger().Infof("End do Backup")
	operator := parseOperatorFromDTO(ope)
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil {
		return nil, errors.New("load cluster aggregation")
	}
	clusterAggregation.CurrentOperator = operator
	cluster := clusterAggregation.Cluster

	flow, _ := CreateFlowWork(clusterId, FlowBackupCluster)

	//todo: only support FULL Physics backup now
	record := &BackupRecord {
		ClusterId: clusterId,
		Range: BackupRangeFull,
		BackupType: BackupTypePhysics,
		OperatorId: operator.Id,
		FilePath: getBackupPath(filePath, clusterId, time.Now().Unix(), string(BackupRangeFull)),
		StartTime: time.Now().Unix(),
	}
	clusterAggregation.LastBackupRecord = record
	_, err =  client.DBClient.SaveBackupRecord(context.TODO(), &db.DBSaveBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			TenantId:    cluster.TenantId,
			ClusterId:   record.ClusterId,
			BackupType: string(record.BackupType),
			BackupRange: string(record.Range),
			OperatorId:  record.OperatorId,
			FilePath:    record.FilePath,
			FlowId:      int64(flow.FlowWork.Id),
		},
	})
	if err != nil {
		getLogger().Errorf("save backup record failed, %s", err.Error())
		return nil, errors.New("save backup record failed")
	}

	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func DeleteBackup(ope *proto.OperatorDTO, clusterId string, bakId int64) error {
	getLogger().Infof("Begin do DeleteBackup, clusterId: %s, bakId: %d", clusterId, bakId)
	defer getLogger().Infof("End do DeleteBackup")
	//todo: parma pre check
	resp, err := client.DBClient.QueryBackupRecords(context.TODO(), &db.DBQueryBackupRecordRequest{ClusterId: clusterId, RecordId: bakId})
	if err != nil {
		getLogger().Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
		return fmt.Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
	}
	getLogger().Infof("query backup record to be deleted, record: %v", resp.GetBackupRecords().GetBackupRecord())
	filePath := resp.GetBackupRecords().GetBackupRecord().GetFilePath() //backup dir

	err = os.RemoveAll(filePath)
	if err != nil {
		getLogger().Errorf("remove backup filePath failed, %s", err.Error())
		return fmt.Errorf("remove backup filePath failed, %s", err.Error())
	}

	_, err = client.DBClient.DeleteBackupRecord(context.TODO(), &db.DBDeleteBackupRecordRequest{Id: bakId})
	if err != nil{
		getLogger().Errorf("delete metadb backup record failed, %s", err.Error())
		return fmt.Errorf("delete metadb backup record failed, %s", err.Error())
	}

	return nil
}

func Recover(ope *proto.OperatorDTO, clusterId string, backupRecordId int64) (*ClusterAggregation, error){
	operator := parseOperatorFromDTO(ope)

	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil {
		return nil, errors.New("load cluster aggregation")
	}
	clusterAggregation.CurrentOperator = operator
	clusterAggregation.LastRecoverRecord = &RecoverRecord{
		ClusterId: clusterId,
		OperatorId: operator.Id,
		BackupRecord: BackupRecord{Id: backupRecordId},
	}

	//currentFlow := clusterAggregation.CurrentWorkFlow
	//if currentFlow != nil && !currentFlow.Finished(){
	//	return clusterAggregation, errors.New("incomplete processing flow")
	//}

	flow, err := CreateFlowWork(clusterId, FlowRecoverCluster)
	if err != nil {
		// todo
	}

	flow.AddContext(contextClusterKey, clusterAggregation)

	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func getBackupPath(filePrefix string, clusterId string, timeStamp int64, backupRange string) string {
	if filePrefix != "" {
		//use user spec filepath
		//local://br_data/[clusterId]/16242354365-FULL/(lock/SST/metadata)
		return fmt.Sprintf("%s/%s/%d-%s", filePrefix, clusterId, timeStamp, backupRange)
	}
	return fmt.Sprintf("%s/%s/%d-%s", defaultPathPrefix, clusterId, timeStamp, backupRange)
}

func backupCluster(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin backupCluster")
	defer getLogger().Info("end backupCluster")

	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	record := clusterAggregation.LastBackupRecord
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	clusterFacade := libbr.ClusterFacade{
		DbConnParameter: libbr.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			Ip:	tidbServer.Host,
			Port: strconv.Itoa(tidbServer.Port),
		},
		DbName: "",	//todo: support db table backup
		TableName: "",
		ClusterId: cluster.Id,
		ClusterName: cluster.ClusterName,
		TaskID: uint64(task.Id),
	}
	storage := libbr.BrStorage{
		StorageType: libbr.StorageTypeLocal,
		Root: record.FilePath,
	}

	getLogger().Infof("begin call brmgr backup api, clusterFacade[%v], storage[%v]", clusterFacade, storage)
	_, err := libbr.BackUp(clusterFacade, storage, uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call backup api failed, %s", err.Error())
		return false
	}
	record.BizId = uint64(task.Id)
	return true
}

func updateBackupRecord(task *TaskEntity, flowContext *FlowContext) bool {
	getLogger().Info("begin updateBackupRecord")
	defer getLogger().Info("end updateBackupRecord")
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
	_, err :=  client.DBClient.UpdateBackupRecord(context.TODO(), &db.DBUpdateBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			Id: record.Id,
			Size: record.Size,
			EndTime: time.Now().Unix(),
		},
	})
	if err != nil {
		getLogger().Errorf("update backup record for cluster[%s] failed, %s", clusterAggregation.Cluster.Id, err.Error())
		return false
	}
	return true
}
/*
func recoverFromSrcCluster(task *TaskEntity, flowContext *FlowContext) bool {
	getLogger().Info("begin recoverFromSrcCluster")
	defer getLogger().Info("end recoverFromSrcCluster")

	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	recoverInfo := cluster.RecoverInfo
	if recoverInfo.SourceClusterId == "" || recoverInfo.BackupRecordId == 0 {
		getLogger().Infof("cluster[%s] no need recover", cluster.Id)
		return true
	}

	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	record, err := client.DBClient.QueryBackupRecords(context.TODO(), &db.DBQueryBackupRecordRequest{ClusterId: recoverInfo.SourceClusterId, RecordId: recoverInfo.BackupRecordId})
	if err != nil {
		getLogger().Errorf("query backup record failed, %s", err.Error())
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
		StorageType: libbr.StorageTypeLocal,
		Root: record.GetBackupRecords().GetBackupRecord().GetFilePath(),
	}
	getLogger().Infof("begin call brmgr restore api, clusterFacade[%v], storage[%v]", clusterFacade, storage)
	_, err = libbr.Restore(clusterFacade, storage, uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call restore api failed, %s", err.Error())
		return false
	}
	return true
}

//for no nfs storage
func mergeBackupFiles(task *TaskEntity, context *FlowContext) bool {
	//copy file from tikv server to backup dir
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	record := clusterAggregation.LastBackupRecord
	tikvDir := record.FilePath
	backupDir := record.FilePath

	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	tikvServers := configModel.TiKVServers

	for index, tikv := range tikvServers {
		err := scpTikvBackupFiles(tikv, tikvDir, backupDir, index)
		if err != nil {
			getLogger().Errorf("scp backup files from tikv server[%s] failed, %s", tikv.Host, err.Error())
			return false
		}
	}

	return true
}

func scpTikvBackupFiles(tikv *spec.TiKVSpec, tikvDir, backupDir string, index int) error {
	sshConfig, err := auth.PrivateKey("root", "/root/.ssh/id_rsa_tiup_test", ssh.InsecureIgnoreHostKey())
	if err != nil {
		getLogger().Errorf("ssh auth remote host failed, %s", err.Error())
		return err
	}

	scpClient := scp.NewClient(fmt.Sprintf("%s:22", tikv.Host), &sshConfig)
	err = scpClient.Connect()
	if err != nil {
		getLogger().Errorf("scp client connect tikv server[%s] failed, %s", tikv.Host, err.Error())
		return err
	}
	defer scpClient.Session.Close()

	backupFile, err := os.Open(fmt.Sprintf("%s/tikv-%d.zip", backupDir, index))
	if err != nil {
		getLogger().Errorf("open backup file failed, %s", err.Error())
		return err
	}
	defer backupFile.Close()

	err = scpClient.CopyFromRemote(backupFile, fmt.Sprintf("%s/tikv.zip", tikvDir))
	if err != nil {
		getLogger().Errorf("copy backup file from tikv server[%s] failed, %s", tikv.Host, err.Error())
		return err
	}
	return nil
}

func cleanLocalTikvBackupDir() error {
	return nil
}
*/

func recoverCluster(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}
