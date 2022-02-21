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
 * @File: changefeed.go
 * @Description:
 * @Author: xieyujie@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/15 21:21
*******************************************************************************/

package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	util "github.com/pingcap-inc/tiem/util/http"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	CDCApiUrl  = "/api/v1/changefeeds"
)

type ChangeFeedCreateReq struct {
	CDCAddress       string   `json:"-"`
	ChangeFeedID     string   `json:"changefeed_id"`
	SinkURI          string   `json:"sink_uri"`
	StartTS          uint64   `json:"start_ts"`
	TargetTS         uint64   `json:"target_ts"`
	IgnoreTxnStartTS []uint64 `json:"ignore_txn_start_ts"`
	FilterRules      []string `json:"filter_rules"`
	SinkConfig       []string `json:"sink_config"`
	MounterWorkerNum int      `json:"mounter_worker_num"`
}

type ChangeFeedUpdateReq struct {
	ChangeFeedID     string   `json:"-"`
	CDCAddress       string   `json:"-"`
	SinkURI          string   `json:"sink_uri"`
	TargetTS         int64    `json:"target_ts"`
	FilterRules      []string `json:"filter_rules"`
	IgnoreTxnStartTS []uint64 `json:"ignore_txn_start_ts"`
	SinkConfig       []string `json:"sink_config"`
	MounterWorkerNum int      `json:"mounter_worker_num"`
}

type ChangeFeedPauseReq struct {
	ChangeFeedID string
	CDCAddress   string
}

type ChangeFeedDeleteReq struct {
	ChangeFeedID string
	CDCAddress   string
}

type ChangeFeedResumeReq struct {
	ChangeFeedID string
	CDCAddress   string
}

type ChangeFeedCmdAcceptResp struct {
	Accepted  bool
	Succeed   bool
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type ChangeFeedQueryReq struct {
	CDCAddress string
	State      string
}

type ChangeFeedQueryResp struct {
	Tasks []ChangeFeedInfo
}

type ChangeFeedDetailReq struct {
	CDCAddress   string
	ChangeFeedID string
}

type ChangeFeedDetailResp struct {
	ChangeFeedInfo
}

type ChangeFeedInfo struct {
	ChangeFeedID  string `json:"id"`
	State         string `json:"state"`
	CheckPointTSO uint64 `json:"checkpoint_tso"`
	ResolvedTSO   uint64 `json:"resolved_tso"`
	Error         string `json:"error"`
}

var CDCService ChangeFeedService

type ChangeFeedService interface {
	CreateChangeFeedTask(ctx context.Context, req ChangeFeedCreateReq) (resp ChangeFeedCmdAcceptResp, err error)
	UpdateChangeFeedTask(ctx context.Context, req ChangeFeedUpdateReq) (resp ChangeFeedCmdAcceptResp, err error)
	PauseChangeFeedTask(ctx context.Context, req ChangeFeedPauseReq) (resp ChangeFeedCmdAcceptResp, err error)
	ResumeChangeFeedTask(ctx context.Context, req ChangeFeedResumeReq) (resp ChangeFeedCmdAcceptResp, err error)
	DeleteChangeFeedTask(ctx context.Context, req ChangeFeedDeleteReq) (resp ChangeFeedCmdAcceptResp, err error)
	QueryChangeFeedTasks(ctx context.Context, req ChangeFeedQueryReq) (resp ChangeFeedQueryResp, err error)
	DetailChangeFeedTask(ctx context.Context, req ChangeFeedDetailReq) (ChangeFeedDetailResp, error)
}

type ChangeFeedServiceImpl struct {}

func init() {
	CDCService = new(ChangeFeedServiceImpl)
}

func (service *ChangeFeedServiceImpl) CreateChangeFeedTask(ctx context.Context, req ChangeFeedCreateReq) (resp ChangeFeedCmdAcceptResp, err error) {
	framework.LogWithContext(ctx).Infof("micro srv create change feed task, req : %v", req)
	url := fmt.Sprintf("http://%s%s", req.CDCAddress, CDCApiUrl)

	bytes, err := json.Marshal(&req)
	if err != nil {
		err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, "", err)
		return
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(bytes, &data)

	if err != nil {
		err = errors.WrapError(errors.TIEM_UNMARSHAL_ERROR, "", err)
		return
	}

	framework.LogWithContext(ctx).Infof("create change feed task, url = %s, data = %s", url, data)
	httpResp, err := util.PostJSON(url, data, map[string]string{})
	if err != nil {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		handleAcceptedCmd(ctx, req.CDCAddress, req.ChangeFeedID, &resp, func(info ChangeFeedInfo) bool {
			return constants.ChangeFeedStatusNormal.EqualCDCState(info.State)
		})

	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}

	return
}

func handleAcceptError(ctx context.Context, httpResp *http.Response, resp *ChangeFeedCmdAcceptResp) {
	resp.Accepted = false
	resp.Succeed = false

	respBody, err := ioutil.ReadAll(httpResp.Body)

	if err != nil {
		framework.LogWithContext(ctx).Errorf("read http response failed, %s", err.Error())
		resp.ErrorCode = ""
		resp.ErrorMsg = err.Error()
		return
	}

	err = json.Unmarshal(respBody, resp)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("unmarshal http response failed, %s", err.Error())
		resp.ErrorCode = ""
		resp.ErrorMsg = err.Error()
	}
}

var changeFeedRetryTimes = 20

func handleAcceptedCmd(ctx context.Context,
	address string, id string,
	resp *ChangeFeedCmdAcceptResp,
	assert func(info ChangeFeedInfo) bool) {
	for i := 0; i < changeFeedRetryTimes; i++ {
		time.Sleep(time.Millisecond * 500)
		task, err := getChangeFeedTaskByID(ctx, address, id)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("execute cdc command failed, err = %s", err.Error())
			resp.Succeed = false
			resp.ErrorMsg = err.Error()
			return
		}

		switch task.State {
		case strings.ToLower(constants.ChangeFeedStatusError.ToString()):
			resp.Succeed = false
			resp.ErrorMsg = task.ChangeFeedInfo.Error
		case strings.ToLower(constants.ChangeFeedStatusFailed.ToString()):
			resp.Succeed = false
			resp.ErrorMsg = task.ChangeFeedInfo.Error
		default:
			if assert(task.ChangeFeedInfo) {
				resp.Succeed = true
				return
			}
		}
	}
}

func (service *ChangeFeedServiceImpl) UpdateChangeFeedTask(ctx context.Context, req ChangeFeedUpdateReq) (resp ChangeFeedCmdAcceptResp, err error) {
	framework.LogWithContext(ctx).Infof("micro srv update change feed task, req : %v", req)
	url := fmt.Sprintf("http://%s%s/%s", req.CDCAddress, CDCApiUrl, req.ChangeFeedID)

	bytes, err := json.Marshal(&req)
	if err != nil {
		err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, "", err)
		return
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(bytes, &data)

	if err != nil {
		err = errors.WrapError(errors.TIEM_UNMARSHAL_ERROR, "", err)
		return
	}

	httpResp, err := util.PutJSON(url, data, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		handleAcceptedCmd(ctx, req.CDCAddress, req.ChangeFeedID, &resp, func(info ChangeFeedInfo) bool {
			return true
		})
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (service *ChangeFeedServiceImpl) PauseChangeFeedTask(ctx context.Context, req ChangeFeedPauseReq) (resp ChangeFeedCmdAcceptResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s/pause", req.CDCAddress, CDCApiUrl, req.ChangeFeedID)
	httpResp, err := util.PostJSON(url, map[string]interface{}{}, map[string]string{})
	if err != nil {
		err = errors.WrapError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, "", err)
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		handleAcceptedCmd(ctx, req.CDCAddress, req.ChangeFeedID, &resp, func(info ChangeFeedInfo) bool {
			return constants.ChangeFeedStatusStopped.EqualCDCState(info.State)
		})
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (service *ChangeFeedServiceImpl) ResumeChangeFeedTask(ctx context.Context, req ChangeFeedResumeReq) (resp ChangeFeedCmdAcceptResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s/resume", req.CDCAddress, CDCApiUrl, req.ChangeFeedID)
	httpResp, err := util.PostJSON(url, map[string]interface{}{}, map[string]string{})
	if err != nil {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
		handleAcceptedCmd(ctx, req.CDCAddress, req.ChangeFeedID, &resp, func(info ChangeFeedInfo) bool {
			return constants.ChangeFeedStatusNormal.EqualCDCState(info.State)
		})
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (service *ChangeFeedServiceImpl) DeleteChangeFeedTask(ctx context.Context, req ChangeFeedDeleteReq) (resp ChangeFeedCmdAcceptResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s", req.CDCAddress, CDCApiUrl, req.ChangeFeedID)
	httpResp, err := util.Delete(url)
	if err != nil {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
		return
	}

	if http.StatusAccepted == httpResp.StatusCode {
		resp.Accepted = true
	} else {
		handleAcceptError(ctx, httpResp, &resp)
	}
	return
}

func (service *ChangeFeedServiceImpl) QueryChangeFeedTasks(ctx context.Context, req ChangeFeedQueryReq) (resp ChangeFeedQueryResp, err error) {
	url := fmt.Sprintf("http://%s%s", req.CDCAddress, CDCApiUrl)
	params := map[string]string{}
	if req.State != "" {
		params["state"] = req.State
	}
	httpResp, err := util.Get(url, params, map[string]string{})

	if err != nil {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
		return
	}

	if http.StatusOK == httpResp.StatusCode {
		respBody, readErr := ioutil.ReadAll(httpResp.Body)
		if readErr != nil {
			framework.LogWithContext(ctx).Errorf("read http response failed, %s", readErr.Error())
			return
		}
		resp.Tasks = make([]ChangeFeedInfo, 0)
		readErr = json.Unmarshal(respBody, &resp.Tasks)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("unmarshal http response failed, %s", readErr.Error())
			return
		}
	} else {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
	}
	return
}

func (service *ChangeFeedServiceImpl) DetailChangeFeedTask(ctx context.Context, req ChangeFeedDetailReq) (ChangeFeedDetailResp, error) {
	return getChangeFeedTaskByID(ctx, req.CDCAddress, req.ChangeFeedID)
}


func getChangeFeedTaskByID(ctx context.Context, pdAddress, id string) (resp ChangeFeedDetailResp, err error) {
	url := fmt.Sprintf("http://%s%s/%s", pdAddress, CDCApiUrl, id)
	httpResp, err := util.Get(url, map[string]string{}, map[string]string{})

	if err != nil {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
		framework.LogWithContext(ctx).Errorf("get change feed task failed, %s", err.Error())
		return
	}

	if http.StatusOK == httpResp.StatusCode {
		respBody, readErr := ioutil.ReadAll(httpResp.Body)
		if readErr != nil {
			framework.LogWithContext(ctx).Errorf("read http response failed, %s", readErr.Error())
			return
		}
		readErr = json.Unmarshal(respBody, &resp)
		if readErr != nil {
			framework.LogWithContext(ctx).Errorf("unmarshal http response failed, %s", readErr.Error())
			return
		}
	} else {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, err.Error())
		framework.LogWithContext(ctx).Errorf(err.Error())
	}
	return
}

