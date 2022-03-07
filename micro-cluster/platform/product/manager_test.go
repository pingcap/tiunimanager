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
 * @File: manager_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import (
	"context"
	"github.com/pingcap-inc/tiem/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManager_QueryVendors(t *testing.T)  {
	t.Run("invalid parameters", func(t *testing.T) {
		_, err := NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{
			VendorIDs: []string{},
		})
		assert.Error(t, err)
		_, err = NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{
			VendorIDs: []string{"111"},
		})
		assert.Error(t, err)
		_, err = NewManager().QueryVendors(context.TODO(), message.QueryVendorInfoReq{
			VendorIDs: []string{"111", "Local"},
		})
		assert.Error(t, err)
	})
}