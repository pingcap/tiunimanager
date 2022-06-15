/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
 * @File: parameter_readwrite.go
 * @Description: cluster parameter read write implements
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 18:58
*******************************************************************************/

package parameter

import (
	"context"
	"time"

	"github.com/pingcap/tiunimanager/common/errors"

	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models/cluster/management"

	"gorm.io/gorm"

	dbCommon "github.com/pingcap/tiunimanager/models/common"
)

type ClusterParameterReadWrite struct {
	dbCommon.GormDB
}

func NewClusterParameterReadWrite(db *gorm.DB) *ClusterParameterReadWrite {
	m := &ClusterParameterReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m ClusterParameterReadWrite) QueryClusterParameter(ctx context.Context, clusterId, parameterName, instanceType string, offset, size int) (paramGroupId string, params []*ClusterParamDetail, total int64, err error) {
	log := framework.LogWithContext(ctx)
	cluster := management.Cluster{}
	err = m.DB(ctx).Where("id = ?", clusterId).First(&cluster).Error
	if err != nil {
		log.Errorf("find params by cluster id err: %v, request cluster id: %v", err.Error(), clusterId)
		err = errors.NewErrorf(errors.TIUNIMANAGER_CLUSTER_NOT_FOUND, err.Error())
		return
	}
	paramGroupId = cluster.ParameterGroupID

	query := m.DB(ctx).Model(&ClusterParameterMapping{}).
		Select("parameters.id, parameters.category, parameters.name, parameters.instance_type, parameters.system_variable, "+
			"parameters.type, parameters.unit, parameters.unit_options, parameters.range, parameters.range_type, "+
			"parameters.has_reboot, parameters.has_apply, parameters.update_source, parameters.read_only, parameters.description, "+
			"parameter_group_mappings.default_value, cluster_parameter_mappings.real_value, parameter_group_mappings.note, "+
			"cluster_parameter_mappings.created_at, cluster_parameter_mappings.updated_at").
		Joins("left join parameters on parameters.id = cluster_parameter_mappings.parameter_id").
		Joins("left join parameter_group_mappings on parameters.id = parameter_group_mappings.parameter_id").
		Where("cluster_parameter_mappings.cluster_id = ? and parameter_group_mappings.parameter_group_id = ?", cluster.ID, paramGroupId)

	// Fuzzy query by parameter name
	if parameterName != "" {
		query.Where("parameters.name like '%" + parameterName + "%'")
	}
	if instanceType != "" {
		query.Where("parameters.instance_type = ?", instanceType)
	}

	err = query.Order("parameters.instance_type desc").
		Count(&total).Offset(offset).Limit(size).
		Scan(&params).Error
	if err != nil {
		log.Errorf("find params by cluster id err: %v", err.Error())
		err = errors.Error(errors.TIUNIMANAGER_CLUSTER_PARAMETER_QUERY_ERROR)
		return
	}
	return
}

func (m ClusterParameterReadWrite) UpdateClusterParameter(ctx context.Context, clusterId string, params []*ClusterParameterMapping) (err error) {
	log := framework.LogWithContext(ctx)

	if clusterId == "" {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "cluster id is empty")
	}

	tx := m.DB(ctx).Begin()

	// batch update cluster_parameter_mapping table
	for i, param := range params {
		params[i].UpdatedAt = time.Now()
		err = tx.Model(&ClusterParameterMapping{}).
			Where("cluster_id = ? and parameter_id = ?", clusterId, param.ParameterID).
			Update("real_value", param.RealValue).Error
		if err != nil {
			log.Errorf("update cluster params err: %v", err.Error())
			tx.Rollback()
			return errors.NewErrorf(errors.TIUNIMANAGER_CLUSTER_PARAMETER_UPDATE_ERROR, err.Error())
		}
	}

	tx.Commit()
	return
}

func (m ClusterParameterReadWrite) ApplyClusterParameter(ctx context.Context, parameterGroupId string, clusterId string, params []*ClusterParameterMapping) (err error) {
	log := framework.LogWithContext(ctx)

	if clusterId == "" || parameterGroupId == "" {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "cluster id or parameter group id is empty")
	}

	tx := m.DB(ctx).Begin()

	// delete cluster_parameter_mapping table
	err = tx.Where("cluster_id = ?", clusterId).Delete(&ClusterParameterMapping{}).Error
	if err != nil {
		log.Errorf("apply param group err: %v", err.Error())
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR, err.Error())
	}

	// update clusters table
	c := management.Cluster{
		Entity: dbCommon.Entity{ID: clusterId},
	}
	err = tx.Model(&c).Update("parameter_group_id", parameterGroupId).Error
	if err != nil {
		log.Errorf("apply param group err: %v", err.Error())
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_CLUSTER_PARAMETER_UPDATE_ERROR, err.Error())
	}

	// batch insert cluster_parameter_mapping table
	for i := range params {
		params[i].ClusterID = clusterId
		params[i].CreatedAt = time.Now()
		params[i].UpdatedAt = time.Now()
	}
	err = tx.CreateInBatches(params, len(params)).Error
	if err != nil {
		log.Errorf("apply param group map err: %v, request param map: %v", err.Error(), params)
		tx.Rollback()
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_CREATE_RELATION_PARAM_ERROR, err.Error())
	}

	tx.Commit()
	return
}
