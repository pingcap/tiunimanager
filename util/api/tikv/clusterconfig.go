/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 * @File: clusterconfig.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/8 15:11
*******************************************************************************/

package tikv

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pingcap/tiunimanager/message/cluster"

	"github.com/pingcap/tiunimanager/library/framework"
	util "github.com/pingcap/tiunimanager/util/http"
)

const (
	TiKVApiUrl = "/config"
)

var ApiService TiKVApiService

type TiKVApiService interface {
	EditConfig(ctx context.Context, apiEditConfigReq cluster.ApiEditConfigReq) (bool, error)
	ShowConfig(ctx context.Context, showConfigReq cluster.ApiShowConfigReq) ([]byte, error)
}

type TiKVApiServiceImpl struct{}

func init() {
	ApiService = new(TiKVApiServiceImpl)
}

func (service *TiKVApiServiceImpl) EditConfig(ctx context.Context, editConfigReq cluster.ApiEditConfigReq) (bool, error) {
	framework.LogWithContext(ctx).Infof("request tikv api edit config, api req: %v", editConfigReq)

	url := fmt.Sprintf("http://%s:%d%s", editConfigReq.InstanceHost, editConfigReq.InstancePort, TiKVApiUrl)
	resp, err := util.PostJSON(url, editConfigReq.ConfigMap, editConfigReq.Headers)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("request tikv api edit config resp err: %v", err.Error())
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		errMsg := fmt.Sprintf("request TiKV api response status code: %v, content: %v", resp.StatusCode, string(b))
		framework.LogWithContext(ctx).Errorf("tikv api edit config, %s", errMsg)
		return false, errors.New(errMsg)
	}
	return resp.StatusCode == http.StatusOK, nil
}

func (service *TiKVApiServiceImpl) ShowConfig(ctx context.Context, showConfigReq cluster.ApiShowConfigReq) ([]byte, error) {
	framework.LogWithContext(ctx).Infof("request tikv api show config, api req: %v", showConfigReq)

	url := fmt.Sprintf("http://%s:%d%s", showConfigReq.InstanceHost, showConfigReq.InstancePort, TiKVApiUrl)
	resp, err := util.Get(url, showConfigReq.Params, showConfigReq.Headers)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("request tikv api show config resp err: %v", err.Error())
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("request TiKV api response status code: %v, content: %v", resp.StatusCode, string(b))
		framework.LogWithContext(ctx).Errorf("tikv api show config, %s", errMsg)
		return nil, errors.New(errMsg)
	}
	return b, nil
}
