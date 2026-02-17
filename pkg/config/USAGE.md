# Dubbo-go 配置中心使用指南

## 概述

dubbo-go已经集成了Nacos配置中心，但是**只管理dubbo自身的配置**（注册中心、协议、服务等），**不包括业务配置**（如Redis、MySQL等）。

对于业务配置，我们需要自己处理。本项目提供了三种方式来获取业务配置。

## 核心要点

### dubbo-go 做了什么？

```go
// 这一行代码会：
dubbo_config.Load()
```

1. 连接Nacos配置中心
2. 拉取配置
3. 自动注册配置监听器
4. 配置变化时自动更新dubbo配置

### dubbo-go 没做什么？

❌ **不会**把配置注入到viper
❌ **不会**自动更新业务配置（Redis、MySQL等）
❌ **没有**提供访问业务配置的全局API

## 获取业务配置的三种方式

### 方式1: 简单的 Get 方法（推荐用于简单配置）

```go
// 初始化应用配置管理器
config.InitAppConfig("go-server", "DEFAULT_GROUP")

// 获取配置值
redisHost := config.GetString("redis.host")      // "192.168.139.230"
redisPort := config.GetInt("redis.port")          // 6379
redisDB := config.GetInt("redis.db")              // 0
poolSize := config.GetInt("redis.pool_size")      // 100

// 检查配置是否存在
if config.IsSet("redis.password") {
    password := config.GetString("redis.password")
}
```

### 方式2: 获取整个配置 Map

```go
// 获取Redis整个配置map
redisConfig := config.GetStringMap("redis")
if redisConfig != nil {
    host := redisConfig["host"].(string)
    port := int(redisConfig["port"].(int))
    fmt.Printf("Redis: %s:%d\n", host, port)
}
```

### 方式3: 解析到结构体（推荐）

```go
// 解析到结构体
redisConfig, err := config.GetRedisConfigFromDubbo()
if err != nil {
    logger.Warnf("Failed to get redis config: %v", err)
} else {
    // 创建 Redis 客户端
    redisClient, err := redisConfig.CreateRedisClient()
    if err != nil {
        logger.Errorf("Failed to create redis client: %v", err)
    } else {
        // 测试连接
        ctx := context.Background()
        if err := redisClient.Ping(ctx).Err(); err != nil {
            logger.Errorf("Redis ping failed: %v", err)
        } else {
            logger.Infof("Redis connected: %s", redisConfig.GetAddr())
        }
        defer redisClient.Close()
    }
}
```

## 完整示例

```go
package main

import (
    "context"
    "helloworld/config"
    "github.com/dubbogo/gost/log/logger"
    "dubbo.apache.org/dubbo-go/v3"
    dubbo_config "dubbo.apache.org/dubbo-go/v3/config"
)

func main() {
    // 创建dubbo实例
    ins, err := dubbo.NewInstance(
        dubbo.WithConfigCenter(
            config_center.WithNacos(),
            config_center.WithDataID("go-server"),
            config_center.WithAddress("192.168.139.230:8848"),
        ),
    )
    if err != nil {
        panic(err)
    }

    // 加载dubbo配置（连接Nacos，拉取配置，注册监听器）
    if err := dubbo_config.Load(); err != nil {
        panic(err)
    }

    // 初始化业务配置管理器
    config.InitAppConfig("go-server", "DEFAULT_GROUP")

    // 获取Redis配置
    redisHost := config.GetString("redis.host")
    redisPort := config.GetInt("redis.port")

    logger.Infof("Redis config: %s:%d", redisHost, redisPort)

    // 或者解析到结构体
    redisCfg, err := config.GetRedisConfigFromDubbo()
    if err == nil {
        redisClient, _ := redisCfg.CreateRedisClient()
        defer redisClient.Close()
    }

    // 启动服务
    srv, _ := ins.NewServer()
    srv.Serve()
}
```

## 配置热更新

当Nacos中的配置发生变化时：

1. **dubbo配置**会自动更新（注册中心、协议等）
2. **业务配置**也会自动更新到内存中

下一次调用 `config.GetString("redis.host")` 时会获取到最新的值。

如果需要在配置变化时执行特定逻辑（如重新创建Redis连接），可以修改 `app_config.go` 中的监听器：

```go
// Process 实现 ConfigurationListener 接口
func (l *appConfigListener) Process(event *config_center.ConfigChangeEvent) {
    logger.Infof("App config changed: %v", event.Value)

    // ... 解析新配置 ...

    // 触发配置变化回调
    onConfigChanged(appConfig.data)
}

func onConfigChanged(newConfig map[string]interface{}) {
    // 检查Redis配置是否变化
    if newRedis, ok := newConfig["redis"].(map[string]interface{}); ok {
        logger.Infof("Redis config changed: %+v", newRedis)
        // 重新创建Redis客户端
        recreateRedisClient(newRedis)
    }
}
```

## Nacos 配置格式

在Nacos配置中心（Data ID: `go-server`, Group: `DEFAULT_GROUP`）配置：

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

mysql:
  host: "192.168.139.230"
  port: 3306
  username: "root"
  password: "password"
  database: "test"
```

## API 参考

### 配置访问方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `Get(key string)` | 获取配置值 | `config.Get("redis.host")` |
| `GetString(key string)` | 获取字符串配置 | `config.GetString("redis.host")` |
| `GetInt(key string)` | 获取整数配置 | `config.GetInt("redis.port")` |
| `GetBool(key string)` | 获取布尔配置 | `config.GetBool("cache.enabled")` |
| `GetStringMap(key string)` | 获取map配置 | `config.GetStringMap("redis")` |
| `IsSet(key string)` | 检查配置是否存在 | `config.IsSet("redis.password")` |
| `GetAll()` | 获取所有配置 | `config.GetAll()` |

### 特定配置获取

| 方法 | 说明 |
|------|------|
| `GetRedisConfigFromDubbo()` | 获取Redis配置结构体 |
| `GetRedisConfigFromViper()` | 从viper获取Redis配置（如果使用了viper集成） |

## 常见问题

### Q: dubbo-go会自动把配置注入到viper吗？

**A: 不会**。dubbo-go使用koanf而不是viper，并且没有把配置暴露给业务代码使用。

### Q: 配置变化后会自动更新吗？

**A: 是的**。我们实现的配置管理器会监听Nacos配置变化，并自动更新内存中的配置。

### Q: 如何在配置变化时重新创建客户端连接？

**A: 修改`app_config.go`中的`Process`方法，添加配置变化的回调逻辑。

### Q: viper.GetString("redis.host") 能用吗？

**A: 不能**。dubbo-go不会把配置注入到viper中，需要使用我们提供的`config.GetString("redis.host")`。

## 总结

- dubbo-go只管理dubbo配置，不管业务配置
- 使用`config.GetString()`等方法获取业务配置
- 配置变化会自动更新，下次调用获取新值
- 如需在配置变化时执行逻辑，修改监听器即可
