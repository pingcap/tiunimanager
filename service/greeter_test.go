package service

import (
	"context"
	greeterPb "tcp/proto/greeter"
	"testing"
)

func TestGreeterHello(t *testing.T) {
	var g Greeter
	req := greeterPb.Request{
		Name: "Test",
	}
	resp := greeterPb.Response{}
	err := g.Hello(context.Background(), &req, &resp)
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
