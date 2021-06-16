package temp

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
	"github.com/pingcap/ticp/database/infrastructure/adapt"
	"github.com/pingcap/ticp/database/interfaces/proto"
)

// Make request
/*
	rsp, err := TicpClient.Hello(context.Background(), &pb.HelloRequest{
		Name: "Foo",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rsp.Greeting)
*/
var DbClient proto.DbService

func init() {
	appendToInitFpArray(initDbClient)
}

func initDbClient(srv micro.Service) {
	DbClient = proto.NewDbService(adapt.TICP_DB_SERVICE_NAME, srv.Client())
}
