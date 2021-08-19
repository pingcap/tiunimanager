package framework

import (
	"context"
	"fmt"
	"testing"

	"github.com/asim/go-micro/v3/server"
	"github.com/pingcap/tiem/library/firstparty/util"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	"github.com/pingcap/tiem/micro-metadb/service"
	"gorm.io/gorm"
)

// TODO: to be a real and serious test

func TestFramework(t *testing.T) {
	f := NewFramework(WithServiceName("go.micro.tiem.test"))
	util.AssertNoErr(f.InitLogger())
	util.AssertNoErr(f.InitConfig())
	util.AssertNoErr(f.InitClient())
	cfg := f.MustGetConfig()
	fmt.Println(f.Run(
		WithRegistryAddrs(cfg.GetRegistryAddress()),
		//WithRegistryAddrs([]string{"127.0.0.1:2379"}),
		WithServiceListenAddr("127.0.0.1:1234"),
		WithCertificateCrtFilePath(cfg.GetCertificateCrtFilePath()),
		WithCertificateKeyFilePath(cfg.GetCertificateKeyFilePath()),
		WithRegisterServiceHandlerFp(func(server server.Server, opts ...server.HandlerOption) error {
			return db.RegisterTiEMDBServiceHandler(server, new(service.DBServiceHandler))
		})),
	)
}

type MetaDBFrameWork struct {
	Framework
	MetaDB *gorm.DB
}

func (p *MetaDBFrameWork) InitSqliteDB() error {
	// init sqlite db and set p.MetaDB
	// NIY!
	return nil
}

func (p *MetaDBFrameWork) MustGetMetaDB() *gorm.DB {
	db := p.MetaDB
	util.AssertWithInfo(db != nil, "MetaDB should not be nil")
	return db
}

func NewMetaDBFrameWork(opts ...Opt) *MetaDBFrameWork {
	f := &MetaDBFrameWork{
		Framework: NewFramework(opts...),
	}
	return f
}

func TestMetaDBFramework(t *testing.T) {
	f := NewMetaDBFrameWork(WithServiceName("go.micro.tiem.db"))
	util.AssertNoErr(f.InitLogger())
	util.AssertNoErr(f.InitConfig())
	util.AssertNoErr(f.InitClient())
	util.AssertNoErr(f.InitSqliteDB())
	cfg := f.MustGetConfig()
	fmt.Println(f.Run(
		WithRegistryAddrs(cfg.GetRegistryAddress()),
		//WithRegistryAddrs([]string{"127.0.0.1:2379"}),
		WithServiceListenAddr("127.0.0.1:1234"),
		WithCertificateCrtFilePath(cfg.GetCertificateCrtFilePath()),
		WithCertificateKeyFilePath(cfg.GetCertificateKeyFilePath()),
		WithRegisterServiceHandlerFp(func(server server.Server, opts ...server.HandlerOption) error {
			return db.RegisterTiEMDBServiceHandler(server, new(service.DBServiceHandler))
		})),
	)
	f.MustGetLogger().GetDefaultLogger().Infoln("info")
	log := f.MustGetLogger()
	ctx := log.NewContextWithField(context.Background(), "testKey", "testValue")
	resp, err := f.MustGetClient().GetClusterClient().Login(
		ctx, &cluster.LoginRequest{
			AccountName: "whoami",
			Password:    "you-know-what",
		},
	)
	log.WithContext(ctx).Infoln("f.MustGetClient().GetClusterClient().Login yield:", resp, err)
}
