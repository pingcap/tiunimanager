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
 ******************************************************************************/

/*******************************************************************************
 * @File: parametergroup_readwrite.go
 * @Description: parameter group read write implements
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 18:50
*******************************************************************************/

package parametergroup

import (
	"context"
	"time"

	"github.com/pingcap-inc/tiem/util/uuidutil"

	"github.com/pingcap-inc/tiem/common/errors"

	"github.com/pingcap-inc/tiem/library/framework"

	"gorm.io/gorm"

	dbCommon "github.com/pingcap-inc/tiem/models/common"
)

type ParameterGroupReadWrite struct {
	dbCommon.GormDB
}

func NewParameterGroupReadWrite(db *gorm.DB) *ParameterGroupReadWrite {
	m := &ParameterGroupReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m ParameterGroupReadWrite) CreateParameterGroup(ctx context.Context, pg *ParameterGroup, pgm []*ParameterGroupMapping) (*ParameterGroup, error) {
	log := framework.LogWithContext(ctx)
	if pg.Name == "" || pg.ClusterSpec == "" || pg.ClusterVersion == "" {
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, errors.TIEM_PARAMETER_INVALID.Explain())
	}

	tx := m.DB(ctx).Begin()
	// gen id
	pg.ID = uuidutil.GenerateID()
	// insert parameter_group table
	pg.CreatedAt = time.Now()
	pg.UpdatedAt = time.Now()
	err := m.DB(ctx).Create(pg).Error
	if err != nil {
		log.Errorf("add param group err: %v, request param: %v", err.Error(), pg)
		tx.Rollback()
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_CREATE_ERROR, errors.TIEM_PARAMETER_GROUP_CREATE_ERROR.Explain())
	}

	// batch insert parameter_group_mapping table
	for i := range pgm {
		pgm[i].ParameterGroupID = pg.ID
		pgm[i].CreatedAt = time.Now()
		pgm[i].UpdatedAt = time.Now()
	}
	err = m.DB(ctx).CreateInBatches(pgm, len(pgm)).Error
	if err != nil {
		log.Errorf("add param group map err: %v, request param map: %v", err.Error(), pgm)
		tx.Rollback()
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_CREATE_ERROR, errors.TIEM_PARAMETER_GROUP_CREATE_ERROR.Explain())
	}
	tx.Commit()
	return pg, nil
}

func (m ParameterGroupReadWrite) DeleteParameterGroup(ctx context.Context, parameterGroupId string) (err error) {
	log := framework.LogWithContext(ctx)
	if parameterGroupId == "" {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, errors.TIEM_PARAMETER_INVALID.Explain())
	}

	tx := m.DB(ctx).Begin()

	// delete parameter_group_mapping table
	err = m.DB(ctx).Where("parameter_group_id = ?", parameterGroupId).Delete(&ParameterGroupMapping{}).Error
	if err != nil {
		log.Errorf("delete param group map err: %v, request param id: %v", err.Error(), parameterGroupId)
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR, errors.TIEM_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR.Explain())
	}

	// delete parameter_group table
	err = m.DB(ctx).Where("id = ?", parameterGroupId).Delete(&ParameterGroup{}).Error
	if err != nil {
		log.Errorf("delete param group err: %v, request param id: %v", err.Error(), parameterGroupId)
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_DELETE_ERROR, errors.TIEM_PARAMETER_GROUP_DELETE_ERROR.Explain())
	}
	tx.Commit()
	return
}

func (m ParameterGroupReadWrite) UpdateParameterGroup(ctx context.Context, pg *ParameterGroup, pgm []*ParameterGroupMapping) (err error) {
	log := framework.LogWithContext(ctx)

	if pg.ID == "" {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, errors.TIEM_PARAMETER_INVALID.Explain())
	}

	tx := m.DB(ctx).Begin()
	// update parameter_group table, set updated_at
	pg.UpdatedAt = time.Now()
	err = m.DB(ctx).Where("id = ?", pg.ID).Updates(pg).Error
	if err != nil {
		log.Errorf("update param group err: %v, request param: %v", err.Error(), pg)
		tx.Rollback()
		return errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_UPDATE_ERROR, errors.TIEM_PARAMETER_GROUP_UPDATE_ERROR.Explain())
	}

	// range update parameter_group_mapping table
	for i := range pgm {
		pgm[i].UpdatedAt = time.Now()
		err = m.DB(ctx).Where("parameter_group_id = ? and parameter_id = ?", pg.ID, pgm[i].ParameterID).Updates(pgm[i]).Error
		if err != nil {
			log.Errorf("update param group map err: %v, request param map: %v", err.Error(), pgm)
			tx.Rollback()
			return errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_UPDATE_RELATION_PARAM_ERROR, errors.TIEM_PARAMETER_GROUP_UPDATE_RELATION_PARAM_ERROR.Explain())
		}
	}
	tx.Commit()
	return
}

func (m ParameterGroupReadWrite) QueryParameterGroup(ctx context.Context, name, clusterSpec, clusterVersion string, dbType, hasDefault int, offset, size int) (groups []*ParameterGroup, total int64, err error) {
	groups = make([]*ParameterGroup, 0)
	log := framework.LogWithContext(ctx)

	query := m.DB(ctx).Model(&ParameterGroup{})
	if name != "" {
		query = query.Where("name like '%" + name + "%'")
	}
	if clusterSpec != "" {
		query = query.Where("cluster_spec = ?", clusterSpec)
	}
	if clusterVersion != "" {
		query = query.Where("cluster_version = ?", clusterVersion)
	}
	if dbType > 0 {
		query = query.Where("db_type = ?", dbType)
	}
	if hasDefault > 0 {
		query = query.Where("has_default = ?", hasDefault)
	}
	err = query.Order("created_at desc").Count(&total).Offset(offset).Limit(size).Find(&groups).Error
	if err != nil {
		log.Errorf("list param group err: %v", err.Error())
		return nil, 0, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_QUERY_ERROR, errors.TIEM_PARAMETER_GROUP_QUERY_ERROR.Explain())
	}
	return groups, total, err
}

func (m ParameterGroupReadWrite) GetParameterGroup(ctx context.Context, parameterGroupId, parameterName string) (group *ParameterGroup, params []*ParamDetail, err error) {
	log := framework.LogWithContext(ctx)
	if parameterGroupId == "" {
		return nil, nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, errors.TIEM_PARAMETER_INVALID.Explain())
	}

	group = &ParameterGroup{}
	err = m.DB(ctx).Where("id = ?", parameterGroupId).First(&group).Error

	if err != nil {
		log.Errorf("get param group err: %v", err.Error())
		return nil, nil, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_QUERY_ERROR, errors.TIEM_PARAMETER_GROUP_QUERY_ERROR.Explain())
	}

	params, err = m.QueryParametersByGroupId(ctx, parameterGroupId, parameterName)
	if err != nil {
		log.Errorf("get param group err: %v", err.Error())
		return nil, nil, errors.NewErrorf(errors.TIEM_PARAMETER_QUERY_ERROR, errors.TIEM_PARAMETER_QUERY_ERROR.Explain())
	}
	return
}

func (m ParameterGroupReadWrite) CreateParameter(ctx context.Context, parameter *Parameter) (*Parameter, error) {
	log := framework.LogWithContext(ctx)
	if parameter.Category == "" || parameter.Name == "" || parameter.InstanceType == "" {
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, errors.TIEM_PARAMETER_INVALID.Explain())
	}
	// gen id
	parameter.ID = uuidutil.GenerateID()
	err := m.DB(ctx).Create(parameter).Error
	if err != nil {
		log.Errorf("add param err: %v, request param: %v", err.Error(), parameter)
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_CREATE_ERROR, errors.TIEM_PARAMETER_CREATE_ERROR.Explain())
	}
	return parameter, err
}

func (m ParameterGroupReadWrite) DeleteParameter(ctx context.Context, parameterId string) (err error) {
	log := framework.LogWithContext(ctx)
	err = m.DB(ctx).Where("id = ?", parameterId).Delete(&Parameter{}).Error
	if err != nil {
		log.Errorf("delete param err: %v, request param id: %v", err.Error(), parameterId)
		return errors.NewErrorf(errors.TIEM_PARAMETER_CREATE_ERROR, errors.TIEM_PARAMETER_CREATE_ERROR.Explain())
	}
	return nil
}

func (m ParameterGroupReadWrite) UpdateParameter(ctx context.Context, parameter *Parameter) (err error) {
	log := framework.LogWithContext(ctx)

	if parameter.ID == "" {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, errors.TIEM_PARAMETER_INVALID.Explain())
	}

	err = m.DB(ctx).Where("id = ?", parameter.ID).Updates(parameter).Error
	if err != nil {
		log.Errorf("update param err: %v, request param id: %v, param object: %v", err.Error(), parameter.ID, parameter)
		return errors.NewErrorf(errors.TIEM_PARAMETER_UPDATE_ERROR, errors.TIEM_PARAMETER_UPDATE_ERROR.Explain())
	}
	return err
}

func (m ParameterGroupReadWrite) QueryParametersByGroupId(ctx context.Context, parameterGroupId, parameterName string) (params []*ParamDetail, err error) {
	log := framework.LogWithContext(ctx)

	query := m.DB(ctx).Model(&Parameter{}).
		Select("parameters.id, parameters.category, parameters.name, parameters.instance_type, parameters.system_variable, "+
			"parameters.type, parameters.unit, parameters.range, parameters.has_reboot, parameters.has_apply, "+
			"parameters.update_source, parameters.description, parameter_group_mappings.default_value, parameter_group_mappings.note, "+
			"parameter_group_mappings.created_at, parameter_group_mappings.updated_at").
		Joins("left join parameter_group_mappings on parameters.id = parameter_group_mappings.parameter_id").
		Where("parameter_group_mappings.parameter_group_id = ?", parameterGroupId)

	// Fuzzy query by parameter name
	if parameterName != "" {
		query.Where("parameters.name like '%" + parameterName + "%'")
	}

	err = query.Order("parameters.instance_type desc").
		Scan(&params).Error
	if err != nil {
		log.Errorf("query parameters by group id err: %v", err.Error())
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_QUERY_ERROR, errors.TIEM_PARAMETER_QUERY_ERROR.Explain())
	}
	return
}

func (m ParameterGroupReadWrite) GetParameter(ctx context.Context, parameterId string) (parameter *Parameter, err error) {
	log := framework.LogWithContext(ctx)
	parameter = &Parameter{}
	err = m.DB(ctx).Where("id = ?", parameterId).First(&parameter).Error
	if err != nil {
		log.Errorf("load param err: %v, request param id: %v", err.Error(), parameterId)
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_DETAIL_ERROR, errors.TIEM_PARAMETER_DETAIL_ERROR.Explain())
	}
	return parameter, err
}
