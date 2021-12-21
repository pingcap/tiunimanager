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
 * @File: lib_api_deprecated.go
 * @Description: tidb component uses api to update parameters
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/2 13:51
*******************************************************************************/

package secondparty

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	util "github.com/pingcap-inc/tiem/library/util/http"

	"github.com/pingcap-inc/tiem/library/spec"

	"github.com/pingcap-inc/tiem/library/framework"
)

// Deprecated
func (secondMicro *SecondMicro) ApiEditConfig(ctx context.Context, apiEditConfigReq ApiEditConfigReq) (bool, error) {
	framework.LogWithContext(ctx).Infof("micro srv api edit config, api req: %v", apiEditConfigReq)
	switch apiEditConfigReq.TiDBClusterComponent {
	case spec.TiDBClusterComponent_TiDB:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, TiDBApiUrl)
		resp, err := util.PostForm(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request tidb api resp err: %v", err.Error())
			return false, err
		}
		framework.LogWithContext(ctx).Infof("apieditconfig, request tidb api resp status code: %v, content length: %v", resp.StatusCode, resp.ContentLength)
		return resp.StatusCode == http.StatusOK, nil
	case spec.TiDBClusterComponent_TiKV:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, TiKVApiUrl)
		resp, err := util.PostJSON(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request tikv api resp err: %v", err.Error())
			return false, err
		}
		framework.LogWithContext(ctx).Infof("apieditconfig, request tikv api resp status code: %v, content length: %v", resp.StatusCode, resp.ContentLength)
		return resp.StatusCode == http.StatusOK, nil
	case spec.TiDBClusterComponent_PD:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, PdApiUrl)
		resp, err := util.PostJSON(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request pd api resp err: %v", err.Error())
			return false, err
		}
		framework.LogWithContext(ctx).Infof("apieditconfig, request pd api resp status code: %v, content length: %v", resp.StatusCode, resp.ContentLength)
		return resp.StatusCode == http.StatusOK, nil
	default:
		framework.LogWithContext(ctx).Warnf("apieditconfig, component %s api update parameters are not supported", apiEditConfigReq.TiDBClusterComponent)
		return false, errors.New(fmt.Sprintf("Component %s API update parameters are not supported", apiEditConfigReq.TiDBClusterComponent))
	}
}
