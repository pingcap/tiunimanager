package api

import (
	"tcp/client"

	greeterPb "tcp/proto/greeter"

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
	if resp, err := client.GreeterClient.Hello(c, req); err != nil {
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
