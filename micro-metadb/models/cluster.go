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

package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/prometheus/client_golang/prometheus"

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
	Exclusive               bool
	Region                  string
	CpuArchitecture         string
	CurrentTopologyConfigId uint
	CurrentDemandId         uint
	CurrentFlowId           uint
	ParamGroupId            uint
}

type ClusterRelation struct {
	Record
	SubjectClusterId string
	ObjectClusterId  string
	RelationType     uint32
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

	FilePath  string
	FlowId    int64
	Size      uint64
	BackupTso uint64

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

	BackupDate string
	StartHour  uint32
	EndHour    uint32
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

func (m *DAOClusterManager) HandleMetrics(funcName string, code int) {
	framework.Current.GetMetrics().SqliteRequestsCounterMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Inc()
}

func (m *DAOClusterManager) SetDb(db *gorm.DB) {
	m.db = db
}

func (m *DAOClusterManager) Db(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

func (m *DAOClusterManager) UpdateClusterStatus(ctx context.Context, clusterId string, status int8) (cluster *Cluster, err error) {
	if clusterId == "" {
		return nil, errors.New(fmt.Sprintf("UpdateClusterStatus has invalid parameter, clusterId: %s, status: %d", clusterId, status))
	}
	cluster = &Cluster{}
	err = m.Db(ctx).Model(cluster).Where("id = ?", clusterId).First(cluster).Update("status", status).Error
	return cluster, err
}

func (m *DAOClusterManager) UpdateComponentDemand(ctx context.Context, clusterId string, content string, tenantId string) (cluster *Cluster, demand *DemandRecord, err error) {
	if "" == clusterId || "" == tenantId {
		return nil, nil, errors.New(fmt.Sprintf("UpdateComponentDemand has invalid parameter, clusterId: %s, content: %s, content: %s", clusterId, tenantId, content))
	}

	cluster = &Cluster{}
	demand = &DemandRecord{
		ClusterId: clusterId,
		Content:   content,
		Record: Record{
			TenantId: tenantId,
		},
	}

	err = m.Db(ctx).Create(demand).Error
	if nil == err {
		err = m.Db(ctx).Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_demand_id", demand.ID).Error
		if nil != err {
			err = errors.New(fmt.Sprintf("update demand faild, clusterId: %s, tenantId: %s, demandId: %d, error: %v", clusterId, tenantId, demand.ID, err))
		}
	} else {
		err = errors.New(fmt.Sprintf("craete demand faild, clusterId: %s, tenantId: %s, demandId: %d, error: %v", clusterId, tenantId, demand.ID, err))
	}
	return cluster, demand, err
}

func (m *DAOClusterManager) UpdateClusterFlowId(ctx context.Context, clusterId string, flowId uint) (cluster *Cluster, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("UpdateClusterFlowId has invalid parameter, clusterId: %s, flowId: %d", clusterId, flowId))
	}
	cluster = &Cluster{}
	return cluster, m.Db(ctx).Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_flow_id", flowId).Error
}

func (m *DAOClusterManager) UpdateClusterInfo(ctx context.Context, clusterId string, name, clusterType, versionCode, tags string, tls bool) (cluster *Cluster, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("UpdateClusterInfo has invalid parameter, clusterId: %s", clusterId))
	}
	cluster = &Cluster{}
	return cluster, m.Db(ctx).Model(cluster).Where("id = ?", clusterId).First(cluster).
		Update("name", name).
		Update("type", clusterType).
		Update("version", versionCode).
		Update("tags", tags).
		Update("tls", tls).
		Error
}
func (m *DAOClusterManager) UpdateTopologyConfig(ctx context.Context, clusterId string, content string, tenantId string) (cluster *Cluster, err error) {
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
	err = m.Db(ctx).Create(record).Error
	if nil == err {
		err = m.Db(ctx).Model(cluster).Where("id = ?", clusterId).First(cluster).Update("current_topology_config_id", record.ID).Error
		if nil != err {
			err = errors.New(fmt.Sprintf("update topology config faild, clusterId: %s, tenantId: %s, topologyId: %d, error: %v", clusterId, tenantId, record.ID, err))
		}
	} else {
		err = errors.New(fmt.Sprintf("craete topology config faild, clusterId: %s, tenantId: %s, topologyId: %d, error: %v", clusterId, tenantId, record.ID, err))
	}
	return cluster, err
}

func (m *DAOClusterManager) DeleteCluster(ctx context.Context, clusterId string) (cluster *Cluster, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("DeleteCluster has invalid parameter, clusterId: %s", clusterId))
	}
	cluster = &Cluster{}
	return cluster, m.Db(ctx).First(cluster, "id = ?", clusterId).Delete(cluster).Error
}

func (m *DAOClusterManager) FetchCluster(ctx context.Context, clusterId string) (result *ClusterFetchResult, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("FetchCluster has invalid parameter, clusterId: %s", clusterId))
	}
	result = &ClusterFetchResult{
		Cluster: &Cluster{},
	}

	err = m.Db(ctx).First(result.Cluster, "id = ?", clusterId).Error
	if nil == err {
		cluster := result.Cluster
		if cluster.CurrentDemandId > 0 {
			result.DemandRecord = &DemandRecord{}
			err = m.Db(ctx).First(result.DemandRecord, "id = ?", cluster.CurrentDemandId).Error
			if err != nil {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query demand record failed, clusterId: %s, demandId: %d, error: %v", clusterId, cluster.CurrentDemandId, err))
			}
		}

		if cluster.CurrentTopologyConfigId > 0 {
			result.TopologyConfig = &TopologyConfig{}
			err = m.Db(ctx).First(result.TopologyConfig, "id = ?", cluster.CurrentTopologyConfigId).Error
			if nil != err {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query demand record failed, clusterId: %s, topologyId:%d, error: %v", clusterId, cluster.CurrentTopologyConfigId, err))
			}
		}

		if cluster.CurrentFlowId > 0 {
			result.Flow = &FlowDO{}
			err = m.Db(ctx).First(result.Flow, "id = ?", cluster.CurrentFlowId).Error
			if nil != err {
				return nil, errors.New(fmt.Sprintf("FetchCluster, query workflow failed, clusterId: %s, workflowId:%d, error: %v", clusterId, cluster.CurrentFlowId, err))
			}
		}

		result.ComponentInstances, err = m.ListComponentInstances(ctx, cluster.ID)
		if nil != err {
			return nil, errors.New(fmt.Sprintf("FetchCluster, ListComponentInstances failed, clusterId: %s, error: %v", clusterId, err))
		}
	} else {
		return nil, errors.New(fmt.Sprintf("FetchCluster, query cluster failed, clusterId: %s, error: %v", clusterId, err))
	}
	return result, nil
}

func (m *DAOClusterManager) ListClusterDetails(ctx context.Context, clusterId, clusterName, clusterType, clusterStatus string,
	clusterTag string, offset int, length int) (result []*ClusterFetchResult, total int64, err error) {

	clusters, total, err := m.ListClusters(ctx, clusterId, clusterName, clusterType, clusterStatus, clusterTag, offset, length)

	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query cluster lists failed, error: %v", err))
	}

	result = make([]*ClusterFetchResult, len(clusters))

	if total == 0 {
		return result, 0, nil
	}

	flowIds := make([]uint, len(clusters))
	demandIds := make([]uint, len(clusters))
	topologyConfigIds := make([]uint, len(clusters))

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

	flows := make([]*FlowDO, len(clusters))
	err = m.Db(ctx).Find(&flows, flowIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query flow lists failed, error: %v", err))
	}
	for _, v := range flows {
		clusterMap[v.BizId].Flow = v
	}

	demands := make([]*DemandRecord, len(clusters))
	err = m.Db(ctx).Find(&demands, demandIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query demand lists failed, error: %v", err))
	}
	for _, v := range demands {
		clusterMap[v.ClusterId].DemandRecord = v
	}

	topologyConfigs := make([]*TopologyConfig, len(clusters))
	err = m.Db(ctx).Find(&topologyConfigs, topologyConfigIds).Error
	if nil != err {
		return nil, 0, errors.New(fmt.Sprintf("ListClusterDetails, query topology config lists failed, error: %v", err))
	}
	for _, v := range topologyConfigs {
		clusterMap[v.ClusterId].TopologyConfig = v
	}
	return result, total, nil
}

func (m *DAOClusterManager) ListClusters(ctx context.Context, clusterId, clusterName, clusterType, clusterStatus string,
	clusterTag string, offset int, length int) (clusters []*Cluster, total int64, err error) {

	clusters = make([]*Cluster, length)
	query := m.Db(ctx).Table(TABLE_NAME_CLUSTER)
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

	query = query.Where("deleted_at is NULL")
	return clusters, total, query.Count(&total).Order("updated_at desc").Offset(offset).Limit(length).Find(&clusters).Error
}

func (m *DAOClusterManager) CreateCluster(ctx context.Context, clusterReq Cluster) (cluster *Cluster, err error) {
	cluster = &Cluster{Entity: Entity{TenantId: clusterReq.TenantId},
		Name:            clusterReq.Name,
		DbPassword:      clusterReq.DbPassword,
		Type:            clusterReq.Type,
		Version:         clusterReq.Version,
		Tls:             clusterReq.Tls,
		Tags:            clusterReq.Tags,
		OwnerId:         clusterReq.OwnerId,
		CpuArchitecture: clusterReq.CpuArchitecture,
		Region:          clusterReq.Region,
		Exclusive:       clusterReq.Exclusive,
	}
	cluster.Code = generateEntityCode(clusterReq.Name)
	err = m.Db(ctx).Create(cluster).Error
	return
}

func (m *DAOClusterManager) SaveParameters(ctx context.Context, tenantId, clusterId, operatorId string, flowId uint, content string) (do *ParametersRecord, err error) {
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

	err = m.Db(ctx).Create(do).Error
	return
}

func (m *DAOClusterManager) GetCurrentParameters(ctx context.Context, clusterId string) (do *ParametersRecord, err error) {
	if "" == clusterId {
		return nil, errors.New(fmt.Sprintf("GetCurrentParameters has invalid parameter,clusterId:%s", clusterId))
	}
	do = &ParametersRecord{}
	return do, m.Db(ctx).Where("cluster_id = ?", clusterId).Last(do).Error
}

func (m *DAOClusterManager) DeleteBackupRecord(ctx context.Context, id uint) (record *BackupRecord, err error) {
	if id <= 0 {
		return nil, errors.New(fmt.Sprintf("DeleteBackupRecord has invalid parameter, Id: %d", id))
	}
	record = &BackupRecord{}
	record.ID = id
	err = m.Db(ctx).Where("id = ?", record.ID).Delete(record).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_RECORD, 0)
	return record, err
}

func (m *DAOClusterManager) SaveBackupRecord(ctx context.Context, record *dbpb.DBBackupRecordDTO) (do *BackupRecord, err error) {
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
	err = m.Db(ctx).Create(do).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_RECORD, 0)
	return
}

func (m *DAOClusterManager) UpdateBackupRecord(ctx context.Context, record *dbpb.DBBackupRecordDTO) error {
	err := m.Db(ctx).Model(&BackupRecord{}).Where("id = ?", record.Id).Updates(BackupRecord{
		Size:      record.GetSize(),
		BackupTso: record.GetBackupTso(),
		EndTime:   time.Unix(record.GetEndTime(), 0),
		Record: Record{
			TenantId:  record.GetTenantId(),
			UpdatedAt: time.Now(),
		}}).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_RECORD, 0)
	if err != nil {
		return err
	}
	return nil
}

func (m *DAOClusterManager) QueryBackupRecord(ctx context.Context, clusterId string, recordId int64) (*BackupRecordFetchResult, error) {
	record := BackupRecord{}
	err := m.Db(ctx).Table(TABLE_NAME_BACKUP_RECORD).Where("id = ? and cluster_id = ?", recordId, clusterId).First(&record).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_RECORD, 0)
	if err != nil {
		return nil, err
	}

	flow := FlowDO{}
	err = m.Db(ctx).Find(&flow, record.FlowId).Error
	m.HandleMetrics(TABLE_NAME_FLOW, 0)
	if err != nil {
		return nil, err
	}

	return &BackupRecordFetchResult{
		BackupRecord: &record,
		Flow:         &flow,
	}, nil
}
func (m *DAOClusterManager) ListBackupRecords(ctx context.Context, clusterId string, startTime, endTime int64, backupMode string, offset, length int) (dos []*BackupRecordFetchResult, total int64, err error) {
	records := make([]*BackupRecord, length)
	db := m.Db(ctx).Table(TABLE_NAME_BACKUP_RECORD).Where("cluster_id = ? and deleted_at is null", clusterId)
	if startTime > 0 {
		db = db.Where("start_time >= ?", time.Unix(startTime, 0))
	}
	if endTime > 0 {
		db = db.Where("end_time <= ?", time.Unix(endTime, 0))
	}
	if backupMode != "" {
		db = db.Where("backup_mode = ?", backupMode)
	}

	err = db.Count(&total).Order("id desc").Offset(offset).Limit(length).
		Find(&records).
		Error
	m.HandleMetrics(TABLE_NAME_BACKUP_RECORD, 0)
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
		err = m.Db(ctx).Find(&flows, flowIds).Error
		m.HandleMetrics(TABLE_NAME_FLOW, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("ListBackupRecord, query record failed, clusterId: %s, error: %s", clusterId, err.Error())
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

func (m *DAOClusterManager) SaveRecoverRecord(ctx context.Context, tenantId, clusterId, operatorId string,
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

	err = m.Db(ctx).Create(do).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_RECORD, 0)
	return
}

func (m *DAOClusterManager) SaveBackupStrategy(ctx context.Context, strategy *dbpb.DBBackupStrategyDTO) (*BackupStrategy, error) {
	strategyDO := BackupStrategy{
		ClusterId:  strategy.GetClusterId(),
		OperatorId: strategy.GetOperatorId(),
		Record: Record{
			TenantId: strategy.GetTenantId(),
		},
		BackupDate: strategy.GetBackupDate(),
		StartHour:  strategy.GetStartHour(),
		EndHour:    strategy.GetEndHour(),
	}
	result := m.Db(ctx).Table(TABLE_NAME_BACKUP_STRATEGY).Where("cluster_id = ?", strategy.ClusterId).First(&strategyDO)
	m.HandleMetrics(TABLE_NAME_BACKUP_STRATEGY, 0)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			strategyDO.CreatedAt = time.Now()
			strategyDO.UpdatedAt = time.Now()
			err := m.Db(ctx).Create(&strategyDO).Error
			m.HandleMetrics(TABLE_NAME_BACKUP_STRATEGY, 0)
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
		err := m.Db(ctx).Model(&strategyDO).Save(&strategyDO).Error
		m.HandleMetrics(TABLE_NAME_BACKUP_STRATEGY, 0)
		if err != nil {
			return nil, err
		}
	}
	return &strategyDO, nil
}

func (m *DAOClusterManager) QueryBackupStartegy(ctx context.Context, clusterId string) (*BackupStrategy, error) {
	strategyDO := BackupStrategy{}
	err := m.Db(ctx).Table(TABLE_NAME_BACKUP_STRATEGY).Where("cluster_id = ?", clusterId).First(&strategyDO).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_STRATEGY, 0)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &strategyDO, nil
}

func (m *DAOClusterManager) QueryBackupStartegyByTime(ctx context.Context, weekday string, startHour uint32) ([]*BackupStrategy, error) {
	var strategyListDO []*BackupStrategy
	err := m.Db(ctx).Table(TABLE_NAME_BACKUP_STRATEGY).Where("start_hour = ?", startHour).Where("backup_date like '%" + weekday + "%'").Find(&strategyListDO).Error
	m.HandleMetrics(TABLE_NAME_BACKUP_STRATEGY, 0)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return strategyListDO, nil
}

func (m *DAOClusterManager) CreateClusterRelation(ctx context.Context, request ClusterRelation) (result *ClusterRelation, err error) {
	if "" == request.TenantId || "" == request.SubjectClusterId || "" == request.ObjectClusterId || request.RelationType <= 0 {
		return nil, errors.New(fmt.Sprintf("CreateClusterRelation failed, has invalid parameter, tenantId: %s, subjectClusterId: %s, objectClusterId: %s, relationType: %d", request.TenantId, request.SubjectClusterId, request.ObjectClusterId, request.RelationType))
	}
	result = &ClusterRelation{
		Record:           Record{TenantId: request.TenantId},
		SubjectClusterId: request.SubjectClusterId,
		ObjectClusterId:  request.ObjectClusterId,
		RelationType:     request.RelationType,
	}
	return result, m.Db(ctx).Create(result).Error
}

func (m *DAOClusterManager) ListClusterRelationBySubjectId(ctx context.Context, subjectId string) (result []*ClusterRelation, err error) {
	if "" == subjectId {
		return nil, errors.New(fmt.Sprintf("ListClusterRelationBySubjectId failed, has invalid parameter, subjectId: %s", subjectId))
	}
	result = make([]*ClusterRelation, 0)
	return result, m.Db(ctx).Table(TABLE_NAME_CLUSTER_RELATION).Where("subject_cluster_id = ?", subjectId).Find(&result).Error
}

func (m *DAOClusterManager) ListClusterRelationByObjectId(ctx context.Context, objectId string) (result []*ClusterRelation, err error) {
	if "" == objectId {
		return nil, errors.New(fmt.Sprintf("ListClusterRelationByObjectId failed, has invalid parameter, objectId: %s", objectId))
	}
	result = make([]*ClusterRelation, 0)
	return result, m.Db(ctx).Table(TABLE_NAME_CLUSTER_RELATION).Where("object_cluster_id = ?", objectId).Find(&result).Error
}

func (m *DAOClusterManager) ListClusterRelationById(ctx context.Context, id uint) (result *ClusterRelation, err error) {
	if id <= 0 {
		return nil, errors.New(fmt.Sprintf("ListClusterRelationById failed, has invalid parameter, id: %d", id))
	}
	result = &ClusterRelation{}
	return result, m.Db(ctx).Model(result).Where("id = ?", id).First(&result).Error
}

func (m *DAOClusterManager) UpdateClusterRelation(ctx context.Context, relationId uint, subjectId, objectId string, relationType uint32) (result *ClusterRelation, err error) {
	if relationId <= 0 || "" == subjectId || "" == objectId || relationType <= 0 {
		return nil, errors.New(fmt.Sprintf("UpdateClusterRelation has invalid parameter, relationId: %d, subjectId: %s, objectId: %s, relationType: %d", relationId, subjectId, objectId, relationType))
	}
	result = &ClusterRelation{}
	return result, m.Db(ctx).Model(result).Where("id = ?", relationId).First(result).
		Update("subject_cluster_id", subjectId).
		Update("object_cluster_id", objectId).
		Update("relation_type", relationType).
		Error
}

func (m *DAOClusterManager) DeleteClusterRelation(ctx context.Context, relationId uint) (result *ClusterRelation, err error) {
	if relationId <= 0 {
		return nil, errors.New(fmt.Sprintf("DeleteClusterRelation has invalid parameter, relationId: %d", relationId))
	}
	result = &ClusterRelation{}
	result.ID = relationId
	err = m.Db(ctx).Where("id = ?", result.ID).Delete(result).Error
	return result, err
}
