package models

import (
	"gorm.io/gorm"
)

type FlowDO struct {
	Data
	Name        string
	StatusAlias string
}

func (do FlowDO) TableName() string {
	return "flows"
}

type TaskDO struct {
	Data
	ParentType int8		`gorm:"default:0"`
	ParentId   string	`gorm:"default:null"`
	Name       string	`gorm:"default:null"`
	ReturnType string	`gorm:"default:null"`
	Parameters string	`gorm:"default:null"`
	Result     string	`gorm:"default:null"`
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

func CreateFlow(flowName string, statusAlias string, bizId string) (flow *FlowDO, err error) {
	flow = &FlowDO{
		Name:        flowName,
		StatusAlias: statusAlias,
		Data: Data{
			BizId: bizId,
		},
	}
	err = MetaDB.Create(&flow).Error
	return
}

func CreateTask(parentType int8, parentId string, taskName, bizId string, taskReturnType string, parameters, result string) (task *TaskDO, err error) {
	task = &TaskDO{
		ParentType: parentType,
		ParentId:   parentId,
		Name:       taskName,
		ReturnType: taskReturnType,

		Parameters: parameters,
		Result: result,
		Data: Data{
			BizId: bizId,
		},
	}
	err = MetaDB.Create(&task).Error
	return
}

func FetchFlow(id uint) (flow FlowDO, err error) {
	err = MetaDB.Find(&flow, id).Error
	return
}

func BatchFetchFlows(ids []uint) (flows []*FlowDO, err error) {
	err = MetaDB.Find(&flows, ids).Error
	return
}

func FetchFlowDetail(id uint) (flow *FlowDO, tasks []*TaskDO, err error) {
	flow = &FlowDO{}
	err = MetaDB.Find(flow, id).Error

	if err != nil {
		return
	}
	err = MetaDB.Where("parent_type = 0 and parent_id = ?", id).Find(&tasks).Error
	return
}

func FetchTask(id uint) (task TaskDO, err error) {
	err = MetaDB.Find(&task, id).Error
	return
}

func QueryTask(bizId string, taskType string) (tasks []TaskDO, err error) {
	err = MetaDB.Find(&tasks, "biz_id = ?" ,bizId).Error
	return
}

func UpdateFlowStatus(flow FlowDO) (FlowDO, error) {
	err := MetaDB.Model(&flow).Where("id = ?", flow.ID).Update("status", flow.Status).Error

	if err != nil {
		return flow, err
	}
	return flow,nil
}

func BatchSaveTasks(tasks []*TaskDO) (returnTasks []*TaskDO, err error) {
	err = MetaDB.Save(tasks).Error
	if err != nil {
		return tasks, err
	}
	return tasks,nil
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

func UpdateTask(task TaskDO)  (returnTask TaskDO, err error) {
	return task, MetaDB.Model(task).Where("id = ?", task.ID).Updates(task).First(&returnTask).Error
}
