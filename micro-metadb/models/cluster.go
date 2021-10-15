package models

import (
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"time"

	"github.com/pingcap/errors"
	"gorm.io/gorm"
)

type Cluster struct {
	Entity
	Name                    string
	DbPassword              string
	Type                    string
	Version                 string
	Tls                     bool
	Tags                    string
	OwnerId                 string `gorm:"not null;type:varchar(22);default:null"`
	CurrentTopologyConfigId uint
	CurrentDemandId         uint
	CurrentFlowId           uint
}

type DemandRecord struct {
	Record
	ClusterId string `gorm:"not null;type:varchar(22);default:null"`
	Content   string `gorm:"type:text"`
}

type TopologyConfig struct {
	Record
	ClusterId string `gorm:"not null;type:varchar(22);default:null"`
	Content   string `gorm:"type:text"`
}

type ClusterFetchResult struct {
	Cluster            *Cluster
	Flow               *FlowDO
	DemandRecord       *DemandRecord
	TopologyConfig     *TopologyConfig
	ComponentInstances []*ComponentInstance
}

type BackupRecord struct {
	Record
	StorageType  string
	ClusterId    string `gorm:"not null;type:varchar(22);default:null"`
	BackupType   string
	BackupMethod string
	BackupMode   string
	OperatorId   string `gorm:"not null;type:varchar(22);default:null"`

	FilePath string
	FlowId   int64
	Size     uint64

	StartTime time.Time
	EndTime   time.Time
}

type RecoverRecord struct {
	Record
	ClusterId      string `gorm:"not null;type:varchar(22);default:null"`
	OperatorId     string `gorm:"not null;type:varchar(22);default:null"`
	BackupRecordId uint
	FlowId         uint
}

type ParametersRecord struct {
	Record
	ClusterId  string `gorm:"not null;type:varchar(22);default:null"`
	OperatorId string `gorm:"not null;type:varchar(22);default:null"`
	Content    string `gorm:"type:text"`
	FlowId     uint
}

type BackupStrategy struct {
	Record
	ClusterId  string `gorm:"not null;type:varchar(22);default:null"`
	OperatorId string `gorm:"not null;type:varchar(22);default:null"`

	BackupDate  string
	StartHour   uint32
	EndHour 	uint32
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
	err = m.Db().Model(cluster).Where("id = ?", clusterId).First(cluster).Update("status", status).Error
	return cluster, err
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

func (m *DAOClusterManager) UpdateTopologyConfig(clusterId string, content string, tenantId string) (cluster *Cluster, err error) {
	if "" == clusterId || "" == tenantId || "" == content {
		return nil, errors.New(fmt.Sprintf("UpdateTopologyConfig has invalid parameter, clusterId: %s, content: %s", clusterId, content))
	}
	cluster = &Cluster{}
	record := &TopologyConfig{
		ClusterId: clusterId,
		Content:   content,
		Record: Record{
			TenantId: tenantId,
		},
	}
	err = m.Db().Create(record).Error
	if nil == err {
		err = m.Db().Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_topology_config_id", record.ID).Error
		if nil != err {
			err = errors.New(fmt.Sprintf("update topology config faild, clusterId: %s, tenantId: %s, topologyId: %d, error: %v", clusterId, tenantId, record.ID, err))
		}
	} else {
		err = errors.New(fmt.Sprintf("craete topology config faild, clusterId: %s, tenantId: %s, topologyId: %d, error: %v", clusterId, tenantId, record.ID, err))
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

		if cluster.CurrentTopologyConfigId > 0 {
			result.TopologyConfig = &TopologyConfig{}
			err = m.Db().First(result.TopologyConfig, "id = ?", cluster.CurrentTopologyConfigId).Error
			if nil != err {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query demand record failed, clusterId: %s, topologyId:%d, error: %v", clusterId, cluster.CurrentTopologyConfigId, err))
			}
		}

		if cluster.CurrentFlowId > 0 {
			result.Flow = &FlowDO{}
			err = m.Db().First(result.Flow, "id = ?", cluster.CurrentFlowId).Error
			if nil != err {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query workflow failed, clusterId: %s, workflowId:%d, error: %v", clusterId, cluster.CurrentFlowId, err))
			}
		}

		result.ComponentInstances, err = m.ListComponentInstances(cluster.ID)
		if nil != err {
			return nil, errors.New(fmt.Sprintf("FetchCluster, ListComponentInstances failed, clusterId: %s, error: %v", clusterId, err))
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
	topologyConfigIds := make([]uint, len(clusters), len(clusters))

	result = make([]*ClusterFetchResult, len(clusters), len(clusters))
	clusterMap := make(map[string]*ClusterFetchResult)

	for i, c := range clusters {
		flowIds[i] = c.CurrentFlowId
		demandIds[i] = c.CurrentDemandId
		topologyConfigIds[i] = c.CurrentTopologyConfigId
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

	topologyConfigs := make([]*TopologyConfig, len(clusters), len(clusters))
	err = m.Db().Find(&topologyConfigs, topologyConfigIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query topology config lists failed, error: %v", err))
	}
	for _, v := range topologyConfigs {
		clusterMap[v.ClusterId].TopologyConfig = v
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
	do = &ParametersRecord{
		Record: Record{
			TenantId: tenantId,
		},
		OperatorId: operatorId,
		ClusterId:  clusterId,
		Content:    content,
		FlowId:     flowId,
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
	record.ID = id
	err = m.Db().Where("id = ?", record.ID).Delete(record).Error
	return record, err
}

func (m *DAOClusterManager) SaveBackupRecord(record *dbpb.DBBackupRecordDTO) (do *BackupRecord, err error) {
	do = &BackupRecord{
		Record: Record{
			TenantId:  record.GetTenantId(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		StorageType:  record.GetStorageType(),
		ClusterId:    record.GetClusterId(),
		OperatorId:   record.GetOperatorId(),
		BackupMethod: record.GetBackupMethod(),
		BackupType:   record.GetBackupType(),
		BackupMode:   record.GetBackupMode(),
		FlowId:       record.GetFlowId(),
		FilePath:     record.GetFilePath(),
		StartTime:    time.Unix(record.GetStartTime(), 0),
		EndTime:      time.Unix(record.GetEndTime(), 0),
	}

	return do, m.Db().Create(do).Error
}

func (m *DAOClusterManager) UpdateBackupRecord(record *dbpb.DBBackupRecordDTO) error {
	err := m.Db().Model(&BackupRecord{}).Where("id = ?", record.Id).Updates(BackupRecord{
		Size:    record.GetSize(),
		EndTime: time.Unix(record.GetEndTime(), 0),
		Record: Record{
			TenantId:  record.GetTenantId(),
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
		Flow:         &flow,
	}, nil
}
func (m *DAOClusterManager) ListBackupRecords(clusterId string, startTime, endTime int64,offset, length int) (dos []*BackupRecordFetchResult, total int64, err error) {

	records := make([]*BackupRecord, length)
	db := m.Db().Table(TABLE_NAME_BACKUP_RECORD).Where("cluster_id = ? and deleted_at is null", clusterId)
	if startTime > 0 {
		db = db.Where("start_time >= ?", time.Unix(startTime, 0))
	}
	if endTime > 0 {
		db = db.Where("end_time <= ?", time.Unix(endTime, 0))
	}

	err =db.Count(&total).Order("id desc").Offset(offset).Limit(length).
		Find(&records).
		Error

	if nil == err {
		// query flows
		flowIds := make([]int64, len(records))
		dos = make([]*BackupRecordFetchResult, len(records))
		for i, r := range records {
			flowIds[i] = r.FlowId
			dos[i] = &BackupRecordFetchResult{
				BackupRecord: r,
			}
		}

		flows := make([]*FlowDO, len(records))
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

func (m *DAOClusterManager) SaveBackupStrategy(strategy *dbpb.DBBackupStrategyDTO) (*BackupStrategy, error) {
	strategyDO := BackupStrategy{
		ClusterId: strategy.GetClusterId(),
		OperatorId: strategy.GetOperatorId(),
		Record: Record{
			TenantId: strategy.GetTenantId(),
		},
		BackupDate:  strategy.GetBackupDate(),
		StartHour:   strategy.GetStartHour(),
		EndHour:     strategy.GetEndHour(),
	}
	result := m.Db().Table(TABLE_NAME_BACKUP_STRATEGY).Where("cluster_id = ?", strategy.ClusterId).First(&strategyDO)
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
		strategyDO.BackupDate = strategy.GetBackupDate()
		strategyDO.StartHour = strategy.GetStartHour()
		strategyDO.EndHour = strategy.GetEndHour()
		err := m.Db().Model(&strategyDO).Save(&strategyDO).Error
		if err != nil {
			return nil, err
		}
	}
	return &strategyDO, nil
}

func (m *DAOClusterManager) QueryBackupStartegy(clusterId string) (*BackupStrategy, error) {
	strategyDO := BackupStrategy{}
	err := m.Db().Table(TABLE_NAME_BACKUP_STRATEGY).Where("cluster_id = ?", clusterId).First(&strategyDO).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &strategyDO, nil
}


func (m *DAOClusterManager) QueryBackupStartegyByTime(weekday string, startHour uint32) ([]*BackupStrategy, error) {
	var strategyListDO []*BackupStrategy
	err := m.Db().Table(TABLE_NAME_BACKUP_STRATEGY).Where("start_hour = ?", startHour).Where("backup_date like '%" + weekday + "%'").Find(&strategyListDO).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return strategyListDO, nil
}
