package models

import (
	"fmt"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/pingcap/errors"
	"gorm.io/gorm"
	"time"
)

type Cluster struct {
	Entity
	Name                string
	DbPassword          string
	Type                string
	Version             string
	Tls                 bool
	Tags                string
	OwnerId             string `gorm:"not null;type:varchar(36);default:null"`
	CurrentTiupConfigId uint
	CurrentDemandId     uint
	CurrentFlowId       uint
}

type DemandRecord struct {
	Record
	ClusterId string `gorm:"not null;type:varchar(36);default:null"`
	Content   string `gorm:"type:text"`
}

type TiUPConfig struct {
	Record
	ClusterId string `gorm:"not null;type:varchar(36);default:null"`
	Content   string `gorm:"type:text"`
}

type ClusterFetchResult struct {
	Cluster      *Cluster
	Flow         *FlowDO
	DemandRecord *DemandRecord
	TiUPConfig   *TiUPConfig
}

type BackupRecord struct {
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

type RecoverRecord struct {
	Record
	ClusterId      string `gorm:"not null;type:varchar(36);default:null"`
	OperatorId     string `gorm:"not null;type:varchar(36);default:null"`
	BackupRecordId uint
	FlowId         uint

}

type ParametersRecord struct {
	Record
	ClusterId  string `gorm:"not null;type:varchar(36);default:null"`
	OperatorId string `gorm:"not null;type:varchar(36);default:null"`
	Content    string `gorm:"type:text"`
	FlowId     uint
}
type BackupStrategy struct {
	Record
	ClusterId 		string		`gorm:"not null;type:varchar(36);default:null"`
	OperatorId 		string		`gorm:"not null;type:varchar(36);default:null"`

	BackupDate  	string
	FilePath    	string
	BackupRange 	string
	BackupType  	string
	Period      	string
}

type BackupRecordFetchResult struct {
	BackupRecord *BackupRecord
	Flow         *FlowDO
}

type DAOClusterManager struct {
	db *gorm.DB
}

func NewDAOClusterManager(d *gorm.DB) *DAOClusterManager {
	m := new(DAOClusterManager)
	m.SetDb(d)
	return m
}

func (m *DAOClusterManager) SetDb(db *gorm.DB) {
	m.db = db
}

func (m *DAOClusterManager) Db() *gorm.DB {
	return m.db
}

func (m *DAOClusterManager) UpdateClusterStatus(clusterId string, status int8) (cluster *Cluster, err error) {
	if clusterId == "" {
		return nil, errors.New(fmt.Sprintf("UpdateClusterStatus has invalid parameter, clusterId: %s, status: %d", clusterId, status))
	}
	cluster = &Cluster{}
	return cluster, m.Db().Model(cluster).Where("id = ?", clusterId).Update("status", status).Find(cluster).Error
}

func (m *DAOClusterManager) UpdateClusterDemand(clusterId string, content string, tenantId string) (cluster *Cluster, demand *DemandRecord, err error) {
	if "" == clusterId || "" == tenantId {
		return nil, nil, errors.New(fmt.Sprintf("UpdateClusterDemand has invalid parameter, clusterId: %s, content: %s, content: %s", clusterId, tenantId, content))
	}

	cluster = &Cluster{}
	demand = &DemandRecord{
		ClusterId: clusterId,
		Content:   content,
		Record: Record{
			TenantId: tenantId,
		},
	}

	err = m.Db().Create(demand).Error
	if nil == err {
		err = m.Db().Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_demand_id", demand.ID).Error
		if nil != err {
			err = errors.New(fmt.Sprintf("update demand faild, clusterId: %s, tenantId: %s, demandId: %d, error: %v", clusterId, tenantId, demand.ID, err))
		}
	} else {
		err = errors.New(fmt.Sprintf("craete demand faild, clusterId: %s, tenantId: %s, demandId: %d, error: %v", clusterId, tenantId, demand.ID, err))
	}
	return cluster, demand, err
}

func (m *DAOClusterManager) UpdateClusterFlowId(clusterId string, flowId uint) (cluster *Cluster, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("UpdateClusterFlowId has invalid parameter, clusterId: %s, flowId: %d", clusterId, flowId))
	}
	cluster = &Cluster{}
	return cluster, m.Db().Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_flow_id", flowId).Error
}

func (m *DAOClusterManager) UpdateTiUPConfig(clusterId string, content string, tenantId string) (cluster *Cluster, err error) {
	if "" == clusterId || "" == tenantId || "" == content {
		return nil, errors.New(fmt.Sprintf("UpdateTiUPConfig has invalid parameter, clusterId: %s, content: %s", clusterId, content))
	}
	cluster = &Cluster{}
	record := &TiUPConfig{
		ClusterId: clusterId,
		Content:   content,
		Record: Record{
			TenantId: tenantId,
		},
	}
	err = m.Db().Create(record).Error
	if nil == err {
		err = m.Db().Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_tiup_config_id", record.ID).Error
		if nil != err {
			err = errors.New(fmt.Sprintf("update tiup config faild, clusterId: %s, tenantId: %s, TiUPId: %d, error: %v", clusterId, tenantId, record.ID, err))
		}
	} else {
		err = errors.New(fmt.Sprintf("craete tiup config faild, clusterId: %s, tenantId: %s, TiUPId: %d, error: %v", clusterId, tenantId, record.ID, err))
	}
	return cluster, err
}

func (m *DAOClusterManager) DeleteCluster(clusterId string) (cluster *Cluster, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("DeleteCluster has invalid parameter, clusterId: %s", clusterId))
	}
	cluster = &Cluster{}
	return cluster, m.Db().First(cluster, "id = ?", clusterId).Delete(cluster).Error
}

func (m *DAOClusterManager) FetchCluster(clusterId string) (result *ClusterFetchResult, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("FetchCluster has invalid parameter, clusterId: %s", clusterId))
	}
	result = &ClusterFetchResult{
		Cluster: &Cluster{},
	}

	err = m.Db().First(result.Cluster, "id = ?", clusterId).Error
	if nil == err {
		cluster := result.Cluster
		if cluster.CurrentDemandId > 0 {
			result.DemandRecord = &DemandRecord{}
			err = m.Db().First(result.DemandRecord, "id = ?", cluster.CurrentDemandId).Error
			if err != nil {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query demand record failed, clusterId: %s, demandId: %d, error: %v", clusterId, cluster.CurrentDemandId, err))
			}
		}

		if cluster.CurrentTiupConfigId > 0 {
			result.TiUPConfig = &TiUPConfig{}
			err = m.Db().First(result.TiUPConfig, "id = ?", cluster.CurrentTiupConfigId).Error
			if nil != err {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query demand record failed, clusterId: %s, TiUPID:%d, error: %v", clusterId, cluster.CurrentTiupConfigId, err))
			}
		}

		if cluster.CurrentFlowId > 0 {
			result.Flow = &FlowDO{}
			err = m.Db().First(result.Flow, "id = ?", cluster.CurrentFlowId).Error
			if nil != err {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query workflow failed, clusterId: %s, workflowId:%d, error: %v", clusterId, cluster.CurrentFlowId, err))
			}
		}
	} else {
		return nil, errors.New(fmt.Sprintf("FetchCluster, query cluster failed, clusterId: %s, error: %v", clusterId, err))
	}
	return result, nil
}

func (m *DAOClusterManager) ListClusterDetails(clusterId, clusterName, clusterType, clusterStatus string,
	clusterTag string, offset int, length int) (result []*ClusterFetchResult, total int64, err error) {

	clusters, total, err := m.ListClusters(clusterId, clusterName, clusterType, clusterStatus, clusterTag, offset, length)

	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query cluster lists failed, error: %v", err))
	}

	flowIds := make([]uint, len(clusters), len(clusters))
	demandIds := make([]uint, len(clusters), len(clusters))
	tiupConfigIds := make([]uint, len(clusters), len(clusters))

	result = make([]*ClusterFetchResult, len(clusters), len(clusters))
	clusterMap := make(map[string]*ClusterFetchResult)

	for i, c := range clusters {
		flowIds[i] = c.CurrentFlowId
		demandIds[i] = c.CurrentDemandId
		tiupConfigIds[i] = c.CurrentTiupConfigId
		result[i] = &ClusterFetchResult{
			Cluster: c,
		}
		clusterMap[c.ID] = result[i]
	}

	flows := make([]*FlowDO, len(clusters), len(clusters))
	err = m.Db().Find(&flows, flowIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query flow lists failed, error: %v", err))
	}
	for _, v := range flows {
		clusterMap[v.BizId].Flow = v
	}

	demands := make([]*DemandRecord, len(clusters), len(clusters))
	err = m.Db().Find(&demands, demandIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query demand lists failed, error: %v", err))
	}
	for _, v := range demands {
		clusterMap[v.ClusterId].DemandRecord = v
	}

	tiupConfigs := make([]*TiUPConfig, len(clusters), len(clusters))
	err = m.Db().Find(&tiupConfigs, tiupConfigIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query TiUP config lists failed, error: %v", err))
	}
	for _, v := range tiupConfigs {
		clusterMap[v.ClusterId].TiUPConfig = v
	}
	return result, total, nil
}

func (m *DAOClusterManager) ListClusters(clusterId, clusterName, clusterType, clusterStatus string,
	clusterTag string, offset int, length int) (clusters []*Cluster, total int64, err error) {

	clusters = make([]*Cluster, length, length)
	query := m.Db().Table(TABLE_NAME_CLUSTER)
	if clusterId != "" {
		query = query.Where("id = ?", clusterId)
	}
	if clusterName != "" {
		query = query.Where("name like '%" + clusterName + "%'")
	}
	if clusterType != "" {
		query = query.Where("type = ?", clusterType)
	}
	if clusterStatus != "" {
		query = query.Where("status = ?", clusterStatus)
	}
	if clusterTag != "" {
		query = query.Where("tags like '%," + clusterTag + ",%'")
	}
	return clusters, total, query.Count(&total).Offset(offset).Limit(length).Find(&clusters).Error
}

func (m *DAOClusterManager) CreateCluster(ClusterName, DbPassword, ClusterType, ClusterVersion string,
	Tls bool, Tags, OwnerId, tenantId string) (cluster *Cluster, err error) {
	cluster = &Cluster{Entity: Entity{TenantId: tenantId},
		Name:       ClusterName,
		DbPassword: DbPassword,
		Type:       ClusterType,
		Version:    ClusterVersion,
		Tls:        Tls,
		Tags:       Tags,
		OwnerId:    OwnerId,
	}
	cluster.Code = generateEntityCode(ClusterName)
	err = m.Db().Create(cluster).Error
	return
}

func (m *DAOClusterManager) SaveParameters(tenantId, clusterId, operatorId string, flowId uint, content string) (do *ParametersRecord, err error) {
	if "" == tenantId || "" == clusterId || "" == operatorId || "" == content {
		return nil, errors.New(fmt.Sprintf("SaveParameters has invalid parameter, tenantId: %s, clusterId:%s, operatorId: %s, content: %s, flowId: %d",
			tenantId, clusterId, operatorId, content, flowId))
	}
	do = &ParametersRecord {
		Record: Record{
			TenantId: tenantId,
		},
		OperatorId: operatorId,
		ClusterId: clusterId,
		Content: content,
		FlowId: flowId,
	}

	err = m.Db().Create(do).Error
	return
}

func (m *DAOClusterManager) GetCurrentParameters(clusterId string) (do *ParametersRecord, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("GetCurrentParameters has invalid parameter,clusterId:%s", clusterId))
	}
	do = &ParametersRecord{}
	return do, m.Db().Where("cluster_id = ?", clusterId).Last(do).Error
}

func (m *DAOClusterManager) DeleteBackupRecord(id uint) (record *BackupRecord, err error) {
	if id <= 0 {
		return nil, errors.New(fmt.Sprintf("DeleteBackupRecord has invalid parameter, Id: %d", id))
	}
	record = &BackupRecord{}
	err = m.Db().First(record, "id = ?", id).Error
	if err != nil {
		err = m.Db().Delete(record).Error
	}
	return record, err
}

func (m *DAOClusterManager) SaveBackupRecord(record *dbPb.DBBackupRecordDTO) (do *BackupRecord, err error) {
	do = &BackupRecord{
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

	return do, m.Db().Create(do).Error
}

func (m *DAOClusterManager) UpdateBackupRecord(record *dbPb.DBBackupRecordDTO) error {
	err := m.Db().Model(&BackupRecord{}).Where("id = ?", record.Id).Updates(BackupRecord{
		Size: record.GetSize(), EndTime: record.GetEndTime(), Record: Record{
			TenantId: record.GetClusterId(),
			UpdatedAt: time.Now(),
		}}).Error

	if err != nil {
		return err
	}
	return nil
}

func (m *DAOClusterManager) QueryBackupRecord(clusterId string, recordId int64) (*BackupRecordFetchResult, error) {
	record := BackupRecord{}
	err := m.Db().Table("backup_records").Where("id = ? and cluster_id = ?", recordId, clusterId).First(&record).Error
	if err != nil {
		return nil, err
	}

	flow := FlowDO{}
	err = m.Db().Find(&flow, record.FlowId).Error
	if err != nil {
		return nil, err
	}

	return &BackupRecordFetchResult{
		BackupRecord: &record,
		Flow: &flow,
	}, nil
}
func (m *DAOClusterManager) ListBackupRecords(clusterId string,
	offset, length int) (dos []*BackupRecordFetchResult, total int64, err error) {

	records := make([]*BackupRecord, length, length)
	err = m.Db().Table(TABLE_NAME_BACKUP_RECORD).
		Where("cluster_id = ?", clusterId).
		Count(&total).Order("id desc").Offset(offset).Limit(length).
		Find(&records).
		Error

	if nil == err {
		// query flows
		flowIds := make([]int64, len(records), len(records))
		dos = make([]*BackupRecordFetchResult, len(records), len(records))
		for i, r := range records {
			flowIds[i] = r.FlowId
			dos[i] = &BackupRecordFetchResult{
				BackupRecord: r,
			}
		}

		flows := make([]*FlowDO, len(records), len(records))
		err = m.Db().Find(&flows, flowIds).Error
		if err != nil {
			return nil, 0, errors.New(fmt.Sprintf("ListBackupRecord, query record failed, clusterId: %s, error: %v", clusterId, err))
		}

		flowMap := make(map[int64]*FlowDO)
		for _, v := range flows {
			flowMap[int64(v.ID)] = v
		}
		for i, v := range records {
			dos[i].BackupRecord = v
			dos[i].Flow = flowMap[v.FlowId]
		}
	}
	return
}

func (m *DAOClusterManager) SaveRecoverRecord(tenantId, clusterId, operatorId string,
	backupRecordId uint,
	flowId uint) (do *RecoverRecord, err error) {
	do = &RecoverRecord{
		Record: Record{
			TenantId: tenantId,
		},
		ClusterId:      clusterId,
		OperatorId:     operatorId,
		FlowId:         flowId,
		BackupRecordId: backupRecordId,
	}
	return do, m.Db().Create(do).Error
}

func (m *DAOClusterManager) SaveBackupStrategy(strategy *dbPb.DBBackupStrategyDTO) (*BackupStrategy, error) {
	strategyDO := BackupStrategy{
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
	result := m.Db().Table("backup_strategy").Where("cluster_id = ?", strategy.ClusterId).First(&strategyDO)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			strategyDO.CreatedAt = time.Now()
			strategyDO.UpdatedAt = time.Now()
			err := m.Db().Create(&strategyDO).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, result.Error
		}
	} else {
		strategyDO.UpdatedAt = time.Now()
		err := m.Db().Model(&BackupStrategy{}).Updates(&BackupStrategy{
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

func  (m *DAOClusterManager) QueryBackupStartegy(clusterId string) (*BackupStrategy, error) {
	strategyDO := BackupStrategy{}
	err := m.Db().Table("backup_strategy").Where("cluster_id = ?", clusterId).First(&strategyDO).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &strategyDO, nil
}
