package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"app/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// 最大请求/响应体大小限制 (5MB)
const MaxBodySize = 5 * 1024 * 1024

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set(logger.RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			if c.Request.ContentLength <= MaxBodySize {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			} else {
				requestBody = []byte(fmt.Sprintf("[请求体太大，大小: %d字节]", c.Request.ContentLength))
			}
		}

		// 构建请求日志字段
		requestFields := []zap.Field{
			logger.String("client_ip", c.ClientIP()),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("query", c.Request.URL.RawQuery),
			logger.String("user_agent", c.Request.UserAgent()),
		}

		// 添加请求体（如果存在）
		if len(requestBody) > 0 {
			addBodyToFields(requestBody, "request_body", &requestFields)
		}

		// 记录请求信息
		logger.Info(c, "收到HTTP请求", requestFields...)

		// 记录请求开始时间
		startTime := time.Now()

		// 创建自定义响应写入器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算请求处理时间
		latency := time.Since(startTime)

		// 构建响应日志字段
		responseFields := []zap.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.Int("status", c.Writer.Status()),
			logger.Duration("latency", latency),
		}

		// 添加响应体（如果存在）
		if blw.body.Len() > 0 {
			if blw.body.Len() <= MaxBodySize {
				addBodyToFields(blw.body.Bytes(), "response_body", &responseFields)
			} else {
				responseFields = append(responseFields, logger.String("response_body",
					fmt.Sprintf("[响应体太大，大小: %d字节]", blw.body.Len())))
			}
		}

		// 根据状态码选择日志级别
		statusCode := c.Writer.Status()
		if statusCode >= http.StatusInternalServerError {
			logger.Error(c, "完成HTTP请求", responseFields...)
		} else if statusCode >= http.StatusBadRequest {
			logger.Warn(c, "完成HTTP请求", responseFields...)
		} else {
			logger.Info(c, "完成HTTP请求", responseFields...)
		}
	}
}

// 添加请求/响应体到日志字段
func addBodyToFields(body []byte, fieldName string, fields *[]zap.Field) {
	if isJSON(body) {
		var jsonMap map[string]interface{}
		if err := json.Unmarshal(body, &jsonMap); err == nil {
			// 敏感字段处理
			sanitizeJSON(jsonMap)
			*fields = append(*fields, logger.Any(fieldName, jsonMap))
		} else {
			*fields = append(*fields, logger.String(fieldName, string(body)))
		}
	} else {
		*fields = append(*fields, logger.String(fieldName, string(body)))
	}
}

// bodyLogWriter 是一个自定义的响应写入器，用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 实现ResponseWriter接口的Write方法
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// isJSON 检查字节数组是否为JSON格式
func isJSON(data []byte) bool {
	return json.Valid(data) && (data[0] == '{' || data[0] == '[')
}

// sensitiveFieldMap 敏感字段映射，用于快速查找
var sensitiveFieldMap = map[string]bool{
	"password":      true,
	"token":         true,
	"secret":        true,
	"authorization": true,
	"auth":          true,
	"key":           true,
}

// sanitizeJSON 处理JSON中的敏感字段
func sanitizeJSON(data map[string]interface{}) {
	for k, v := range data {
		// 转换为小写进行检查
		lowerKey := strings.ToLower(k)

		// 检查是否为敏感字段或包含敏感字段
		isSensitive := false
		for field := range sensitiveFieldMap {
			if strings.Contains(lowerKey, field) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			data[k] = "[REDACTED]"
			continue
		}

		// 递归处理嵌套的map
		if nestedMap, ok := v.(map[string]interface{}); ok {
			sanitizeJSON(nestedMap)
			continue
		}

		// 处理数组中的map
		if nestedSlice, ok := v.([]interface{}); ok {
			for _, item := range nestedSlice {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					sanitizeJSON(nestedMap)
				}
			}
		}
	}
}
