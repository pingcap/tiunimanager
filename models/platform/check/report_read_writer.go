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
 * @File: report_read_writer
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/14
*******************************************************************************/

package check

import (
	"context"
	"encoding/json"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
)

type ReaderWriter interface {
	CreateReport(ctx context.Context, report *CheckReport) (*CheckReport, error)
	DeleteReport(ctx context.Context, checkID string) error
	GetReport(ctx context.Context, checkID string) (interface{}, string, error)
	QueryReports(ctx context.Context) (map[string]structs.CheckReportMeta, error)
	UpdateReport(ctx context.Context, checkID, report string) error
	UpdateStatus(ctx context.Context, checkID, status string) error
}

type ReportReadWrite struct {
	dbCommon.GormDB
}

func NewReportReadWrite(db *gorm.DB) *ReportReadWrite {
	return &ReportReadWrite{
		dbCommon.WrapDB(db),
	}
}

func (rrw *ReportReadWrite) CreateReport(ctx context.Context, report *CheckReport) (*CheckReport, error) {
	if "" == report.Report {
		framework.LogWithContext(ctx).Errorf("create report %v, parameter invalid", report)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "create report %v, parameter invalid", report)
	}
	err := rrw.DB(ctx).Create(report).Error

	return report, err
}

func (rrw *ReportReadWrite) DeleteReport(ctx context.Context, checkID string) error {
	if "" == checkID {
		framework.LogWithContext(ctx).Errorf("delete report checkID %s, parameter invalid", checkID)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "delete report checkID %s, parameter invalid", checkID)
	}
	return rrw.DB(ctx).Where("id = ?", checkID).Unscoped().Delete(&CheckReport{}).Error
}

func (rrw *ReportReadWrite) GetReport(ctx context.Context, checkID string) (interface{}, string, error) {
	if "" == checkID {
		framework.LogWithContext(ctx).Errorf("get report checkID %s, parameter invalid", checkID)
		return nil, "", errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "get report checkID %s, parameter invalid", checkID)
	}

	report := &CheckReport{}
	err := rrw.DB(ctx).First(report, "id = ?", checkID).Error
	if err != nil {
		return nil, "", err
	}
	var info interface{}
	if report.Type == string(constants.PlatformReport) {
		info = &structs.CheckPlatformReportInfo{}
		err = json.Unmarshal([]byte(report.Report), info)
		if err != nil {
			return nil, "", err
		}
	} else if report.Type == string(constants.ClusterReport) {
		info = &structs.CheckClusterReportInfo{}
		err = json.Unmarshal([]byte(report.Report), info)
		if err != nil {
			return nil, "", err
		}
	} else {
		return nil, "", errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "report type %s not supported", report.Type)
	}
	return info, report.Type, nil
}

func (rrw *ReportReadWrite) QueryReports(ctx context.Context) (map[string]structs.CheckReportMeta, error) {
	reportMetas := make(map[string]structs.CheckReportMeta)

	var ids []string
	err := rrw.DB(ctx).Table("check_reports").Select("id").Find(&ids).Error
	if err != nil {
		return reportMetas, err
	}

	for _, id := range ids {
		report := &CheckReport{}
		err := rrw.DB(ctx).First(report, "id = ?", id).Error
		if err != nil {
			return reportMetas, err
		}
		_, ok := reportMetas[report.ID]
		if !ok {
			reportMetas[report.ID] = structs.CheckReportMeta{
				ID:        report.ID,
				Creator:   report.Creator,
				Type:      report.Type,
				Status:    report.Status,
				CreatedAt: report.CreatedAt,
			}
		}
	}
	return reportMetas, nil
}

func (rrw *ReportReadWrite) UpdateReport(ctx context.Context, checkID, report string) error {
	if "" == checkID || "" == report {
		framework.LogWithContext(ctx).Errorf("update %s report: %s, parameter invalid", checkID, report)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "update %s report: %s, parameter invalid", checkID, report)
	}

	return rrw.DB(ctx).Model(&CheckReport{}).Where("id = ?", checkID).Update("report", report).Error
}

func (rrw *ReportReadWrite) UpdateStatus(ctx context.Context, checkID, status string) error {
	if "" == checkID || "" == status {
		framework.LogWithContext(ctx).Errorf("update %s check report status %s, parameter invalid", checkID, status)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "update %s check report status %s, parameter invalid", checkID, status)
	}
	return rrw.DB(ctx).Model(&CheckReport{}).Where("id = ?", checkID).Update("status", status).Error
}
