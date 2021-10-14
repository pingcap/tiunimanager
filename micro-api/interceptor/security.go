package interceptor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func TlsHandler(addr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     addr,
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		}

		c.Next()
	}
}
