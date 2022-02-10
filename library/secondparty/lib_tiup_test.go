/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: lib_tiup_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/9
*******************************************************************************/

package secondparty

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/tiup"
	"github.com/pingcap-inc/tiem/test/mockmodels/mocktiupconfig"
	asserts "github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTiUPHomeForComponent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tiupRW := mocktiupconfig.NewMockReaderWriter(ctrl)
		models.SetTiUPConfigReaderWriter(tiupRW)
		tiupRW.EXPECT().QueryByComponentType(gomock.Any(), gomock.Any()).Return(&tiup.TiupConfig{TiupHome: "testdata"}, nil).AnyTimes()

		result := GetTiUPHomeForComponent(context.TODO(), TiEMComponentTypeStr)
		asserts.Equal(t, "testdata", result)
		result = GetTiUPHomeForComponent(context.TODO(), ClusterComponentTypeStr)
		asserts.Equal(t, "testdata", result)
	})

	t.Run("fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tiupRW := mocktiupconfig.NewMockReaderWriter(ctrl)
		models.SetTiUPConfigReaderWriter(tiupRW)
		tiupRW.EXPECT().QueryByComponentType(gomock.Any(), gomock.Any()).Return(nil, errors.New("cannot get")).AnyTimes()

		result := GetTiUPHomeForComponent(context.TODO(), TiEMComponentTypeStr)
		asserts.Equal(t, "", result)
	})
}
