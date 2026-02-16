# config 包使用指南

## 概述

`config` 包提供统一的配置管理和客户端初始化功能，支持从 dubbo-go 的 Nacos 配置中心获取业务配置并初始化 Redis 和 MySQL 连接。

## 核心功能

### 1. 配置获取

支持三种方式获取配置：

```go
// 方式1: 获取单个配置值
redisHost := config.GetString("redis.host")
redisPort := config.GetInt("redis.port")

// 方式2: 获取整个配置 map
redisConfig := config.GetStringMap("redis")

// 方式3: 解析到结构体（推荐）
redisCfg, err := config.GetRedisConfigFromDubbo()
mysqlCfg, err := config.GetMySQLConfigFromDubbo()
```

### 2. 统一客户端初始化（推荐）

使用 `InitializeClients` 一次性初始化所有客户端：

```go
import "helloworld/config"

func main() {
    // 初始化所有客户端（Redis + MySQL）
    clients, err := config.InitializeClients(appName, group)
    if err != nil {
        log.Warnf("Failed to initialize clients: %v", err)
    }
    defer config.CloseClients(clients)

    // 使用客户端
    if clients.Redis != nil {
        // 使用 clients.Redis
        val, err := clients.Redis.Get(ctx, "key").Result()
    }

    if clients.MySQL != nil {
        // 使用 clients.MySQL
        db := clients.MySQL
        // ... GORM 操作
    }
}
```

## go-server 使用示例

```go
package main

import (
    "helloworld/config"
    greet "helloworld/greet"
    "dubbo.apache.org/dubbo-go/v3"
    "dubbo.apache.org/dubbo-go/v3/config_center"
    _ "dubbo.apache.org/dubbo-go/v3/imports"
    "dubbo.apache.org/dubbo-go/v3/protocol"
    "dubbo.apache.org/dubbo-go/v3/registry"
    "github.com/dubbogo/gost/log/logger"
)

type GreetTripleServer struct{}

func main() {
    cfg, err := config.ParseConfig("")
    if err != nil {
        panic(err)
    }

    // 创建 dubbo 实例
    ins, err := dubbo.NewInstance(
        dubbo.WithName(cfg.AppName),
        dubbo.WithConfigCenter(
            config_center.WithNacos(),
            config_center.WithDataID(cfg.AppName),
            config_center.WithAddress(cfg.Nacos.Address),
            config_center.WithNamespace(cfg.Nacos.Namespace),
            config_center.WithGroup(cfg.Nacos.Group),
        ),
        dubbo.WithRegistry(
            registry.WithNacos(),
            registry.WithAddress(cfg.Nacos.Address),
        ),
        dubbo.WithProtocol(
            protocol.WithTriple(),
            protocol.WithPort(cfg.AppPort),
        ),
    )
    if err != nil {
        panic(err)
    }

    // 统一初始化 Redis 和 MySQL
    clients, err := config.InitializeClients(cfg.AppName, cfg.Nacos.Group)
    if err != nil {
        logger.Warnf("Failed to initialize clients: %v", err)
    }
    defer config.CloseClients(clients)

    // 创建 server 并注册服务
    srv, err := ins.NewServer()
    if err != nil {
        panic(err)
    }

    if err := greet.RegisterGreetServiceHandler(srv, &GreetTripleServer{}); err != nil {
        panic(err)
    }

    if err := srv.Serve(); err != nil {
        panic(err)
    }
}
```

## go-client 使用示例

```go
package main

import (
    "context"
    "helloworld/config"
    greet "helloworld/greet"
    "log"

    "dubbo.apache.org/dubbo-go/v3"
    "dubbo.apache.org/dubbo-go/v3/config_center"
    "dubbo.apache.org/dubbo-go/v3/registry"
    "github.com/dubbogo/gost/log/logger"
)

func main() {
    // 创建 dubbo 实例（配置中心）
    ins, err := dubbo.NewInstance(
        dubbo.WithConfigCenter(
            config_center.WithNacos(),
            config_center.WithDataID("go-client"),
            config_center.WithAddress("192.168.139.230:8848"),
            config_center.WithNamespace("public"),
            config_center.WithGroup("DEFAULT_GROUP"),
        ),
        dubbo.WithRegistry(
            registry.WithNacos(),
            registry.WithAddress("192.168.139.230:8848"),
        ),
    )
    if err != nil {
        panic(err)
    }

    // 初始化 Redis 和 MySQL 客户端（复用 server 的配置）
    clients, err := config.InitializeClients("go-server", "DEFAULT_GROUP")
    if err != nil {
        logger.Warnf("Failed to initialize clients: %v", err)
    }
    defer config.CloseClients(clients)

    // 使用 Redis
    if clients.Redis != nil {
        ctx := context.Background()
        val, err := clients.Redis.Get(ctx, "mykey").Result()
        if err != nil {
            logger.Errorf("Redis get failed: %v", err)
        } else {
            logger.Infof("Redis value: %s", val)
        }
    }

    // 使用 MySQL
    if clients.MySQL != nil {
        // GORM 操作
        var results []User
        clients.MySQL.Find(&results)
        logger.Infof("MySQL query results: %+v", results)
    }

    // 创建 client 并调用服务
    cli, err := ins.NewClient()
    if err != nil {
        panic(err)
    }

    svc, err := greet.NewGreetService(cli)
    if err != nil {
        panic(err)
    }

    resp, err := svc.Greet(context.Background(), &greet.GreetRequest{Name: "World"})
    if err != nil {
        logger.Errorf("Greet failed: %v", err)
    } else {
        logger.Infof("Greet response: %s", resp.Greeting)
    }
}
```

## API 文档

### InitializeClients

```go
func InitializeClients(appName, group string) (*Clients, error)
```

初始化所有客户端连接。

**参数:**
- `appName`: 应用名称（Nacos Data ID）
- `group`: 配置分组（通常是 "DEFAULT_GROUP"）

**返回:**
- `*Clients`: 客户端实例集合
- `error`: 错误信息（如果某个客户端初始化失败，不会中断其他客户端）

### CloseClients

```go
func CloseClients(clients *Clients)
```

关闭所有客户端连接。

**参数:**
- `clients`: 客户端实例集合

### Clients 结构体

```go
type Clients struct {
    Redis *redis.Client  // Redis 客户端
    MySQL *gorm.DB       // MySQL/GORM 客户端
}
```

## 配置要求

Nacos 配置中心需要包含以下配置（YAML 格式）：

```yaml
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
  password: "your_password"
  database: "order_service"
  charset: "utf8mb4"
  location: "Asia/Shanghai"
  conn_timeout: 3s
  read_timeout: "5s"
  write_timeout: "5s"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: "1h"
```

## 错误处理

- `InitializeClients` 不会因为单个客户端初始化失败而中断
- 建议检查 `clients.Redis != nil` 和 `clients.MySQL != nil` 后再使用
- 使用 `defer config.CloseClients(clients)` 确保连接正确关闭

## 优势

1. **代码复用**: go-client 和 go-server 使用相同的初始化逻辑
2. **简洁**: 一行代码初始化所有客户端
3. **统一管理**: 所有连接的生命周期统一管理
4. **配置热更新**: 配置变更后自动生效（需要重新初始化）
5. **容错性**: 单个客户端失败不影响其他客户端
