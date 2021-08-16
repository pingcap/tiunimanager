package framework

import (
	"fmt"
	"testing"

	"github.com/asim/go-micro/v3/server"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	"github.com/pingcap/tiem/micro-metadb/service"
)

// TODO: to be a real and serious test
func TestFramework(t *testing.T) {
	f := NewFramework(
		WithInitFp(func() error {
			fmt.Println("initFp executing")
			return nil // fmt.Errorf("err")
		}),
		WithCertificateCrtFilePath("rctPath"),
		WithCertificateKeyFilePath("keyPath"),
		WithServiceName("go.micro.tiem.db"),
		WithRegisterServiceHandlerFp(func(server server.Server, opts ...server.HandlerOption) error {
			return db.RegisterTiEMDBServiceHandler(server, new(service.DBServiceHandler))
		}),
	)
	fmt.Println(f.Run())
}
