package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-Id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = "req_" + uuid.NewString()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set(headerRequestID, rid)
		c.Next()
	}
}
