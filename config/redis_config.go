package config

import (
	"fmt"
	"strings"
	"time"

	conf "dubbo.apache.org/dubbo-go/v3/common/config"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	"github.com/dubbogo/gost/log/logger"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

// RedisConfig 结构体定义
type RedisConfig struct {
	Host          string `json:"host" yaml:"host"`
	Port          int    `json:"port" yaml:"port"`
	Password      string `json:"password" yaml:"password"`
	DB            int    `json:"db" yaml:"db"`
	PoolSize      int    `json:"pool_size" yaml:"pool_size"`
	MinIdleConns  int    `json:"min_idle_conns" yaml:"min_idle_conns"`
	ConnTimeout   string `json:"conn_timeout" yaml:"conn_timeout"`
	ReadTimeout   string `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout  string `json:"write_timeout" yaml:"write_timeout"`
	PoolTimeout   string `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout   string `json:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFreq string `json:"idle_check_freq" yaml:"idle_check_freq"`
	MaxConnAge    string `json:"max_conn_age" yaml:"max_conn_age"`
}

// GetAddr 获取 Redis 连接地址
func (rc *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", rc.Host, rc.Port)
}

// parseDuration 解析时间字符串为 time.Duration
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	// 移除可能的引号
	s = strings.Trim(s, `"`)
	d, err := time.ParseDuration(s)
	if err != nil {
		logger.Warnf("Failed to parse duration %s: %v, using default 0", s, err)
		return 0
	}
	return d
}

// GetRedisConfigFromDubbo 从 dubbo-go 配置中心获取 Redis 配置
func GetRedisConfigFromDubbo() (*RedisConfig, error) {
	// 从 dubbo-go 环境实例获取动态配置
	dynamicConfig := conf.GetEnvInstance().GetDynamicConfiguration()
	if dynamicConfig == nil {
		return nil, fmt.Errorf("dynamic configuration not initialized, please ensure config center is configured")
	}

	// 获取配置内容 (使用空字符串作为key，获取整个配置文件)
	// 或者使用 dataId 作为 key
	content, err := dynamicConfig.GetProperties("go-server", config_center.WithGroup("DEFAULT_GROUP"))

	if err != nil {
		return nil, fmt.Errorf("failed to get config from dubbo config center: %w", err)
	}

	if content == "" {
		return nil, fmt.Errorf("config content is empty")
	}

	logger.Infof("Got config from dubbo config center: %s", content)

	// 解析 YAML 配置
	var configMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &configMap); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	// 提取 redis 配置
	return ParseRedisConfig(configMap)
}

// ParseRedisConfig 从配置 map 中解析 Redis 配置
func ParseRedisConfig(params map[string]interface{}) (*RedisConfig, error) {
	redisMap, ok := params["redis"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("redis config not found or invalid format")
	}

	config := &RedisConfig{}

	// 解析各个字段
	if v, ok := redisMap["host"].(string); ok {
		config.Host = v
	}
	if v, ok := redisMap["port"].(int); ok {
		config.Port = v
	}
	if v, ok := redisMap["password"].(string); ok {
		config.Password = v
	}
	if v, ok := redisMap["db"].(int); ok {
		config.DB = v
	}
	if v, ok := redisMap["pool_size"].(int); ok {
		config.PoolSize = v
	}
	if v, ok := redisMap["min_idle_conns"].(int); ok {
		config.MinIdleConns = v
	}
	if v, ok := redisMap["conn_timeout"].(string); ok {
		config.ConnTimeout = v
	}
	if v, ok := redisMap["read_timeout"].(string); ok {
		config.ReadTimeout = v
	}
	if v, ok := redisMap["write_timeout"].(string); ok {
		config.WriteTimeout = v
	}
	if v, ok := redisMap["pool_timeout"].(string); ok {
		config.PoolTimeout = v
	}
	if v, ok := redisMap["idle_timeout"].(string); ok {
		config.IdleTimeout = v
	}
	if v, ok := redisMap["idle_check_freq"].(string); ok {
		config.IdleCheckFreq = v
	}
	if v, ok := redisMap["max_conn_age"].(string); ok {
		config.MaxConnAge = v
	}

	logger.Infof("Parsed Redis config: %+v", config)
	return config, nil
}

// CreateRedisClient 创建 Redis 客户端
func (rc *RedisConfig) CreateRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            rc.GetAddr(),
		Password:        rc.Password,
		DB:              rc.DB,
		PoolSize:        rc.PoolSize,
		MinIdleConns:    rc.MinIdleConns,
		DialTimeout:     parseDuration(rc.ConnTimeout),
		ReadTimeout:     parseDuration(rc.ReadTimeout),
		WriteTimeout:    parseDuration(rc.WriteTimeout),
		PoolTimeout:     parseDuration(rc.PoolTimeout),
		ConnMaxIdleTime: parseDuration(rc.IdleTimeout),
		ConnMaxLifetime: parseDuration(rc.MaxConnAge),
	})

	logger.Infof("Redis client created: addr=%s, db=%d", rc.GetAddr(), rc.DB)
	return client, nil
}

// GetRedisConfigFromNacos 从 Nacos 获取 Redis 配置 (兼容旧版本)
// 推荐使用 GetRedisConfigFromDubbo() 从 dubbo-go 配置中心获取
func GetRedisConfigFromNacos(nacosAddr, namespace, group, dataID string) (*RedisConfig, error) {
	return GetRedisConfigFromDubbo()
}
