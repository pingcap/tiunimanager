/*******************************************************************************
 * @File: report
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/14
*******************************************************************************/

package check

import (
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/gorm"
	"time"
)

type CheckReport struct {
	ID        string    `gorm:"primarykey"`
	Report    string    `gorm:"default:null;not null;"`
	Creator   string    `gorm:"default:null;not null;"`
	Status    string    `gorm:"default:null;not null;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (report *CheckReport) BeforeCreate(tx *gorm.DB) (err error) {
	if len(report.ID) == 0 {
		report.ID = uuidutil.GenerateID()
	}

	return nil
}
