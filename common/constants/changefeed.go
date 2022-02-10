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

package constants

import (
	"github.com/pingcap-inc/tiem/common/errors"
	"strings"
)

type ChangeFeedStatus string

const (
	ChangeFeedStatusInitial  ChangeFeedStatus = "Initial"
	ChangeFeedStatusNormal   ChangeFeedStatus = "Normal"
	ChangeFeedStatusStopped  ChangeFeedStatus = "Stopped"
	ChangeFeedStatusFinished ChangeFeedStatus = "Finished"
	ChangeFeedStatusError    ChangeFeedStatus = "Error"
	ChangeFeedStatusFailed   ChangeFeedStatus = "Failed"
	ChangeFeedStatusUnknown  ChangeFeedStatus = "Unknown"
)

func (s ChangeFeedStatus) EqualCDCState(state string) bool {
	return strings.ToLower(s.ToString()) == state
}

func (s ChangeFeedStatus) IsFinal() bool {
	return ChangeFeedStatusFinished == s || ChangeFeedStatusFailed == s
}

func (s ChangeFeedStatus) ToString() string {
	return string(s)
}

func UnfinishedChangeFeedStatus() []ChangeFeedStatus{
	return []ChangeFeedStatus{
		ChangeFeedStatusInitial,ChangeFeedStatusNormal,ChangeFeedStatusStopped,ChangeFeedStatusError,
	}
}

func IsValidChangeFeedStatus(s string) bool {
	return ChangeFeedStatusInitial.ToString() == s ||
		ChangeFeedStatusNormal.ToString() == s ||
		ChangeFeedStatusStopped.ToString() == s ||
		ChangeFeedStatusFinished.ToString() == s ||
		ChangeFeedStatusError.ToString() == s ||
		ChangeFeedStatusFailed.ToString() == s
}

func ConvertChangeFeedStatus(s string) (status ChangeFeedStatus, err error) {
	if IsValidChangeFeedStatus(s) {
		return ChangeFeedStatus(s), nil
	} else {
		return ChangeFeedStatusUnknown, errors.NewError(errors.TIEM_PARAMETER_INVALID, "unexpected change feed status")
	}
}

type DownstreamType string

const (
	DownstreamTypeTiDB  DownstreamType = "tidb"
	DownstreamTypeKafka DownstreamType = "kafka"
	DownstreamTypeMysql DownstreamType = "mysql"
)
