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
	"helloworld/api"

	"dubbo.apache.org/dubbo-go/v3"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"github.com/dubbogo/gost/log/logger"
)

type GreeterProvider struct {
}

func (s *GreeterProvider) SayHello(ctx context.Context, in *api.HelloRequest) (*api.User, error) {
	logger.Infof("Dubbo3 GreeterV2Handler get user name = %s\n", in.Name)
	return &api.User{Name: "Hello " + in.Name, Id: "12345", Age: 21}, nil
}

func (s *GreeterProvider) SayHelloStream(ctx context.Context, stream api.Greeter_SayHelloStreamServer) error {
	logger.Info("SayHelloStream method called")

	for {
		req, err := stream.Recv()
		if err != nil {
			logger.Infof("SayHelloStream recv finished or error: %v", err)
			break
		}

		logger.Infof("Received stream request: %s", req.Name)

		user := &api.User{
			Name: "Hello " + req.Name,
			Id:   "stream-12345",
			Age:  21,
		}

		if err := stream.Send(user); err != nil {
			logger.Errorf("SayHelloStream send error: %v", err)
			return err
		}
	}

	return nil
}

func main() {
	// 创建 dubbo 实例（使用配置文件 conf/dubbogo.yaml）
	ins, err := dubbo.NewInstance()
	if err != nil {
		logger.Errorf("new dubbo instance failed: %v", err)
		panic(err)
	}

	// 创建 server
	srv, err := ins.NewServer()
	if err != nil {
		logger.Errorf("new server failed: %v", err)
		panic(err)
	}

	// 注册服务（使用 V2 接口）
	if err := api.RegisterGreeterHandler(srv, &GreeterProvider{}); err != nil {
		logger.Errorf("register greeter v2 handler failed: %v", err)
		panic(err)
	}

	// 启动服务
	if err := srv.Serve(); err != nil {
		logger.Errorf("server serve failed: %v", err)
		panic(err)
	}
}
