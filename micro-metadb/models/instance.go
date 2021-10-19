package models

import (
	"context"
	"errors"
	"fmt"
)

type ComponentInstance struct {
	Entity
	ClusterId 			string	`gorm:"not null;type:varchar(22);default:null"`
	ComponentType		string 	`gorm:"not null;"`

	Role     string
	Spec     string 	`gorm:"not null;"`
	Version  string 	`gorm:"not null;"`
	HostId   string		`gorm:"type:varchar(22);default:null"`
	DiskId   string		`gorm:"type:varchar(22);default:null"`
	PortInfo string
	AllocRequestId string	`gorm:"not null;type:varchar(22);default:null"`
}

func (m *DAOClusterManager) ListComponentInstances(ctx context.Context, clusterId string) (componentInstances []*ComponentInstance, err error) {
	if clusterId == "" {
		return nil, errors.New(fmt.Sprintf("ListComponentInstances has invalid parameter, clusterId: %s", clusterId))
	}
	componentInstances = make([]*ComponentInstance, 0, 10)

	err = m.Db(ctx).Table(TABLE_NAME_COMPONENT_INSTANCE).Where("cluster_id = ?", clusterId).Find(&componentInstances).Error

	return componentInstances, err
}

func (m *DAOClusterManager) ListComponentInstancesByHost(ctx context.Context, hostId string) (componentInstances []*ComponentInstance, err error) {
	if hostId == "" {
		return nil, errors.New(fmt.Sprintf("ListComponentInstancesByHost has invalid parameter, hostId: %s", hostId))
	}
	componentInstances = make([]*ComponentInstance, 0, 0)

	err = m.Db(ctx).Table(TABLE_NAME_COMPONENT_INSTANCE).Where("host_id = ?", hostId).Find(&componentInstances).Error

	return componentInstances, err
}

func (m *DAOClusterManager) AddClusterComponentInstance(ctx context.Context, clusterId string, componentInstances []*ComponentInstance) ([]*ComponentInstance, error) {
	if clusterId == "" {
		return nil, errors.New(fmt.Sprintf("AddClusterComponentInstance has invalid parameter, clusterId: %s", clusterId))
	}
	if componentInstances == nil || len(componentInstances) == 0 {
		return nil, errors.New(fmt.Sprintf("AddClusterComponentInstance has invalid parameter, componentInstances: %v", componentInstances))
	}
	err := m.db.CreateInBatches(componentInstances, len(componentInstances)).Error
	return componentInstances, err
}