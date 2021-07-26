package models

import "errors"

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

func UpdateClusterDemand(clusterId string, content string, tenantId string) (cluster *ClusterDO, err error) {
	record := &DemandRecordDO{
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

	cluster = &ClusterDO{}
	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("current_demand_id", record.ID).Error
	return
}

func UpdateClusterFlowId(clusterId string, flowId uint) (cluster *ClusterDO, err error) {
	if clusterId == ""{
		return nil, errors.New("cluster id is empty")
	}
	cluster = &ClusterDO{}

	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("current_flow_id", flowId).Error

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

	err = MetaDB.Model(cluster).Where("id = ?", clusterId).Update("current_tiup_config_id", record.ID).Error

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