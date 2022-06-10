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
	"encoding/json"
	"time"

	"github.com/pingcap-inc/tiunimanager/message"

	"github.com/pingcap-inc/tiunimanager/models/cluster/management"

	"github.com/pingcap-inc/tiunimanager/util/uuidutil"

	"github.com/pingcap-inc/tiunimanager/common/errors"

	"github.com/pingcap-inc/tiunimanager/library/framework"

	"gorm.io/gorm"

	dbCommon "github.com/pingcap-inc/tiunimanager/models/common"
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

func (m ParameterGroupReadWrite) CreateParameterGroup(ctx context.Context, pg *ParameterGroup, pgm []*ParameterGroupMapping, addParameters []message.ParameterInfo) (*ParameterGroup, error) {
	log := framework.LogWithContext(ctx)
	if pg.Name == "" || pg.ClusterSpec == "" || pg.ClusterVersion == "" {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "name or clusterSpec or cluster version is empty")
	}

	tx := m.DB(ctx).Begin()
	// gen id
	pg.ID = uuidutil.GenerateID()
	// insert parameter_group table
	pg.CreatedAt = time.Now()
	pg.UpdatedAt = time.Now()
	err := tx.Create(pg).Error
	if err != nil {
		log.Errorf("add param group err: %v, request param: %v", err.Error(), pg)
		tx.Rollback()
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_CREATE_ERROR, err.Error())
	}

	// batch insert parameter_group_mapping table
	for i := range pgm {
		pgm[i].ParameterGroupID = pg.ID
		pgm[i].CreatedAt = time.Now()
		pgm[i].UpdatedAt = time.Now()
	}
	err = tx.CreateInBatches(pgm, len(pgm)).Error
	if err != nil {
		log.Errorf("add param group map err: %v, request param map: %v", err.Error(), pgm)
		tx.Rollback()
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_CREATE_ERROR, err.Error())
	}

	// Add extended parameters when creating parameter groups
	if err = m.addParameters(dbCommon.CtxWithTransaction(ctx, tx), pg.ID, addParameters); err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return pg, nil
}

func (m ParameterGroupReadWrite) DeleteParameterGroup(ctx context.Context, parameterGroupId string) (err error) {
	log := framework.LogWithContext(ctx)
	if parameterGroupId == "" {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter group id is empty")
	}
	tx := m.DB(ctx).Begin()

	// check if the parameter group manages the cluster
	var total int64 = 0
	err = tx.Model(&management.Cluster{}).Where("parameter_group_id = ?", parameterGroupId).Count(&total).Error
	if err != nil {
		log.Errorf("query cluster count err: %v, request param id: %v", err.Error(), parameterGroupId)
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR, err.Error())
	}
	if total > 0 {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_RELATION_CLUSTER_NOT_DEL, "parameter group id: %s", parameterGroupId)
	}
	// delete parameter_group_mapping table
	err = tx.Where("parameter_group_id = ?", parameterGroupId).Delete(&ParameterGroupMapping{}).Error
	if err != nil {
		log.Errorf("delete param group map err: %v, request param id: %v", err.Error(), parameterGroupId)
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR, err.Error())
	}

	// delete parameter_group table
	err = tx.Where("id = ?", parameterGroupId).Delete(&ParameterGroup{}).Error
	if err != nil {
		log.Errorf("delete param group err: %v, request param id: %v", err.Error(), parameterGroupId)
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_ERROR, err.Error())
	}
	tx.Commit()
	return
}

func (m ParameterGroupReadWrite) UpdateParameterGroup(ctx context.Context, pg *ParameterGroup, pgm []*ParameterGroupMapping, addParameters []message.ParameterInfo, delParameters []string) (err error) {
	log := framework.LogWithContext(ctx)

	if pg.ID == "" {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter group id is empty")
	}

	// update parameter_group table, set updated_at
	tx := m.DB(ctx).Begin()
	pg.UpdatedAt = time.Now()
	err = tx.Where("id = ?", pg.ID).Updates(pg).Error
	if err != nil {
		log.Errorf("update param group err: %v, request param: %v", err.Error(), pg)
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_UPDATE_ERROR, err.Error())
	}

	// range update parameter_group_mapping table
	for i := range pgm {
		pgm[i].UpdatedAt = time.Now()
		err = tx.Where("parameter_group_id = ? and parameter_id = ?", pg.ID, pgm[i].ParameterID).Updates(pgm[i]).Error
		if err != nil {
			log.Errorf("update param group map err: %v, request param map: %v", err.Error(), pgm)
			tx.Rollback()
			return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_UPDATE_RELATION_PARAM_ERROR, err.Error())
		}
	}

	// Add extended parameters when creating parameter groups
	if err = m.addParameters(dbCommon.CtxWithTransaction(ctx, tx), pg.ID, addParameters); err != nil {
		tx.Rollback()
		return err
	}

	// Delete parameter mappings
	if len(delParameters) > 0 {
		for _, parameterId := range delParameters {
			// delete parameter_group_mapping table
			err = tx.Where("parameter_group_id = ? and parameter_id = ?", pg.ID, parameterId).Delete(&ParameterGroupMapping{}).Error
			if err != nil {
				tx.Rollback()
				return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR, err.Error())
			}
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
	if clusterVersion != "" {
		query = query.Where("cluster_version like '%" + clusterVersion + "%'")
	}
	if clusterSpec != "" {
		query = query.Where("cluster_spec = ?", clusterSpec)
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
		return nil, 0, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_QUERY_ERROR, err.Error())
	}
	return groups, total, err
}

func (m ParameterGroupReadWrite) GetParameterGroup(ctx context.Context, parameterGroupId, parameterName, instanceType string) (group *ParameterGroup, params []*ParamDetail, err error) {
	log := framework.LogWithContext(ctx)
	if parameterGroupId == "" {
		return nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter group id is empty")
	}

	group = &ParameterGroup{}
	err = m.DB(ctx).Where("id = ?", parameterGroupId).First(&group).Error

	if err != nil {
		log.Errorf("get param group err: %v", err.Error())
		return nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_QUERY_ERROR, err.Error())
	}

	params, err = m.QueryParametersByGroupId(ctx, parameterGroupId, parameterName, instanceType)
	if err != nil {
		log.Errorf("get param group err: %v", err.Error())
		return nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR, err.Error())
	}
	return
}

func (m ParameterGroupReadWrite) CreateParameter(ctx context.Context, parameter *Parameter) (*Parameter, error) {
	log := framework.LogWithContext(ctx)
	if parameter.Category == "" || parameter.Name == "" || parameter.InstanceType == "" {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "category or name or instanceType is empty")
	}
	// gen id
	parameter.ID = uuidutil.GenerateID()
	err := m.DB(ctx).Create(parameter).Error
	if err != nil {
		log.Errorf("add parameter err: %v, request param: %v", err.Error(), parameter)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_CREATE_ERROR, err.Error())
	}
	return parameter, err
}

func (m ParameterGroupReadWrite) DeleteParameter(ctx context.Context, parameterId string) (err error) {
	log := framework.LogWithContext(ctx)
	err = m.DB(ctx).Where("id = ?", parameterId).Delete(&Parameter{}).Error
	if err != nil {
		log.Errorf("delete parameter err: %v, request param id: %v", err.Error(), parameterId)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_CREATE_ERROR, err.Error())
	}
	return nil
}

func (m ParameterGroupReadWrite) UpdateParameter(ctx context.Context, parameter *Parameter) (err error) {
	log := framework.LogWithContext(ctx)

	if parameter.ID == "" {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter id is empty")
	}

	err = m.DB(ctx).Where("id = ?", parameter.ID).Updates(parameter).Error
	if err != nil {
		log.Errorf("update parameter err: %v, request param id: %v, param object: %v", err.Error(), parameter.ID, parameter)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_UPDATE_ERROR, err.Error())
	}
	return err
}

func (m ParameterGroupReadWrite) QueryParameters(ctx context.Context, offset, size int) (params []*Parameter, total int64, err error) {
	params = make([]*Parameter, 0)
	log := framework.LogWithContext(ctx)

	err = m.DB(ctx).Model(&Parameter{}).Count(&total).Offset(offset).Limit(size).Find(&params).Error
	if err != nil {
		log.Errorf("list parameters err: %v", err.Error())
		return nil, 0, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR, err.Error())
	}
	return params, total, err
}

func (m ParameterGroupReadWrite) QueryParametersByGroupId(ctx context.Context, parameterGroupId, parameterName, instanceType string) (params []*ParamDetail, err error) {
	log := framework.LogWithContext(ctx)

	query := m.DB(ctx).Model(&Parameter{}).
		Select("parameters.id, parameters.category, parameters.name, parameters.instance_type, parameters.system_variable, "+
			"parameters.type, parameters.unit, parameters.unit_options, parameters.range, parameters.range_type, parameters.has_reboot, parameters.has_apply, "+
			"parameters.update_source, parameters.read_only, parameters.description, parameter_group_mappings.default_value, parameter_group_mappings.note, "+
			"parameter_group_mappings.created_at, parameter_group_mappings.updated_at").
		Joins("left join parameter_group_mappings on parameters.id = parameter_group_mappings.parameter_id").
		Where("parameter_group_mappings.parameter_group_id = ?", parameterGroupId)

	// Fuzzy query by parameter name
	if parameterName != "" {
		query.Where("parameters.name like '%" + parameterName + "%'")
	}
	if instanceType != "" {
		query.Where("parameters.instance_type = ?", instanceType)
	}

	err = query.Order("parameters.instance_type desc").
		Scan(&params).Error
	if err != nil {
		log.Errorf("query parameters by group id err: %v", err.Error())
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR, err.Error())
	}
	return
}

func (m ParameterGroupReadWrite) GetParameter(ctx context.Context, parameterId string) (parameter *Parameter, err error) {
	log := framework.LogWithContext(ctx)
	parameter = &Parameter{}
	err = m.DB(ctx).Where("id = ?", parameterId).First(&parameter).Error
	if err != nil {
		log.Errorf("load parameter err: %v, request param id: %v", err.Error(), parameterId)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_DETAIL_ERROR, err.Error())
	}
	return parameter, err
}

func (m ParameterGroupReadWrite) ExistsParameter(ctx context.Context, category, name, instanceType string) (parameter *Parameter, err error) {
	log := framework.LogWithContext(ctx)
	parameter = &Parameter{}
	err = m.DB(ctx).Model(&Parameter{}).Where("category = ? and name = ? and instance_type = ?", category, name, instanceType).Find(&parameter).Error
	if err != nil {
		log.Errorf("exists parameter err: %v", err.Error())
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR, err.Error())
	}
	return parameter, err
}

func (m ParameterGroupReadWrite) addParameters(ctx context.Context, pgID string, addParameters []message.ParameterInfo) (err error) {
	if len(addParameters) > 0 {
		for _, addParameter := range addParameters {
			var rangeByte []byte
			if addParameter.Range != nil && len(addParameter.Range) > 0 {
				rangeByte, err = json.Marshal(addParameter.Range)
				if err != nil {
					return errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, err.Error())
				}
			}
			var unitOptionsByte []byte
			if addParameter.UnitOptions != nil && len(addParameter.UnitOptions) > 0 {
				unitOptionsByte, err = json.Marshal(addParameter.UnitOptions)
				if err != nil {
					return errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, err.Error())
				}
			}
			// add parameters
			parameter, err := m.CreateParameter(ctx, &Parameter{
				Category:       addParameter.Category,
				Name:           addParameter.Name,
				InstanceType:   addParameter.InstanceType,
				SystemVariable: addParameter.SystemVariable,
				Type:           addParameter.Type,
				Unit:           addParameter.Unit,
				UnitOptions:    string(unitOptionsByte),
				Range:          string(rangeByte),
				RangeType:      addParameter.RangeType,
				HasReboot:      addParameter.HasReboot,
				HasApply:       addParameter.HasApply,
				UpdateSource:   addParameter.UpdateSource,
				ReadOnly:       addParameter.ReadOnly,
				Description:    addParameter.Description,
			})
			if err != nil {
				return err
			}

			// insert parameter group mapping table
			err = m.DB(ctx).Create(ParameterGroupMapping{
				ParameterGroupID: pgID,
				ParameterID:      parameter.ID,
				DefaultValue:     addParameter.DefaultValue,
				Note:             addParameter.Note,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}).Error
			if err != nil {
				return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_CREATE_ERROR, err.Error())
			}
		}
	}
	return nil
}
