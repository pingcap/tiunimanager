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

package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/crypto/rand"
	"io/ioutil"
	"net/http"
	"time"
)

type LoginRequest struct {
	Type     int    `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

type Dashboard struct {
	ClusterId string `json:"clusterId"`
	Url       string `json:"url"`
	Token     string `json:"token"`
}

var loginUrlSuffix string = "api/user/login"

func DescribeDashboard(ctx context.Context, ope *clusterpb.OperatorDTO, clusterId string) (*Dashboard, error) {
	//todo: check operator and clusterId
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil || clusterAggregation.Cluster == nil {
		return nil, errors.New("load cluster aggregation failed")
	}

	url := getDashboardUrlFromCluser(clusterAggregation)
	token, err := getLoginToken(ctx, url, "root", "") //todo: replace by real data
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		ClusterId: clusterId,
		Url:       url,
		Token:     token,
	}

	return dashboard, nil
}

func getDashboardUrlFromCluser(clusterAggregation *ClusterAggregation) string {
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel
	pdNum := len(configModel.PDServers)
	pdServer := configModel.PDServers[rand.Intn(pdNum)]
	pdClientPort := pdServer.ClientPort
	if pdClientPort == 0 {
		pdClientPort = DefaultPDClientPort
	}
	return fmt.Sprintf("http://%s:%d/dashboard/", pdServer.Host, pdClientPort)
}

func getLoginToken(ctx context.Context, dashboardUrl, userName, password string) (string, error) {
	url := fmt.Sprintf("%s%s", dashboardUrl, loginUrlSuffix)
	body := &LoginRequest{
		Username: userName,
		Password: password,
	}
	resp, err := post(ctx, url, body, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var loginResp LoginResponse
	err = json.Unmarshal(data, &loginResp)
	if err != nil {
		return "", err
	}
	getLoggerWithContext(ctx).Infof("getLoginToken resp: %v", loginResp)

	return loginResp.Token, nil
}

func post(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			return nil, errors.New("http post body to json failed")
		}
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, errors.New("new request is fail: %v \n")
	}
	req.Header.Set("Content-type", "application/json")
	//add headers

	for key, val := range headers {
		req.Header.Add(key, val)
	}

	//http client
	client := &http.Client{}
	getLoggerWithContext(ctx).Infof("%s URL : %s \n", http.MethodPost, req.URL.String())
	return client.Do(req)
}
