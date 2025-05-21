// Package database 提供数据库连接和操作的封装，基于GORM框架
package database

import (
	"fmt"
	"time"

	"app/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// Init 初始化数据库连接并配置连接池
func Init() error {
	cfg := config.GetDatabaseConfig()

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	// 解析连接时间配置
	connMaxLifetime, _ := time.ParseDuration(cfg.ConnMaxLifetime)
	connMaxIdleTime, _ := time.ParseDuration(cfg.ConnMaxIdleTime)
	maxIdleConns := cfg.MaxConnections / 4

	// GORM配置
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层SQL DB连接并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层SQL连接失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 设置全局数据库实例
	DB = db
	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层SQL连接失败: %w", err)
	}

	if err = sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	DB = nil
	return nil
}

// CheckDBHealth 检查数据库健康状态
func CheckDBHealth() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层SQL连接失败: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库健康检查失败: %w", err)
	}

	return GetDBStats()
}

// GetDBStats 获取数据库连接池统计信息
func GetDBStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层SQL连接失败: %w", err)
	}

	// 获取连接池统计信息
	stats := sqlDB.Stats()
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
