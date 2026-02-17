# Viper + Nacos 配置中心集成使用指南

## 概述

本项目已集成 Viper 配置管理框架，并实现了与 Nacos 配置中心的实时同步。当 Nacos 中的配置发生变化时，会自动更新到 Viper，并触发配置变化监听器。

## 功能特性

1. **自动同步**: Nacos 配置自动同步到 Viper
2. **实时监听**: Nacos 配置变化实时更新到 Viper
3. **配置监听器**: 支持添加自定义配置变化监听器
4. **类型安全**: 支持直接解析配置到结构体
5. **线程安全**: 所有操作都是线程安全的

## 使用方式

### 1. 在 Nacos 中配置

在 Nacos 配置中心添加配置（Data ID: `go-server`, Group: `DEFAULT_GROUP`）：

```yaml
dubbo:
  application:
    name: go-server
  registries:
    nacos:
      protocol: nacos
      address: 192.168.139.230:8848

redis:
  host: "192.168.139.230"
  port: 6379
  password: ""
  db: 0
  pool_size: 100
  min_idle_conns: 5
  conn_timeout: 3s
  read_timeout: "3s"
  write_timeout: "3s"
  pool_timeout: "30s"
  idle_timeout: "5m"
  idle_check_freq: "1m"
  max_conn_age: "1h"

mysql:
  host: "192.168.139.230"
  port: 3306
  username: "root"
  password: "password"
  database: "test"
```

### 2. 初始化 Viper

```go
import "helloworld/config"

// 在 dubbo instance 创建后初始化 viper
if err := config.InitViperFromNacos(cfg.AppName, cfg.Nacos.Group); err != nil {
    logger.Warnf("Failed to init viper from nacos: %v", err)
}
```

### 3. 获取配置值

#### 方式1: 直接获取配置值

```go
vc := config.GetViper()

// 获取字符串
redisHost := vc.GetString("redis.host")
// 输出: "192.168.139.230"

// 获取整数
redisPort := vc.GetInt("redis.port")
// 输出: 6379

// 获取布尔值
enableCache := vc.GetBool("cache.enabled")

// 获取数组
servers := vc.GetStringSlice("servers")

// 检查配置是否存在
if vc.IsSet("redis.password") {
    password := vc.GetString("redis.password")
}
```

#### 方式2: 解析到结构体

```go
// 定义配置结构体
type MySQLConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Database string `yaml:"database"`
}

// 从 viper 解析
vc := config.GetViper()
var mysqlConfig MySQLConfig
if err := vc.UnmarshalKey("mysql", &mysqlConfig); err != nil {
    logger.Errorf("Failed to unmarshal mysql config: %v", err)
}

logger.Infof("MySQL config: %+v", mysqlConfig)
```

#### 方式3: 使用封装的方法（推荐）

```go
// 获取 Redis 配置
redisConfig, err := config.GetRedisConfigFromViper()
if err != nil {
    logger.Errorf("Failed to get redis config: %v", err)
} else {
    // 创建 Redis 客户端
    redisClient, err := redisConfig.CreateRedisClient()
    // ...
}
```

### 4. 监听配置变化

```go
vc := config.GetViper()

// 添加配置变化监听器
vc.AddChangeListener(func(oldValue, newValue interface{}) {
    logger.Infof("Config changed in viper!")
    oldMap, _ := oldValue.(map[string]interface{})
    newMap, _ := newValue.(map[string]interface{})

    // 检查特定配置是否变化
    if oldRedis, ok := oldMap["redis"]; ok {
        if newRedis, ok := newMap["redis"]; ok {
            logger.Infof("Redis config changed: %+v -> %+v", oldRedis, newRedis)
            // 重新创建 Redis 客户端
            handleRedisConfigChange()
        }
    }

    if oldMySQL, ok := oldMap["mysql"]; ok {
        if newMySQL, ok := newMap["mysql"]; ok {
            logger.Infof("MySQL config changed: %+v -> %+v", oldMySQL, newMySQL)
            // 重新创建 MySQL 连接
            handleMySQLConfigChange()
        }
    }
})

func handleRedisConfigChange() {
    // 获取新配置并重新创建客户端
    redisConfig, err := config.GetRedisConfigFromViper()
    if err != nil {
        logger.Errorf("Failed to get new redis config: %v", err)
        return
    }

    newClient, err := redisConfig.CreateRedisClient()
    if err != nil {
        logger.Errorf("Failed to create new redis client: %v", err)
        return
    }

    // 替换旧客户端（注意线程安全）
    // ...
}
```

### 5. 获取所有配置

```go
vc := config.GetViper()

// 获取所有配置
allSettings := vc.AllSettings()
for key, value := range allSettings {
    logger.Infof("Config: %s = %v", key, value)
}

// 或者使用底层 viper 实例
v := vc.GetViper()
allSettings = v.AllSettings()
```

## API 参考

### ViperConfig 方法

| 方法 | 说明 |
|------|------|
| `Get(key string) interface{}` | 获取配置值 |
| `GetString(key string) string` | 获取字符串配置 |
| `GetInt(key string) int` | 获取整数配置 |
| `GetInt64(key string) int64` | 获取 int64 配置 |
| `GetFloat64(key string) float64` | 获取 float64 配置 |
| `GetBool(key string) bool` | 获取布尔配置 |
| `GetStringSlice(key string) []string` | 获取字符串数组配置 |
| `GetStringMap(key string) map[string]interface{}` | 获取 map 配置 |
| `IsSet(key string) bool` | 检查配置是否已设置 |
| `AllSettings() map[string]interface{}` | 获取所有配置 |
| `Unmarshal(rawVal interface{}) error` | 解析所有配置到结构体 |
| `UnmarshalKey(key string, rawVal interface{}) error` | 解析指定 key 的配置到结构体 |
| `AddChangeListener(listener ConfigChangeListener)` | 添加配置变化监听器 |
| `GetViper() *viper.Viper` | 获取底层 viper 实例 |

### 全局函数

| 函数 | 说明 |
|------|------|
| `GetViper() *ViperConfig` | 获取全局 viper 配置实例（单例） |
| `InitViperFromNacos(dataID, group string) error` | 从 nacos 初始化 viper 配置 |

## 配置优先级

Nacos 配置中心的配置会覆盖本地配置文件。当 Nacos 配置发生变化时，会自动更新到 Viper。

## 注意事项

1. **初始化顺序**: 必须在创建 dubbo instance 后再初始化 viper，因为需要使用 dubbo-go 的配置中心连接
2. **线程安全**: ViperConfig 的所有方法都是线程安全的，可以在 goroutine 中使用
3. **配置监听**: 配置监听器会在独立的 goroutine 中执行，注意处理并发问题
4. **默认值**: 建议在结构体中设置合理的默认值
5. **配置验证**: 获取配置后建议进行验证，确保配置的合法性

## 完整示例

```go
package main

import (
    "context"
    "helloworld/config"
    "github.com/dubbogo/gost/log/logger"
    "github.com/redis/go-redis/v9"
)

func main() {
    // ... dubbo instance 创建代码 ...

    // 初始化 viper
    if err := config.InitViperFromNacos("go-server", "DEFAULT_GROUP"); err != nil {
        logger.Warnf("Failed to init viper: %v", err)
        return
    }

    // 获取 Redis 配置
    redisConfig, err := config.GetRedisConfigFromViper()
    if err != nil {
        logger.Errorf("Failed to get redis config: %v", err)
        return
    }

    // 创建 Redis 客户端
    redisClient, err := redisConfig.CreateRedisClient()
    if err != nil {
        logger.Errorf("Failed to create redis client: %v", err)
        return
    }
    defer redisClient.Close()

    // 测试连接
    ctx := context.Background()
    if err := redisClient.Ping(ctx).Err(); err != nil {
        logger.Errorf("Redis ping failed: %v", err)
        return
    }

    logger.Infof("Redis connected successfully")

    // 监听配置变化
    vc := config.GetViper()
    vc.AddChangeListener(func(oldValue, newValue interface{}) {
        logger.Info("Redis config changed, recreating client...")
        // 重新创建客户端逻辑
    })

    // ... 其他业务逻辑 ...
}
```

## 测试配置热更新

1. 启动应用
2. 在 Nacos 控制台修改配置
3. 保存发布
4. 查看应用日志，会看到配置变化的日志
5. 应用会自动使用新配置

## 扩展

你可以基于这个集成框架，轻松添加其他配置源，如：

- Consul
- Etcd
- 文件系统
- 环境变量
- 命令行参数

所有配置源都会统一到 Viper 中管理。
