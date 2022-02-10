/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: lib_tiup_v2
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"context"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
)

type TiUPComponentTypeStr string

const (
	ClusterComponentTypeStr TiUPComponentTypeStr = "cluster"
	DMComponentTypeStr      TiUPComponentTypeStr = "dm"
	TiEMComponentTypeStr    TiUPComponentTypeStr = "tiem"
	CTLComponentTypeStr     TiUPComponentTypeStr = "ctl"
	DefaultComponentTypeStr TiUPComponentTypeStr = "default"
)

func GetTiUPHomeForComponent(ctx context.Context, tiUPComponent TiUPComponentTypeStr) string {
	var component string
	switch tiUPComponent {
	case TiEMComponentTypeStr:
		component = string(TiEMComponentTypeStr)
	default:
		component = string(DefaultComponentTypeStr)
	}
	tiUPConfig, err := models.GetTiUPConfigReaderWriter().QueryByComponentType(context.Background(), component)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("fail get tiup_home for %s: %s", component, err.Error())
		return ""
	} else {
		return tiUPConfig.TiupHome
	}
}
