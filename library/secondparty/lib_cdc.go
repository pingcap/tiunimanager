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
 ******************************************************************************/

package secondparty

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	util "github.com/pingcap-inc/tiem/library/util/http"
	"net/http"
	"time"
)

const CDCApiUrl  = "/api/v1/changefeeds"

func (secondMicro *SecondPartyManager) CreateChangeFeedTask(ctx context.Context, req ChangeFeedCreateReq) (resp ChangeFeedCmdAcceptResp, err error) {
	framework.LogWithContext(ctx).Infof("micro srv create change feed task, req : %v", req)
	url := fmt.Sprintf("http://%s%s", req.PD, CDCApiUrl)

	bytes, err := json.Marshal(&req)
	if err != nil {
		err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, "", err)
		return
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(bytes, data)

	if err != nil {
		err = errors.WrapError(errors.TIEM_UNMARSHAL_ERROR, "", err)
		return
	}

	httpResp, err := util.PostJSON(url, data, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		resp.Succeed = true
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}

	return
}

func handleAcceptError(ctx context.Context, httpResp *http.Response, resp *ChangeFeedCmdAcceptResp) {
	resp.Accepted = false
	resp.Succeed = false

	respBody := make([]byte, 0)
	length, err := httpResp.Body.Read(respBody)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("read http response failed, %s", err.Error())
		resp.ErrorCode = ""
		resp.ErrorMsg = err.Error()
		return
	}
	
	if length == 0 {
		framework.LogWithContext(ctx).Errorf("read http response empty")
		return
	}

	err = json.Unmarshal(respBody, resp)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("unmarshal http response failed, %s", err.Error())
		resp.ErrorCode = ""
		resp.ErrorMsg = err.Error()
	}
}

var changeFeedRetryTimes = 10

func handleAcceptedCmd(ctx context.Context,
	pdAddress string, id string,
	resp *ChangeFeedCmdAcceptResp,
	asserts func(info ChangeFeedInfo) bool) {
	for i := 0; i < changeFeedRetryTimes; i++ {
		time.Sleep(time.Second)
		task, err := getChangeFeedTaskByID(ctx, pdAddress, id)
		if err != nil {
			resp.Succeed = false
			resp.ErrorMsg = err.Error()
			return
		}

		if task.State == constants.ChangeFeedStatusError.ToString() ||
			task.State == constants.ChangeFeedStatusFailed.ToString() ||
			task.State == constants.ChangeFeedStatusFinished.ToString() {
			resp.Succeed = false
			return
		}

	}
}

func (secondMicro *SecondPartyManager) UpdateChangeFeedTask(ctx context.Context, req ChangeFeedUpdateReq) (resp ChangeFeedCmdAcceptResp, err error) {
	framework.LogWithContext(ctx).Infof("micro srv update change feed task, req : %v", req)
	url := fmt.Sprintf("http://%s%s/%s", req.PD, CDCApiUrl, req.ChangeFeedID)

	bytes, err := json.Marshal(&req)
	if err != nil {
		err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, "", err)
		return
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(bytes, data)

	if err != nil {
		err = errors.WrapError(errors.TIEM_UNMARSHAL_ERROR, "", err)
		return
	}

	httpResp, err := util.PostJSON(url, data, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		resp.Succeed = true
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (secondMicro *SecondPartyManager) PauseChangeFeedTask(ctx context.Context, req ChangeFeedPauseReq) (resp ChangeFeedCmdAcceptResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s/pause", req.PD, CDCApiUrl, req.ChangeFeedID)
	httpResp, err := util.PostJSON(url, map[string]interface{}{}, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		resp.Succeed = true
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (secondMicro *SecondPartyManager) ResumeChangeFeedTask(ctx context.Context, req ChangeFeedResumeReq) (resp ChangeFeedCmdAcceptResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s/resume", req.PD, CDCApiUrl, req.ChangeFeedID)
	httpResp, err := util.PostJSON(url, map[string]interface{}{}, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		resp.Succeed = true
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (secondMicro *SecondPartyManager) DeleteChangeFeedTask(ctx context.Context, req ChangeFeedDeleteReq) (resp ChangeFeedCmdAcceptResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s", req.PD, CDCApiUrl, req.ChangeFeedID)
	// todo delete
	httpResp, err := util.PostJSON(url, map[string]interface{}{}, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		resp.Succeed = true
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (secondMicro *SecondPartyManager) QueryChangeFeedTasks(ctx context.Context, req ChangeFeedQueryReq) (resp ChangeFeedQueryResp, err error) {
	url := fmt.Sprintf("http://%s%s", req.PD, CDCApiUrl)
	params := map[string]string{}
	if req.State != "" {
		params["state"] = req.State
	}
	httpResp, err := util.Get(url, params, map[string]string{})

	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusOK == httpResp.StatusCode {
		respBody := make([]byte, 0)
		_, err = httpResp.Body.Read(respBody)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("read http response failed, %s", err.Error())
			return
		}
		resp.Tasks = make([]ChangeFeedInfo, 0)
		err = json.Unmarshal(respBody, &resp.Tasks)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("unmarshal http response failed, %s", err.Error())
			return
		}
	} else {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
	}
	return
}

func (secondMicro *SecondPartyManager) DetailChangeFeedTask(ctx context.Context, req ChangeFeedDetailReq) (ChangeFeedDetailResp, error) {
	return getChangeFeedTaskByID(ctx, req.PD, req.ChangeFeedID)
}

func getChangeFeedTaskByID(ctx context.Context, pdAddress, id string) (resp ChangeFeedDetailResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s", pdAddress, CDCApiUrl, id)
	httpResp, err := util.Get(url, map[string]string{}, map[string]string{})

	if err != nil {
		// todo connect
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if http.StatusOK == httpResp.StatusCode {
		respBody := make([]byte, 0)
		_, err = httpResp.Body.Read(respBody)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("read http response failed, %s", err.Error())
			return
		}
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("unmarshal http response failed, %s", err.Error())
			return 
		}
	} else {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
	}
	return 
}

