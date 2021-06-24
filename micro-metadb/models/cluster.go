package models

import (
	"gorm.io/gorm"
)

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

func CreateCluster(tenantId uint, name, dbPassword, Version string, status,tidbCount,tikvCount,pdCount int) (*Cluster, error) {
	cluster := &Cluster{
		TenantId: tenantId,
		Name: name,
		DbPassword: dbPassword,
		Version: Version,
		Status: status,
		TidbCount: tidbCount,
		TikvCount: tikvCount,
		PdCount: pdCount,
	}
	MetaDB.Create(cluster)
	return cluster, nil
}

func UpdateClusterTiUPConfig(clusterId uint, configContent string) (*Cluster, *TiUPConfig, error) {
	cluster := new(Cluster)
	MetaDB.First(cluster, clusterId)

	// 保存配置
	config := &TiUPConfig{
		TenantId:  cluster.TenantId,
		ClusterId: clusterId,
		Latest:    true,
		Content:   configContent,
	}
	MetaDB.Create(config)

	// 更新旧配置的Latest标签
	if cluster.ConfigID != 0 {
		MetaDB.First(new(TiUPConfig), cluster.ConfigID).Update("latest", false)
	}

	// 更新集群的配置版本
	cluster.ConfigID = config.ClusterId
	MetaDB.Save(cluster)

	return cluster, config, nil
}