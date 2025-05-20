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
	UserIDKey    = "userID"
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

// 自定义时间编码器
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// Init 初始化日志系统
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

	// 创建输出目标
	var writeSyncer zapcore.WriteSyncer

	// 检查是否需要同时输出到控制台
	if cfg.Console {
		consoleSyncer := zapcore.AddSync(os.Stdout)
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), consoleSyncer)
	} else {
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		level,
	)

	// 创建日志记录器
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	// 配置调用栈
	// 解析调用栈级别
	var stacktraceLevel zapcore.Level
	switch strings.ToLower(cfg.StacktraceLevel) {
	case DebugLevel:
		stacktraceLevel = zapcore.DebugLevel
	case InfoLevel:
		stacktraceLevel = zapcore.InfoLevel
	case WarnLevel:
		stacktraceLevel = zapcore.WarnLevel
	case ErrorLevel:
		stacktraceLevel = zapcore.ErrorLevel
	case FatalLevel:
		stacktraceLevel = zapcore.FatalLevel
	default:
		stacktraceLevel = zapcore.ErrorLevel
	}

	// 根据配置决定是否启用调用栈
	if cfg.EnableStacktrace {
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

// Close 关闭日志记录器
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
