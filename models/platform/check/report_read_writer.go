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
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type ReaderWriter interface {
	CreateReport(ctx context.Context, report *CheckReport) (*CheckReport, error)
	DeleteReport(ctx context.Context, checkID string) error
	GetReport(ctx context.Context, checkID string) (structs.CheckReportInfo, error)
	QueryReports(ctx context.Context) (map[string]structs.CheckReportMeta, error)
	UpdateReport(ctx context.Context, checkID, report string) error
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
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "create report %v, parameter invalid", report)
	}
	err := rrw.DB(ctx).Create(report).Error

	return report, err
}

func (rrw *ReportReadWrite) DeleteReport(ctx context.Context, checkID string) error {
	if "" == checkID {
		framework.LogWithContext(ctx).Errorf("delete report checkID %s, parameter invalid", checkID)
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "delete report checkID %s, parameter invalid", checkID)
	}
	return rrw.DB(ctx).Where("id = ?", checkID).Unscoped().Delete(&CheckReport{}).Error
}

func (rrw *ReportReadWrite) GetReport(ctx context.Context, checkID string) (structs.CheckReportInfo, error) {
	info := structs.CheckReportInfo{}
	if "" == checkID {
		framework.LogWithContext(ctx).Errorf("get report checkID %s, parameter invalid", checkID)
		return info, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "get report checkID %s, parameter invalid", checkID)
	}

	report := &CheckReport{}
	err := rrw.DB(ctx).First(report, "id = ?", checkID).Error
	if err != nil {
		return info, err
	}

	err = json.Unmarshal([]byte(report.Report), &info)
	if err != nil {
		return info, err
	}

	return info, nil
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
				CreatedAt: report.CreatedAt,
			}
		}
	}
	return reportMetas, nil
}

func (rrw *ReportReadWrite) UpdateReport(ctx context.Context, checkID, report string) error {
	if "" == checkID || "" == report {
		framework.LogWithContext(ctx).Errorf("update %s report: %s, parameter invalid", checkID, report)
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "update %s report: %s, parameter invalid", checkID, report)
		return nil
	}

	return rrw.DB(ctx).Model(&CheckReport{}).Where("id = ?", checkID).Update("report", report).Error
}
