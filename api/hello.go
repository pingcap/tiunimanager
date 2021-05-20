package api

import (
	"tcp/client"

	tcpPb "tcp/proto/tcp"

	"github.com/opentracing/opentracing-go"

	"github.com/gin-gonic/gin"
)

func Hello(c *gin.Context) {
	req := new(tcpPb.HelloRequest)
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{
			"code": "400",
			"msg":  "bad param",
		})
		return
	}
	var resp *tcpPb.HelloResponse
	var err error
	{
		v, existFlag := c.Get("ParentSpan")
		var parentSpan opentracing.Span
		var ok bool
		if existFlag {
			parentSpan, ok = v.(opentracing.Span)
		}
		if existFlag && ok {
			ctx := opentracing.ContextWithSpan(c, parentSpan)
			resp, err = client.TcpClient.Hello(ctx, req)
		} else {
			resp, err = client.TcpClient.Hello(c, req)
		}
	}
	if err != nil {
		c.JSON(500, gin.H{
			"code": "500",
			"msg":  err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"code": "200",
			"data": resp,
		})
	}
}
