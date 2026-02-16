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

	"dubbo.apache.org/dubbo-go/v3"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"dubbo.apache.org/dubbo-go/v3/registry"
	"github.com/dubbogo/gost/log/logger"

	"helloworld/greet"
)

func main() {
	// nacos 地址
	cfg, err := config.ParseConfig("")
	if err != nil {
		logger.Errorf("parse config failed: %v", err)
		panic(err)
	}
	logger.Infof("Starting server with config: %+v", cfg)

	// 创建 dubbo 实例，配置 nacos 注册中心
	ins, err := dubbo.NewInstance(
		dubbo.WithName(cfg.AppName),
		dubbo.WithRegistry(
			registry.WithNacos(),
			registry.WithAddress(cfg.Nacos.Address),
		),
	)
	if err != nil {
		logger.Errorf("new dubbo instance failed: %v", err)
		panic(err)
	}

	// 创建 client
	cli, err := ins.NewClient()
	if err != nil {
		logger.Errorf("new client failed: %v", err)
		panic(err)
	}

	// 创建 greeterV2 服务客户端
	greeterClient, err := greet.NewGreetService(cli)
	if err != nil {
		logger.Errorf("new greeter v2 client failed: %v", err)
		panic(err)
	}

	// 调用服务
	logger.Info("start to test dubbo")
	req := &greet.GreetRequest{
		Name: "laurence",
	}
	reply, err := greeterClient.Greet(context.Background(), req)
	if err != nil {
		logger.Errorf("call SayHello failed: %v", err)
		panic(err)
	}
	logger.Infof("client response result: %v\n", reply)

	// 保持程序运行
	select {}
}
