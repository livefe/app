package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	SMS      SMSConfig      `mapstructure:"sms"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxConnections  int    `mapstructure:"max_connections"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idle_time"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey   string `mapstructure:"secret_key"`
	ExpiresTime string `mapstructure:"expires_time"`
	Issuer      string `mapstructure:"issuer"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level            string `mapstructure:"level"`
	Format           string `mapstructure:"format"`
	OutputPath       string `mapstructure:"output_path"`
	MaxSize          int    `mapstructure:"max_size"`
	MaxAge           int    `mapstructure:"max_age"`
	MaxBackups       int    `mapstructure:"max_backups"`
	Compress         bool   `mapstructure:"compress"`
	Console          bool   `mapstructure:"console"`           // 是否同时输出到控制台
	EnableStacktrace bool   `mapstructure:"enable_stacktrace"` // 是否启用调用栈
	StacktraceLevel  string `mapstructure:"stacktrace_level"`  // 记录调用栈的最低日志级别
	StacktraceDepth  int    `mapstructure:"stacktrace_depth"`  // 调用栈深度
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	Aliyun AliyunSMSConfig `mapstructure:"aliyun"`
}

// AliyunSMSConfig 阿里云短信服务配置
type AliyunSMSConfig struct {
	AccessKeyID     string            `mapstructure:"access_key_id"`
	AccessKeySecret string            `mapstructure:"access_key_secret"`
	Endpoint        string            `mapstructure:"endpoint"`
	SignName        string            `mapstructure:"sign_name"`
	Templates       map[string]string `mapstructure:"templates"` // 短信模板代码映射
}

var config *Config

// Init 初始化配置
func Init() error {
	// 创建viper实例
	v := viper.New()

	// 设置配置文件名和路径
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 加载.env文件
	envFile := "./config/.env"
	if _, err := os.Stat(envFile); err == nil {
		// 使用gotenv库加载.env文件
		if err := gotenv.Load(envFile); err != nil {
			return fmt.Errorf("读取.env文件失败: %w", err)
		}
	}

	// 读取环境变量
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 将配置映射到结构体
	config = &Config{}
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}

// GetConfig 获取配置
func GetConfig() *Config {
	return config
}

// GetServerConfig 获取服务器配置
func GetServerConfig() ServerConfig {
	return config.Server
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() DatabaseConfig {
	return config.Database
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig() RedisConfig {
	return config.Redis
}

// GetJWTConfig 获取JWT配置
func GetJWTConfig() JWTConfig {
	return config.JWT
}

// GetLoggerConfig 获取日志配置
func GetLoggerConfig() LoggerConfig {
	return config.Logger
}

// GetSMSConfig 获取短信服务配置
func GetSMSConfig() SMSConfig {
	return config.SMS
}
