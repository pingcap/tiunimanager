package temp

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
	"github.com/pingcap/ticp/micro-cluster/infrastructure/adapt"
	proto2 "github.com/pingcap/ticp/micro-cluster/proto"
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
var CommonClient proto2.CommonService

func init() {
	appendToInitFpArray(initCommonClient)
}

func initCommonClient(srv micro.Service) {
	CommonClient = proto2.NewCommonService(adapt.TICP_COMMON_SERVICE_NAME, srv.Client())
}
