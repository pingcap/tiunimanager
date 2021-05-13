package api

import (
	"tcp/client"

	greeterPb "tcp/proto/greeter"

	"github.com/opentracing/opentracing-go"

	"github.com/gin-gonic/gin"
)

func Greeter(c *gin.Context) {
	req := new(greeterPb.Request)
	if err := c.BindJSON(req); err != nil {
		c.JSON(200, gin.H{
			"code": "500",
			"msg":  "bad param",
		})
		return
	}
	var resp *greeterPb.Response
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
			resp, err = client.GreeterClient.Hello(ctx, req)
		} else {
			resp, err = client.GreeterClient.Hello(c, req)
		}
	}
	if err != nil {
		c.JSON(200, gin.H{
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
