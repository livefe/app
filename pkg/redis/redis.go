// Package redis 提供Redis数据库操作的封装
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"app/config"

	"github.com/redis/go-redis/v9"
)

// Client 全局Redis客户端实例
var Client *redis.Client

// 错误常量
var (
	// ErrNil 表示Redis中不存在该键
	ErrNil = redis.Nil
	// ErrKeyNotFound 表示键不存在
	ErrKeyNotFound = errors.New("键不存在")
	// ErrInvalidType 表示类型无效
	ErrInvalidType = errors.New("无效的数据类型")
	// ErrInvalidBitOp 表示不支持的位操作类型
	ErrInvalidBitOp = errors.New("不支持的位操作类型")
	// ErrInvalidBitOpParams 表示位操作参数无效
	ErrInvalidBitOpParams = errors.New("位操作参数无效")
)

// 默认上下文超时时间
const defaultTimeout = 5 * time.Second

// RedisConfig Redis连接配置结构体
type RedisConfig struct {
	Addr         string        // Redis服务器地址
	Password     string        // Redis密码
	DB           int           // 数据库索引
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	DialTimeout  time.Duration // 连接超时时间
	ReadTimeout  time.Duration // 读取超时时间
	WriteTimeout time.Duration // 写入超时时间
}

// Init 初始化Redis连接并测试连接可用性
func Init() error {
	// 获取并解析Redis配置
	redisConfig, err := parseRedisConfig()
	if err != nil {
		return fmt.Errorf("解析Redis配置失败: %w", err)
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := getContext()
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		_ = client.Close()
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}

	// 设置全局Client实例
	Client = client

	return nil
}

// parseRedisConfig 解析Redis配置
func parseRedisConfig() (*RedisConfig, error) {
	cfg := config.GetRedisConfig()
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// 解析超时时间配置
	dialTimeout, err := time.ParseDuration(cfg.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("解析连接超时时间失败: %w", err)
	}

	readTimeout, err := time.ParseDuration(cfg.ReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("解析读取超时时间失败: %w", err)
	}

	writeTimeout, err := time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("解析写入超时时间失败: %w", err)
	}

	return &RedisConfig{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}, nil
}

// Close 安全地关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// getContext 创建带默认超时的上下文
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// 字符串操作

// Set 设置键值对并指定过期时间
func Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := getContext()
	defer cancel()
	return Client.Set(ctx, key, value, expiration).Err()
}

// SetNX 当键不存在时设置键值对并指定过期时间，常用于实现分布式锁
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()
	return Client.SetNX(ctx, key, value, expiration).Result()
}

// Get 获取字符串类型的键值
func Get(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	result, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

// GetObj 获取JSON对象并反序列化到指定结构
func GetObj(key string, obj interface{}) error {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrKeyNotFound
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), obj)
}

// SetObj 设置对象（序列化后存储）
func SetObj(key string, obj interface{}, expiration time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	ctx, cancel := getContext()
	defer cancel()

	return Client.Set(ctx, key, data, expiration).Err()
}

// Del 删除键
func Del(keys ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Del(ctx, keys...).Result()
}

// Exists 检查键是否存在
func Exists(keys ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Expire(ctx, key, expiration).Result()
}

// 哈希表操作

// HSet 设置哈希表字段
func HSet(key string, values ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希表字段
func HGet(key, field string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// HGetAll 获取哈希表所有字段和值
func HGetAll(key string) (map[string]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希表字段
func HDel(key string, fields ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.HDel(ctx, key, fields...).Result()
}

// 列表操作

// LPush 将一个或多个值插入到列表头部
func LPush(key string, values ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.LPush(ctx, key, values...).Result()
}

// RPush 将一个或多个值插入到列表尾部
func RPush(key string, values ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.RPush(ctx, key, values...).Result()
}

// LPop 移出并获取列表的第一个元素
func LPop(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// RPop 移出并获取列表的最后一个元素
func RPop(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// LRange 获取列表指定范围内的元素
func LRange(key string, start, stop int64) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.LRange(ctx, key, start, stop).Result()
}

// 集合操作

// SAdd 向集合添加一个或多个成员
func SAdd(key string, members ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SAdd(ctx, key, members...).Result()
}

// SMembers 获取集合所有成员
func SMembers(key string) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SMembers(ctx, key).Result()
}

// SRem 移除集合中一个或多个成员
func SRem(key string, members ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SRem(ctx, key, members...).Result()
}

// 有序集合操作

// ZAdd 向有序集合添加一个或多个成员
func ZAdd(key string, members ...redis.Z) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ZAdd(ctx, key, members...).Result()
}

// ZRange 通过索引区间返回有序集合成员
func ZRange(key string, start, stop int64) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 移除有序集合中的一个或多个成员
func ZRem(key string, members ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ZRem(ctx, key, members...).Result()
}

// 计数器操作

// Incr 将 key 中储存的数字值增一
func Incr(key string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Incr(ctx, key).Result()
}

// IncrBy 将 key 中储存的数字值增加指定增量值
func IncrBy(key string, value int64) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.IncrBy(ctx, key, value).Result()
}

// IncrByFloat 将 key 中储存的数字值增加指定浮点数增量值
func IncrByFloat(key string, value float64) (float64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.IncrByFloat(ctx, key, value).Result()
}

// Decr 将 key 中储存的数字值减一
func Decr(key string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Decr(ctx, key).Result()
}

// DecrBy 将 key 中储存的数字值减去指定减量值
func DecrBy(key string, value int64) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.DecrBy(ctx, key, value).Result()
}

// Scan 迭代器操作

// Scan 迭代当前数据库中的数据库键
func Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	ctx, cancel := getContext()
	defer cancel()

	keys, nextCursor, err := Client.Scan(ctx, cursor, match, count).Result()
	return keys, nextCursor, err
}

// HScan 迭代哈希表中的键值对
func HScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	ctx, cancel := getContext()
	defer cancel()

	values, nextCursor, err := Client.HScan(ctx, key, cursor, match, count).Result()
	return values, nextCursor, err
}

// SScan 迭代集合中的元素
func SScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	ctx, cancel := getContext()
	defer cancel()

	members, nextCursor, err := Client.SScan(ctx, key, cursor, match, count).Result()
	return members, nextCursor, err
}

// ZScan 迭代有序集合中的元素
func ZScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	ctx, cancel := getContext()
	defer cancel()

	values, nextCursor, err := Client.ZScan(ctx, key, cursor, match, count).Result()
	return values, nextCursor, err
}

// 发布订阅操作

// Publish 将信息发送到指定的频道
func Publish(channel string, message interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Publish(ctx, channel, message).Result()
}

// Subscribe 订阅给定的一个或多个频道的信息
func Subscribe(channels ...string) *redis.PubSub {
	return Client.Subscribe(context.Background(), channels...)
}

// PSubscribe 订阅一个或多个符合给定模式的频道
func PSubscribe(patterns ...string) *redis.PubSub {
	return Client.PSubscribe(context.Background(), patterns...)
}

// 事务操作

// TxPipeline 创建一个事务管道
func TxPipeline() redis.Pipeliner {
	return Client.TxPipeline()
}

// Watch 监视一个或多个key，如果在事务执行之前这个key被其他命令所改动，那么事务将被打断
func Watch(fn func(*redis.Tx) error, keys ...string) error {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Watch(ctx, fn, keys...)
}

// 键管理命令

// Keys 查找所有符合给定模式的键
func Keys(pattern string) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Keys(ctx, pattern).Result()
}

// Type 返回键所储存的值的类型
func Type(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Type(ctx, key).Result()
}

// TTL 返回键的剩余生存时间
func TTL(key string) (time.Duration, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.TTL(ctx, key).Result()
}

// Rename 修改键的名称
func Rename(key, newkey string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Rename(ctx, key, newkey).Result()
}

// RenameNX 仅当 newkey 不存在时修改键的名称
func RenameNX(key, newkey string) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.RenameNX(ctx, key, newkey).Result()
}

// 位图操作

// SetBit 对key所储存的字符串值，设置或清除指定偏移量上的位
func SetBit(key string, offset int64, value int) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SetBit(ctx, key, offset, value).Result()
}

// GetBit 对key所储存的字符串值，获取指定偏移量上的位
func GetBit(key string, offset int64) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.GetBit(ctx, key, offset).Result()
}

// BitCount 计算字符串中被设置为1的比特位的数量
func BitCount(key string, bitCount *redis.BitCount) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.BitCount(ctx, key, bitCount).Result()
}

// 管道操作

// Pipeline 创建一个管道，用于一次性执行多个命令
func Pipeline() redis.Pipeliner {
	return Client.Pipeline()
}

// Pipelined 在管道中执行命令
func Pipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Pipelined(ctx, fn)
}

// 地理位置操作

// GeoAdd 将指定的地理空间位置（纬度、经度、名称）添加到指定的key中
func GeoAdd(key string, geoLocation ...*redis.GeoLocation) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.GeoAdd(ctx, key, geoLocation...).Result()
}

// GeoPos 从key里返回所有给定位置元素的位置（经度和纬度）
func GeoPos(key string, members ...string) ([]*redis.GeoPos, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.GeoPos(ctx, key, members...).Result()
}

// GeoDist 返回两个给定位置之间的距离
func GeoDist(key string, member1, member2, unit string) (float64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.GeoDist(ctx, key, member1, member2, unit).Result()
}

// GeoRadius 以给定的经纬度为中心， 返回键包含的位置元素当中， 与中心的距离不超过给定最大距离的所有位置元素
func GeoRadius(key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.GeoRadius(ctx, key, longitude, latitude, query).Result()
}

// GeoRadiusByMember 以给定的成员为中心， 返回键包含的位置元素当中， 与中心的距离不超过给定最大距离的所有位置元素
func GeoRadiusByMember(key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.GeoRadiusByMember(ctx, key, member, query).Result()
}

// HyperLogLog操作

// PFAdd 将任意数量的元素添加到指定的HyperLogLog中
func PFAdd(key string, els ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.PFAdd(ctx, key, els...).Result()
}

// PFCount 返回给定HyperLogLog的基数估算值
func PFCount(keys ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.PFCount(ctx, keys...).Result()
}

// PFMerge 将多个HyperLogLog合并为一个HyperLogLog
func PFMerge(dest string, keys ...string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.PFMerge(ctx, dest, keys...).Result()
}

// 脚本执行

// Eval 执行Lua脚本
func Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Eval(ctx, script, keys, args...).Result()
}

// EvalSha 执行Lua脚本（通过SHA1校验和）
func EvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptLoad 将脚本加载到脚本缓存中
func ScriptLoad(script string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ScriptLoad(ctx, script).Result()
}

// ScriptExists 检查脚本是否已经被保存在缓存中
func ScriptExists(scripts ...string) ([]bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ScriptExists(ctx, scripts...).Result()
}

// ScriptFlush 从脚本缓存中移除所有脚本
func ScriptFlush() (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ScriptFlush(ctx).Result()
}

// 位操作

// BitOp 对一个或多个保存二进制位的字符串键执行位元操作，并将结果保存到 destkey 上
func BitOp(op string, destKey string, keys ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	// 根据操作类型调用相应的方法
	switch op {
	case "AND", "and":
		return Client.BitOpAnd(ctx, destKey, keys...).Result()
	case "OR", "or":
		return Client.BitOpOr(ctx, destKey, keys...).Result()
	case "XOR", "xor":
		return Client.BitOpXor(ctx, destKey, keys...).Result()
	case "NOT", "not":
		if len(keys) != 1 {
			return 0, ErrInvalidBitOpParams
		}
		return Client.BitOpNot(ctx, destKey, keys[0]).Result()
	default:
		return 0, ErrInvalidBitOp
	}
}

// BitPos 返回字符串里面第一个被设置为1或者0的bit位
func BitPos(key string, bit int64, pos ...int64) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.BitPos(ctx, key, bit, pos...).Result()
}

// 流操作

// XAdd 将消息添加到流
func XAdd(a *redis.XAddArgs) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XAdd(ctx, a).Result()
}

// XDel 从流中删除消息
func XDel(stream string, ids ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XDel(ctx, stream, ids...).Result()
}

// XLen 获取流包含的元素数量
func XLen(stream string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XLen(ctx, stream).Result()
}

// XRange 获取流中的消息范围
func XRange(stream, start, stop string) ([]redis.XMessage, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XRange(ctx, stream, start, stop).Result()
}

// XRevRange 反向获取流中的消息范围
func XRevRange(stream, start, stop string) ([]redis.XMessage, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XRevRange(ctx, stream, start, stop).Result()
}

// XRead 从流中读取数据
func XRead(a *redis.XReadArgs) ([]redis.XStream, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XRead(ctx, a).Result()
}

// XGroupCreate 创建消费者组
func XGroupCreate(stream, group, start string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XGroupCreate(ctx, stream, group, start).Result()
}

// XReadGroup 读取消费者组中的消息
func XReadGroup(a *redis.XReadGroupArgs) ([]redis.XStream, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.XReadGroup(ctx, a).Result()
}

// 集群操作

// ClusterSlots 获取集群节点的插槽映射
func ClusterSlots() ([]redis.ClusterSlot, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ClusterSlots(ctx).Result()
}

// 其他实用命令

// FlushDB 清空当前数据库中的所有key
func FlushDB() (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.FlushDB(ctx).Result()
}

// FlushAll 清空整个 Redis 服务器的数据
func FlushAll() (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.FlushAll(ctx).Result()
}

// Time 返回当前服务器时间
func Time() (time.Time, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Time(ctx).Result()
}

// DBSize 返回当前数据库的key数量
func DBSize() (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.DBSize(ctx).Result()
}

// Info 获取Redis服务器的各种信息和统计数值
func Info(section ...string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Info(ctx, section...).Result()
}
