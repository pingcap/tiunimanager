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
 * @File: productupgradepath
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/3
*******************************************************************************/

package upgrade

import (
	"context"

	"github.com/pingcap-inc/tiunimanager/common/constants"
)

type ReaderWriter interface {
	// Create
	// @Description: create a product upgrade path
	// @Receiver m
	// @Parameter ctx
	// @Parameter task
	// @return *ProductUpgradePath
	// @return error
	Create(ctx context.Context, upgradeType constants.UpgradeType, emProductIDType constants.EMProductIDType,
		srcVersion string, dstVersion string) (*ProductUpgradePath, error)

	// Get
	// @Description: get from id
	// @Receiver m
	// @Parameter ctx
	// @Parameter pathId
	// @return *ProductUpgradePath
	// @return error if task non-existent
	Get(ctx context.Context, id string) (*ProductUpgradePath, error)

	// QueryBySrcVersion
	// @Description: query ProductUpgradePath s for given srcVersion
	// @Receiver m
	// @Parameter ctx
	// @Parameter srcVersion
	// @return []*ProductUpgradePath
	// @return error if srcVersion is invalid
	QueryBySrcVersion(ctx context.Context, srcVersion string) (paths []*ProductUpgradePath, err error)

	// Delete
	// @Description: delete a product upgrade path
	// @Receiver m
	// @Parameter ctx
	// @Parameter id
	// @return err if task non-existent
	Delete(ctx context.Context, id string) (err error)
}
