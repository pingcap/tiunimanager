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

package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/crypto/rand"
	"io/ioutil"
	"net/http"
	"time"
)

type loginRequest struct {
	Type     int    `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

const loginUrlSuffix string = "api/user/login"

func GetDashboardInfo(ctx context.Context, request *cluster.GetDashboardInfoReq) (*cluster.GetDashboardInfoResp, error) {
	meta, err := handler.Get(ctx, request.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster %s meta failed: %s", request.ClusterID, err.Error())
		return nil, err
	}

	url := getDashboardUrlFromCluser(meta)
	token, err := getLoginToken(ctx, url, "root", "") //todo: get username passwd from meta
	if err != nil {
		return nil, err
	}

	dashboard := &cluster.GetDashboardInfoResp{
		ClusterID: request.ClusterID,
		Url:       url,
		Token:     token,
	}

	return dashboard, nil
}

func getDashboardUrlFromCluser(meta *handler.ClusterMeta) string {
	instances := meta.GetPDClientAddresses()
	pdNum := len(instances)
	pdServer := instances[rand.Intn(pdNum)]
	pdClientPort := pdServer.Port
	if pdClientPort == 0 {
		pdClientPort = common.DefaultPDClientPort
	}
	return fmt.Sprintf("http://%s:%d/dashboard/", pdServer.IP, pdClientPort)
}

func getLoginToken(ctx context.Context, dashboardUrl, userName, password string) (string, error) {
	url := fmt.Sprintf("%s%s", dashboardUrl, loginUrlSuffix)
	body := &loginRequest{
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
	var loginResp loginResponse
	err = json.Unmarshal(data, &loginResp)
	if err != nil {
		return "", err
	}
	framework.LogWithContext(ctx).Infof("getLoginToken resp: %v", loginResp)

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
	framework.LogWithContext(ctx).Infof("%s URL : %s \n", http.MethodPost, req.URL.String())
	return client.Do(req)
}
