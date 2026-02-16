# 测试指南

## 前置条件

1. **Nacos 服务运行中**
   - 地址: 192.168.139.230:8848
   - 控制台: http://192.168.139.230:8848/nacos

2. **Redis 服务运行中**（可选，用于测试连接）
   - 地址: 192.168.139.230:6379

## 第一步：在 Nacos 中配置

登录 Nacos 控制台：http://192.168.139.230:8848/nacos

### 创建配置

点击「配置管理」->「配置列表」->「+」创建配置：

- **Data ID**: `go-server`
- **Group**: `DEFAULT_GROUP`
- **配置格式**: `YAML`
- **配置内容**:

```yaml
dubbo:
  application:
    name: go-server
  registries:
    nacos:
      protocol: nacos
      address: 192.168.139.230:8848
  protocols:
    triple:
      name: tri
      port: 20001

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

## 第二步：运行服务

```bash
cd /Users/mac/code/goroot/newDemo/go-server/cmd
go run server.go -nacos-addr 192.168.139.230:8848 -app-name go-server
```

## 预期输出

如果一切正常，你应该看到类似以下的输出：

```
Starting server with config: &{Nacos:... AppName:go-server ...}
[Config Center] Config center doesn't start
或
2025/02/16 23:30:00 [INFO] Config center initialized

# 以下是我们添加的配置获取输出：
Redis host: 192.168.139.230
Redis port: 6379
完整Redis配置: map[conn_timeout:3s db:0 host:192.168.139.230 idle_check_freq:1m ...]

2025/02/16 23:30:00 [INFO] Loaded Redis config from dubbo: &{Host:192.168.139.230 Port:6379 ...}
2025/02/16 23:30:00 [INFO] Redis client created: addr=192.168.139.230:6379, db=0
2025/02/16 23:30:00 [INFO] Redis connected successfully: 192.168.139.230:6379

2025/02/16 23:30:00 [INFO] Dubbo server is running...
```

## 第三步：测试配置热更新

1. 保持服务运行
2. 在 Nacos 控制台修改 Redis 配置，例如将 `port` 从 6379 改为 6380
3. 点击「发布」
4. 查看服务日志，应该看到：

```
2025/02/16 23:31:00 [INFO] App config changed: <配置内容>
2025/02/16 23:31:00 [INFO] App config updated successfully
```

5. 再次调用 `config.GetString("redis.port")` 会返回新值 6380

## 测试代码说明

代码中有三种获取配置的方式：

### 方式1：简单的 Get 方法
```go
fmt.Println("Redis host:", config.GetString("redis.host"))
fmt.Println("Redis port:", config.GetInt("redis.port"))
```

### 方式2：获取整个配置 Map
```go
redisConfig := config.GetStringMap("redis")
fmt.Println("完整Redis配置:", redisConfig)
```

### 方式3：解析到结构体（推荐）
```go
redisCfg, err := config.GetRedisConfigFromDubbo()
redisClient, err := redisCfg.CreateRedisClient()
redisClient.Ping(ctx) // 测试连接
```

## 故障排查

### 问题1：无法连接Nacos
```
Failed to init app config: failed to get config from center
```

**解决方案**：
- 检查 Nacos 服务是否运行
- 检查地址和端口是否正确
- 检查防火墙设置

### 问题2：配置未找到
```
Failed to get redis config: redis config not found
```

**解决方案**：
- 确认在 Nacos 中创建了配置
- 确认 Data ID 是 `go-server`
- 确认 Group 是 `DEFAULT_GROUP`

### 问题3：Redis连接失败
```
Redis ping failed: dial tcp 192.168.139.230:6379: connect: connection refused
```

**解决方案**：
- 检查 Redis 服务是否运行
- 检查地址和端口是否正确
- 如果 Redis 不运行，这部分错误是预期的，不会影响服务启动

## 下一步

测试成功后，你可以：

1. 在你的业务代码中使用 `config.GetString("xxx")` 获取配置
2. 添加更多业务配置（如MySQL、Kafka等）
3. 实现配置变化时的回调逻辑（在 `app_config.go` 的 `Process` 方法中）
