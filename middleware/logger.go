package middleware

import (
	"doris"
	"fmt"
	"strconv"
	"time"
)

// 中间件函数的格式
func Logger() doris.HandlerFunc {
	return func(c *doris.Context) error {
		// 前向执行部分
		begin := time.Now()
		c.Next()

		// 计算处理时间
		elapsed := time.Since(begin)

		// 获取请求信息
		logs := strconv.Itoa(c.Response.Status()) + " | " +
			doris.HTTPErrorMessages[c.Response.Status()].Error() + " | " +
			fmt.Sprint(elapsed) + " | " +
			c.Request.Host + " | " +
			c.Request.RemoteAddr + " | " +
			c.Request.UserAgent() + " | " +
			c.Request.Method + " | " +
			c.Request.RequestURI
		l := c.Doris.Logger

		if c.Response.Status() >= 400 {
			l.Error(logs)
		} else {
			l.Info(logs)
		}

		return nil
	}
}
