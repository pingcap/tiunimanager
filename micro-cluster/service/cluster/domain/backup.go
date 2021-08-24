package domain

import (
	"context"
	"errors"
	"fmt"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/pingcap-inc/tiem/library/firstparty/client"
	"github.com/pingcap-inc/tiem/library/secondparty/libbr"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"golang.org/x/crypto/ssh"
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
	operator := parseOperatorFromDTO(ope)
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil {
		return nil, errors.New("load cluster aggregation")
	}
	clusterAggregation.CurrentOperator = operator

	record := &BackupRecord {
		ClusterId: clusterId,
		Range: BackupRange(backupRange),
		BackupType: BackupType(backupType),
		OperatorId: operator.Id,
		FilePath: getBackupPath(filePath, clusterId, time.Now().Unix(), backupRange),
		StartTime: time.Now().Unix(),
	}
	clusterAggregation.LastBackupRecord = record

	flow, _ := CreateFlowWork(clusterId, FlowBackupCluster)
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
}

func DeleteBackup(ope *proto.OperatorDTO, clusterId string, bakId int64) error {
	//todo: parma pre check
	resp, err := client.DBClient.QueryBackupRecords(context.TODO(), &db.DBQueryBackupRecordRequest{ClusterId: clusterId, RecordId: bakId})
	if err != nil {
		log.Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
		return fmt.Errorf("query backup record %d of cluster %s failed, %s", bakId, clusterId, err.Error())
	}
	log.Infof("query backup record to be deleted, record: %v", resp.GetBackupRecords().GetBackupRecord())
	filePath := resp.GetBackupRecords().GetBackupRecord().GetFilePath() //backup dir

	err = os.RemoveAll(filePath)
	if err != nil {
		log.Errorf("remove backup filePath failed, %s", err.Error())
		return fmt.Errorf("remove backup filePath failed, %s", err.Error())
	}

	_, err = client.DBClient.DeleteBackupRecord(context.TODO(), &db.DBDeleteBackupRecordRequest{Id: bakId})
	if err != nil{
		log.Errorf("delete metadb backup record failed, %s", err.Error())
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
	log.Info("begin backupCluster")
	defer log.Info("end backupCluster")

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

	log.Infof("begin call brmgr backup api, clusterFacade[%v], storage[%v]", clusterFacade, storage)
	_, err := libbr.BackUp(clusterFacade, storage, uint64(task.Id))
	if err != nil {
		log.Errorf("call backup api failed, %s", err.Error())
		return false
	}
	record.BizId = uint64(task.Id)
	return true
}

func updateBackupRecord(task *TaskEntity, flowContext *FlowContext) bool {
	log.Info("begin updateBackupRecord")
	defer log.Info("end updateBackupRecord")
	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)
	record := clusterAggregation.LastBackupRecord
	/*
	//todo: update size
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
	resp := libbr.ShowBackUpInfo(clusterFacade, uint64(task.Id))
	record.Size = resp.Size
	*/
	_, err :=  client.DBClient.UpdateBackupRecord(context.TODO(), &db.DBUpdateBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			Id: record.Id,
			Size: record.Size,
		},
	})
	if err != nil {
		log.Errorf("update backup record for cluster[%s] failed, %s", clusterAggregation.Cluster.Id, err.Error())
		return false
	}
	return true
}

func recoverFromSrcCluster(task *TaskEntity, flowContext *FlowContext) bool {
	log.Info("begin recoverFromSrcCluster")
	defer log.Info("end recoverFromSrcCluster")

	clusterAggregation := flowContext.value(contextClusterKey).(*ClusterAggregation)
	cluster := clusterAggregation.Cluster
	recoverInfo := cluster.RecoverInfo
	if recoverInfo.SourceClusterId == "" || recoverInfo.BackupRecordId == 0 {
		log.Infof("cluster[%s] no need recover", cluster.Id)
		return true
	}

	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	record, err := client.DBClient.QueryBackupRecords(context.TODO(), &db.DBQueryBackupRecordRequest{ClusterId: recoverInfo.SourceClusterId, RecordId: recoverInfo.BackupRecordId})
	if err != nil {
		log.Errorf("query backup record failed, %s", err.Error())
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
	log.Infof("begin call brmgr restore api, clusterFacade[%v], storage[%v]", clusterFacade, storage)
	_, err = libbr.Restore(clusterFacade, storage, uint64(task.Id))
	if err != nil {
		log.Errorf("call restore api failed, %s", err.Error())
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
			log.Errorf("scp backup files from tikv server[%s] failed, %s", tikv.Host, err.Error())
			return false
		}
	}

	return true
}

func scpTikvBackupFiles(tikv *spec.TiKVSpec, tikvDir, backupDir string, index int) error {
	sshConfig, err := auth.PrivateKey("root", "/root/.ssh/id_rsa_tiup_test", ssh.InsecureIgnoreHostKey())
	if err != nil {
		log.Errorf("ssh auth remote host failed, %s", err.Error())
		return err
	}

	scpClient := scp.NewClient(fmt.Sprintf("%s:22", tikv.Host), &sshConfig)
	err = scpClient.Connect()
	if err != nil {
		log.Error("scp client connect tikv server[%s] failed, %s", tikv.Host, err.Error())
		return err
	}
	defer scpClient.Session.Close()

	backupFile, err := os.Open(fmt.Sprintf("%s/tikv-%s.zip", backupDir, index))
	if err != nil {
		log.Error("open backup file failed, %s", err.Error())
		return err
	}
	defer backupFile.Close()

	err = scpClient.CopyFromRemote(backupFile, fmt.Sprintf("%s/tikv.zip", tikvDir))
	if err != nil {
		log.Error("copy backup file from tikv server[%s] failed, %s", tikv.Host, err.Error())
		return err
	}
	return nil
}

func cleanLocalTikvBackupDir() error {
	return nil
}

func recoverCluster(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}
