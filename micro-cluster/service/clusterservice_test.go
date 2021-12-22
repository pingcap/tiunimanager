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

package service

import (
	"context"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

type TestErrorStruct struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

func (p TestErrorStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.NewError(errors.TIEM_MARSHAL_ERROR, "")
}

func Test_handleRequest(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		req := &clusterpb.RpcRequest{
			Request: "{\n    \"name\": \"aaa\",\n    \"type\": 4\n}",
		}
		resp := &clusterpb.RpcResponse{}
		data := TestStruct{}
		succeed := handleRequest(context.TODO(), req, resp, &data)
		assert.True(t, succeed)
	})
	t.Run("unmarshal error", func(t *testing.T) {
		req := &clusterpb.RpcRequest{
			Request: "\n    \"name\": \"aaa\",\n    \"type\": 4\n}",
		}
		resp := &clusterpb.RpcResponse{}
		data := TestStruct{}
		succeed := handleRequest(context.TODO(), req, resp, &data)
		assert.False(t, succeed)
		assert.Equal(t, int32(common.TIEM_UNMARSHAL_ERROR), resp.Code)
	})
}

func Test_handleResponse(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		resp := &clusterpb.RpcResponse{}
		data := TestStruct{
			Name: "aaa",
			Type: 4,
		}
		handleResponse(context.TODO(), resp, nil, data, &clusterpb.RpcPage{
			Page:     4,
			PageSize: 8,
			Total:    32,
		})

		assert.Equal(t, int32(0), resp.Code)
		assert.Equal(t, int32(4), resp.GetPage().Page)
		assert.Equal(t, int32(8), resp.GetPage().PageSize)
		assert.Equal(t, int32(32), resp.GetPage().Total)
		assert.Equal(t, "{\"name\":\"aaa\",\"type\":4}", resp.GetResponse())
	})

	t.Run("Error", func(t *testing.T) {
		resp := &clusterpb.RpcResponse{}
		data := TestStruct{
			Name: "aaa",
			Type: 4,
		}
		handleResponse(context.TODO(), resp, framework.SimpleError(common.TIEM_CLUSTER_NOT_FOUND), data, &clusterpb.RpcPage{
			Page:     4,
			PageSize: 8,
			Total:    32,
		})

		assert.Equal(t, int32(common.TIEM_CLUSTER_NOT_FOUND), resp.Code)
		assert.Empty(t, resp.GetPage())
		assert.Empty(t, resp.GetResponse())
	})

	t.Run("marshal error", func(t *testing.T) {
		resp := &clusterpb.RpcResponse{}

		handleResponse(context.TODO(), resp, nil, TestErrorStruct{}, &clusterpb.RpcPage{
			Page:     4,
			PageSize: 8,
			Total:    32,
		})

		assert.Equal(t, int32(errors.TIEM_MARSHAL_ERROR), resp.Code)
		assert.Empty(t, resp.GetPage())
		assert.Empty(t, resp.GetResponse())
	})
}

func Test_handlePanic(t *testing.T) {
	t.Run("succeed", func(t *testing.T) {
		resp := &clusterpb.RpcResponse{
		}

		func () {
			defer handlePanic(context.TODO(), "create", resp)
		}()
		assert.Equal(t, int32(0), resp.Code)
		assert.Empty(t, resp.Message)
	})

}
