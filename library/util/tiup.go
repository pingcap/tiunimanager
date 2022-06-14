/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 * @File: tiup
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/21
*******************************************************************************/

package util

import (
	"context"
	"fmt"
	"github.com/pingcap/tiunimanager/deployment"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models"
)

func GetTiUPHomeForComponent(ctx context.Context, tiUPComponent deployment.TiUPComponentType) string {
	var component string
	switch tiUPComponent {
	case deployment.TiUPComponentTypeEM:
		component = string(deployment.TiUPComponentTypeEM)
	default:
		component = string(deployment.TiUPComponentTypeDefault)
	}
	config, err := models.GetConfigReaderWriter().GetConfig(context.Background(), fmt.Sprintf("%s_tiup_home", component))
	if err != nil {
		framework.LogWithContext(ctx).Warnf("fail get tiup_home for %s: %s", component, err.Error())
		return ""
	} else {
		return config.ConfigValue
	}
}
