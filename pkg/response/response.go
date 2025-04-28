package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code       int         `json:"code"`                 // 状态码
	Message    string      `json:"message"`              // 消息
	Data       interface{} `json:"data"`                 // 数据
	Error      string      `json:"error,omitempty"`      // 错误信息，仅在出错时返回
	Timestamp  int64       `json:"timestamp,omitempty"`  // 响应时间戳
}

// Success 成功响应
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Fail 失败响应
func Fail(c *gin.Context, statusCode int, message string, err error) {
	response := Response{
		Code:      statusCode,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
	}

	// 如果有错误信息，则添加到响应中
	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(statusCode, response)
}

// BadRequest 请求参数错误响应
func BadRequest(c *gin.Context, message string, err error) {
	Fail(c, http.StatusBadRequest, message, err)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string, err error) {
	Fail(c, http.StatusUnauthorized, message, err)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string, err error) {
	Fail(c, http.StatusForbidden, message, err)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string, err error) {
	Fail(c, http.StatusNotFound, message, err)
}

// InternalServerError 服务器内部错误响应
func InternalServerError(c *gin.Context, message string, err error) {
	Fail(c, http.StatusInternalServerError, message, err)
}
