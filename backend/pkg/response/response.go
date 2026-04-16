package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeOK              = 0
	CodeBadRequest      = 40001
	CodeInvalidValue    = 40002
	CodeUnauthorized    = 40101
	CodeForbidden       = 40301
	CodeNotFound        = 40401
	CodeConflict        = 40901
	CodeBizValidation   = 42201
	CodeSoftLimit       = 42202
	CodeInternal        = 50001
	CodeAmapFailure     = 50201
	CodeAmapInvalid     = 50202
	CodeDoubaoFailure   = 50203
	CodeDoubaoInvalid   = 50204
)

type Body struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id,omitempty"`
}

func requestID(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:      CodeOK,
		Message:   "ok",
		Data:      data,
		RequestID: requestID(c),
	})
}

func Fail(c *gin.Context, httpStatus, code int, message string, data interface{}) {
	c.AbortWithStatusJSON(httpStatus, Body{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestID(c),
	})
}

func BadRequest(c *gin.Context, message string, data interface{}) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message, data)
}

func Internal(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeInternal, message, nil)
}
