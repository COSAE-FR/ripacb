package acb

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func UserAgentMiddleware(c *gin.Context) {
	ua := c.GetHeader("User-Agent")
	if len(ua) > 0 {
		parts := strings.Split(ua, "/")
		if len(parts) == 2 {
			c.Set("client_platform", parts[0])
			c.Set("client_version", parts[1])
		}
	}
}
