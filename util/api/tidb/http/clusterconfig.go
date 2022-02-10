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
 * @Description: edit config by tidb api
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/8 15:11
*******************************************************************************/

package http

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/library/framework"
	util "github.com/pingcap-inc/tiem/util/http"
)

const (
	TiDBApiUrl = "/settings"
)

var ApiService TiDBApiService

type TiDBApiService interface {
	ApiEditConfig(ctx context.Context, apiEditConfigReq cluster.ApiEditConfigReq) (bool, error)
}

type TiDBApiServiceImpl struct{}

func init() {
	ApiService = new(TiDBApiServiceImpl)
}

func (service *TiDBApiServiceImpl) ApiEditConfig(ctx context.Context, apiEditConfigReq cluster.ApiEditConfigReq) (bool, error) {
	framework.LogWithContext(ctx).Infof("micro srv api edit config, api req: %v", apiEditConfigReq)

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
}
