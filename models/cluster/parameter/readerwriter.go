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
 * @File: readerwriter.go
 * @Description: parameter group reader and writer interface define
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:31
*******************************************************************************/

package parameter

import "context"

// ReaderWriter
// @Description: cluster parameter reader and writer interface
type ReaderWriter interface {

	// QueryClusterParameter
	// @Description: query cluster parameter
	// @param ctx
	// @param clusterId
	// @param parameterName
	// @param page
	// @param pageSize
	// @return paramGroupId
	// @return total
	// @return params
	// @return err
	QueryClusterParameter(ctx context.Context, clusterId, parameterName string, offset, size int) (paramGroupId string, params []*ClusterParamDetail, total int64, err error)

	// UpdateClusterParameter
	// @Description: update cluster parameters
	// @param ctx
	// @param clusterId
	// @param params
	// @return err
	UpdateClusterParameter(ctx context.Context, clusterId string, params []*ClusterParameterMapping) (err error)

	// ApplyClusterParameter
	// @Description: applying a parameter group to a cluster
	// @param ctx
	// @param parameterGroupId
	// @param clusterId
	// @param params
	// @return err
	ApplyClusterParameter(ctx context.Context, parameterGroupId string, clusterId string, params []*ClusterParameterMapping) (err error)
}
