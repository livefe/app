package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"app/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 定义上下文键
const (
	RequestIDKey = "request_id"
	UserIDKey    = "user_id"
)

// 定义日志级别
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
)

// 定义日志格式
const (
	JSONFormat    = "json"
	ConsoleFormat = "console"
)

// 最大请求/响应体大小限制 (5MB)
const MaxBodySize = 5 * 1024 * 1024

// Logger 全局日志实例
var logger *zap.Logger

// SugaredLogger 提供更便捷的API
var SugaredLogger *zap.SugaredLogger

// Init 初始化日志系统
func Init() error {
	cfg := config.GetLoggerConfig()

	// 创建日志目录
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 配置日志级别
	var level zapcore.Level
	switch strings.ToLower(cfg.Level) {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置日志编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	switch strings.ToLower(cfg.Format) {
	case JSONFormat:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case ConsoleFormat:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 配置日志轮转
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.OutputPath,
		MaxSize:    cfg.MaxSize,    // 每个日志文件的最大大小（MB）
		MaxBackups: cfg.MaxBackups, // 保留的旧日志文件的最大数量
		MaxAge:     cfg.MaxAge,     // 保留旧日志文件的最大天数
		Compress:   cfg.Compress,   // 是否压缩旧日志文件
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(lumberJackLogger),
		level,
	)

	// 创建日志记录器
	logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 创建SugaredLogger
	SugaredLogger = logger.Sugar()

	return nil
}

// Close 关闭日志记录器
func Close() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// WithContext 从上下文中获取请求ID和用户ID，并添加到日志字段中
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}

	// 构建基础日志字段
	fields := []zap.Field{}

	// 添加请求ID（如果存在）
	if id, ok := ctx.Value(RequestIDKey).(string); ok && id != "" {
		fields = append(fields, zap.String("request_id", id))
	}

	// 添加用户ID（如果存在）
	if id, ok := ctx.Value(UserIDKey).(string); ok && id != "" {
		fields = append(fields, zap.String("user_id", id))
	} else if id, ok := ctx.Value(UserIDKey).(uint); ok && id > 0 {
		fields = append(fields, zap.String("user_id", fmt.Sprintf("%d", id)))
	}

	// 返回带有字段的日志记录器
	return logger.With(fields...)
}

// WithContextS 从上下文中获取请求ID和用户ID，并添加到SugaredLogger字段中
func WithContextS(ctx context.Context) *zap.SugaredLogger {
	return WithContext(ctx).Sugar()
}

// Debug 记录调试级别日志
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// Info 记录信息级别日志
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// Error 记录错误级别日志
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Error(msg, fields...)
}

// Fatal 记录致命级别日志
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Fatal(msg, fields...)
}

// Debugf 记录调试级别日志（格式化）
func Debugf(ctx context.Context, format string, args ...interface{}) {
	WithContextS(ctx).Debugf(format, args...)
}

// Infof 记录信息级别日志（格式化）
func Infof(ctx context.Context, format string, args ...interface{}) {
	WithContextS(ctx).Infof(format, args...)
}

// Warnf 记录警告级别日志（格式化）
func Warnf(ctx context.Context, format string, args ...interface{}) {
	WithContextS(ctx).Warnf(format, args...)
}

// Errorf 记录错误级别日志（格式化）
func Errorf(ctx context.Context, format string, args ...interface{}) {
	WithContextS(ctx).Errorf(format, args...)
}

// Fatalf 记录致命级别日志（格式化）
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	WithContextS(ctx).Fatalf(format, args...)
}
