package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/pkg/response"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				response.Internal(c, fmt.Sprintf("internal panic: %v", r))
			}
		}()
		c.Next()
	}
}
