package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hass-security/hass-security/webapp/backend/pkg/config"
)

func ConfigMiddleware(appConfig config.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("CONFIG", appConfig)
		c.Next()
	}
}
