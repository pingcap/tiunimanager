package domain

import (
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_calculateNextBackupTime_case1(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Monday,Tuesday,Thursday"
	startHour := 10
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 20, 10, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case2(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Monday,Tuesday,Thursday,Sunday"
	startHour := 10
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 19, 10, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case3(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Tuesday,Friday"
	startHour := 11
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 21, 11, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case4(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Tuesday,Friday"
	startHour := 12
	nextBackupTime, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2021, 9, 17, 12, 0, 0, 0, time.Local), nextBackupTime)
}

func Test_calculateNextBackupTime_case5(t *testing.T) {
	now := time.Date(2021, 9, 17, 11, 0, 0, 0, time.Local)
	weekdayStr := "Tuesday,Weekday"
	startHour := 12
	_, err := calculateNextBackupTime(now, weekdayStr, startHour)
	assert.NotNil(t, err)
}

func Test_SaveBackupStrategyPreCheck_case1(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "14:00-16:00",
		BackupDate: "Monday,Tuesday,Thursday,Sunday",
	})
	assert.NoError(t, err)
}

func Test_SaveBackupStrategyPreCheck_case3(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "14:0016:00",
		BackupDate: "Monday,Tuesday,Thursday,Sunday",
	})
	assert.NotNil(t, err)
}

func Test_SaveBackupStrategyPreCheck_case4(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "18-16",
		BackupDate: "Monday,Tuesday,Thursday,Sunday",
	})
	assert.NotNil(t, err)
}

func Test_SaveBackupStrategyPreCheck_case5(t *testing.T) {
	err := SaveBackupStrategyPreCheck(&proto.OperatorDTO{}, &proto.BackupStrategy{
		Period:     "14:00-16:00",
		BackupDate: "Monday,weekday,Thursday,Sunday",
	})
	assert.NotNil(t, err)
}
