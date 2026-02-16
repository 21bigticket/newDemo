/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"helloworld/config"
	greet "helloworld/greet"

	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/registry"
	"github.com/dubbogo/gost/log/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GreetTripleServer struct {
	redisClient *redis.Client
}

func (srv *GreetTripleServer) Greet(ctx context.Context, req *greet.GreetRequest) (*greet.GreetResponse, error) {
	logger.Infof("dobbo-do-service receive: %v", req)
	resp := &greet.GreetResponse{Greeting: req.Name}
	return resp, nil
}

func main() {
	cfg, err := config.ParseConfig("")
	if err != nil {
		logger.Errorf("parse config failed: %v", err)
		panic(err)
	}
	logger.Infof("Starting server with config: %+v", cfg)

	// 创建 dubbo 实例（纯代码配置）
	// 必须先创建 dubbo instance 初始化配置中心后，才能获取业务配置
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
		logger.Errorf("new dubbo instance failed: %v", err)
		panic(err)
	}

	// 初始化应用配置管理器（从 dubbo-go 的配置中心获取业务配置）
	if err := config.InitAppConfig(cfg.AppName, cfg.Nacos.Group); err != nil {
		logger.Warnf("Failed to init app config: %v", err)
	}

	// 获取业务配置的三种方式：

	// 方式1: 使用简单的 Get 方法
	fmt.Println("Redis host:", config.GetString("redis.host"))
	fmt.Println("Redis port:", config.GetInt("redis.port"))

	// 方式2: 获取整个配置 map
	redisConfig := config.GetStringMap("redis")
	if redisConfig != nil {
		fmt.Println("完整Redis配置:", redisConfig)
	}

	// 方式3: 解析到结构体（推荐）
	redisCfg, err := config.GetRedisConfigFromDubbo()
	if err != nil {
		logger.Warnf("Failed to get redis config: %v", err)
	} else {
		// 创建 Redis 客户端
		redisClient, err := redisCfg.CreateRedisClient()
		if err != nil {
			logger.Errorf("Failed to create redis client: %v", err)
		} else {
			// 测试 Redis 连接
			ctx := context.Background()
			if err := redisClient.Ping(ctx).Err(); err != nil {
				logger.Errorf("Redis ping failed: %v", err)
			} else {
				logger.Infof("Redis connected successfully: %s", redisCfg.GetAddr())
			}
			defer redisClient.Close()
		}
	}

	// MySQL 数据库初始化
	var db *gorm.DB
	mysqlConfig, err := config.GetMySQLConfigFromDubbo()
	if err != nil {
		logger.Warnf("Failed to get mysql config: %v", err)
	} else {
		// 创建 GORM 数据库连接
		db, err = mysqlConfig.CreateDB()
		if err != nil {
			logger.Errorf("Failed to create database connection: %v", err)
		} else {
			// 测试数据库连接
			sqlDB, _ := db.DB()
			if err := sqlDB.Ping(); err != nil {
				logger.Errorf("Database ping failed: %v", err)
			} else {
				logger.Infof("MySQL connected successfully: %s@%s:%d/%s",
					mysqlConfig.Username, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database)
			}
			defer sqlDB.Close()
		}
	}

	// 创建 server
	srv, err := ins.NewServer()
	if err != nil {
		logger.Errorf("new server failed: %v", err)
		panic(err)
	}

	// 注册服务（使用 V2 接口）
	if err := greet.RegisterGreetServiceHandler(srv, &GreetTripleServer{}); err != nil {
		logger.Errorf("register greeter v2 handler failed: %v", err)
		panic(err)
	}

	// 启动服务
	if err := srv.Serve(); err != nil {
		logger.Errorf("server serve failed: %v", err)
		panic(err)
	}
}
