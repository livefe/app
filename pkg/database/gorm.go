// Package database 提供数据库连接和操作的封装
// 基于GORM框架，支持MySQL数据库的连接管理和操作
package database

import (
	"fmt"
	"time"

	"app/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// GormDB 全局GORM数据库连接实例
// 在应用程序中可直接使用此变量进行数据库操作
var GormDB *gorm.DB

// GormConfig GORM数据库连接配置结构体
// 包含连接字符串、连接池设置和日志级别等配置项
type GormConfig struct {
	DSN             string          // 数据库连接字符串(Data Source Name)
	MaxOpenConns    int             // 最大打开连接数，控制并发连接数量
	MaxIdleConns    int             // 最大空闲连接数，控制连接池大小
	ConnMaxLifetime time.Duration   // 连接最大生存时间，超时后会被关闭重建
	ConnMaxIdleTime time.Duration   // 空闲连接最大生存时间，超时后会被关闭
	LogLevel        logger.LogLevel // 日志级别，控制SQL日志输出详细程度
}

// Init 初始化GORM数据库连接并配置连接池
// 成功时返回nil，失败时返回带上下文的错误信息
func Init() error {
	// 获取并解析数据库配置
	gormConfig, err := parseGormConfig()
	if err != nil {
		return fmt.Errorf("解析GORM数据库配置失败: %w", err)
	}

	// 配置GORM日志
	gormLogger := logger.New(
		nil, // 不使用默认的日志输出，由应用自行处理日志
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢SQL阈值，超过此时间的查询会被记录
			LogLevel:                  logger.Silent,          // 静默日志级别，不输出SQL日志
			IgnoreRecordNotFoundError: true,                   // 忽略记录未找到错误，减少日志噪音
			Colorful:                  false,                  // 关闭彩色打印，适合生产环境
		},
	)

	// GORM配置
	gormDialectorConfig := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，与数据库表名一致
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束，提高迁移性能
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(gormConfig.DSN), gormDialectorConfig)
	if err != nil {
		return fmt.Errorf("GORM连接数据库失败: %w", err)
	}

	// 获取底层SQL DB连接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取GORM底层SQL DB连接失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(gormConfig.MaxOpenConns)       // 设置最大打开连接数
	sqlDB.SetMaxIdleConns(gormConfig.MaxIdleConns)       // 设置最大空闲连接数
	sqlDB.SetConnMaxLifetime(gormConfig.ConnMaxLifetime) // 设置连接最大生存时间
	sqlDB.SetConnMaxIdleTime(gormConfig.ConnMaxIdleTime) // 设置空闲连接最大生存时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("GORM数据库连接测试失败: %w", err)
	}

	// 设置全局GormDB实例
	GormDB = db

	return nil
}

// parseGormConfig 解析GORM数据库配置
func parseGormConfig() (*GormConfig, error) {
	// 获取数据库配置
	cfg := config.GetDatabaseConfig()

	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name)

	// 解析连接最大生存时间
	connMaxLifetime, err := time.ParseDuration(cfg.ConnMaxLifetime)
	if err != nil {
		connMaxLifetime = time.Hour // 默认值为1小时
		// 解析连接最大生存时间失败，使用默认值1小时
	}

	// 解析空闲连接最大生存时间
	connMaxIdleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime)
	if err != nil {
		connMaxIdleTime = time.Minute * 30 // 默认值为30分钟
		// 解析空闲连接最大生存时间失败，使用默认值30分钟
	}

	// 计算最大空闲连接数
	maxIdleConns := cfg.MaxConnections / 4 // 空闲连接数设为最大连接数的1/4
	if maxIdleConns < 2 {
		maxIdleConns = 2 // 确保至少有2个空闲连接
	}

	return &GormConfig{
		DSN:             dsn,
		MaxOpenConns:    cfg.MaxConnections,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
		LogLevel:        logger.Silent, // 静默日志级别
	}, nil
}

// GetGormDB 获取GORM数据库连接实例
func GetGormDB() *gorm.DB {
	return GormDB
}

// CloseGormDB 关闭GORM数据库连接
func CloseGormDB() error {
	if GormDB != nil {
		// 正在关闭GORM数据库连接

		// 获取底层SQL DB连接
		sqlDB, err := GormDB.DB()
		if err != nil {
			return fmt.Errorf("获取GORM底层SQL DB连接失败: %w", err)
		}

		// 关闭连接
		err = sqlDB.Close()
		if err != nil {
			return fmt.Errorf("关闭GORM数据库连接失败: %w", err)
		}

		// GORM数据库连接已关闭
		GormDB = nil
		return nil
	}
	return nil
}

// CheckGormDBHealth 检查GORM数据库健康状态
func CheckGormDBHealth() (map[string]interface{}, error) {
	// 检查数据库是否已初始化
	if GormDB == nil {
		return nil, fmt.Errorf("GORM数据库未初始化")
	}

	// 获取底层SQL DB连接
	sqlDB, err := GormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取GORM底层SQL DB连接失败: %w", err)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		// 数据库健康检查失败
		return nil, fmt.Errorf("GORM数据库健康检查失败: %w", err)
	}

	// 获取并返回连接池状态
	return GetDBStats()
}

// GetDBStats 获取数据库连接池统计信息
func GetDBStats() (map[string]interface{}, error) {
	// 检查数据库是否已初始化
	if GormDB == nil {
		return nil, fmt.Errorf("GORM数据库未初始化")
	}

	// 获取底层SQL DB连接
	sqlDB, err := GormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取GORM底层SQL DB连接失败: %w", err)
	}

	// 获取连接池统计信息
	stats := sqlDB.Stats()

	// 转换为map以便于序列化
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
