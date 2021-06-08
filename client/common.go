package client

import (
	commonPb "github.com/pingcap/ticp/proto/common"
	"github.com/pingcap/ticp/service"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
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
var CommonClient commonPb.CommonService

func init() {
	appendToInitFpArray(initCommonClient)
}

func initCommonClient(srv micro.Service) {
	CommonClient = commonPb.NewCommonService(service.TICP_COMMON_SERVICE_NAME, srv.Client())
}
