package api

import (
	"github.com/pingcap/ticp/client"

	commonPb "github.com/pingcap/ticp/proto/common"

	"github.com/opentracing/opentracing-go"

	"github.com/gin-gonic/gin"
)

func Hello(c *gin.Context) {
	req := new(commonPb.HelloRequest)
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{
			"code": "400",
			"msg":  "bad param",
		})
		return
	}
	var resp *commonPb.HelloResponse
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
			resp, err = client.CommonClient.Hello(ctx, req)
		} else {
			resp, err = client.CommonClient.Hello(c, req)
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
