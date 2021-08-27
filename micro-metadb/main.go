package main

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	dbService "github.com/pingcap-inc/tiem/micro-metadb/service"
)

func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.MetaDBService,
		defaultPortForLocal,
	)

	log := framework.GetLogger()

	f.PrepareService(func(service micro.Service) error {
		dao := new (models.DAOManager)
		dao.SetFramework(f)
		err := dao.InitDB()
		if nil != err {
			return err
		}

		handler := new (dbService.DBServiceHandler)
		handler.SetDao(dao)
		handler.SetLog(log)

		return dbPb.RegisterTiEMDBServiceHandler(service.Server(), handler) })

	err := f.StartService()
	if nil != err {
		log.Errorf("start meta-server failed, error: %v", err)
	}
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = common.DefaultMicroMetaDBPort
	}
	return nil
}