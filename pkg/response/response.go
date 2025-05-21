// Package response 提供统一的HTTP响应处理功能，确保API返回标准格式的响应。
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 定义统一的API响应结构
type Response struct {
	Code      int         `json:"code"`            // HTTP状态码
	Message   string      `json:"message"`         // 响应消息
	Data      interface{} `json:"data"`            // 响应数据
	Error     string      `json:"error,omitempty"` // 错误详情
	Timestamp int64       `json:"timestamp"`       // 响应时间戳
}

// NewResponse 创建标准响应对象
func NewResponse(statusCode int, message string, data interface{}, err error) Response {
	resp := Response{
		Code:      statusCode,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}

// Success 返回HTTP 200成功响应
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, NewResponse(http.StatusOK, message, data, nil))
}

// Fail 返回指定HTTP状态码的失败响应
func Fail(c *gin.Context, statusCode int, message string, err error) {
	c.JSON(statusCode, NewResponse(statusCode, message, nil, err))
}

// BadRequest 返回400错误（请求参数错误）
func BadRequest(c *gin.Context, message string, err error) {
	Fail(c, http.StatusBadRequest, message, err)
}

// Unauthorized 返回401错误（未授权）
func Unauthorized(c *gin.Context, message string, err error) {
	Fail(c, http.StatusUnauthorized, message, err)
}

// Forbidden 返回403错误（禁止访问）
func Forbidden(c *gin.Context, message string, err error) {
	Fail(c, http.StatusForbidden, message, err)
}

// NotFound 返回404错误（资源不存在）
func NotFound(c *gin.Context, message string, err error) {
	Fail(c, http.StatusNotFound, message, err)
}

// InternalServerError 返回500错误（服务器内部错误）
func InternalServerError(c *gin.Context, message string, err error) {
	Fail(c, http.StatusInternalServerError, message, err)
}
