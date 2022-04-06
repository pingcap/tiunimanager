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
 * @File: report_read_writer_test
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/14
*******************************************************************************/

package check

import (
	ctx "context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReportReadWrite_CreateReport(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		report := &CheckReport{Report: ""}
		_, err := testRW.CreateReport(ctx.TODO(), report)
		assert.Error(t, err)
	})
	t.Run("normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  "report",
			Creator: "admin",
			Status:  "Running",
		}
		got, err := testRW.CreateReport(ctx.TODO(), report)
		assert.NoError(t, err)
		assert.Equal(t, got.Report, report.Report)
		assert.Equal(t, got.Creator, report.Creator)
		assert.NotEmpty(t, got.ID)
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestReportReadWrite_DeleteReport(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.DeleteReport(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  "report",
			Creator: "admin",
			Status:  "Running",
		}
		got, err := testRW.CreateReport(ctx.TODO(), report)
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestReportReadWrite_GetReport(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		_, _, err := testRW.GetReport(ctx.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("platform normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  `{"tenants": {}, "hosts": {}}`,
			Creator: "admin",
			Type:    string(constants.PlatformReport),
			Status:  "Running",
		}
		got, err := testRW.CreateReport(ctx.TODO(), report)
		assert.NoError(t, err)
		reportInfo, reportType, err := testRW.GetReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
		assert.Equal(t, reportType, string(constants.PlatformReport))
		platformReportInfo := reportInfo.(*structs.CheckPlatformReportInfo)
		assert.NotNil(t, platformReportInfo)
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})

	t.Run("cluster normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  `{"clusterID": "123"}`,
			Creator: "admin",
			Type:    string(constants.ClusterReport),
			Status:  "Running",
		}

		got, err := testRW.CreateReport(ctx.TODO(), report)
		assert.NoError(t, err)
		reportInfo, reportType, err := testRW.GetReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
		assert.Equal(t, reportType, string(constants.ClusterReport))
		clusterReportInfo := reportInfo.(*structs.CheckClusterReportInfo)
		assert.Equal(t, clusterReportInfo.ID, "123")
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestReportReadWrite_QueryReports(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  `{"tenants": {}, "hosts": {}}`,
			Creator: "admin",
			Status:  "Running",
		}
		got, err := testRW.CreateReport(ctx.TODO(), report)
		assert.NoError(t, err)
		infos, err := testRW.QueryReports(ctx.TODO())
		assert.NoError(t, err)
		assert.Equal(t, len(infos), 1)
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestReportReadWrite_UpdateReport(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateReport(ctx.TODO(), "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  `{"tenants": {}, "hosts": {}}`,
			Creator: "admin",
			Status:  "Running",
		}
		got, err := testRW.CreateReport(ctx.TODO(), report)
		assert.NoError(t, err)
		err = testRW.UpdateReport(ctx.TODO(), got.ID, "report")
		assert.NoError(t, err)
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}

func TestReportReadWrite_UpdateStatus(t *testing.T) {
	t.Run("invalid parameter", func(t *testing.T) {
		err := testRW.UpdateStatus(ctx.TODO(), "", "")
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		report := &CheckReport{
			Report:  `{"tenants": {}, "hosts": {}}`,
			Creator: "admin",
			Status:  "Running",
		}
		got, err := testRW.CreateReport(ctx.TODO(), report)
		assert.NoError(t, err)
		err = testRW.UpdateStatus(ctx.TODO(), got.ID, "Completed")
		assert.NoError(t, err)
		err = testRW.DeleteReport(ctx.TODO(), got.ID)
		assert.NoError(t, err)
	})
}
