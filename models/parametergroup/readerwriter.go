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
 * @File: readerwriter.go
 * @Description: parameter group reader and writer interface define
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:31
*******************************************************************************/

package parametergroup

import (
	"context"

	"github.com/pingcap/tiunimanager/message"
)

// ReaderWriter
// @Description: parameter group reader and writer interface
type ReaderWriter interface {

	// CreateParameterGroup
	// @Description: create a new parameter group
	// @param ctx
	// @param pg
	// @param pgm
	// @param addParameters
	// @return *ParameterGroup
	// @return error
	CreateParameterGroup(ctx context.Context, pg *ParameterGroup, pgm []*ParameterGroupMapping, addParameters []message.ParameterInfo) (*ParameterGroup, error)

	// DeleteParameterGroup
	// @Description: delete a parameter group
	// @param ctx
	// @param parameterGroupId
	// @return err
	DeleteParameterGroup(ctx context.Context, parameterGroupId string) (err error)

	// UpdateParameterGroup
	// @Description: update a parameter group
	// @param ctx
	// @param pg
	// @param pgm
	// @param addParameters
	// @param delParameters
	// @return err
	UpdateParameterGroup(ctx context.Context, pg *ParameterGroup, pgm []*ParameterGroupMapping, addParameters []message.ParameterInfo, delParameters []string) (err error)

	// QueryParameterGroup
	// @Description: query parameter group list
	// @param ctx
	// @param name
	// @param clusterSpec
	// @param clusterVersion
	// @param dbType
	// @param hasDefault
	// @param offset
	// @param size
	// @return groups
	// @return total
	// @return err
	QueryParameterGroup(ctx context.Context, name, clusterSpec, clusterVersion string, dbType, hasDefault int, offset, size int) (groups []*ParameterGroup, total int64, err error)

	// GetParameterGroup
	// @Description: get parameter group by id
	// @param ctx
	// @param parameterGroupId
	// @param parameterName
	// @param instanceType
	// @return group
	// @return params
	// @return err
	GetParameterGroup(ctx context.Context, parameterGroupId, parameterName, instanceType string) (group *ParameterGroup, params []*ParamDetail, err error)

	// CreateParameter
	// @Description: create a new parameter
	// @param ctx
	// @param parameter
	// @return *Parameter
	// @return error
	CreateParameter(ctx context.Context, parameter *Parameter) (*Parameter, error)

	// DeleteParameter
	// @Description: delete a parameter
	// @param ctx
	// @param parameterId
	// @return err
	DeleteParameter(ctx context.Context, parameterId string) (err error)

	// UpdateParameter
	// @Description: update a parameter
	// @param ctx
	// @param parameter
	// @return err
	UpdateParameter(ctx context.Context, parameter *Parameter) (err error)

	// QueryParameters
	// @Description: query parameters list
	// @param ctx
	// @param offset
	// @param size
	// @return params
	// @return err
	QueryParameters(ctx context.Context, offset, size int) (params []*Parameter, total int64, err error)

	// QueryParametersByGroupId
	// @Description: query parameters by parameter group id
	// @param ctx
	// @param parameterGroupId
	// @param parameterName
	// @param instanceType
	// @return params
	// @return err
	QueryParametersByGroupId(ctx context.Context, parameterGroupId, parameterName, instanceType string) (params []*ParamDetail, err error)

	// GetParameter
	// @Description: get parameter by id
	// @param ctx
	// @param parameterId
	// @return parameter
	// @return err
	GetParameter(ctx context.Context, parameterId string) (parameter *Parameter, err error)

	// ExistsParameter
	// @Description: determine if the parameter exists
	// @param ctx
	// @param category
	// @param name
	// @param instanceType
	ExistsParameter(ctx context.Context, category, name, instanceType string) (parameter *Parameter, err error)
}
