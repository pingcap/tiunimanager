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
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	util "github.com/pingcap-inc/tiem/library/util/http"
	"net/http"
	"strconv"
)

const CDCApiUrl  = "/api/v1/changefeeds"

func (secondMicro *SecondMicro) CreateChangeFeedTask(ctx context.Context, req ChangeFeedCreateReq) (resp ChangeFeedCmdAcceptResp, err error) {
	framework.LogWithContext(ctx).Infof("micro srv create change feed task, req : %v", req)
	url := fmt.Sprintf("http://%s%s", CDCApiUrl)

	bytes, err := json.Marshal(&req)
	if err != nil {
		err = framework.WrapError(common.TIEM_MARSHAL_ERROR, "", err)
		return
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(bytes, data)

	if err != nil {
		err = framework.WrapError(common.TIEM_UNMARSHAL_ERROR, "", err)
		return
	}

	httpResp, err := util.PostJSON(url, data, map[string]string{})
	if err != nil {
		err = framework.WrapError(common.TIEM_CHANGE_FEED_CONNECT_ERROR, "", err)
		return
	}

	if strconv.Itoa(http.StatusAccepted) == httpResp.Status {
		resp.Succeed = true
	} else {
		resp.Succeed = false
		return
	}

	return
}

func (secondMicro *SecondMicro) UpdateChangeFeedTask(ctx context.Context, req ChangeFeedUpdateReq) (ChangeFeedCmdAcceptResp, error) {
	return ChangeFeedCmdAcceptResp{}, nil
}

func (secondMicro *SecondMicro) PauseChangeFeedTask(ctx context.Context, req ChangeFeedPauseReq) (ChangeFeedCmdAcceptResp, error) {
	return ChangeFeedCmdAcceptResp{}, nil
}

func (secondMicro *SecondMicro) ResumeChangeFeedTask(ctx context.Context, req ChangeFeedResumeReq) (ChangeFeedCmdAcceptResp, error) {
	return ChangeFeedCmdAcceptResp{}, nil
}

func (secondMicro *SecondMicro) DeleteChangeFeedTask(ctx context.Context, req ChangeFeedDeleteReq) (ChangeFeedCmdAcceptResp, error) {
	return ChangeFeedCmdAcceptResp{}, nil
}

func (secondMicro *SecondMicro) QueryChangeFeedTasks(ctx context.Context, req ChangeFeedQueryReq) (ChangeFeedQueryResp, error) {
	return ChangeFeedQueryResp{}, nil
}

func (secondMicro *SecondMicro) DetailChangeFeedTask(ctx context.Context, req ChangeFeedDetailReq) (ChangeFeedDetailResp, error) {
	return ChangeFeedDetailResp{}, nil
}



