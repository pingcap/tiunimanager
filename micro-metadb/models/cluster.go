package models

import (
	"gorm.io/gorm"
)

type ClusterDO struct {
	Entity
	Demand 					ClusterDemandDO

	ClusterName 			string
	DbPassword 				string
	ClusterType 			string
	ClusterVersion 			string
	Tls 					bool
	Tags           			string
	OwnerId 				string
	Status 					uint
	CurrentFlowId			uint
}

type ClusterDemandDO struct {
	Record
	ClusterId 			string
	Content 			string		`gorm:"type:text"`
}

type TiUPConfigDO struct {
	Record
	ClusterId			string
	Content 			string		`gorm:"type:text"`
}

type Cluster struct {
	gorm.Model

	TenantId   uint 	`gorm:"size:32"`
	Name       string 	`gorm:"size:32"`
	DbPassword string 	`gorm:"size:32"`
	Version    string 	`gorm:"size:32"`
	Status     int 		`gorm:"size:32"`
	TidbCount  int  	`gorm:"size:32"`
	TikvCount  int  	`gorm:"size:32"`
	PdCount    int  	`gorm:"size:32"`
	ConfigID   uint 	`gorm:"size:32"`
}

type TiUPConfig struct {
	gorm.Model

	TenantId  uint   `gorm:"size:32"`
	ClusterId uint   `gorm:"size:32"`
	Latest    bool   `gorm:"size:32"`
	Content   string `gorm:"type:text"`
}

func FetchCluster(clusterId uint) (cluster *Cluster, err error) {
	MetaDB.First(cluster, clusterId)
	return
}

func FetchTiUPConfig(configId uint) (config *TiUPConfig, err error){
	MetaDB.First(config, configId)
	return
}

func CreateCluster(tenantId uint, name, dbPassword, Version string, status,tidbCount,tikvCount,pdCount int) (cluster Cluster, err error) {

	cluster.TenantId = tenantId
	cluster.Name = name
	cluster.DbPassword = dbPassword
	cluster.Version = Version
	cluster.Status = status
	cluster.TikvCount = tikvCount
	cluster.TidbCount = tidbCount
	cluster.PdCount = pdCount

	MetaDB.Create(&cluster)
	return
}

func UpdateClusterTiUPConfig(clusterId uint, configContent string) (cluster Cluster, config TiUPConfig, err error) {
	MetaDB.First(&cluster, clusterId)

	// 保存配置
	config.TenantId = cluster.TenantId
	config.ClusterId = cluster.ID
	config.Latest = true
	config.Content = configContent
	MetaDB.Create(&config)

	// 更新旧配置的Latest标签
	if cluster.ConfigID != 0 {
		MetaDB.First(new(TiUPConfig), cluster.ConfigID).Update("latest", false)
	}

	// 更新集群的配置版本
	cluster.ConfigID = config.ClusterId
	MetaDB.Save(cluster)

	return cluster, config, nil
}

func ListClusters(page, pageSize int) (clusters []Cluster, err error) {
	MetaDB.Find(&clusters).Offset((page - 1) * pageSize).Limit(pageSize)
	return
}