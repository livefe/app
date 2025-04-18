package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"app/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 定义日志级别
const (
	DebugLevel  = "debug"
	InfoLevel   = "info"
	WarnLevel   = "warn"
	ErrorLevel  = "error"
	DPanicLevel = "dpanic"
	PanicLevel  = "panic"
	FatalLevel  = "fatal"
)

// Logger 定义日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Sync() error

	// 结构化日志方法
	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
}

// ZapLogger 基于zap的日志实现
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	fields map[string]interface{} // 存储结构化日志字段
}

// 全局日志实例
var globalLogger Logger

// Init 初始化日志
func Init() error {
	logCfg := config.GetLoggerConfig()

	// 创建日志目录
	if err := os.MkdirAll(logCfg.OutputPath, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 设置日志级别
	var level zapcore.Level
	switch logCfg.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case DPanicLevel:
		level = zapcore.DPanicLevel
	case PanicLevel:
		level = zapcore.PanicLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 设置日志编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if logCfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志输出
	// 按日期生成日志文件
	todayLogFile := filepath.Join(logCfg.OutputPath, time.Now().Format("2006-01-02")+".log")

	// 使用lumberjack进行日志轮转
	lumberJackLogger := &lumberjack.Logger{
		Filename:   todayLogFile,
		MaxSize:    logCfg.MaxSize,    // 单个日志文件最大大小，单位MB
		MaxBackups: logCfg.MaxBackups, // 最大保留日志文件数量
		MaxAge:     logCfg.MaxAge,     // 日志文件最大保存天数
		Compress:   logCfg.Compress,   // 是否压缩日志文件
	}

	// 同时输出到控制台和文件
	writeSyncer := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(lumberJackLogger),
	)

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 创建ZapLogger实例
	globalLogger = &ZapLogger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
		fields: make(map[string]interface{}),
	}

	return nil
}

// GetLogger 获取全局日志实例
func GetLogger() Logger {
	return globalLogger
}

// 以下是全局函数，提供更简洁的API

// WithFields 添加多个字段到日志
func WithFields(fields map[string]interface{}) Logger {
	return globalLogger.WithFields(fields)
}

// WithField 添加单个字段到日志
func WithField(key string, value interface{}) Logger {
	return globalLogger.WithField(key, value)
}

// WithError 添加错误信息到日志
func WithError(err error) Logger {
	return globalLogger.WithError(err)
}

// Debug 输出Debug级别日志
func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}

// Debugf 输出Debug级别格式化日志
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info 输出Info级别日志
func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

// Infof 输出Info级别格式化日志
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn 输出Warn级别日志
func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

// Warnf 输出Warn级别格式化日志
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error 输出Error级别日志
func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

// Errorf 输出Error级别格式化日志
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// DPanic 输出DPanic级别日志
func DPanic(args ...interface{}) {
	globalLogger.DPanic(args...)
}

// DPanicf 输出DPanic级别格式化日志
func DPanicf(format string, args ...interface{}) {
	globalLogger.DPanicf(format, args...)
}

// Panic 输出Panic级别日志
func Panic(args ...interface{}) {
	globalLogger.Panic(args...)
}

// Panicf 输出Panic级别格式化日志
func Panicf(format string, args ...interface{}) {
	globalLogger.Panicf(format, args...)
}

// Fatal 输出Fatal级别日志
func Fatal(args ...interface{}) {
	globalLogger.Fatal(args...)
}

// Fatalf 输出Fatal级别格式化日志
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

// Sync 同步日志
func Sync() error {
	return globalLogger.Sync()
}

// Debug 输出Debug级别日志
func (l *ZapLogger) Debug(args ...interface{}) {
	l.getSugarWithFields().Debug(args...)
}

// Debugf 输出Debug级别格式化日志
func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.getSugarWithFields().Debugf(format, args...)
}

// Info 输出Info级别日志
func (l *ZapLogger) Info(args ...interface{}) {
	l.getSugarWithFields().Info(args...)
}

// Infof 输出Info级别格式化日志
func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.getSugarWithFields().Infof(format, args...)
}

// Warn 输出Warn级别日志
func (l *ZapLogger) Warn(args ...interface{}) {
	l.getSugarWithFields().Warn(args...)
}

// Warnf 输出Warn级别格式化日志
func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	l.getSugarWithFields().Warnf(format, args...)
}

// Error 输出Error级别日志
func (l *ZapLogger) Error(args ...interface{}) {
	l.getSugarWithFields().Error(args...)
}

// Errorf 输出Error级别格式化日志
func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.getSugarWithFields().Errorf(format, args...)
}

// DPanic 输出DPanic级别日志
func (l *ZapLogger) DPanic(args ...interface{}) {
	l.getSugarWithFields().DPanic(args...)
}

// DPanicf 输出DPanic级别格式化日志
func (l *ZapLogger) DPanicf(format string, args ...interface{}) {
	l.getSugarWithFields().DPanicf(format, args...)
}

// Panic 输出Panic级别日志
func (l *ZapLogger) Panic(args ...interface{}) {
	l.getSugarWithFields().Panic(args...)
}

// Panicf 输出Panic级别格式化日志
func (l *ZapLogger) Panicf(format string, args ...interface{}) {
	l.getSugarWithFields().Panicf(format, args...)
}

// Fatal 输出Fatal级别日志
func (l *ZapLogger) Fatal(args ...interface{}) {
	l.getSugarWithFields().Fatal(args...)
}

// Fatalf 输出Fatal级别格式化日志
func (l *ZapLogger) Fatalf(format string, args ...interface{}) {
	l.getSugarWithFields().Fatalf(format, args...)
}

// Sync 同步日志
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// WithFields 添加多个字段到日志
func (l *ZapLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := &ZapLogger{
		logger: l.logger,
		sugar:  l.sugar,
		fields: make(map[string]interface{}, len(l.fields)+len(fields)),
	}

	// 复制现有字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// 添加新字段
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithField 添加单个字段到日志
func (l *ZapLogger) WithField(key string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

// WithError 添加错误信息到日志
func (l *ZapLogger) WithError(err error) Logger {
	if err != nil {
		return l.WithField("error", err.Error())
	}
	return l
}

// 获取带有字段的sugar logger
func (l *ZapLogger) getSugarWithFields() *zap.SugaredLogger {
	if len(l.fields) == 0 {
		return l.sugar
	}

	args := make([]interface{}, 0, len(l.fields)*2)
	for k, v := range l.fields {
		args = append(args, k, v)
	}

	return l.sugar.With(args...)
}
