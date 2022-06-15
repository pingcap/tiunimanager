/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
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
	Token  structs.SensitiveText `json:"token"`
	Expire time.Time             `json:"expire"`
}

const loginUrlSuffix string = "api/user/login"

func GetDashboardInfo(ctx context.Context, request cluster.GetDashboardInfoReq) (resp cluster.GetDashboardInfoResp, err error) {
	meta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		errMsg := fmt.Sprintf("get cluster %s meta failed: %s", request.ClusterID, err.Error())
		framework.LogWithContext(ctx).Errorf(errMsg)
		return resp, errors.WrapError(errors.TIUNIMANAGER_CLUSTER_NOT_FOUND, errMsg, err)
	}

	tidbUserInfo, err := meta.GetDBUserNamePassword(ctx, constants.Root)
	if err != nil {
		return resp, errors.WrapError(errors.TIUNIMANAGER_USER_NOT_FOUND,
			fmt.Sprintf("get cluster %s user info from meta failed: %s", request.ClusterID, err.Error()), err)
	}
	framework.LogWithContext(ctx).Infof("get cluster %s user info from meta", meta.Cluster.ID)

	url, err := getDashboardUrlFromCluster(ctx, meta)
	if err != nil {
		return resp, errors.WrapError(errors.TIUNIMANAGER_DASHBOARD_NOT_FOUND,
			fmt.Sprintf("find cluster %s dashboard failed: %s", request.ClusterID, err.Error()), err)
	}
	token, err := getLoginToken(ctx, url, tidbUserInfo.Name, tidbUserInfo.Password.Val)
	if err != nil {
		return resp, errors.WrapError(errors.TIUNIMANAGER_DASHBOARD_NOT_FOUND,
			fmt.Sprintf("get cluster %s dashboard login token failed: %s", request.ClusterID, err.Error()), err)
	}

	resp.ClusterID = request.ClusterID
	resp.Url = url
	resp.Token = structs.SensitiveText(token)
	return resp, nil
}

func getDashboardUrlFromCluster(ctx context.Context, meta *meta.ClusterMeta) (string, error) {
	pdAddress := meta.GetPDClientAddresses()
	if len(pdAddress) == 0 {
		framework.LogWithContext(ctx).Errorf("get pd address from meta failed, empty address")
		return "", fmt.Errorf("get pd address from meta failed, empty address")
	}
	framework.LogWithContext(ctx).Infof("get cluster %s tidb address from meta, %+v", meta.Cluster.ID, pdAddress)
	pdNum := len(pdAddress)
	pdServer := pdAddress[rand.Intn(pdNum)]
	pdClientPort := pdServer.Port

	return fmt.Sprintf("http://%s:%d/dashboard/", pdServer.IP, pdClientPort), nil
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

	return string(loginResp.Token), nil
}

func post(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("http post body to json failed")
		}
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, fmt.Errorf("new request is fail: %+v", req)
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
