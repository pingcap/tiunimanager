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

/*******************************************************************************
 * @File: param_group.go
 * @Description: param group models
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/18 15:11
*******************************************************************************/

package models

import (
	"context"
	"fmt"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"

	"gorm.io/gorm"
)

type DAOParamManager struct {
	db *gorm.DB
}

func NewDAOParamManager(d *gorm.DB) *DAOParamManager {
	m := new(DAOParamManager)
	m.SetDB(d)
	return m
}

func (m *DAOParamManager) SetDB(d *gorm.DB) {
	m.db = d
}

func (m *DAOParamManager) DB(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

type ParamDO struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"uniqueIndex;not null;type:varchar(22);"`
	ComponentType string `gorm:"not null;"`
	Type          int    `gorm:"default:0"`
	Unit          string
	Range         string
	HasReboot     int `gorm:"default:0"`
	Source        int `gorm:"default:0"`
	Description   string
	CreatedAt     time.Time `gorm:"<-:create"`
	UpdatedAt     time.Time
}

func (do ParamDO) TableName() string {
	return TABLE_NAME_PARAM
}

type ParamGroupDO struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex;not null;type:varchar(22);"`
	ParentId   uint
	Spec       string `gorm:"not null;"`
	HasDefault int    `gorm:"default:1"`
	DbType     int    `gorm:"default:0"`
	GroupType  int    `gorm:"default:0"`
	Version    string
	Note       string
	CreatedAt  time.Time `gorm:"<-:create"`
	UpdatedAt  time.Time
}

func (do ParamGroupDO) TableName() string {
	return TABLE_NAME_PARAM_GROUP
}

type ParamGroupMapDO struct {
	ParamGroupId uint
	ParamId      uint
	DefaultValue string `gorm:"not null;"`
	Note         string
	CreatedAt    time.Time `gorm:"<-:create"`
	UpdatedAt    time.Time
}

func (do ParamGroupMapDO) TableName() string {
	return TABLE_NAME_PARAM_GROUP_MAP
}

type ClusterParamMapDO struct {
	ClusterId string `gorm:"not null;type:varchar(22);default:null"`
	ParamId   uint
	RealValue string    `gorm:"not null;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (do ClusterParamMapDO) TableName() string {
	return TABLE_NAME_CLUSTER_PARAM_MAP
}

func (m *DAOParamManager) AddParamGroup(ctx context.Context, pg *ParamGroupDO, pgm []*ParamGroupMapDO) (id uint, err error) {
	log := framework.LogWithContext(ctx)
	if pg.Name == "" || pg.Spec == "" || pg.Version == "" {
		err = fmt.Errorf(fmt.Sprintf("param valid err! require param name: %s, spec: %s, version: %s", pg.Name, pg.Spec, pg.Version))
		return
	}

	tx := m.DB(ctx).Begin()

	// insert param_groups table
	pg.CreatedAt = time.Now()
	pg.UpdatedAt = time.Now()
	err = m.DB(ctx).Create(pg).Error
	if err != nil {
		log.Errorf("add param group err: %v, request param: %v", err.Error(), pg)
		tx.Rollback()
		return
	}

	// batch insert param_group_map table
	for i := range pgm {
		pgm[i].ParamGroupId = pg.ID
		pgm[i].CreatedAt = time.Now()
		pgm[i].UpdatedAt = time.Now()
	}
	err = m.DB(ctx).CreateInBatches(pgm, len(pgm)).Error
	if err != nil {
		log.Errorf("add param group map err: %v, request param map: %v", err.Error(), pgm)
		tx.Rollback()
		return
	}
	tx.Commit()
	return pg.ID, err
}

func (m *DAOParamManager) UpdateParamGroup(ctx context.Context, id uint, name, spec, version, note string, pgm []*ParamGroupMapDO) (err error) {
	log := framework.LogWithContext(ctx)
	tx := m.DB(ctx).Begin()

	// update param_groups table
	pg := &ParamGroupDO{
		Name:      name,
		Spec:      spec,
		Version:   version,
		Note:      note,
		UpdatedAt: time.Now(),
	}
	err = m.DB(ctx).Where("id = ?", id).Updates(pg).Error
	if err != nil {
		log.Errorf("update param group err: %v, request param: %v", err.Error(), pg)
		tx.Rollback()
		return
	}

	// range update param_group_map table
	for i := range pgm {
		pgm[i].UpdatedAt = time.Now()
		err = m.DB(ctx).Where("param_group_id = ? and param_id = ?", id, pgm[i].ParamId).Updates(pgm[i]).Error
		if err != nil {
			log.Errorf("update param group map err: %v, request param map: %v", err.Error(), pgm)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return
}

func (m *DAOParamManager) DeleteParamGroup(ctx context.Context, id uint) (err error) {
	log := framework.LogWithContext(ctx)
	tx := m.DB(ctx).Begin()

	// delete param_group_map table
	err = m.DB(ctx).Where("param_group_id = ?", id).Delete(&ParamGroupMapDO{}).Error
	if err != nil {
		log.Errorf("delete param group map err: %v, request param id: %v", err.Error(), id)
		tx.Rollback()
		return
	}

	// delete param_groups table
	err = m.DB(ctx).Where("id = ?", id).Delete(&ParamGroupDO{}).Error
	if err != nil {
		log.Errorf("delete param group err: %v, request param id: %v", err.Error(), id)
		tx.Rollback()
		return
	}

	tx.Commit()
	return
}

func (m *DAOParamManager) ListParamGroup(ctx context.Context, name, spec, version string, dbType, hasDefault int32, offset, size int) (groups []*ParamGroupDO, total int64, err error) {
	log := framework.LogWithContext(ctx)

	groups = make([]*ParamGroupDO, size)
	query := m.DB(ctx).Table(TABLE_NAME_PARAM_GROUP)
	if name != "" {
		query = query.Where("name like '%" + name + "%'")
	}
	if spec != "" {
		query = query.Where("spec = ?", spec)
	}
	if version != "" {
		query = query.Where("version = ?", version)
	}
	if dbType > 0 {
		query = query.Where("db_type = ?", dbType)
	}
	if hasDefault > 0 {
		query = query.Where("has_default = ?", hasDefault)
	}
	err = query.Order("id asc").Count(&total).Offset(offset).Limit(size).Find(&groups).Error
	if err != nil {
		log.Errorf("list param group err: %v", err.Error())
		return
	}
	return groups, total, err
}

type ParamDetail struct {
	Id            uint
	Name          string
	ComponentType string
	Type          int
	Unit          string
	Range         string
	HasReboot     int
	Source        int
	Description   string
	DefaultValue  string
	Note          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (m *DAOParamManager) LoadParamGroup(ctx context.Context, id uint) (group *ParamGroupDO, params []*ParamDetail, err error) {
	log := framework.LogWithContext(ctx)
	group = &ParamGroupDO{}
	err = m.DB(ctx).Find(group, id).Error

	if err != nil {
		log.Errorf("load param group err: %v", err.Error())
		return
	}

	params, err = m.LoadParamsByGroupId(ctx, id)
	if err != nil {
		log.Errorf("load param group err: %v", err.Error())
		return
	}
	return
}

func (m *DAOParamManager) LoadParamsByGroupId(ctx context.Context, id uint) (params []*ParamDetail, err error) {
	log := framework.LogWithContext(ctx)

	err = m.DB(ctx).Model(&ParamDO{}).
		Select("params.id, params.name, params.component_type, params.type, params.unit, params.range, "+
			"params.has_reboot, params.source, params.description, param_group_map.default_value, param_group_map.note, "+
			"param_group_map.created_at, param_group_map.updated_at").
		Joins("left join param_group_map on params.id = param_group_map.param_id").
		Where("param_group_map.param_group_id = ?", id).Scan(&params).Error
	if err != nil {
		log.Errorf("load params err: %v", err.Error())
		return
	}
	return
}

func (m *DAOParamManager) ApplyParamGroup(ctx context.Context, id uint, clusterId string, params []*ClusterParamMapDO) (err error) {
	log := framework.LogWithContext(ctx)
	tx := m.DB(ctx).Begin()

	// delete cluster_param_map table
	err = m.DB(ctx).Where("cluster_id = ?", clusterId).Delete(&ClusterParamMapDO{}).Error
	if err != nil {
		log.Errorf("apply param group err: %v", err.Error())
		tx.Rollback()
		return
	}

	// update clusters table
	err = m.DB(ctx).Model(&Cluster{}).Where("id = ?", clusterId).Update("param_group_id", id).Error
	if err != nil {
		log.Errorf("apply param group err: %v", err.Error())
		tx.Rollback()
		return
	}

	// batch insert cluster_param_map table
	for i := range params {
		params[i].ClusterId = clusterId
		params[i].CreatedAt = time.Now()
		params[i].UpdatedAt = time.Now()
	}
	err = m.DB(ctx).CreateInBatches(params, len(params)).Error
	if err != nil {
		log.Errorf("apply param group map err: %v, request param map: %v", err.Error(), params)
		tx.Rollback()
		return
	}

	tx.Commit()
	return
}

type ClusterParamDetail struct {
	Id            uint
	Name          string
	ComponentType string
	Type          int
	Unit          string
	Range         string
	HasReboot     int
	Source        int
	Description   string
	DefaultValue  string
	Note          string
	RealValue     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (m *DAOParamManager) FindParamsByClusterId(ctx context.Context, clusterId string, offset int, size int) (paramGroupId uint, total int64, params []*ClusterParamDetail, err error) {
	log := framework.LogWithContext(ctx)
	cluster := Cluster{}
	err = m.DB(ctx).Where("id = ?", clusterId).First(&cluster).Error
	if err != nil {
		log.Errorf("find params by cluster id err: %v, request cluster id: %v", err.Error(), clusterId)
		return
	}
	paramGroupId = cluster.ParamGroupId

	err = m.DB(ctx).Model(&ClusterParamMapDO{}).
		Select("params.id, params.name, params.component_type, params.type, params.unit, params.range, "+
			"params.has_reboot, params.source, params.description, param_group_map.default_value, param_group_map.note, "+
			"cluster_param_map.real_value, cluster_param_map.created_at, cluster_param_map.updated_at").
		Joins("left join params on params.id = cluster_param_map.param_id").
		Joins("left join param_group_map on params.id = param_group_map.param_id").
		Where("cluster_param_map.cluster_id = ? and param_group_map.param_group_id = ?", cluster.ID, paramGroupId).
		Count(&total).Offset(offset).Limit(size).
		Scan(&params).Error
	if err != nil {
		log.Errorf("find params by cluster id err: %v", err.Error())
		return
	}
	return
}

func (m *DAOParamManager) UpdateClusterParams(ctx context.Context, clusterId string, params []*ClusterParamMapDO) (err error) {
	log := framework.LogWithContext(ctx)
	tx := m.DB(ctx).Begin()

	// batch update cluster_param_map table
	for i, param := range params {
		params[i].UpdatedAt = time.Now()
		err = m.DB(ctx).Model(&ClusterParamMapDO{}).
			Where("cluster_id = ? and param_id = ?", clusterId, param.ParamId).
			Update("real_value", param.RealValue).Error
		if err != nil {
			log.Errorf("update cluster params err: %v", err.Error())
			tx.Rollback()
			return
		}
	}

	tx.Commit()
	return
}

func (m *DAOParamManager) AddParam(ctx context.Context, p *ParamDO) (id uint, err error) {
	log := framework.LogWithContext(ctx)
	err = m.DB(ctx).Create(p).Error
	if err != nil {
		log.Errorf("add param err: %v, request param: %v", err.Error(), p)
	}
	return p.ID, err
}

func (m *DAOParamManager) DeleteParam(ctx context.Context, paramId uint) (err error) {
	log := framework.LogWithContext(ctx)
	err = m.DB(ctx).Where("id = ?", paramId).Delete(&ParamDO{}).Error
	if err != nil {
		log.Errorf("delete param err: %v, request param id: %v", err.Error(), paramId)
		return err
	}
	return nil
}

func (m *DAOParamManager) UpdateParam(ctx context.Context, paramId uint, p *ParamDO) (err error) {
	log := framework.LogWithContext(ctx)
	err = m.DB(ctx).Where("id = ?", paramId).Updates(p).Error
	if err != nil {
		log.Errorf("update param err: %v, request param id: %v, param object: %v", err.Error(), paramId, p)
	}
	return err
}

func (m *DAOParamManager) LoadParamById(ctx context.Context, paramId uint) (p ParamDO, err error) {
	log := framework.LogWithContext(ctx)
	p = ParamDO{}
	err = m.DB(ctx).Where("id = ?", paramId).First(&p).Error
	if err != nil {
		log.Errorf("find param err: %v, request param id: %v", err.Error(), paramId)
	}
	return p, err
}
