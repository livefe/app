package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 以下是字段构造函数，封装zap的字段构造功能

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

// Error 创建错误类型的日志字段
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
