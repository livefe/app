// Package response 提供统一的HTTP响应处理功能
// 包含成功和各种错误情况的标准化响应格式
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一API响应结构体
// 用于确保所有API返回一致的JSON格式
type Response struct {
	Code      int         `json:"code"`                // HTTP状态码
	Message   string      `json:"message"`             // 响应消息
	Data      interface{} `json:"data"`                // 响应数据，成功时包含实际数据，失败时为null
	Error     string      `json:"error,omitempty"`     // 错误详情，仅在出错时返回
	Timestamp int64       `json:"timestamp,omitempty"` // 响应时间戳（Unix时间戳）
}

// Success 返回成功响应
// 参数:
//   - c: Gin上下文
//   - message: 成功消息
//   - data: 响应数据，可以是任意类型
//
// 固定返回HTTP 200状态码
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Fail 返回失败响应
// 参数:
//   - c: Gin上下文
//   - statusCode: HTTP状态码
//   - message: 错误消息
//   - err: 错误对象，可以为nil
//
// 返回指定的HTTP状态码和错误信息
func Fail(c *gin.Context, statusCode int, message string, err error) {
	// 构建基本响应
	response := Response{
		Code:      statusCode,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
	}

	// 如果有错误对象，添加详细错误信息
	if err != nil {
		response.Error = err.Error()
	}

	// 返回JSON响应
	c.JSON(statusCode, response)
}

// BadRequest 返回400错误 - 请求参数错误
// 用于请求格式错误、参数验证失败等场景
func BadRequest(c *gin.Context, message string, err error) {
	Fail(c, http.StatusBadRequest, message, err)
}

// Unauthorized 返回401错误 - 未授权
// 用于未登录、令牌无效等场景
func Unauthorized(c *gin.Context, message string, err error) {
	Fail(c, http.StatusUnauthorized, message, err)
}

// Forbidden 返回403错误 - 禁止访问
// 用于权限不足、禁止操作等场景
func Forbidden(c *gin.Context, message string, err error) {
	Fail(c, http.StatusForbidden, message, err)
}

// NotFound 返回404错误 - 资源不存在
// 用于请求的资源不存在的场景
func NotFound(c *gin.Context, message string, err error) {
	Fail(c, http.StatusNotFound, message, err)
}

// InternalServerError 返回500错误 - 服务器内部错误
// 用于服务器内部异常、未预期的错误等场景
func InternalServerError(c *gin.Context, message string, err error) {
	Fail(c, http.StatusInternalServerError, message, err)
}
