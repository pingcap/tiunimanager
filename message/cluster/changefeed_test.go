/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

package cluster

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestChangeFeedTaskInfo_AcceptUpstreamUpdateTS(t *testing.T) {
	t.Run("current", func(t *testing.T) {
		currentUnix := time.Now().UnixNano() / int64(time.Millisecond)
		cf := ChangeFeedTaskInfo{}
		cf.AcceptUpstreamUpdateTS(uint64(currentUnix << 18))
		assert.Equal(t, fmt.Sprintf(strconv.FormatInt(int64(uint64(currentUnix << 18)), 10)), cf.UpstreamUpdateTS)
		assert.Equal(t, currentUnix, cf.UpstreamUpdateUnix)
	})
	t.Run("empty", func(t *testing.T) {
		cf := ChangeFeedTaskInfo{}
		cf.AcceptUpstreamUpdateTS(0)
		assert.Equal(t, "0", cf.UpstreamUpdateTS)
		assert.Equal(t, int64(0), cf.UpstreamUpdateUnix)
	})
}

func TestChangeFeedTaskInfo_AcceptDownstreamFetchTS(t *testing.T) {
	t.Run("current", func(t *testing.T) {
		currentUnix := time.Now().UnixNano() / int64(time.Millisecond)
		cf := ChangeFeedTaskInfo{}
		cf.AcceptDownstreamFetchTS(uint64(currentUnix << 18))
		assert.Equal(t, fmt.Sprintf(strconv.FormatInt(int64(uint64(currentUnix << 18)), 10)), cf.DownstreamFetchTS)
		assert.Equal(t, currentUnix, cf.DownstreamFetchUnix)
	})
	t.Run("empty", func(t *testing.T) {
		cf := ChangeFeedTaskInfo{}
		cf.AcceptDownstreamFetchTS(0)
		assert.Equal(t, "0", cf.DownstreamFetchTS)
		assert.Equal(t, int64(0), cf.DownstreamFetchUnix)
	})
}

func TestChangeFeedTaskInfo_AcceptDownstreamSyncTS(t *testing.T) {
	t.Run("current", func(t *testing.T) {
		currentUnix := time.Now().UnixNano() / int64(time.Millisecond)
		cf := ChangeFeedTaskInfo{}
		cf.AcceptDownstreamSyncTS(uint64(currentUnix << 18))
		assert.Equal(t, fmt.Sprintf(strconv.FormatInt(int64(uint64(currentUnix << 18)), 10)), cf.DownstreamSyncTS)
		assert.Equal(t, currentUnix, cf.DownstreamSyncUnix)
	})
	t.Run("empty", func(t *testing.T) {
		cf := ChangeFeedTaskInfo{}
		cf.AcceptDownstreamSyncTS(0)
		assert.Equal(t, "0", cf.DownstreamSyncTS)
		assert.Equal(t, int64(0), cf.DownstreamSyncUnix)
	})
}