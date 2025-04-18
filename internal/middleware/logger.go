package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"app/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggerMiddleware 创建日志中间件，记录HTTP请求信息
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 生成请求ID
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 记录请求体（可选，对于大型请求体可能影响性能）
		var requestBody string
		if c.Request.Method != http.MethodGet && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = string(bodyBytes)
		}

		// 创建响应体捕获器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 获取响应状态
		statusCode := c.Writer.Status()

		// 记录日志
		logger.WithFields(map[string]interface{}{
			"request_id":    requestID,
			"method":        method,
			"path":          path,
			"ip":            ip,
			"user_agent":    userAgent,
			"status_code":   statusCode,
			"latency":       latency.String(),
			"latency_ms":    float64(latency.Microseconds()) / 1000.0,
			"request_body":  requestBody,
			"response_body": blw.body.String(),
			"error":         c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}).Info("HTTP请求")
	}
}

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，同时写入到原始writer和缓冲区
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
