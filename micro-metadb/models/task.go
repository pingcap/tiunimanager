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
 *                                                                            *
 ******************************************************************************/

package models

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type FlowDO struct {
	Data
	Name        string
	StatusAlias string
	Operator    string `gorm:"not null;type:varchar(22);default:null"`
}

func (do FlowDO) TableName() string {
	return "flows"
}

type TaskDO struct {
	Data
	ParentType int8   `gorm:"default:0"`
	ParentId   string `gorm:"default:null;index"`
	Name       string `gorm:"default:null"`
	ReturnType string `gorm:"default:null"`
	Parameters string `gorm:"default:null"`
	Result     string `gorm:"default:null"`
	StartTime  sql.NullTime
	EndTime    sql.NullTime
}

func (do TaskDO) TableName() string {
	return "tasks"
}

func (do *TaskDO) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (do *FlowDO) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func CreateFlow(db *gorm.DB, flowName string, statusAlias string, bizId string, Operator string) (flow *FlowDO, err error) {
	flow = &FlowDO{
		Name:        flowName,
		StatusAlias: statusAlias,
		Operator:    Operator,
		Data: Data{
			BizId: bizId,
		},
	}
	err = db.Create(&flow).Error
	return
}

func CreateTask(db *gorm.DB, parentType int8, parentId string, taskName, bizId string, taskReturnType string, parameters, result string, unixTime int64) (task *TaskDO, err error) {
	task = &TaskDO{
		ParentType: parentType,
		ParentId:   parentId,
		Name:       taskName,
		ReturnType: taskReturnType,
		StartTime:  sql.NullTime{
			Time:  time.Unix(unixTime, 0),
			Valid: unixTime == 0,
		},
		Parameters: parameters,
		Result:     result,
		Data: Data{
			BizId: bizId,
		},
	}
	err = db.Create(&task).Error
	return
}

func FetchFlow(db *gorm.DB, id uint) (flow FlowDO, err error) {
	err = db.Find(&flow, id).Error
	return
}

func ListFlows(db *gorm.DB, bizId, keyword string, status int, offset int, length int) (flows []*FlowDO, total int64, err error) {
	flows = make([]*FlowDO, length)
	query := db.Table(TABLE_NAME_FLOW)
	if bizId != "" {
		query = query.Where("biz_id = ?", bizId)
	}
	if keyword != "" {
		query = query.Where("name like '%" + keyword + "%'")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err = query.Order("id desc").Count(&total).Offset(offset).Limit(length).Find(&flows).Error
	return flows, total, err
}

func BatchFetchFlows(db *gorm.DB, ids []uint) (flows []*FlowDO, err error) {
	err = db.Find(&flows, ids).Error
	return
}

func FetchFlowDetail(db *gorm.DB, id uint) (flow *FlowDO, tasks []*TaskDO, err error) {
	flow = &FlowDO{}
	err = db.Find(flow, id).Error

	if err != nil {
		return
	}
	err = db.Where("parent_type = 0 and parent_id = ?", id).Find(&tasks).Error
	return
}

func FetchTask(db *gorm.DB, id uint) (task TaskDO, err error) {
	err = db.Find(&task, id).Error
	return
}

func QueryTask(db *gorm.DB, bizId string, taskType string) (tasks []TaskDO, err error) {
	err = db.Find(&tasks, "biz_id = ?", bizId).Error
	return
}

func UpdateFlowStatus(db *gorm.DB, flow FlowDO) (FlowDO, error) {
	err := db.Model(&flow).Where("id = ?", flow.ID).Update("status", flow.Status).Error

	if err != nil {
		return flow, err
	}
	return flow, nil
}

func BatchSaveTasks(db *gorm.DB, tasks []*TaskDO) (returnTasks []*TaskDO, err error) {
	err = db.Save(tasks).Error
	if err != nil {
		return tasks, err
	}
	return tasks, nil
}

//func UpdateFlowAndTasks(flow *FlowDO, tasks []*TaskDO) (*FlowDO, []*TaskDO, error) {
//	err := MetaDB.Model(&flow).Where("id = ?", flow.ID).Update("status", flow.Status).First(&flow).Error
//
//	if err != nil {
//		return flow, tasks, err
//	}
//
//	newTasks := make([]*TaskDO, 0, len(tasks))
//	needUpdateTasks := make([]*TaskDO, 0, len(tasks))
//	for _,t := range tasks {
//		if t.ID != 0 {
//			needUpdateTasks = append(needUpdateTasks, t)
//		} else {
//			newTasks = append(newTasks, t)
//		}
//	}
//
//	if len(newTasks) > 0 {
//		err = MetaDB.CreateInBatches(&newTasks, len(newTasks)).Error
//	}
//
//	for _, t := range needUpdateTasks {
//		w, err := UpdateTask(*t)
//	}
//
//	return flow, tasks, nil
//}

func UpdateTask(db *gorm.DB, task TaskDO) (returnTask TaskDO, err error) {
	return task, db.Model(task).Where("id = ?", task.ID).Updates(task).First(&returnTask).Error
}
