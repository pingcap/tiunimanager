package models

import (
	"errors"
	dbPb "github.com/pingcap/tiem/micro-metadb/proto"
	"gorm.io/gorm"
	"time"
)

type ClusterDO struct {
	Entity
	ClusterName 			string
	DbPassword 				string
	ClusterType 			string
	ClusterVersion 			string
	Tls 					bool
	Tags           			string
	OwnerId 				string		`gorm:"not null;type:varchar(36);default:null"`
	CurrentTiupConfigId     uint
	CurrentDemandId 		uint
	CurrentFlowId			uint
}

func (d ClusterDO) TableName() string {
	return "clusters"
}

type DemandRecordDO struct {
	Record
	ClusterId 			string		`gorm:"not null;type:varchar(36);default:null"`
	Content 			string		`gorm:"type:text"`
}

func (d DemandRecordDO) TableName() string {
	return "demand_records"
}

type TiUPConfigDO struct {
	Record
	ClusterId			string		`gorm:"not null;type:varchar(36);default:null"`
	Content 			string		`gorm:"type:text"`
}

func (d TiUPConfigDO) TableName() string {
	return "tiup_configs"
}

func UpdateClusterStatus(clusterId string, status int8) (cluster *ClusterDO, err error) {
	if clusterId == ""{
		return nil, errors.New("cluster id is empty")
	}
	cluster = &ClusterDO{}
	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("status", status).Error
	return
}

func UpdateClusterDemand(clusterId string, content string, tenantId string) (cluster *ClusterDO, demand *DemandRecordDO, err error) {
	demand = &DemandRecordDO{
		ClusterId: clusterId,
		Content: content,
		Record: Record{
			TenantId: tenantId,
		},
	}

	err = MetaDB.Create(demand).Error
	if err != nil {
		return
	}

	cluster = &ClusterDO{}
	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("current_demand_id", demand.ID).Find(cluster).Error
	return
}

func UpdateClusterFlowId(clusterId string, flowId uint) (cluster *ClusterDO, err error) {
	if clusterId == "" {
		return nil, errors.New("cluster id is empty")
	}
	cluster = &ClusterDO{}

	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("current_flow_id", flowId).Find(cluster).Error

	return
}

func UpdateTiUPConfig(clusterId string, content string, tenantId string) (cluster *ClusterDO, err error) {
	cluster = &ClusterDO{}
	record := &TiUPConfigDO{
		ClusterId: clusterId,
		Content: content,
		Record: Record{
			TenantId: tenantId,
		},
	}

	err = MetaDB.Create(record).Error
	if err != nil {
		return
	}

	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("current_tiup_config_id", record.ID).Find(cluster).Error

	return
}

func DeleteCluster(clusterId string) (cluster *ClusterDO, err error) {
	if clusterId == ""{
		 return nil, errors.New("empty cluster id")
	}
	cluster = &ClusterDO{}
	err = MetaDB.Find(cluster, "id = ?", clusterId).Error

	if err != nil {
		return
	}

	err = MetaDB.Delete(cluster).Error
	return
}

func FetchCluster(clusterId string) (result *ClusterFetchResult, err error) {
	result = &ClusterFetchResult{
		Cluster: &ClusterDO{},
	}

	err = MetaDB.Find(result.Cluster, "id = ?", clusterId).Error
	if err != nil {
		return
	}

	cluster := result.Cluster
	if cluster.CurrentDemandId > 0 {
		result.DemandRecord = &DemandRecordDO{}
		err = MetaDB.Find(result.DemandRecord, "id = ?", cluster.CurrentDemandId).Error
		if err != nil {
			return
		}
	}

	if cluster.CurrentTiupConfigId > 0 {
		result.TiUPConfig = &TiUPConfigDO{}
		err = MetaDB.Find(result.TiUPConfig, "id = ?", cluster.CurrentTiupConfigId).Error
		if err != nil {
			return
		}
	}

	if cluster.CurrentFlowId > 0 {
		result.Flow = &FlowDO{}
		err = MetaDB.Find(result.Flow, "id = ?", cluster.CurrentFlowId).Error
	}
	return
}

type ClusterFetchResult struct {
	Cluster *ClusterDO
	Flow *FlowDO
	DemandRecord *DemandRecordDO
	TiUPConfig *TiUPConfigDO
}

func ListClusterDetails(clusterId string,
	clusterName string,
	clusterType string,
	clusterStatus string,
	clusterTag string,
	offset int, length int) (result []*ClusterFetchResult, total int64, err error){

	clusters, total, err := ListClusters(clusterId, clusterName, clusterType, clusterStatus, clusterTag, offset, length)

	flowIds := make([]uint, len(clusters), len(clusters))
	demandIds := make([]uint, len(clusters), len(clusters))
	tiupConfigIds := make([]uint, len(clusters), len(clusters))

	result = make([]*ClusterFetchResult, len(clusters), len(clusters))
	clusterMap := make(map[string]*ClusterFetchResult)

	for i,c := range clusters {
		flowIds[i] = c.CurrentFlowId
		demandIds[i] = c.CurrentDemandId
		tiupConfigIds[i] = c.CurrentTiupConfigId
		result[i] = &ClusterFetchResult{
			Cluster: c,
		}
		clusterMap[c.ID] = result[i]
	}

	flows := make([]*FlowDO, len(clusters), len(clusters))
	err = MetaDB.Find(&flows, flowIds).Error
	for _,v := range flows {
		clusterMap[v.BizId].Flow = v
	}

	demands := make([]*DemandRecordDO, len(clusters), len(clusters))
	err = MetaDB.Find(&demands, demandIds).Error
	for _,v := range demands {
		clusterMap[v.ClusterId].DemandRecord = v
	}

	tiupConfigs := make([]*TiUPConfigDO, len(clusters), len(clusters))
	err = MetaDB.Find(&tiupConfigs, tiupConfigIds).Error
	for _,v := range tiupConfigs {
		clusterMap[v.ClusterId].TiUPConfig = v
	}

	return
}

func ListClusters(clusterId string,
	clusterName string,
	clusterType string,
	clusterStatus string,
	clusterTag string,
	offset int, length int) (clusters []*ClusterDO, total int64, err error){

	clusters = make([]*ClusterDO, length, length)

	db := MetaDB.Table("clusters")

	if clusterId != ""{
		db = db.Where("id = ?", clusterId)
	}

	if clusterName != ""{
		db = db.Where("cluster_name like '%" + clusterName + "%'")
	}

	if clusterType != ""{
		db = db.Where("cluster_type = ?", clusterType)
	}

	if clusterStatus != ""{
		db = db.Where("status = ?", clusterStatus)
	}

	if clusterTag != ""{
		db = db.Where("tags like '%," + clusterTag + ",%'")
	}

	err = db.Count(&total).Offset(offset).Limit(length).Find(&clusters).Error

	return
}

func CreateCluster(
		ClusterName 			string,
		DbPassword 				string,
		ClusterType 			string,
		ClusterVersion 			string,
		Tls 					bool,
		Tags           			string,
		OwnerId 				string,
		TenantId    			string,
	) (cluster *ClusterDO, err error){
	cluster = &ClusterDO{}
	cluster.ClusterName = ClusterName
	cluster.DbPassword = DbPassword
	cluster.ClusterType = ClusterType
	cluster.ClusterVersion = ClusterVersion
	cluster.Tls = Tls
	cluster.Tags = Tags
	cluster.OwnerId = OwnerId
	cluster.TenantId = TenantId

	err = MetaDB.Create(cluster).Error
	if err != nil {
		return
	}

	return
}

type BackupRecordDO struct {
	Record
	ClusterId   string		`gorm:"not null;type:varchar(36);default:null"`
	BackupRange string
	BackupType  string
	OperatorId  string		`gorm:"not null;type:varchar(36);default:null"`

	FilePath 	string
	FlowId		int64
	Size 		uint64

	StartTime	int64
	EndTime 	int64
}

func (d BackupRecordDO) TableName() string {
	return "backup_records"
}

type RecoverRecordDO struct {
	Record
	ClusterId 		string		`gorm:"not null;type:varchar(36);default:null"`

	OperatorId 		string		`gorm:"not null;type:varchar(36);default:null"`
	BackupRecordId  uint
	FlowId			uint
}



func (d RecoverRecordDO) TableName() string {
	return "recover_records"
}

type BackupStrategyDO struct {
	Record
	ClusterId 		string		`gorm:"not null;type:varchar(36);default:null"`
	OperatorId 		string		`gorm:"not null;type:varchar(36);default:null"`

	BackupDate  	string
	FilePath    	string
	BackupRange 	string
	BackupType  	string
	Period      	string
}

func (d BackupStrategyDO) TableName() string {
	return "backup_strategy"
}

type ParametersRecordDO struct {
	Record
	ClusterId 		string		`gorm:"not null;type:varchar(36);default:null"`

	OperatorId 		string		`gorm:"not null;type:varchar(36);default:null"`
	Content 		string		`gorm:"type:text"`
	FlowId 			uint
}

func (d ParametersRecordDO) TableName() string {
	return "parameters_records"
}

func SaveParameters(tenantId, clusterId, operatorId string, flowId uint, content string) (do *ParametersRecordDO, err error) {
	do = &ParametersRecordDO {
		Record: Record{
			TenantId: tenantId,
		},
		OperatorId: operatorId,
		ClusterId: clusterId,
		Content: content,
		FlowId: flowId,
	}

	err = MetaDB.Create(do).Error
	return
}

func GetCurrentParameters(clusterId string) (do *ParametersRecordDO, err error) {
	do = &ParametersRecordDO{}
	err = MetaDB.Where("cluster_id = ?", clusterId).Last(do).Error
	return
}

func DeleteBackupRecord(id uint) (record *BackupRecordDO, err error) {
	record = &BackupRecordDO{}
	err = MetaDB.Find(record, "id = ?", id).Error

	if err != nil {
		return
	}

	err = MetaDB.Delete(record).Error
	return
}

func SaveBackupRecord(record *dbPb.DBBackupRecordDTO) (do *BackupRecordDO, err error){
	do = &BackupRecordDO{
		Record: Record{
			TenantId: record.GetClusterId(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ClusterId: record.GetClusterId(),
		OperatorId: record.GetOperatorId(),
		BackupRange: record.GetBackupRange(),
		BackupType: record.GetBackupType(),
		FlowId: record.GetFlowId(),
		FilePath: record.GetFilePath(),
		StartTime: record.GetStartTime(),
	}

	err = MetaDB.Create(do).Error
	return
}

func UpdateBackupRecord(record *dbPb.DBBackupRecordDTO) error {
	err := MetaDB.Model(&BackupRecordDO{}).Where("id = ?", record.Id).Updates(BackupRecordDO{
		Size: record.GetSize(), EndTime: record.GetEndTime(), Record: Record{
			TenantId: record.GetClusterId(),
			UpdatedAt: time.Now(),
		}}).Error

	if err != nil {
		return err
	}
	return nil
}

type BackupRecordFetchResult struct {
	BackupRecordDO *BackupRecordDO
	Flow *FlowDO
}

func QueryBackupRecord(clusterId string, recordId int64) (*BackupRecordFetchResult, error) {
	record := BackupRecordDO{}
	err := MetaDB.Table("backup_records").Where("id = ? and cluster_id", recordId, clusterId).First(&record).Error
	if err != nil {
		return nil, err
	}

	flow := FlowDO{}
	err = MetaDB.Find(&flow, record.FlowId).Error
	if err != nil {
		return nil, err
	}

	return &BackupRecordFetchResult{
		BackupRecordDO: &record,
		Flow: &flow,
	}, nil
}

func ListBackupRecords(clusterId string,
	offset, length int) (dos []*BackupRecordFetchResult, total int64, err error) {

	records := make([]*BackupRecordDO, length, length)
	err = MetaDB.Table("backup_records").
		Where("cluster_id = ?", clusterId).
		Count(&total).Order("id desc").Offset(offset).Limit(length).
		Find(&records).
		Error

	if err != nil {return}
	// query flows
	flowIds := make([]int64, len(records), len(records))

	dos = make([]*BackupRecordFetchResult, len(records), len(records))

	for i,r := range records {
		flowIds[i] = r.FlowId
		dos[i] = &BackupRecordFetchResult{
			BackupRecordDO: r,
		}
	}

	flows := make([]*FlowDO, len(records), len(records))
	err = MetaDB.Find(&flows, flowIds).Error
	if err != nil {return}

	flowMap := make(map[int64]*FlowDO)

	for _,v := range flows {
		flowMap[int64(v.ID)] = v
	}

	for i,v := range records {
		dos[i].BackupRecordDO = v
		dos[i].Flow = flowMap[v.FlowId]
	}

	return
}

func SaveRecoverRecord(tenantId, clusterId, operatorId string,
	backupRecordId uint,
	flowId uint) (do *RecoverRecordDO, err error) {
	do = &RecoverRecordDO{
		Record: Record{
			TenantId: tenantId,
		},
		ClusterId: clusterId,
		OperatorId: operatorId,
		FlowId: flowId,
		BackupRecordId: backupRecordId,
	}

	err = MetaDB.Create(do).Error
	return
}

func SaveBackupStrategy(strategy *dbPb.DBBackupStrategyDTO) (*BackupStrategyDO, error) {
	strategyDO := BackupStrategyDO{
		ClusterId: strategy.ClusterId,
		Record: Record{
			TenantId: strategy.TenantId,
		},
		BackupDate: strategy.BackupDate,
		BackupRange: strategy.BackupRange,
		BackupType: strategy.BackupType,
		Period: strategy.Period,
		FilePath: strategy.FilePath,
	}
	result := MetaDB.Table("backup_strategy").Where("cluster_id = ?", strategy.ClusterId).First(&strategyDO)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			strategyDO.CreatedAt = time.Now()
			strategyDO.UpdatedAt = time.Now()
			err := MetaDB.Create(&strategyDO).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, result.Error
		}
	} else {
		strategyDO.UpdatedAt = time.Now()
		err := MetaDB.Model(&BackupStrategyDO{}).Updates(&BackupStrategyDO{
			BackupDate: strategy.BackupDate,
			BackupRange: strategy.BackupRange,
			BackupType: strategy.BackupType,
			Period: strategy.Period,
			FilePath: strategy.FilePath,
		}).Error
		if err != nil {
			return nil, err
		}
	}
	return &strategyDO, nil
}

func QueryBackupStartegy(clusterId string) (*BackupStrategyDO, error) {
	strategyDO := BackupStrategyDO{}
	err := MetaDB.Table("backup_strategy").Where("cluster_id = ?", clusterId).First(&strategyDO).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &strategyDO, nil
}