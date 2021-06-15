package service

import (
	"context"
	"testing"

	commonPb "github.com/pingcap/ticp/proto/common"
)

func TestTcpHello(t *testing.T) {
	var c Common
	req := commonPb.HelloRequest{
		Name: "Test",
	}
	resp := commonPb.HelloResponse{}
	err := c.Hello(context.Background(), &req, &resp)
	if err != nil {
		t.Error("got an error:", err)
		return
	}
	want := "Hello " + req.Name
	if resp.Greeting != want {
		t.Error("got an unwanted response:", resp.Greeting, "want:", want)
		return
	}
}
