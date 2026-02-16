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

	// 初始化 Redis 和 MySQL 客户端（统一管理）
	clients, err := config.InitializeClients(cfg.AppName, cfg.Nacos.Group)
	if err != nil {
		logger.Warnf("Failed to initialize some clients: %v", err)
	}
	defer config.CloseClients(clients)

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
