package service

import (
	"context"
	tcpPb "tcp/proto/tcp"
	"testing"
)

func TestTcpHello(t *testing.T) {
	var tcp Tcp
	req := tcpPb.HelloRequest{
		Name: "Test",
	}
	resp := tcpPb.HelloResponse{}
	err := tcp.Hello(context.Background(), &req, &resp)
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
