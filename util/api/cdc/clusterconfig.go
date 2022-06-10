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

package cdc

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pingcap-inc/tiunimanager/library/framework"
	"github.com/pingcap-inc/tiunimanager/message/cluster"
	util "github.com/pingcap-inc/tiunimanager/util/http"
)

const (
	CdcLogApiUrl = "/api/v1/log"
)

var ApiService CDCApiService

type CDCApiService interface {
	EditConfig(ctx context.Context, apiEditConfigReq cluster.ApiEditConfigReq) (bool, error)
}

type CDCApiServiceImpl struct{}

func init() {
	ApiService = new(CDCApiServiceImpl)
}

func (service *CDCApiServiceImpl) EditConfig(ctx context.Context, editConfigReq cluster.ApiEditConfigReq) (bool, error) {
	framework.LogWithContext(ctx).Infof("request cdc api edit config, api req: %v", editConfigReq)

	url := fmt.Sprintf("http://%s:%d%s", editConfigReq.InstanceHost, editConfigReq.InstancePort, CdcLogApiUrl)
	resp, err := util.PostJSON(url, editConfigReq.ConfigMap, editConfigReq.Headers)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("request cdc api edit config resp err: %v", err.Error())
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		errMsg := fmt.Sprintf("request CDC api response status code: %v, content: %v", resp.StatusCode, string(b))
		framework.LogWithContext(ctx).Errorf("cdc api edit config, %s", errMsg)
		return false, errors.New(errMsg)
	}
	return resp.StatusCode == http.StatusOK, nil
}
