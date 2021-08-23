package framework
//
//import (
//	"context"
//	"fmt"
//	"os"
//	"testing"
//
//	"github.com/asim/go-micro/v3/server"
//	"github.com/micro/cli/v2"
//	"github.com/pingcap/tiem/library/firstparty/util"
//	cluster "github.com/pingcap/tiem/micro-cluster/proto"
//	db "github.com/pingcap/tiem/micro-metadb/proto"
//	"github.com/pingcap/tiem/micro-metadb/service"
//	"gorm.io/gorm"
//)
//
//// TODO: to be a real and serious test
//
//func TestFramework(t *testing.T) {
//	f := NewDefaultFramework(WithServiceName("go.micro.tiem.test"))
//	util.AssertNoErr(f.InitLogger())
//	util.AssertNoErr(f.InitConfig())
//	util.AssertNoErr(f.InitClient())
//	cfg := f.MustGetConfig()
//	fmt.Println(f.Run(
//		WithRegistryAddrs(cfg.GetRegistryAddress()),
//		//WithRegistryAddrs([]string{"127.0.0.1:2379"}),
//		WithServiceListenAddr("127.0.0.1:1234"),
//		WithCertificateCrtFilePath(cfg.GetCertificateCrtFilePath()),
//		WithCertificateKeyFilePath(cfg.GetCertificateKeyFilePath()),
//		WithRegisterServiceHandlerFp(func(server server.Server, opts ...server.HandlerOption) error {
//			return db.RegisterTiEMDBServiceHandler(server, new(service.DBServiceHandler))
//		})),
//	)
//}
//
//type MetaDBFrameWork struct {
//	Framework
//	MetaDB *gorm.DB
//
//	ArgConfigPath string
//}
//
//func (p *MetaDBFrameWork) InitSqliteDB() error {
//	// init sqlite db and set p.MetaDB
//	// NIY!
//	return nil
//}
//
//func (p *MetaDBFrameWork) MustGetMetaDB() *gorm.DB {
//	db := p.MetaDB
//	util.AssertWithInfo(db != nil, "MetaDB should not be nil")
//	return db
//}
//
//func NewMetaDBFrameWork(opts ...Opt) *MetaDBFrameWork {
//	f := &MetaDBFrameWork{
//		Framework: NewDefaultFramework(opts...),
//	}
//	return f
//}
//
//func TestMetaDBFramework(t *testing.T) {
//	f := NewMetaDBFrameWork(WithServiceName("go.micro.tiem.db"))
//	f.AddOpts(WithArgFlags([]cli.Flag{
//		&cli.StringFlag{
//			Name:        "tiem-conf-file",
//			Value:       "",
//			Usage:       "specify the configure file path of tidb cloud platform",
//			Destination: &f.ArgConfigPath,
//		},
//	}))
//	util.AssertNoErr(f.ArgsParse())
//	util.AssertNoErr(f.InitLogger())
//	util.AssertNoErr(f.InitConfig())
//	util.AssertNoErr(f.InitClient())
//	util.AssertNoErr(f.InitSqliteDB())
//
//	log := f.MustGetLogger()
//	log.GetDefaultLogger().Infoln("ArgConfigPath:", f.ArgConfigPath)
//
//	util.AssertNoErr(f.SetupQuitSignalHandler(func() {
//		log.GetDefaultLogger().Infoln("recieved a quit signal and quit now")
//		os.Exit(0)
//	}))
//
//	cfg := f.MustGetConfig()
//	fmt.Println(f.Run(
//		WithRegistryAddrs(cfg.GetRegistryAddress()),
//		WithServiceListenAddr("127.0.0.1:1234"),
//		WithCertificateCrtFilePath(cfg.GetCertificateCrtFilePath()),
//		WithCertificateKeyFilePath(cfg.GetCertificateKeyFilePath()),
//		WithRegisterServiceHandlerFp(func(server server.Server, opts ...server.HandlerOption) error {
//			return db.RegisterTiEMDBServiceHandler(server, new(service.DBServiceHandler))
//		})),
//	)
//	ctx := log.NewContextWithField(context.Background(), "testKey", "testValue")
//	resp, err := f.MustGetClient().GetClusterClient().Login(
//		ctx, &cluster.LoginRequest{
//			AccountName: "whoami",
//			Password:    "you-know-what",
//		},
//	)
//	log.WithContext(ctx).Infoln("f.MustGetClient().GetClusterClient().Login yield:", resp, err)
//}
