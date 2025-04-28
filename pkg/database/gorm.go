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
var GormDB *gorm.DB

// GormConfig GORM数据库连接配置
type GormConfig struct {
	DSN             string          // 数据库连接字符串
	MaxOpenConns    int             // 最大打开连接数
	MaxIdleConns    int             // 最大空闲连接数
	ConnMaxLifetime time.Duration   // 连接最大生存时间
	ConnMaxIdleTime time.Duration   // 空闲连接最大生存时间
	LogLevel        logger.LogLevel // 日志级别
}

// InitGormDB 初始化GORM数据库连接
func InitGormDB() error {
	// 获取并解析数据库配置
	gormConfig, err := parseGormConfig()
	if err != nil {
		return fmt.Errorf("解析GORM数据库配置失败: %w", err)
	}

	// 配置GORM日志（已简化）
	gormLogger := logger.New(
		nil, // 移除日志输出
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢SQL阈值
			LogLevel:                  logger.Silent,          // 静默日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略记录未找到错误
			Colorful:                  false,                  // 关闭彩色打印
		},
	)

	// GORM配置
	gormDialectorConfig := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
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
	sqlDB.SetMaxOpenConns(gormConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(gormConfig.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(gormConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(gormConfig.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("GORM数据库连接测试失败: %w", err)
	}

	// 数据库连接成功

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
