package config

import (
	"context"

	"github.com/dubbogo/gost/log/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Clients 全局客户端实例
type Clients struct {
	Redis *redis.Client
	MySQL *gorm.DB
}

// InitializeClients 初始化所有客户端连接
// 这个函数可以从 Nacos 获取配置并初始化 Redis 和 MySQL 连接
// 方便 go-client 和 go-server 等多个项目复用
func InitializeClients(appName, group string) (*Clients, error) {
	// 初始化应用配置管理器
	if err := InitAppConfig(appName, group); err != nil {
		logger.Warnf("Failed to init app config: %v", err)
		return nil, err
	}

	//  初始化日志系统
	logCfg, err := GetLogConfigFromNacos()
	if err != nil {
		logger.Warnf("Failed to get log config: %v", err)
	} else {
		if err := InitLogger(logCfg); err != nil {
			logger.Warnf("Failed to init logger: %v", err)
		}
	}

	clients := &Clients{}

	// 初始化 Redis
	redisClient, err := initRedis()
	if err != nil {
		logger.Warnf("Failed to init redis: %v", err)
		// Redis 失败不阻塞，继续初始化 MySQL
	} else {
		clients.Redis = redisClient
	}

	// 初始化 MySQL
	db, err := initMySQL()
	if err != nil {
		logger.Warnf("Failed to init mysql: %v", err)
		// MySQL 失败不阻塞
	} else {
		clients.MySQL = db
	}
	logger.Infof("Clients initialized: Redis=%v, MySQL=%v",
		clients.Redis != nil, clients.MySQL != nil)

	return clients, nil
}

// initRedis 初始化 Redis 连接
func initRedis() (*redis.Client, error) {
	redisCfg, err := GetRedisConfigFromDubbo()
	if err != nil {
		return nil, err
	}

	redisClient, err := redisCfg.CreateRedisClient()
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		return nil, err
	}

	logger.Infof("Redis initialized successfully: %s", redisCfg.GetAddr())
	return redisClient, nil
}

// initMySQL 初始化 MySQL 连接
func initMySQL() (*gorm.DB, error) {
	mysqlCfg, err := GetMySQLConfigFromDubbo()
	if err != nil {
		return nil, err
	}

	db, err := mysqlCfg.CreateDB()
	if err != nil {
		return nil, err
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	logger.Infof("MySQL initialized successfully: %s@%s:%d/%s",
		mysqlCfg.Username, mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database)

	return db, nil
}

// CloseClients 关闭所有客户端连接
func CloseClients(clients *Clients) {
	if clients == nil {
		return
	}

	if clients.Redis != nil {
		if err := clients.Redis.Close(); err != nil {
			logger.Errorf("Failed to close redis client: %v", err)
		}
	}

	if clients.MySQL != nil {
		sqlDB, err := clients.MySQL.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Errorf("Failed to close mysql connection: %v", err)
			}
		}
	}

	logger.Info("All clients closed")
}
