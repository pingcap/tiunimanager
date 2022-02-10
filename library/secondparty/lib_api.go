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
 * @File: lib_api
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/15
*******************************************************************************/

package secondparty

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	util "github.com/pingcap-inc/tiem/util/http"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/spec"
)

const (
	PdApiUrl     = "/pd/api/v1/config"
	TiKVApiUrl   = "/config"
	TiDBApiUrl   = "/settings"
	CdcLogApiUrl = "/api/v1/log"
)

func (manager *SecondPartyManager) ApiEditConfig(ctx context.Context, apiEditConfigReq ApiEditConfigReq) (bool, error) {
	framework.LogWithContext(ctx).Infof("micro srv api edit config, api req: %v", apiEditConfigReq)
	switch apiEditConfigReq.TiDBClusterComponent {
	case spec.TiDBClusterComponent_TiDB:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, TiDBApiUrl)
		resp, err := util.PostForm(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request tidb api resp err: %v", err.Error())
			return false, err
		}
		if resp.StatusCode != http.StatusOK {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}
			errMsg := fmt.Sprintf("request TiDB api response status code: %v, content: %v", resp.StatusCode, string(b))
			framework.LogWithContext(ctx).Errorf("apieditconfig, %s", errMsg)
			return false, errors.New(errMsg)
		}
		return resp.StatusCode == http.StatusOK, nil
	case spec.TiDBClusterComponent_TiKV:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, TiKVApiUrl)
		resp, err := util.PostJSON(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request tikv api resp err: %v", err.Error())
			return false, err
		}
		if resp.StatusCode != http.StatusOK {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}
			errMsg := fmt.Sprintf("request TiKV api response status code: %v, content: %v", resp.StatusCode, string(b))
			framework.LogWithContext(ctx).Errorf("apieditconfig, %s", errMsg)
			return false, errors.New(errMsg)
		}
		return resp.StatusCode == http.StatusOK, nil
	case spec.TiDBClusterComponent_PD:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, PdApiUrl)
		resp, err := util.PostJSON(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request pd api resp err: %v", err.Error())
			return false, err
		}
		if resp.StatusCode != http.StatusOK {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}
			errMsg := fmt.Sprintf("request PD api response status code: %v, content: %v", resp.StatusCode, string(b))
			framework.LogWithContext(ctx).Errorf("apieditconfig, %s", errMsg)
			return false, errors.New(errMsg)
		}
		return resp.StatusCode == http.StatusOK, nil
	case spec.TiDBClusterComponent_CDC:
		url := fmt.Sprintf("http://%s:%d%s", apiEditConfigReq.InstanceHost, apiEditConfigReq.InstancePort, CdcLogApiUrl)
		resp, err := util.PostJSON(url, apiEditConfigReq.ConfigMap, apiEditConfigReq.Headers)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apieditconfig, request pd api resp err: %v", err.Error())
			return false, err
		}
		if resp.StatusCode != http.StatusOK {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}
			errMsg := fmt.Sprintf("request CDC api response status code: %v, content: %v", resp.StatusCode, string(b))
			framework.LogWithContext(ctx).Errorf("apieditconfig, %s", errMsg)
			return false, errors.New(errMsg)
		}
		return resp.StatusCode == http.StatusOK, nil
	default:
		framework.LogWithContext(ctx).Warnf("apieditconfig, component %s api update parameters are not supported", apiEditConfigReq.TiDBClusterComponent)
		return false, fmt.Errorf("component %s API update parameters are not supported", apiEditConfigReq.TiDBClusterComponent)
	}
}
