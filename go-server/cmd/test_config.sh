#!/bin/bash

echo "======================================"
echo "Dubbo-go 配置中心测试脚本"
echo "======================================"
echo ""

# 检查 Nacos 配置是否存在
echo "检查步骤1：确认 Nacos 配置"
echo "------------------------"
echo "请确认已在 Nacos 控制台创建以下配置："
echo "  - Data ID: go-server"
echo "  - Group: DEFAULT_GROUP"
echo "  - 配置格式: YAML"
echo ""
echo "配置内容应包含："
cat <<'EOF'
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
EOF
echo ""
read -p "按 Enter 继续测试..."

# 测试编译
echo ""
echo "检查步骤2：编译代码"
echo "------------------------"
cd /Users/mac/code/goroot/newDemo/go-server/cmd
go build -o /tmp/go-server server.go
if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 显示测试命令
echo ""
echo "检查步骤3：运行测试"
echo "------------------------"
echo "运行以下命令启动服务："
echo ""
echo "  cd /Users/mac/code/goroot/newDemo/go-server/cmd"
echo "  go run server.go -nacos-addr 192.168.139.230:8848 -app-name go-server"
echo ""
echo "或者使用编译后的二进制："
echo ""
echo "  /tmp/go-server -nacos-addr 192.168.139.230:8848 -app-name go-server"
echo ""

# 检查 Nacos 连接
echo "检查步骤4：测试 Nacos 连接"
echo "------------------------"
read -p "是否现在启动服务进行测试？(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "正在启动服务..."
    echo "按 Ctrl+C 停止服务"
    echo ""
    go run server.go -nacos-addr 192.168.139.230:8848 -app-name go-server
else
    echo ""
    echo "跳过运行。你可以稍后手动运行测试命令。"
fi

echo ""
echo "======================================"
echo "测试完成"
echo "======================================"
