package service

import (
	"context"

	"github.com/pingcap/tcp/addon/logger"
	"github.com/pingcap/tcp/models"

	dbPb "github.com/pingcap/tcp/proto/db"
)

type Db struct{}

func (d *Db) FindUserByName(ctx context.Context, req *dbPb.FindUserByNameRequest, rsp *dbPb.FindUserByNameResponse) error {
	u, e := models.FindUserByName(ctx, req.Name)
	if e == nil {
		uPb := dbPb.User{
			Name: u.Name,
		}
		rsp.U = &uPb
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (d *Db) CheckUser(ctx context.Context, req *dbPb.CheckUserRequest, rsp *dbPb.CheckUserResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckUser"})
	log := logger.WithContext(ctx)
	e := models.CheckUser(ctx, req.Name, req.Passwd)
	if e == nil {
		log.Debug("CheckUser success")
	} else {
		log.Errorf("CheckUser failed: %s", e)
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (d *Db) CreateUser(ctx context.Context, req *dbPb.CreateUserRequest, rsp *dbPb.CreateUserResponse) error {
	e := models.CreateUser(ctx, req.Name, req.Passwd)
	if e == nil {
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}
