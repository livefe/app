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

// 上下文键常量，用于从上下文中提取标识信息
const (
	// RequestIDKey 请求ID的上下文键名
	RequestIDKey = "request_id"
	// UserIDKey 用户ID的上下文键名
	UserIDKey = "userID"
)

// 日志级别常量
const (
	// DebugLevel 调试级别
	DebugLevel = "debug"
	// InfoLevel 信息级别
	InfoLevel = "info"
	// WarnLevel 警告级别
	WarnLevel = "warn"
	// ErrorLevel 错误级别
	ErrorLevel = "error"
	// FatalLevel 致命级别
	FatalLevel = "fatal"
)

// 日志格式常量
const (
	// JSONFormat JSON格式输出
	JSONFormat = "json"
	// ConsoleFormat 控制台格式输出
	ConsoleFormat = "console"
)

// MaxBodySize 最大请求/响应体大小限制 (5MB)
const MaxBodySize = 5 * 1024 * 1024

// 全局日志实例
var (
	// logger 原始zap日志实例
	logger *zap.Logger
	// SugaredLogger 提供更便捷的API的sugar日志实例
	SugaredLogger *zap.SugaredLogger
)

// timeEncoder 自定义时间编码器，格式化为"2006-01-02 15:04:05.000"格式
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// Init 初始化日志系统
// 根据配置创建日志记录器，设置日志级别、格式、输出路径等
// 返回初始化过程中可能出现的错误
func Init() error {
	// 获取配置
	cfg := config.GetLoggerConfig()

	// 确保日志路径是绝对路径
	if !filepath.IsAbs(cfg.OutputPath) {
		// 获取当前工作目录
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("获取当前工作目录失败: %w", err)
		}
		cfg.OutputPath = filepath.Join(cwd, cfg.OutputPath)
	}

	// 创建日志目录
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 配置日志级别
	level := getZapLevel(cfg.Level)

	// 配置日志编码器
	encoder := createEncoder(cfg.Format)

	// 配置日志轮转
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.OutputPath,
		MaxSize:    cfg.MaxSize,    // 每个日志文件的最大大小（MB）
		MaxBackups: cfg.MaxBackups, // 保留的旧日志文件的最大数量
		MaxAge:     cfg.MaxAge,     // 保留旧日志文件的最大天数
		Compress:   cfg.Compress,   // 是否压缩旧日志文件
	}

	// 创建输出目标
	writeSyncer := createWriteSyncer(lumberJackLogger, cfg.Console)

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建日志记录器选项
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	// 配置调用栈
	if cfg.EnableStacktrace {
		// 解析调用栈级别
		stacktraceLevel := getZapLevel(cfg.StacktraceLevel)
		if stacktraceLevel == zapcore.InvalidLevel {
			stacktraceLevel = zapcore.ErrorLevel
		}

		// 添加调用栈选项
		options = append(options, zap.AddStacktrace(stacktraceLevel))

		// 设置调用栈深度（如果配置了）
		if cfg.StacktraceDepth > 0 {
			// Zap没有直接设置调用栈深度的选项
			// 但可以通过Development模式获取更详细的调用栈信息
			options = append(options, zap.Development())
		}
	}

	// 创建日志记录器
	logger = zap.New(core, options...)

	// 创建SugaredLogger
	SugaredLogger = logger.Sugar()

	return nil
}

// getZapLevel 将字符串日志级别转换为zap日志级别
func getZapLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// createEncoder 根据格式创建日志编码器
func createEncoder(format string) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	switch strings.ToLower(format) {
	case JSONFormat:
		return zapcore.NewJSONEncoder(encoderConfig)
	case ConsoleFormat:
		return zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return zapcore.NewJSONEncoder(encoderConfig)
	}
}

// createWriteSyncer 创建日志输出同步器
func createWriteSyncer(fileLogger *lumberjack.Logger, enableConsole bool) zapcore.WriteSyncer {
	if enableConsole {
		consoleSyncer := zapcore.AddSync(os.Stdout)
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileLogger), consoleSyncer)
	}
	return zapcore.AddSync(fileLogger)
}

// Close 关闭日志记录器，确保所有日志都被写入
// 返回同步过程中可能出现的错误
func Close() error {
	if logger != nil {
		err := logger.Sync()
		// 忽略标准输出/标准错误的同步错误，这些错误通常在应用关闭时发生
		// 当标准输出已关闭但日志系统仍尝试同步时会出现这些错误
		if err != nil && (strings.Contains(err.Error(), "sync /dev/stdout") ||
			strings.Contains(err.Error(), "sync /dev/stderr") ||
			strings.Contains(err.Error(), "The handle is invalid")) {
			return nil
		}
		return err
	}
	return nil
}

// WithContext 从上下文中获取请求ID和用户ID，并添加到日志字段中
// 返回带有上下文信息的日志记录器
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger.With(
			String("request_id", ""),
			String("userID", ""),
		)
	}

	// 构建基础日志字段
	fields := []zap.Field{}

	// 添加请求ID（始终添加，如果不存在则为空字符串）
	requestID := ""
	if id, ok := ctx.Value(RequestIDKey).(string); ok && id != "" {
		requestID = id
	}
	fields = append(fields, String("request_id", requestID))

	// 添加用户ID（始终添加，如果不存在则为空字符串）
	userID := ""
	if id, ok := ctx.Value(UserIDKey).(string); ok && id != "" {
		userID = id
	} else if id, ok := ctx.Value(UserIDKey).(uint); ok && id > 0 {
		userID = fmt.Sprintf("%d", id)
	}
	fields = append(fields, String("userID", userID))

	// 返回带有字段的日志记录器
	return logger.With(fields...)
}

// WithContextS 从上下文中获取请求ID和用户ID，并添加到SugaredLogger字段中
// 返回带有上下文信息的SugaredLogger
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

// String 创建字符串类型的日志字段
func String(key string, val string) zap.Field {
	return zap.String(key, val)
}

// Int 创建整数类型的日志字段
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int64 创建int64类型的日志字段
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Uint 创建uint类型的日志字段
func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

// Uint64 创建uint64类型的日志字段
func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

// Float64 创建float64类型的日志字段
func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// Bool 创建布尔类型的日志字段
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Time 创建时间类型的日志字段
func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

// Duration 创建时间间隔类型的日志字段
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

// Err 创建错误类型的日志字段
func Err(err error) zap.Field {
	return zap.Error(err)
}

// Any 创建任意类型的日志字段
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// Reflect 创建通过反射序列化的日志字段
func Reflect(key string, val interface{}) zap.Field {
	return zap.Reflect(key, val)
}

// Namespace 创建命名空间
func Namespace(key string) zap.Field {
	return zap.Namespace(key)
}

// Array 创建数组类型的日志字段
func Array(key string, val zapcore.ArrayMarshaler) zap.Field {
	return zap.Array(key, val)
}

// Object 创建对象类型的日志字段
func Object(key string, val zapcore.ObjectMarshaler) zap.Field {
	return zap.Object(key, val)
}

// Stringer 创建实现了Stringer接口的日志字段
func Stringer(key string, val fmt.Stringer) zap.Field {
	return zap.Stringer(key, val)
}
