package domain

import (
	"context"
	"errors"
	"fmt"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/pingcap/tiem/library/secondparty/libbr"
	proto "github.com/pingcap/tiem/micro-cluster/proto"
	"github.com/pingcap/tiem/micro-metadb/client"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"os"
	"strconv"
)

type BackupRecord struct {
	Id         int64
	ClusterId  string
	Range      BackupRange
	BackupType BackupType
	OperatorId string
	Size       float32
	FilePath   string
	Status 	   string
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

var backupPathPrefix string = "/tmp/tiem/backup"
var localBackupPathPrefix string = "/tmp/tiem/backup"

const (
	BackupStatusSuccess string = "Success"
	BackupStatusFailed string = "Failed"
	BackupStatusRunning string = "Running"
)

func Backup(ope *proto.OperatorDTO, clusterId string, backupRange string, backupType string) (*ClusterAggregation, error){
	operator := parseOperatorFromDTO(ope)
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil {
		return nil, errors.New("load cluster aggregation")
	}
	clusterAggregation.CurrentOperator = operator
	cluster := clusterAggregation.Cluster

	record := &BackupRecord {
		ClusterId: clusterId,
		Range: BackupRange(backupRange),
		BackupType: BackupType(backupType),
		OperatorId: operator.Id,
		Size: 0,
		Status: BackupStatusRunning,
	}
	resp, err :=  client.DBClient.SaveBackupRecord(context.TODO(), &db.DBSaveBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			TenantId:    cluster.TenantId,
			ClusterId:   record.ClusterId,
			BackupType:  string(record.BackupType),
			BackupRange: string(record.Range),
			OperatorId:  record.OperatorId,
			FilePath:    record.FilePath,
			//FlowId:      int64(clusterAggregation.CurrentWorkFlow.Id),
			Status:		 record.Status,
		},
	})
	if err != nil {
		// todo
		return clusterAggregation, errors.New("create backup record failed")
	}
	record.Id = resp.GetBackupRecord().GetId()
	record.FilePath = getBackupPath(clusterId, record.Id)
	clusterAggregation.LastBackupRecord = record

	flow, _ := CreateFlowWork(clusterId, FlowBackupCluster)
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.Start()

	clusterAggregation.updateWorkFlow(flow.FlowWork)
	ClusterRepo.Persist(clusterAggregation)
	return clusterAggregation, nil
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

func getBackupPath(clusterId string, backupId int64) string {
	return fmt.Sprintf("%s/%s/%s", backupPathPrefix, clusterId, backupId)
}

func getLocalBackupPath(clusterId string, backupId int64) string {
	return fmt.Sprintf("%s/%s/%s", localBackupPathPrefix, clusterId, backupId)
}

func backupCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
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
	}
	storage := libbr.BrStorage{
		StorageType: libbr.StorageTypeLocal,
		Root: getLocalBackupPath(cluster.Id, record.Id),	//local backup dir
	}
	err := libbr.BackUp(clusterFacade, storage, task.Id)
	if err != nil {
		log.Errorf("call backup api failed, %s", err.Error())
		return false
	}
	return true
}

//for no nfs storage
func mergeBackupFiles(task *TaskEntity, context *FlowContext) bool {
	//copy file from tikv server to backup dir
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	cluster := clusterAggregation.Cluster
	record := clusterAggregation.LastBackupRecord
	tikvDir := getLocalBackupPath(cluster.Id, record.Id)
	backupDir := getBackupPath(cluster.Id, record.Id)

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

func backupSuccess(task *TaskEntity, flowContext *FlowContext) bool {
	clusterAggregation := flowContext.value(contextClusterKey).(ClusterAggregation)
	record := clusterAggregation.LastBackupRecord
	record.Status = BackupStatusSuccess
	_, err :=  client.DBClient.UpdateBackupRecord(context.TODO(), &db.DBUpdateBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			Id:			 record.Id,
			Status:		 BackupStatusSuccess,
			Size:		 0,//todo: update real size of backup file
		},
	})
	if err != nil {
		log.Errorf("update backup record failed. %s", err.Error())
		return false
	}
	return DefaultEnd(task, flowContext)
}

func backupFail(task *TaskEntity, flowContext *FlowContext) bool {
	clusterAggregation := flowContext.value(contextClusterKey).(ClusterAggregation)
	record := clusterAggregation.LastBackupRecord
	record.Status = BackupStatusSuccess
	_, err :=  client.DBClient.UpdateBackupRecord(context.TODO(), &db.DBUpdateBackupRecordRequest{
		BackupRecord: &db.DBBackupRecordDTO{
			Id:			 record.Id,
			Status:		 BackupStatusFailed,
		},
	})
	if err != nil {
		log.Errorf("update backup record failed. %s", err.Error())
		return false
	}
	return DefaultFail(task, flowContext)
}

func recoverCluster(task *TaskEntity, context *FlowContext) bool {
	task.Success(nil)
	return true
}