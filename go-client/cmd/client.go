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

	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"dubbo.apache.org/dubbo-go/v3/registry"
	"github.com/dubbogo/gost/log/logger"

	"helloworld/api"
)

func main() {
	// nacos 地址
	nacosAddr := "192.168.139.230:8848"

	// 创建 dubbo 实例，配置 nacos 注册中心
	ins, err := dubbo.NewInstance(
		dubbo.WithName("go-client"),
		dubbo.WithConfigCenter(
			config_center.WithNacos(),
			config_center.WithAddress(nacosAddr),
			config_center.WithDataID("go-client-config"),
			config_center.WithGroup("DEFAULT_GROUP"),
			config_center.WithFileExtYaml(),
		),
		dubbo.WithRegistry(
			registry.WithNacos(),
			registry.WithAddress(nacosAddr),
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

	// 创建 GreeterV2 服务客户端
	greeterClient, err := api.NewGreeter(cli)
	if err != nil {
		logger.Errorf("new greeter v2 client failed: %v", err)
		panic(err)
	}

	// 调用服务
	logger.Info("start to test dubbo")
	req := &api.HelloRequest{
		Name: "laurence",
	}
	reply, err := greeterClient.SayHello(context.Background(), req)
	if err != nil {
		logger.Errorf("call SayHello failed: %v", err)
		panic(err)
	}
	logger.Infof("client response result: %v\n", reply)

	// 保持程序运行
	select {}
}
