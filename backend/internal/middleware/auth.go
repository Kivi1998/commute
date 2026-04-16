package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/pkg/auth"
	"github.com/haojia/commute/pkg/response"
)

const (
	ctxUserID = "user_id"
	ctxEmail  = "user_email"
)

// RequireAuth JWT 鉴权中间件
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			unauthorized(c, "请先登录")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		if token == "" {
			unauthorized(c, "token 为空")
			return
		}
		claims, err := auth.Parse(secret, token)
		if err != nil {
			unauthorized(c, "token 无效或已过期")
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxEmail, claims.Email)
		c.Next()
	}
}

func unauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(401, response.Body{
		Code:      response.CodeUnauthorized,
		Message:   msg,
		Data:      nil,
		RequestID: c.GetString("request_id"),
	})
}

// GetUserID 从 gin.Context 取出登录用户 id；若未设置返回 0
func GetUserID(c *gin.Context) int64 {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return 0
	}
	id, _ := v.(int64)
	return id
}
