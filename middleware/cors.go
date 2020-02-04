// cors is a cross domain middleware
package middleware

import (
	"github.com/leaderwolfpipi/doris"
)

func Cors() doris.HandlerFunc {
	return func(c *doris.Context) error {
		c.Response.Writer.Header().Set(doris.HeaderAccessControlAllowOrigin, "*")
		c.Response.Writer.Header().Set(doris.HeaderAccessControlAllowCredentials, "true")
		c.Response.Writer.Header().Set(doris.HeaderAccessControlAllowHeaders, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token, Language, From")
		c.Response.Writer.Header().Set(doris.HeaderAccessControlAllowMethods, "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return nil
		}

		c.Next()
		return nil
	}
}
